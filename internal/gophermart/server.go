package gophermart

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go-market/internal/gophermart/handlers"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	logger     *zap.Logger
	httpServer *http.Server
	cfg        Config
}

func New(
	cfg Config,
	registrationService handlers.RegistrationService,
	authorizationService handlers.AuthorizationService,
	logger *zap.Logger,
) *Server {
	srv := &http.Server{
		Addr: cfg.ServerAddress,
		Handler: createMux(
			registrationService,
			authorizationService,
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
	registrationService handlers.RegistrationService,
	authorizationService handlers.AuthorizationService,
	logger *zap.Logger,
) *chi.Mux {

	registrationHandler := handlers.NewRegisterHandler(registrationService, logger)
	authorizationHandler := handlers.NewAuthorizationHandler(authorizationService, logger)

	router := chi.NewRouter()

	router.Route("/api/user/", func(router chi.Router) {
		router.Post("/register", registrationHandler.ServeHTTP)
		router.Post("/login", authorizationHandler.ServeHTTP)
	})

	return router
}
