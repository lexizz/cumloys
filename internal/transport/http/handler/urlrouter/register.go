package urlrouter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lexizz/cumloys/internal/service"
	"github.com/lexizz/cumloys/internal/service/createuserservice"
)

func (route *urlRouter) RegistrationHandler(createUserSrv service.CreateUserServiceInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected `/register` === ")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			route.logger.Errorf("---> ERROR ReadAll body: %v\n", err)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		bodyInString := string(body)
		route.logger.Info("=== Body: ", bodyInString)

		errBodyEmpty := checkBodyOnEmpty(body)
		if errBodyEmpty != nil {
			route.logger.Errorf("---> ERROR: Body empty: %+v\n", bodyInString)
			http.Error(writer, errBodyEmpty.Error(), http.StatusBadRequest)
			return
		}

		authorizationData := authorization{}

		errDecode := json.Unmarshal(body, &authorizationData)
		if errDecode != nil {
			route.logger.Errorf("---> ERROR JSON: %+v; BODY:[%v]\n", errDecode, bodyInString)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		errRequireFields := checkLoginAndPasswordOnEmpty(authorizationData)
		if errRequireFields != nil {
			route.logger.Errorf("---> ERROR: Empty fields: %+v\n", bodyInString)
			http.Error(writer, errRequireFields.Error(), http.StatusBadRequest)
			return
		}

		lastInsertID, errCUS := createUserSrv.Handle(request.Context(), authorizationData.Login, authorizationData.Password)
		if errCUS != nil {
			if errors.Is(errCUS, createuserservice.ErrUserExists) {
				http.Error(writer, errCUS.Error(), http.StatusConflict)
				return
			}

			route.logger.Errorf("---> ERROR: createUserService: %v\n", errCUS.Error())
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		claims := map[string]interface{}{
			"user_id": lastInsertID.String(),
		}

		resToken, errToken := route.jwt.Encode(claims)
		if errToken != nil {
			route.logger.Error("---> ERROR token: " + errToken.Error())
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		route.logger.Infof("=== Token: %v === \n", resToken)

		writer.Header().Set("Authorization", resToken)

		sendResponse(writer, []byte("ok"), http.StatusOK, route.logger)
	}
}
