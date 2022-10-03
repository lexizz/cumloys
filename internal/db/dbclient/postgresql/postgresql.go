package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/lexizz/cumloys/internal/config"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/pkg/utils"
)

func NewClient(ctx context.Context, maxAttempts int, config config.PostgresqlConfig, logger logger.Logger) (*pgxpool.Pool, error) {
	var dsn string

	logger.Info("=== GET DSN ...")

	switch {
	case config.DSN != "":
		dsn = config.DSN
	case config.Username != "" && config.Password != "":
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database)
	default:
		return nil, errors.New("data for connect not found")
	}

	logger.Infof("=== DSN: %v", dsn)

	var pool *pgxpool.Pool

	errDoWithTries := utils.DoWithTries(func() error {
		logger.Info("=== Getting a connect to db ...")

		ctxTimeout, cancelCtxTimeout := context.WithTimeout(ctx, 5*time.Second)
		defer cancelCtxTimeout()

		var err error

		pool, err = pgxpool.Connect(ctxTimeout, dsn)
		if err != nil {
			return err
		}

		logger.Info("=== Getting a connect to db - done")

		return nil
	}, maxAttempts, 5*time.Second)

	if errDoWithTries != nil {
		logger.Error("---> ERROR at the time connecting:", errDoWithTries)

		return nil, errDoWithTries
	}

	if pool == nil {
		return nil, errors.New("error connect to database")
	}

	if errPing := pool.Ping(ctx); errPing != nil {
		logger.Error("---> ERROR connect:", errPing)
		return nil, errPing
	}

	return pool, nil
}
