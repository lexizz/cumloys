package findwithdrawpointsservice

import (
	"context"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
)

type findWithdrawPointsService struct {
	transactionRepository repository.TransactionRepositoryInterface
	logger                logger.Logger
}

func New(
	transactionRepository repository.TransactionRepositoryInterface,
	logger logger.Logger,
) *findWithdrawPointsService {
	return &findWithdrawPointsService{
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (service *findWithdrawPointsService) Handle(ctx context.Context, userID uuid.UUID) []models.ScoreWithdraw {
	transactions, _ := service.transactionRepository.GetAllFundsWithdrawn(ctx, userID)

	return transactions
}
