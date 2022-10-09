package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file" // driver to open file with migrations

	configPackage "github.com/lexizz/cumloys/internal/config"
	"github.com/lexizz/cumloys/internal/db/dbclient/postgresql"
	"github.com/lexizz/cumloys/internal/models"
	pkgLogger "github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository/orderrepository"
	"github.com/lexizz/cumloys/internal/repository/scorerepository"
	"github.com/lexizz/cumloys/internal/repository/transactionrepository"
	"github.com/lexizz/cumloys/internal/repository/userrepository"
	"github.com/lexizz/cumloys/internal/server"
	"github.com/lexizz/cumloys/internal/service"
	"github.com/lexizz/cumloys/internal/service/createorderservice"
	"github.com/lexizz/cumloys/internal/service/createuserservice"
	"github.com/lexizz/cumloys/internal/service/findbalanceservice"
	"github.com/lexizz/cumloys/internal/service/findorderservice"
	"github.com/lexizz/cumloys/internal/service/finduserservice"
	"github.com/lexizz/cumloys/internal/service/findwithdrawpointsservice"
	"github.com/lexizz/cumloys/internal/service/gettingpointsservice"
	"github.com/lexizz/cumloys/internal/service/withdrawpointsservice"
	"github.com/lexizz/cumloys/internal/transport/http/handler"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := configPackage.Init()
	logger := pkgLogger.Init()

	poolConnection, errorConnectDB := postgresql.NewClient(ctx, 5, config.Postgresql, logger)
	if errorConnectDB != nil {
		logger.Errorf("---> ERROR: failed connect to database: %v\n", errorConnectDB)
		return
	}

	resultInitialize := InitializingDatabase(config.Postgresql, logger)
	if !resultInitialize {
		return
	}

	userRepo := userrepository.New(poolConnection, logger)
	orderRepo := orderrepository.New(poolConnection, logger)
	scoreRepo := scorerepository.New(poolConnection, logger)
	transactionRepo := transactionrepository.New(poolConnection, logger)

	createUserService := createuserservice.New(userRepo, logger)
	findUserService := finduserservice.New(userRepo, logger)
	createOrderService := createorderservice.New(orderRepo, transactionRepo, logger)
	findOrderService := findorderservice.New(orderRepo, logger)
	findBalanceService := findbalanceservice.New(scoreRepo, transactionRepo, logger)
	gettingPointsService := gettingpointsservice.New(config, &http.Client{}, createOrderService, orderRepo, scoreRepo, transactionRepo, logger)
	withdrawPointsService := withdrawpointsservice.New(orderRepo, scoreRepo, transactionRepo, logger)
	findWithdrawPointsService := findwithdrawpointsservice.New(transactionRepo, logger)

	services := service.Services{
		CreateUserService:         createUserService,
		FindUserService:           findUserService,
		CreateOrderService:        createOrderService,
		FindOrderService:          findOrderService,
		FindBalanceService:        findBalanceService,
		GettingPointsService:      gettingPointsService,
		WithdrawPointsService:     withdrawPointsService,
		FindWithdrawPointsService: findWithdrawPointsService,
	}

	jwt, errToken := models.NewJWT(config.JWT.SignatureAlgorithm, config.JWT.SecretKeyJWT, config.JWT.ExpiryIn)
	if errToken != nil {
		logger.Errorf("---> ERROR: create token: %v ======\n", errorConnectDB)
		return
	}

	handlers := handler.New(config, logger, &services, jwt)
	srv := server.New(ctx, config, handlers.Init(), logger)
	if srv == nil {
		logger.Error("---> ERROR: failed starting server")
		return
	}

	signalChanel := make(chan os.Signal, 1)
	defer close(signalChanel)

	logger.Info("=== Starting server ...")

	go func() {
		logger.Info("=== Server started ===")

		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("---> ERROR: failed run http server: %s\n", err.Error())
		}

		signalChanel <- os.Interrupt
	}()

	signal.Notify(signalChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-signalChanel
	cancel()

	logger.Info("=== HandleSignals: signal detected - ", sig.String(), "; Stopping server...")

	ctxTimeout, cancelCtxTimeout := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		logger.Info("=== soft server shutdown process has started... ===")
		cancelCtxTimeout()
		logger.Info("=== soft server shutdown has finished!!! ===")
	}()

	if err := srv.Stop(ctxTimeout); err != nil {
		logger.Errorf("---> ERROR: server shutdown failed: %v", err)
		return
	}
}

func InitializingDatabase(cfg configPackage.PostgresqlConfig, logger pkgLogger.Logger) bool {
	logger.Info("=== Initializing the database... ")

	db, errOpen := sql.Open("pgx", cfg.DSN)
	if errOpen != nil {
		logger.Errorf("---> ERROR: MIGRATE: failed open connect to database: %v\n", errOpen)
		return false
	}

	driver, errPGXInstance := pgx.WithInstance(db, &pgx.Config{})
	if errPGXInstance != nil {
		logger.Errorf("---> ERROR: MIGRATE: failed instance pgx driver: %v\n", errPGXInstance)
		return false
	}

	urlParsed, errParseURL := url.Parse(cfg.DSN)
	if errParseURL != nil {
		logger.Errorf("---> ERROR: MIGRATE: failed parse dns for getting dbname: %v\n", errParseURL)
		return false
	}

	dbName := strings.ReplaceAll(urlParsed.Path, "/", "")

	migrateInstance, errInst := migrate.NewWithDatabaseInstance("file://internal/db/migrations", dbName, driver)
	// migrateInstance, errInst := migrate.NewWithDatabaseInstance("file://../../internal/db/migrations", dbName, driver)
	// migrateInstance, errInst := migrate.NewWithDatabaseInstance("file://../cumloys/internal/db/migrations", dbName, driver)
	if errInst != nil {
		logger.Errorf("---> ERROR: MIGRATE: failed instance migrate: %v\n", errInst)
		return false
	}

	errUp := migrateInstance.Up()
	if errUp != nil {
		logger.Info("=== changes for migrate not found [" + errUp.Error() + "]")
	}

	errCloseSource, errDBClose := migrateInstance.Close()
	if errCloseSource != nil || errDBClose != nil {
		logger.Errorf("---> ERROR: MIGRATE: failed close db for migrate: %v; %v\n", errCloseSource, errDBClose)
		return false
	}

	logger.Info("=== Initializing finished === ")

	return true
}
