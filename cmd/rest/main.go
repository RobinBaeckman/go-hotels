package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/robinbaeckman/go-hotels/api"
	"github.com/robinbaeckman/go-hotels/internal/config"
	"github.com/robinbaeckman/go-hotels/internal/hotel"
	"github.com/robinbaeckman/go-hotels/internal/store"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"
	"github.com/robinbaeckman/go-hotels/internal/telemetry"
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

	// Context (you can reuse this for OTEL init)
	ctx := context.Background()

	// Start OpenTelemetry (send to Alloy)
	shutdownOTEL, err := telemetry.Setup(ctx, telemetry.Config{
		Endpoint:       os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), // prefer env; fallback if empty
		Insecure:       true,                                     // plaintext to Alloy in compose
		ServiceName:    "api",
		ServiceVersion: "0.1.0",
		Environment:    cfg.AppEnv, // if you have it; otherwise hardcode "local"
	})
	if err != nil {
		slog.Error("failed to init telemetry", "err", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownOTEL(context.Background()) }()

	// Load OpenAPI spec
	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenAPI spec", "err", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	router := chi.NewRouter()

	// 1) OTel server instrumentation (skapar span + metrics)
	router.Use(otelhttp.NewMiddleware(
		"api",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			if rc := chi.RouteContext(r.Context()); rc != nil {
				if pat := rc.RoutePattern(); pat != "" {
					return r.Method + " " + pat
				}
			}
			return r.Method + " " + r.URL.Path
		}),

		otelhttp.WithFilter(func(r *http.Request) bool {
			p := r.URL.Path
			return p != "/health" && p != "/ready"
		}),
	))

	// 2) ➜ Nytt: våra egna metrics med http_route/method/status
	router.Use(rest.HTTPMetricsMiddleware())

	// 3) Logger
	router.Use(rest.LoggerMiddleware(logger))
	router.Use(middleware.OapiRequestValidator(swagger))

	// DB
	pool, err := pg.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("closing database connection pool")
		pool.Close()
	}()

	// Services / handlers
	q := pg.New(pool)
	st := store.NewPostgresStore(q)
	svc := hotel.NewService(st, logger)
	handler := rest.NewHandler(svc, pool, logger)
	api.HandlerFromMux(handler, router)

	// HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx // server ärver main's context
		},
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
