package findbalanceservice

import (
	"context"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
	"github.com/lexizz/cumloys/internal/service"
)

var _ service.FindBalanceServiceInterface = &findBalanceService{}

type findBalanceService struct {
	scoreRepository       repository.ScoreRepositoryInterface
	transactionRepository repository.TransactionRepositoryInterface
	logger                logger.Logger
}

func New(
	scoreRepository repository.ScoreRepositoryInterface,
	transactionRepository repository.TransactionRepositoryInterface,
	logger logger.Logger,
) *findBalanceService {
	return &findBalanceService{
		scoreRepository:       scoreRepository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (service *findBalanceService) GetBalanceByUserID(ctx context.Context, userID uuid.UUID) *models.TotalScoreWithdraw {
	totalScore := models.TotalScoreWithdraw{}

	score, _ := service.scoreRepository.GetScoreByUserID(ctx, userID)
	if score == nil {
		totalScore.Total = 0
	} else {
		totalScore.Total = score.Total
	}

	withdrawPoints, _ := service.transactionRepository.GetSumFundsWithdrawn(ctx, userID)
	totalScore.Withdraw = withdrawPoints

	return &totalScore
}
