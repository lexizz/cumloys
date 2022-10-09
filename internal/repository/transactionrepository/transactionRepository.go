package transactionrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/lexizz/cumloys/internal/db/dbclient"
	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/repository"
)

const (
	IncreaseNumberPointsType int = 1
	DecreaseNumberPointsType int = 2
)

type transactionRepository struct {
	client  dbclient.ClientInterface
	rwMutex *sync.RWMutex
	logger  logger.Logger
}

var _ repository.TransactionRepositoryInterface = &transactionRepository{}

func New(client dbclient.ClientInterface, logger logger.Logger) *transactionRepository {
	rwMutex := sync.RWMutex{}

	transactRepository := transactionRepository{
		client:  client,
		rwMutex: &rwMutex,
		logger:  logger,
	}

	return &transactRepository
}

func (rep *transactionRepository) Insert(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, points float32, typeTransaction int) error {
	query := `INSERT INTO transactions (user_id, order_id, points, type, created_at) VALUES ($1, $2, $3, $4, $5)`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	_, err := rep.client.Exec(ctx, query,
		userID.String(),
		orderID.String(),
		points,
		typeTransaction,
		utils.GetCurrentDatetimeUTC(),
	)
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: insert transaction: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return err
	}

	rep.logger.Info("====== Insert Transaction: OK ======")

	return nil
}

func (rep *transactionRepository) GetSumFundsWithdrawn(ctx context.Context, userID uuid.UUID) (float32, error) {
	query := `SELECT SUM(points) FROM transactions WHERE user_id = $1 AND type = $2;`

	var withdrawPoints sql.NullFloat64

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	err := rep.client.QueryRow(ctx, query, userID.String(), models.DecreasePointsType).Scan(&withdrawPoints)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: GetSumFundsWithdrawn: %v; Type:%[1]T\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return 0, err
	}

	return float32(withdrawPoints.Float64), nil
}

func (rep *transactionRepository) GetAllFundsWithdrawn(ctx context.Context, userID uuid.UUID) ([]models.ScoreWithdraw, error) {
	query := `SELECT t.points, o.number, t.created_at
			FROM transactions AS t
			INNER JOIN orders o on o.id = t.order_id
			WHERE t.user_id = $1 AND t.type = $2 ORDER BY t.created_at ASC`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	rows, errQuery := rep.client.Query(ctx, query, userID.String(), models.DecreasePointsType)
	if errQuery != nil {
		rep.logger.Errorf("---> ERROR: transactionRepository: query in GetAllFundsWithdrawn: %v\n", errQuery)
		return nil, errQuery
	}

	defer rows.Close()

	scoreWithdraws := make([]models.ScoreWithdraw, 0)

	for rows.Next() {
		var scoreWithdraw models.ScoreWithdraw

		err := rows.Scan(
			&scoreWithdraw.SumWithdraw,
			&scoreWithdraw.NumberOrder,
			&scoreWithdraw.CreatedAt,
		)
		if err != nil {
			var pgErr pgconn.PgError

			errorMessage := fmt.Sprintf("---> ERROR: get row from scan: %v\n", err)

			if errors.Is(err, &pgErr) {
				errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
					pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			}

			rep.logger.Errorf(errorMessage)

			return nil, err
		}

		newTime := time.Unix(scoreWithdraw.CreatedAt.Unix(), 0).Format(time.RFC3339)
		newUpdatedAt, errTimeParse := time.Parse(time.RFC3339, newTime)
		if errTimeParse != nil {
			rep.logger.Errorf("---> ERROR: GetAllFundsWithdrawn: failed convert update_at: %v\n", errTimeParse)
			newUpdatedAt = scoreWithdraw.CreatedAt
		}

		scoreWithdraw.CreatedAt = newUpdatedAt

		scoreWithdraws = append(scoreWithdraws, scoreWithdraw)
	}

	if errRows := rows.Err(); errRows != nil {
		rep.logger.Errorf("---> ERROR: GetAllFundsWithdrawn: rows next: %v\n", errRows)
		return nil, errRows
	}

	return scoreWithdraws, nil
}
