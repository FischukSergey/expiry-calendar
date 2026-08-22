// Command server запускает HTTP API Duekeep.
package main

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"duekeep/internal/db"
	"duekeep/internal/handler"
	"duekeep/internal/repository"
	"duekeep/internal/service"
	"duekeep/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	slog.Info("config",
		"http_addr", cfg.HTTPAddr,
		"database_url", redactDSN(cfg.DatabaseURL),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close(pool)

	if err := db.Migrate(ctx, pool, migrations.FS, "."); err != nil {
		return err
	}

	api := handler.New(service.NewHealth(repository.NewHealth(pool)))
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listen", "addr", cfg.HTTPAddr)
		if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case serveErr := <-errCh:
		return serveErr
	}
}

type config struct {
	HTTPAddr    string
	DatabaseURL string
}

func loadConfig() (config, error) {
	cfg := config{
		HTTPAddr:    cmp.Or(os.Getenv("HTTP_ADDR"), ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func redactDSN(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "invalid"
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}
