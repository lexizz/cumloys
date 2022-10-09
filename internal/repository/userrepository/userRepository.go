package userrepository

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

type userRepository struct {
	client  dbclient.ClientInterface
	rwMutex *sync.RWMutex
	logger  logger.Logger
}

var _ repository.UserRepositoryInterface = &userRepository{}

func New(client dbclient.ClientInterface, logger logger.Logger) *userRepository {
	rwMutex := sync.RWMutex{}

	uRepository := userRepository{
		client:  client,
		rwMutex: &rwMutex,
		logger:  logger,
	}

	return &uRepository
}

func (rep *userRepository) IsExists(ctx context.Context, login string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE login=$1`

	var numberOfUsers int

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	err := rep.client.QueryRow(ctx, query, login).Scan(&numberOfUsers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR IsUserExists: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return true, err
	}

	return numberOfUsers > 0, nil
}

func (rep *userRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password, created_at FROM users WHERE login=$1`

	var user models.User

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	err := rep.client.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR GetUserByLogin: %v; Type:%[1]T\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return nil, err
	}

	if len(user.ID.String()) == 0 {
		return nil, nil
	}

	return &user, nil
}

func (rep *userRepository) Insert(ctx context.Context, newLogin string, newPassword string) (*uuid.UUID, error) {
	query := `INSERT INTO users (login, password, created_at) VALUES ($1, $2, $3) RETURNING id`

	rep.rwMutex.Lock()
	defer rep.rwMutex.Unlock()

	var lastInsertID uuid.UUID
	var lastInsert string

	err := rep.client.QueryRow(ctx, query, newLogin, newPassword, utils.GetCurrentDatetimeUTC()).Scan(&lastInsert)
	if err != nil {
		var pgErr pgconn.PgError

		errorMessage := fmt.Sprintf("---> ERROR: insert new user: %v\n", err)

		if errors.Is(err, &pgErr) {
			errorMessage += fmt.Sprintf("---> SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		}

		rep.logger.Errorf(errorMessage)

		return nil, err
	}

	rep.logger.Info("====== Insert User: OK ======")

	lastInsertID, errParse := uuid.Parse(lastInsert)
	if errParse != nil {
		rep.logger.Errorf("---> ERROR: %v\n", errParse)
	}

	return &lastInsertID, nil
}
