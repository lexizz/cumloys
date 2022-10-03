package urlrouter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/service"
	"github.com/lexizz/cumloys/internal/service/withdrawpointsservice"
)

func (route *urlRouter) GettingCurrentBalanceHandler(findBalanceService service.FindBalanceServiceInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected /api/user/balance` (GET) === ")

		userUUID, errUUID := route.getUserUUID(request)
		if errUUID != nil {
			route.logger.Errorf("---> ERROR: GettingCurrentBalanceHandler: getting user id from token: %v", errUUID)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		totalScoreWithdraw := findBalanceService.GetBalanceByUserID(request.Context(), *userUUID)

		totalBalanceForResponse, errEncode := json.Marshal(totalScoreWithdraw)
		if errEncode != nil {
			route.logger.Errorf("---> ERROR: GettingOrdersHandler: failed encode to json: %v", errEncode)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		sendResponse(writer, totalBalanceForResponse, http.StatusOK, route.logger)
	}
}

func (route *urlRouter) WithdrawPointsHandler(
	createOrderService service.CreateOrderServiceInterface,
	findOrderService service.FindOrderServiceInterface,
	withdrawPointsService service.WithdrawPointsServiceInterface,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected /api/user/balance/withdraw` === ")

		userUUID, errUUID := route.getUserUUID(request)
		if errUUID != nil {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: getting user id from token: %v", errUUID)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: readAll body: %v\n", err)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		bodyInString := string(body)
		route.logger.Info("=== Body: ", bodyInString)

		errBodyEmpty := checkBodyOnEmpty(body)
		if errBodyEmpty != nil {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: Body empty: %v\n", bodyInString)
			http.Error(writer, errBodyEmpty.Error(), http.StatusBadRequest)
			return
		}

		withdrawPointData := withdrawPoint{}

		errDecode := json.Unmarshal(body, &withdrawPointData)
		if errDecode != nil {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: json decode: %v\n", errDecode)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		isValidNumber := utils.CheckNumberOrder(withdrawPointData.NumberOrder)
		if !isValidNumber {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: failed number of order: %v", withdrawPointData.NumberOrder)
			http.Error(writer, "wrong number of order", http.StatusUnprocessableEntity)
			return
		}

		var orderID uuid.UUID

		isExistsOrder, orderExistsID, _ := findOrderService.IsExistsOrder(request.Context(), withdrawPointData.NumberOrder)
		if !isExistsOrder {
			route.logger.Errorf("---> ERROR: WithdrawPointsHandler: order not found: %v", withdrawPointData.NumberOrder)

			order, errCreateOrder := createOrderService.Handle(request.Context(), withdrawPointData.NumberOrder, *userUUID)
			if errCreateOrder != nil {
				route.logger.Errorf("---> ERROR: WithdrawPointsHandler: error creating order: %v", withdrawPointData.NumberOrder)
				http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
				return
			}

			orderID = order.ID
		} else {
			orderID = *orderExistsID
		}

		_, errWithdraw := withdrawPointsService.Handle(request.Context(), withdrawPointData.Points, orderID, *userUUID)
		if errWithdraw != nil {
			if errors.Is(errWithdraw, withdrawpointsservice.ErrBalanceZero) {
				route.logger.Errorf("---> ERROR: balance has already zero: %v", errWithdraw)
				http.Error(writer, errWithdraw.Error(), http.StatusPaymentRequired)
				return
			}

			route.logger.Errorf("---> ERROR: handle withdraw points: %v", errWithdraw)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(writer, []byte("ok"), http.StatusOK, route.logger)
	}
}

func (route *urlRouter) GettingInfoAboutBalanceHandler(
	findWithdrawPointsService service.FindWithdrawPointsServiceInterface,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		route.logger.Infof("%v | %v | %v", request.Method, request.Host, request.URL.Path)
		route.logger.Info("=== Part url was detected /api/user/withdrawals` === ")

		userUUID, errUUID := route.getUserUUID(request)
		if errUUID != nil {
			route.logger.Errorf("---> ERROR: GettingInfoAboutBalanceHandler: getting user id from token: %v", errUUID)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		scoreWithdrawals := findWithdrawPointsService.Handle(request.Context(), *userUUID)
		if len(scoreWithdrawals) == 0 {
			route.logger.Error("---> ERROR: GettingInfoAboutBalanceHandler: withdrawals not found: %v")
			http.Error(writer, " withdrawals not found", http.StatusNoContent)
			return
		}

		scoreWithdrawalsForResponse, errEncode := json.Marshal(scoreWithdrawals)
		if errEncode != nil {
			route.logger.Errorf("---> ERROR: GettingOrdersHandler: failed encode to json: %v", errEncode)
			http.Error(writer, ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		sendResponse(writer, scoreWithdrawalsForResponse, http.StatusOK, route.logger)
	}
}
