package createorderservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
	"github.com/lexizz/cumloys/internal/service"
)

var (
	ErrOrderExists   = errors.New("order has already exists")
	ErrOrderCreation = errors.New("failed creation order, this order wasn't created")
)

var _ service.CreateOrderServiceInterface = &createOrderService{}

type createOrderService struct {
	orderRepository       repository.OrderRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	logger                logger.Logger
}

func New(
	orderRepository repository.OrderRepositoryInterface,
	transactionRepository repository.TransactionRepositoryInterface,
	logger logger.Logger,
) *createOrderService {
	return &createOrderService{
		orderRepository:       orderRepository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (service *createOrderService) Handle(ctx context.Context, numberOrder string, userID uuid.UUID) (*models.Order, error) {
	if len(numberOrder) < 1 {
		return nil, errors.New("number order empty")
	}

	isOrderExists, _, _, err := service.orderRepository.IsExists(ctx, numberOrder)
	if err != nil {
		return nil, err
	}

	if isOrderExists {
		return nil, ErrOrderExists
	}

	lastInsertID, errInsert := service.orderRepository.Insert(ctx, numberOrder, userID)
	if errInsert != nil {
		return nil, ErrOrderCreation
	}

	return &models.Order{
		ID:     *lastInsertID,
		Number: numberOrder,
		UserID: userID,
	}, nil
}
