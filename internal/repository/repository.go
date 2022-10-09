package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
)

type UserRepositoryInterface interface {
	Insert(ctx context.Context, newLogin string, newPassword string) (*uuid.UUID, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	IsExists(ctx context.Context, login string) (bool, error)
}

type OrderRepositoryInterface interface {
	Insert(ctx context.Context, number string, userID uuid.UUID) (*uuid.UUID, error)
	Update(ctx context.Context, number string, userID uuid.UUID, status string, points float32) error
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	IsExists(ctx context.Context, number string) (bool, *uuid.UUID, *uuid.UUID, error)
}

type ScoreRepositoryInterface interface {
	Insert(ctx context.Context, userID uuid.UUID, points float32) (*uuid.UUID, error)
	Update(ctx context.Context, userID uuid.UUID, points float32) error
	GetScoreByUserID(ctx context.Context, userID uuid.UUID) (*models.Score, error)
}

type TransactionRepositoryInterface interface {
	Insert(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, points float32, typeTransaction int) error
	GetSumFundsWithdrawn(ctx context.Context, userID uuid.UUID) (float32, error)
	GetAllFundsWithdrawn(ctx context.Context, userID uuid.UUID) ([]models.ScoreWithdraw, error)
}
