package gettingpointsservice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/config"
	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
	"github.com/lexizz/cumloys/internal/service"
)

var _ service.GettingPointsServiceInterface = &gettingPointsService{}

const requestTimeout time.Duration = 2

type gettingPointsService struct {
	cfg                   *config.Config
	httpClient            *http.Client
	createOrderService    service.CreateOrderServiceInterface
	orderRepository       repository.OrderRepositoryInterface
	scoreRepository       repository.ScoreRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	logger                logger.Logger
}

type responseOrderData struct {
	Number string  `json:"order,omitempty"`
	Status string  `json:"status,omitempty"`
	Points float32 `json:"accrual,omitempty"`
}

func New(
	cfg *config.Config,
	httpClient *http.Client,
	createOrderService service.CreateOrderServiceInterface,
	orderRepository repository.OrderRepositoryInterface,
	scoreRepository repository.ScoreRepositoryInterface,
	transactionRepository repository.TransactionRepositoryInterface,
	logger logger.Logger,
) *gettingPointsService {
	return &gettingPointsService{
		cfg:                   cfg,
		httpClient:            httpClient,
		createOrderService:    createOrderService,
		orderRepository:       orderRepository,
		scoreRepository:       scoreRepository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (service *gettingPointsService) Handle(ctx context.Context, numberOrder string, userID uuid.UUID) error {
	order, errCreate := service.createOrderService.Handle(ctx, numberOrder, userID)
	if errCreate != nil {
		return errors.New("error creating order: " + errCreate.Error())
	}

	go func(numberOrder string, userID uuid.UUID) {
		ctxTimeout, cancelCtxTimeout := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCtxTimeout()

		responseData, err := service.sendRequest(numberOrder)
		if err != nil {
			return
		}

		errUpdateOrder := service.orderRepository.Update(ctxTimeout, numberOrder, userID, responseData.Status, responseData.Points)
		if errUpdateOrder != nil {
			return
		}

		score, errScore := service.scoreRepository.GetScoreByUserID(ctxTimeout, userID)
		if errScore != nil {
			return
		}

		if score != nil {
			score.Total += responseData.Points

			errScoreUpdate := service.scoreRepository.Update(ctxTimeout, userID, score.Total)
			if errScoreUpdate != nil {
				return
			}
		} else {
			_, errScoreInsert := service.scoreRepository.Insert(ctxTimeout, userID, responseData.Points)
			if errScoreInsert != nil {
				return
			}
		}

		errTransactionInsert := service.transactionRepository.Insert(
			ctxTimeout,
			userID,
			order.ID,
			responseData.Points,
			models.IncreasePointsType,
		)
		if errTransactionInsert != nil {
			return
		}
	}(numberOrder, userID)

	return nil
}

func (service *gettingPointsService) sendRequest(numberOrder string) (*responseOrderData, error) {
	accrualAddress := service.cfg.IncomingParams.AccrualSystemAddress
	url := accrualAddress + "/api/orders/" + numberOrder

	service.logger.Infof("=== Url accrual: %v", url)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout*time.Second)
	defer cancel()

	var buf io.Reader

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, buf)
	if err != nil {
		service.logger.Errorf("---> ERROR: gettingPointsService: NewRequestWithContext: %v\n", err)
		return nil, err
	}

	response, err := service.httpClient.Do(request)
	if err != nil {
		service.logger.Errorf("---> ERROR: gettingPointsService: request to accrual: %v\n", err)
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		service.logger.Errorf("---> ERROR: gettingPointsService: failed read response body: %v\n", err)
		return nil, err
	}

	service.logger.Infof("=== Response accrual status: %v | Body: %v", response.Status, string(body))

	errorResponseClose := response.Body.Close()
	if errorResponseClose != nil {
		service.logger.Errorf("---> ERROR: gettingPointsService: failed response body close: %v\n", err)
		return nil, errorResponseClose
	}

	if response.StatusCode > http.StatusNoContent {
		service.logger.Errorf("---> ERROR: gettingPointsService: failed request: status: %v", response.Status)
		return nil, errors.New("data not found")
	}

	responseData := responseOrderData{}

	if response.StatusCode == http.StatusNoContent {
		service.logger.Infof("=== points by this number of order not found: %v", response.Status)

		responseData.Number = numberOrder
		responseData.Points = 0
		responseData.Status = "NEW"
	}

	errDecode := json.Unmarshal(body, &responseData)
	if errDecode != nil {
		service.logger.Errorf("---> ERROR: gettingPointsService: json decode: %v\n", errDecode)
		return nil, errDecode
	}

	service.logger.Infof("=== Response data: %+v", responseData)

	return &responseData, nil
}
