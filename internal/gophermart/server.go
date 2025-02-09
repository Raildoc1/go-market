package gophermart

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go-market/internal/gophermart/handlers"
	"go-market/internal/gophermart/middleware"
	"go-market/pkg/logging"
	"net/http"
)

type Server struct {
	logger     *logging.ZapLogger
	httpServer *http.Server
	cfg        Config
}

type AuthorizationService interface {
	handlers.RegistrationService
	handlers.AuthorizationService
}

type OrdersService interface {
	handlers.OrderLoadingService
	handlers.OrderGettingService
}

type WalletService interface {
	handlers.BalanceGettingService
}

func NewServer(
	cfg Config,
	tokenAuth *jwtauth.JWTAuth,
	authorizationService AuthorizationService,
	ordersService OrdersService,
	walletService WalletService,
	logger *logging.ZapLogger,
) *Server {
	srv := &http.Server{
		Addr: cfg.ServerAddress,
		Handler: createMux(
			tokenAuth,
			authorizationService,
			ordersService,
			walletService,
			logger,
		),
	}

	res := &Server{
		cfg:        cfg,
		logger:     logger,
		httpServer: srv,
	}

	return res
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server ListenAndServe failed: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	return nil
}

func createMux(
	tokenAuth *jwtauth.JWTAuth,
	authorizationService AuthorizationService,
	ordersService OrdersService,
	walletService WalletService,
	logger *logging.ZapLogger,
) *chi.Mux {

	registrationHandler := handlers.NewRegisterHandler(authorizationService, logger)
	authorizationHandler := handlers.NewAuthorizationHandler(authorizationService, logger)
	orderLoadingHandler := handlers.NewOrderLoadingHandler(ordersService, logger)
	orderGettingHandler := handlers.NewOrderGettingHandler(ordersService, logger)
	balanceGettingHandler := handlers.NewBalanceGettingHandler(walletService, logger)

	loggerContextMiddleware := middleware.NewLoggerContext()
	panicRecover := middleware.NewPanicRecover(logger)

	router := chi.NewRouter()

	router.Use(loggerContextMiddleware.CreateHandler)
	router.Use(panicRecover.CreateHandler)
	router.Route("/api/user/", func(router chi.Router) {
		router.Post("/register", registrationHandler.ServeHTTP)
		router.Post("/login", authorizationHandler.ServeHTTP)
		router.With(
			jwtauth.Verifier(tokenAuth),
			jwtauth.Authenticator(tokenAuth),
		).Route("/", func(router chi.Router) {
			router.Post("/orders", orderLoadingHandler.ServeHTTP)
			router.Get("/orders", orderGettingHandler.ServeHTTP)
			router.Get("/balance", balanceGettingHandler.ServeHTTP)
		})
	})

	return router
}
