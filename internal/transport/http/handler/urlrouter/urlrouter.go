package urlrouter

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
)

type urlRouter struct {
	logger logger.Logger
	jwt    *models.JWT
}

type authorization struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type withdrawPoint struct {
	NumberOrder string  `json:"order,omitempty"`
	Points      float32 `json:"sum,omitempty"`
}

var (
	ErrRequireFieldsMissing = errors.New("required fields are missing")
	ErrInternalServer       = errors.New("internal server error")
	ErrWrongLoginOrPassword = errors.New("wrong login or password")
)

func New(jwt *models.JWT, logger logger.Logger) *urlRouter {
	return &urlRouter{
		jwt:    jwt,
		logger: logger,
	}
}

func checkBodyOnEmpty(body []byte) error {
	if len(body) == 0 {
		return ErrRequireFieldsMissing
	}

	return nil
}

func checkLoginAndPasswordOnEmpty(authorizationData authorization) error {
	if len(authorizationData.Login) == 0 || len(authorizationData.Password) == 0 {
		return ErrRequireFieldsMissing
	}

	return nil
}

func (route *urlRouter) getUserUUID(request *http.Request) (*uuid.UUID, error) {
	token := request.Header.Get("Authorization")

	tokenJWT, errParse := route.jwt.Parse(token)
	if errParse != nil {
		route.logger.Errorf("---> ERROR parse token: %v; Token: %v\n", errParse, token)
		return nil, errParse
	}

	userFromToken, ok := tokenJWT.Get("user_id")
	userID := fmt.Sprint(userFromToken)
	if !ok || len(userID) == 0 {
		route.logger.Error("---> ERROR: failed getting user id from jwt")
		return nil, errors.New("failed getting user id from token")
	}

	userUUID, errParseUUID := uuid.Parse(userID)
	if errParseUUID != nil {
		route.logger.Errorf("---> ERROR parse user id to uuid: %v", errParse)
		return nil, errParseUUID
	}

	return &userUUID, nil
}

func sendResponse(writer http.ResponseWriter, message []byte, statusCode int, logger logger.Logger) {
	writer.WriteHeader(statusCode)

	_, writeError := writer.Write(message)
	if writeError != nil {
		logger.Errorf("---> ERROR: Write: %v", writeError)
		http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
}
