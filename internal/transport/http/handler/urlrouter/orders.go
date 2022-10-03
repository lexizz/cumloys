package urlrouter

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/service"
)

func (route *urlRouter) GettingOrdersHandler(findOrderService service.FindOrderServiceInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected `/api/user/orders` (GET) === ")

		userUUID, errUUID := route.getUserUUID(request)
		if errUUID != nil {
			route.logger.Errorf("---> ERROR: GettingOrdersHandler: getting user id from token: %v", errUUID)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		orders, err := findOrderService.GetOrdersByUserID(request.Context(), *userUUID)
		if err != nil {
			route.logger.Errorf("---> ERROR: GettingOrdersHandler: getting orders: %v", err)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		ordersForResponse, errEncode := json.Marshal(orders)
		if errEncode != nil {
			route.logger.Errorf("---> ERROR: GettingOrdersHandler: failed encode to json: %v", errEncode)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		sendResponse(writer, ordersForResponse, http.StatusOK, route.logger)
	}
}

func (route *urlRouter) AddingOrdersHandler(
	findOrderService service.FindOrderServiceInterface,
	gettingPointsService service.GettingPointsServiceInterface,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected `/api/user/orders` (POST) === ")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: readAll body: %v\n", err)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		bodyInString := string(body)
		route.logger.Info("=== Body, order:", bodyInString)

		errBodyEmpty := checkBodyOnEmpty(body)
		if errBodyEmpty != nil {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: Body empty: %v\n", bodyInString)
			http.Error(writer, errBodyEmpty.Error(), http.StatusBadRequest)
			return
		}

		numberOrder := strings.ReplaceAll(bodyInString, " ", "")

		isValidNumber := utils.CheckNumberOrder(numberOrder)
		if !isValidNumber {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: failed number of order: %v", numberOrder)
			http.Error(writer, "wrong number of order", http.StatusUnprocessableEntity)
			return
		}

		userUUID, errUUID := route.getUserUUID(request)
		if errUUID != nil {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: getting user id from token: %v", errUUID)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		isExistsOrder, _, currentUserID := findOrderService.IsExistsOrder(request.Context(), numberOrder)
		if isExistsOrder && userUUID.String() != currentUserID.String() {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: order has already exists OTHER user: %v", numberOrder)
			http.Error(writer, "this order has already exists", http.StatusConflict)
			return
		}

		if isExistsOrder && userUUID.String() == currentUserID.String() {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler: order has already exists THIS user: %v", numberOrder)

			sendResponse(writer, []byte("this order has already exists"), http.StatusOK, route.logger)

			return
		}

		errPoints := gettingPointsService.Handle(request.Context(), numberOrder, *userUUID)
		if errPoints != nil {
			route.logger.Errorf("---> ERROR: AddingOrdersHandler handle points: %v", errPoints)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(writer, []byte("ok orders"), http.StatusAccepted, route.logger)
	}
}
