package urlrouter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/service"
	"github.com/lexizz/cumloys/internal/service/finduserservice"
)

func (route *urlRouter) AuthenticationHandler(findUserService service.FindUserServiceInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected `/login` === ")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			route.logger.Errorf("---> ERROR: readAll body: %v\n", err)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		bodyInString := string(body)
		route.logger.Info("=== Body: ", bodyInString)

		errBodyEmpty := checkBodyOnEmpty(body)
		if errBodyEmpty != nil {
			route.logger.Errorf("---> ERROR: Body empty: %v\n", bodyInString)
			http.Error(writer, errBodyEmpty.Error(), http.StatusBadRequest)
			return
		}

		authorizationData := authorization{}

		errDecode := json.Unmarshal(body, &authorizationData)
		if errDecode != nil {
			route.logger.Errorf("---> ERROR: json decode: %v\n", errDecode)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		errRequireFields := checkLoginAndPasswordOnEmpty(authorizationData)
		if errRequireFields != nil {
			route.logger.Errorf("---> ERROR: empty fields login or pwd: %v\n", errRequireFields)
			http.Error(writer, errRequireFields.Error(), http.StatusBadRequest)
			return
		}

		userFromDB, errLogin := findUserService.GetUserByLogin(request.Context(), authorizationData.Login)
		if errLogin != nil {
			if errors.Is(errLogin, finduserservice.ErrInternal) {
				route.logger.Errorf("---> ERROR: getting user: %v\n", errLogin)
				http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
				return
			}

			route.logger.Errorf("---> ERROR: user not found: %v\n", errLogin)
			http.Error(writer, ErrWrongLoginOrPassword.Error(), http.StatusUnauthorized)
			return
		}

		isValidPwd := utils.IsValidPassword(authorizationData.Password, userFromDB.Password)
		if !isValidPwd {
			route.logger.Errorf("---> ERROR: wrong password: %v\n", authorizationData.Password)
			http.Error(writer, ErrWrongLoginOrPassword.Error(), http.StatusUnauthorized)
			return
		}

		claims := map[string]interface{}{
			"user_id": userFromDB.ID.String(),
		}

		resToken, errToken := route.jwt.Encode(claims)
		if errToken != nil {
			route.logger.Errorf("---> ERROR: encode token: %v\n", errToken.Error())
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		route.logger.Infof("=== Token: %v\n", resToken)

		writer.Header().Set("Authorization", resToken)

		sendResponse(writer, []byte("ok"), http.StatusOK, route.logger)
	}
}
