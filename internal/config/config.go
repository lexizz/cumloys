package config

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

const (
	defaultHTTPPort         = "8000"
	defaultHTTPRWTimeout    = 10 * time.Second
	defaultRateLimiterRPS   = 10
	defaultRateLimiterBurst = 2
	defaultRateLimiterTTL   = 10 * time.Minute
	defaultStateDebugMode   = true
)

type (
	Config struct {
		HTTP           HTTPConfig
		Postgresql     PostgresqlConfig
		IncomingParams IncomingParams
		Limiter        LimiterConfig
		JWT            JWTConfig
	}

	IncomingParams struct {
		ServerAddress         string        `env:"RUN_ADDRESS"`
		DatabaseDSN           string        `env:"DATABASE_URI"`
		AccrualSystemAddress  string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
		IsDebugModeEnabled    bool          `env:"DEBUG_ENABLED"`
		SignatureAlgorithmJWT string        `env:"ALG_JWT"`
		SecretKeyJWT          string        `env:"SECRET_KEY_JWT"`
		ExpiryInJWT           time.Duration `env:"EXPIRY_JWT"`
	}

	PostgresqlConfig struct {
		DSN      string
		Username string
		Password string
		Host     string
		Port     string
		Database string
	}

	HTTPConfig struct {
		Address            string
		ReadTimeout        time.Duration
		WriteTimeout       time.Duration
		MaxHeaderMegabytes int
	}

	LimiterConfig struct {
		RPS   int
		Burst int
		TTL   time.Duration
	}

	JWTConfig struct {
		SignatureAlgorithm string
		SecretKeyJWT       string
		ExpiryIn           time.Duration
	}
)

func Init() *Config {
	var config Config

	fillConfigByEnvironments(&config)

	fillConfigByFlags(&config)

	config.HTTP = HTTPConfig{
		Address:            config.IncomingParams.ServerAddress,
		ReadTimeout:        defaultHTTPRWTimeout,
		WriteTimeout:       defaultHTTPRWTimeout,
		MaxHeaderMegabytes: http.DefaultMaxHeaderBytes,
	}

	config.Postgresql.DSN = config.IncomingParams.DatabaseDSN

	config.Limiter = LimiterConfig{
		RPS:   defaultRateLimiterRPS,
		Burst: defaultRateLimiterBurst,
		TTL:   defaultRateLimiterTTL,
	}

	config.JWT = JWTConfig{
		SignatureAlgorithm: config.IncomingParams.SignatureAlgorithmJWT,
		SecretKeyJWT:       config.IncomingParams.SecretKeyJWT,
		ExpiryIn:           config.IncomingParams.ExpiryInJWT,
	}

	return &config
}

func fillConfigByEnvironments(config *Config) {
	errEnv := env.Parse(&config.IncomingParams)
	if errEnv != nil {
		log.Println("=== ERROR FILLING FROM ENVIRONMENT", errEnv)
	}
}

func fillConfigByFlags(config *Config) {
	flagSet := pflag.FlagSet{}

	address := flagSet.StringP("http-address", "a", ":"+defaultHTTPPort, "address for listening via server")
	databaseDSN := flagSet.StringP("db-dsn", "d", "", "DSN of database")
	accrualAddress := flagSet.StringP("accrual-system-address", "r", "http://127.0.0.1:8081", "address of the accrual system")

	signatureAlgorithmJWT := flagSet.StringP("signature-alg-jwt", "g", "HS256", "signature algorithm for jwt")
	secretKeyJWT := flagSet.StringP("secret-key-jwt", "k", "default-key", "secret key for authentificate client")
	expiryInJWT := flagSet.DurationP("expiry-in-jwt", "e", 10*time.Minute, "expiry in for jwt")

	flagSet.BoolVar(&config.IncomingParams.IsDebugModeEnabled, "debug-mode-enabled", defaultStateDebugMode, "show additional logs")

	flagSet.Usage = func() {
		_, err := fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		if err != nil {
			return
		}

		flagSet.PrintDefaults()
		os.Exit(0)
	}

	errFlag := flagSet.Parse(os.Args[1:])
	if errFlag != nil {
		log.Printf("---> ERROR: failed flag parse: %+v; Args: %v\n", errFlag, os.Args)
	}

	if config.IncomingParams.ServerAddress == "" {
		config.IncomingParams.ServerAddress = *address
	}

	if config.IncomingParams.DatabaseDSN == "" {
		config.IncomingParams.DatabaseDSN = *databaseDSN
	}

	if config.IncomingParams.AccrualSystemAddress == "" {
		config.IncomingParams.AccrualSystemAddress = *accrualAddress
	}

	if config.IncomingParams.SignatureAlgorithmJWT == "" {
		config.IncomingParams.SignatureAlgorithmJWT = *signatureAlgorithmJWT
	}

	if config.IncomingParams.SecretKeyJWT == "" {
		config.IncomingParams.SecretKeyJWT = *secretKeyJWT
	}

	if config.IncomingParams.ExpiryInJWT == 0 {
		config.IncomingParams.ExpiryInJWT = *expiryInJWT
	}
}
