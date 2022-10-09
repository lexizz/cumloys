package orderrepository

import (
	"context"
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

type orderRepository struct {
	client  dbclient.ClientInterface
	rwMutex *sync.RWMutex
	logger  logger.Logger
}

var _ repository.OrderRepositoryInterface = &orderRepository{}

func New(client dbclient.ClientInterface, logger logger.Logger) *orderRepository {
	rwMutex := sync.RWMutex{}

	ordRepository := orderRepository{
		client:  client,
		rwMutex: &rwMutex,
		logger:  logger,
	}

	return &ordRepository
}

func (rep *orderRepository) IsExists(ctx context.Context, number string) (bool, *uuid.UUID, *uuid.UUID, error) {
	query := `SELECT id, user_id FROM orders WHERE number = $1;`

	var orderID uuid.UUID
	var userID uuid.UUID

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	err := rep.client.QueryRow(ctx, query, number).Scan(&orderID, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil, nil
		}

		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: IsExists order: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return false, nil, nil, err
	}

	return true, &orderID, &userID, nil
}

func (rep *orderRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	query := `SELECT number, status, points, updated_at 
			FROM orders 
			WHERE user_id = $1 
			ORDER BY created_at DESC;`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	rows, errQuery := rep.client.Query(ctx, query, userID.String())
	if errQuery != nil {
		rep.logger.Errorf("---> ERROR: orderRepository: query in GetAllByUserID: %v\n", errQuery)
		return nil, errQuery
	}

	defer rows.Close()

	orders := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order

		err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Points,
			&order.UpdatedAt,
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

		newTime := time.Unix(order.UpdatedAt.Unix(), 0).Format(time.RFC3339)
		newUpdatedAt, errTimeParse := time.Parse(time.RFC3339, newTime)
		if errTimeParse != nil {
			rep.logger.Errorf("---> ERROR: GetAllByUserID: failed convert update_at: %v\n", errTimeParse)
			newUpdatedAt = order.UpdatedAt
		}

		order.UpdatedAt = newUpdatedAt

		orders = append(orders, order)
	}

	if errRows := rows.Err(); errRows != nil {
		rep.logger.Errorf("---> ERROR: GetAllByUserID: rows next: %v\n", errRows)
		return nil, errRows
	}

	return orders, nil
}

func (rep *orderRepository) Insert(ctx context.Context, number string, userID uuid.UUID) (*uuid.UUID, error) {
	query := `INSERT INTO orders (number, user_id, created_at, updated_at) 
				VALUES ($1, $2, $3, $4) RETURNING id`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	var lastInsertID uuid.UUID
	var lastInsert string

	currentDatetime := utils.GetCurrentDatetimeUTC()

	err := rep.client.QueryRow(
		ctx,
		query,
		number,
		userID.String(),
		currentDatetime,
		currentDatetime,
	).Scan(&lastInsert)
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: insert order: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return nil, err
	}

	rep.logger.Info("====== Insert Order: OK ======")

	lastInsertID, errParse := uuid.Parse(lastInsert)
	if errParse != nil {
		rep.logger.Errorf("---> ERROR: %v\n", errParse)
	}

	return &lastInsertID, nil
}

func (rep *orderRepository) Update(ctx context.Context, number string, userID uuid.UUID, status string, points float32) error {
	query := `UPDATE orders SET (status, points, updated_at) = ($1, $2, $3) WHERE number = $4 AND user_id = $5;`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	_, err := rep.client.Exec(ctx, query, status, points, utils.GetCurrentDatetimeUTC(), number, userID.String())
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: failed update order: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return err
	}

	return nil
}
