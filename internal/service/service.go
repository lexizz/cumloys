package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
)

type Services struct {
	CreateUserService         CreateUserServiceInterface
	FindUserService           FindUserServiceInterface
	CreateOrderService        CreateOrderServiceInterface
	FindOrderService          FindOrderServiceInterface
	FindBalanceService        FindBalanceServiceInterface
	GettingPointsService      GettingPointsServiceInterface
	WithdrawPointsService     WithdrawPointsServiceInterface
	FindWithdrawPointsService FindWithdrawPointsServiceInterface
}

type (
	CreateUserServiceInterface interface {
		Handle(ctx context.Context, newLogin string, newPwd string) (*uuid.UUID, error)
	}

	FindUserServiceInterface interface {
		GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	}

	CreateOrderServiceInterface interface {
		Handle(ctx context.Context, numberOrder string, userID uuid.UUID) (*models.Order, error)
	}

	FindOrderServiceInterface interface {
		GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
		IsExistsOrder(ctx context.Context, numberOrder string) (bool, *uuid.UUID, *uuid.UUID)
	}

	FindBalanceServiceInterface interface {
		GetBalanceByUserID(ctx context.Context, userID uuid.UUID) *models.TotalScoreWithdraw
	}

	GettingPointsServiceInterface interface {
		Handle(ctx context.Context, numberOrder string, userID uuid.UUID) error
	}

	WithdrawPointsServiceInterface interface {
		Handle(ctx context.Context, sumWithdrawPoints float32, orderID uuid.UUID, userID uuid.UUID) (bool, error)
	}

	FindWithdrawPointsServiceInterface interface {
		Handle(ctx context.Context, userID uuid.UUID) []models.ScoreWithdraw
	}
)
