package config

import (
	"flag"
	"go-market/internal/gophermart"
	"go-market/internal/gophermart/data/database"
	"os"
	"time"
)

const (
	serverAddressFlag           = "a"
	serverAddressEnv            = "RUN_ADDRESS"
	serverAddressDefault        = "localhost:8080"
	accrualSystemAddressFlag    = "r"
	accrualSystemAddressEnv     = "ACCRUAL_SYSTEM_ADDRESS"
	accrualSystemAddressDefault = "localhost:8081"
	dbConnectionStringFlag      = "d"
	dbConnectionStringEnv       = "DATABASE_URI"
	dbConnectionStringDefault   = ""
)

type Config struct {
	Server          gophermart.Config
	JWTConfig       JWTConfig
	DB              database.Config
	ShutdownTimeout time.Duration
}

type JWTConfig struct {
	Algorithm      string
	Secret         string
	ExpirationTime time.Duration
}

func Load() (*Config, error) {
	serverAddress := flag.String(
		serverAddressFlag,
		serverAddressDefault,
		"Server address host:port",
	)

	accrualSystemAddress := flag.String(
		accrualSystemAddressFlag,
		accrualSystemAddressDefault,
		"Accrual system address host:port",
	)

	dbConnectionString := flag.String(
		dbConnectionStringFlag,
		dbConnectionStringDefault,
		"PostgreSQL connection string",
	)

	flag.Parse()

	if valStr, ok := os.LookupEnv(serverAddressEnv); ok {
		*serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(accrualSystemAddressEnv); ok {
		*accrualSystemAddress = valStr
	}

	if valStr, ok := os.LookupEnv(dbConnectionStringEnv); ok {
		*dbConnectionString = valStr
	}

	return &Config{
		Server: gophermart.Config{
			ServerAddress:   *serverAddress,
			ShutdownTimeout: time.Second * 5,
		},
		JWTConfig: JWTConfig{
			Algorithm:      "HS256",
			Secret:         "secret",
			ExpirationTime: time.Hour,
		},
		DB: database.Config{
			ConnectionString: *dbConnectionString,
		},
		ShutdownTimeout: time.Second * 5,
	}, nil
}
