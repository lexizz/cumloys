package withdrawpointsservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
)

var ErrBalanceZero = errors.New("insufficient funds")

type withdrawPointsService struct {
	orderRepository       repository.OrderRepositoryInterface
	scoreRepository       repository.ScoreRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	logger                logger.Logger
}

func New(
	orderRepository repository.OrderRepositoryInterface,
	scoreRepository repository.ScoreRepositoryInterface,
	transactionRepository repository.TransactionRepositoryInterface,
	logger logger.Logger,
) *withdrawPointsService {
	return &withdrawPointsService{
		orderRepository:       orderRepository,
		scoreRepository:       scoreRepository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (service *withdrawPointsService) Handle(ctx context.Context, sumWithdrawPoints float32, orderID uuid.UUID, userID uuid.UUID) (bool, error) {
	score, errGetPoints := service.scoreRepository.GetScoreByUserID(ctx, userID)
	if errGetPoints != nil {
		return false, errGetPoints
	}

	if score == nil {
		return false, ErrBalanceZero
	}

	if score.Total <= sumWithdrawPoints {
		return false, ErrBalanceZero
	}

	resultScore := score.Total - sumWithdrawPoints

	errUpdateScore := service.scoreRepository.Update(ctx, userID, resultScore)
	if errUpdateScore != nil {
		return false, errUpdateScore
	}

	errInsertTransaction := service.transactionRepository.Insert(ctx, userID, orderID, sumWithdrawPoints, models.DecreasePointsType)
	if errInsertTransaction != nil {
		return false, errInsertTransaction
	}

	return true, nil
}
