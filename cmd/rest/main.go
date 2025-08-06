package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/config"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/store"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"
	"github.com/robinbaeckman/go-hotels/internal/transport/rest"
)

func main() {
	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Load OpenAPI spec
	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenAPI spec", "err", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	// Router
	router := chi.NewRouter()
	router.Use(middleware.OapiRequestValidator(swagger))

	// Connect to DB for the app itself
	ctx := context.Background()
	pool, err := pg.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing database connection pool")
		pool.Close()
	}()

	// Init app services
	q := pg.New(pool)
	st := store.NewPostgresStore(q)
	svc := hotel.NewService(st)
	handler := rest.NewHandler(svc, pool)
	api.HandlerFromMux(handler, router)

	// HTTP server
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// Signal handling
	ctxShutdown, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("🚀 Server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctxShutdown.Done()
	slog.Warn("shutdown signal received")

	// Graceful shutdown
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	} else {
		slog.Info("server shut down cleanly")
	}
}
