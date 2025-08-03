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

	// Load OpenAPI
	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenAPI spec", "err", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	// Router
	router := chi.NewRouter()
	router.Use(middleware.OapiRequestValidator(swagger))

	// Connect to DB
	ctx := context.Background()
	pool, err := pg.Connect(ctx)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing database connection pool")
		pool.Close()
	}()

	// Init app
	q := pg.New(pool)
	st := store.NewPostgresStore(q)
	svc := hotel.NewService(st)
	handler := rest.NewHandler(svc)
	api.HandlerFromMux(handler, router)

	// Create server with timeouts
	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// Signal context
	ctxShutdown, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("🚀 Server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for signal
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
