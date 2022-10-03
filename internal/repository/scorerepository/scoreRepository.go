package scorerepository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/lexizz/cumloys/internal/db/dbclient"
	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/repository"
)

type scoreRepository struct {
	client  dbclient.ClientInterface
	rwMutex *sync.RWMutex
	logger  logger.Logger
}

var _ repository.ScoreRepositoryInterface = &scoreRepository{}

func New(client dbclient.ClientInterface, logger logger.Logger) *scoreRepository {
	rwMutex := sync.RWMutex{}

	scrRepository := scoreRepository{
		client:  client,
		rwMutex: &rwMutex,
		logger:  logger,
	}

	return &scrRepository
}

func (rep *scoreRepository) GetScoreByUserID(ctx context.Context, userID uuid.UUID) (*models.Score, error) {
	query := `SELECT id, total, user_id, created_at, updated_at FROM score WHERE user_id=$1`

	var score models.Score

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	err := rep.client.QueryRow(ctx, query, userID.String()).Scan(
		&score.ID,
		&score.Total,
		&score.UserID,
		&score.CreatedAt,
		&score.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: scoreRepository: GetScoreByUserId: %v; Type:%[1]T\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return nil, err
	}

	if len(score.ID.String()) == 0 {
		return nil, nil
	}

	return &score, nil
}

func (rep *scoreRepository) Insert(ctx context.Context, userID uuid.UUID, points float32) (*uuid.UUID, error) {
	query := `INSERT INTO score (user_id, total, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	var lastInsertID uuid.UUID
	currentDatetime := utils.GetCurrentDatetimeUTC()

	err := rep.client.QueryRow(ctx, query, userID.String(), points, currentDatetime, currentDatetime).Scan(&lastInsertID)
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: failed insert score: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return nil, err
	}

	rep.logger.Info("====== Insert Score: OK ======")

	return &lastInsertID, nil
}

func (rep *scoreRepository) Update(ctx context.Context, userID uuid.UUID, points float32) error {
	query := `UPDATE score SET (user_id, total, updated_at) = ($1, $2, $3) WHERE user_id = $4;`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	_, err := rep.client.Exec(ctx, query, userID.String(), points, utils.GetCurrentDatetimeUTC(), userID.String())
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: failed update score: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return err
	}

	return nil
}
