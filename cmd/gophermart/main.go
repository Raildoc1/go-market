package main

import (
	"context"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"go-market/cmd/gophermart/Config"
	"go-market/internal/gophermart"
	"go-market/internal/gophermart/data/database"
	"go-market/internal/gophermart/data/dbrepository"
	"go-market/internal/gophermart/service"
	"go-market/pkg/jwtfactory"
	"go-market/pkg/logging"
	"go-market/pkg/pgxstorage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logging.NewZapLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}

	dbFactory := database.NewPgxDatabaseFactory(
		database.Config{
			ConnectionString: cfg.DB.ConnectionString,
		},
	)
	storage, err := pgxstorage.New(dbFactory)
	if err != nil {
		log.Fatal(err)
	}
	repository := dbrepository.New(storage, logger)
	transactionManager := pgxstorage.NewTransactionsManager(storage)

	tokenAuth := jwtauth.New(cfg.JWTConfig.Algorithm, []byte(cfg.JWTConfig.Secret), nil)
	tokenFactory := jwtfactory.New(tokenAuth, cfg.JWTConfig.ExpirationTime)

	registrationService := service.NewRegistration(repository, transactionManager, tokenFactory)
	loginService := service.NewLogin(repository, transactionManager, tokenFactory)

	server := gophermart.NewServer(cfg.Server, tokenAuth, registrationService, loginService, logger)

	rootCtx, cancelCtx := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
	defer cancelCtx()

	if err := run(rootCtx, cfg, server, logger); err != nil {
		logger.ErrorCtx(rootCtx, "Server shutdown with error", zap.Error(err))
	} else {
		logger.InfoCtx(rootCtx, "Server shutdown gracefully")
	}
}

func run(rootCtx context.Context, cfg *config.Config, server *gophermart.Server, logger *logging.ZapLogger) error {
	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the server")
	})

	g.Go(func() error {
		if err := server.Run(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer logger.InfoCtx(ctx, "Shutting down server")
		<-ctx.Done()
		if err := server.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("goroutine error occured: %w", err)
	}

	return nil
}
