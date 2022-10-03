package findorderservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
	"github.com/lexizz/cumloys/internal/service"
)

var _ service.FindOrderServiceInterface = &findOrderService{}

type findOrderService struct {
	orderRepository repository.OrderRepositoryInterface
	logger          logger.Logger
}

var (
	ErrInternal      = errors.New("internal error")
	ErrOrderNotFound = errors.New("order not found")
)

func New(orderRepository repository.OrderRepositoryInterface, logger logger.Logger) *findOrderService {
	return &findOrderService{
		orderRepository: orderRepository,
		logger:          logger,
	}
}

func (service *findOrderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	orders, errQuery := service.orderRepository.GetAllByUserID(ctx, userID)
	if errQuery != nil {
		return nil, ErrInternal
	}

	if orders != nil {
		return orders, nil
	}

	return nil, ErrOrderNotFound
}

func (service *findOrderService) IsExistsOrder(ctx context.Context, numberOrder string) (bool, *uuid.UUID, *uuid.UUID) {
	isExists, orderID, userID, _ := service.orderRepository.IsExists(ctx, numberOrder)

	return isExists, orderID, userID
}
