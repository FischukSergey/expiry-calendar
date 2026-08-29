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

	duekeep "duekeep"
	"duekeep/internal/clock"
	"duekeep/internal/db"
	"duekeep/internal/handler"
	"duekeep/internal/repository"
	"duekeep/internal/seed"
	"duekeep/internal/service"
	"duekeep/internal/sse"
	"duekeep/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// run поднимает slog, пул, goose, seed и HTTP. По SIGINT/SIGTERM — Shutdown за 10 с.
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
		"jwt_access_ttl", cfg.AccessTTL.String(),
		"jwt_refresh_ttl", cfg.RefreshTTL.String(),
		"cookie_secure", cfg.CookieSecure,
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
	if err := seed.Run(ctx, pool, clock.Real{}); err != nil {
		return err
	}

	users := repository.NewUsers(pool)
	refresh := repository.NewRefreshTokens(pool)
	runTx := func(ctx context.Context, fn func(context.Context) error) error {
		return db.RunTx(ctx, pool, fn)
	}
	clk := clock.Real{}
	auth := service.NewAuth(users, refresh, runTx, clk, service.AuthConfig{
		Secret:     []byte(cfg.JWTSecret),
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
	})
	kindsRepo := repository.NewKinds(pool)
	catsRepo := repository.NewCategories(pool)
	itemsRepo := repository.NewItems(pool)
	notesRepo := repository.NewNotifications(pool)
	itemsSvc := service.NewItem(
		itemsRepo, kindsRepo, catsRepo,
		repository.NewRenewals(pool), repository.NewAudit(pool), runTx, clk,
	)
	hub := sse.NewHub()
	api := handler.New(handler.Deps{
		Health:        service.NewHealth(repository.NewHealth(pool)),
		Auth:          auth,
		Kinds:         service.NewKind(kindsRepo),
		Categories:    service.NewCategory(catsRepo),
		Items:         itemsSvc,
		Notifications: service.NewNotification(notesRepo),
		Hub:           hub,
		Spec:          duekeep.OpenAPISpec,
		JWTSecret:     []byte(cfg.JWTSecret),
		CookieSecure:  cfg.CookieSecure,
		RefreshTTL:    cfg.RefreshTTL,
	})
	tkr := service.NewTicker(itemsRepo, notesRepo, runTx, clk, hub)
	go tkr.Run(ctx, 60*time.Second)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 5 * time.Second, // отсекает зависший заголовок (slowloris).
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

// config — env процесса. JWT_SECRET без порога длины: в local compose 19 символов.
type config struct {
	HTTPAddr     string
	DatabaseURL  string
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	CookieSecure bool
}

// loadConfig: HTTP_ADDR, DATABASE_URL, JWT_SECRET обязателен; TTL с дефолтами 15m / 336h.
func loadConfig() (config, error) {
	accessTTL, err := time.ParseDuration(cmp.Or(os.Getenv("JWT_ACCESS_TTL"), "15m"))
	if err != nil {
		return config{}, errors.New("JWT_ACCESS_TTL is invalid")
	}
	refreshTTL, err := time.ParseDuration(cmp.Or(os.Getenv("JWT_REFRESH_TTL"), "336h"))
	if err != nil {
		return config{}, errors.New("JWT_REFRESH_TTL is invalid")
	}
	cfg := config{
		HTTPAddr:     cmp.Or(os.Getenv("HTTP_ADDR"), ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
		CookieSecure: os.Getenv("COOKIE_SECURE") == "true",
	}
	if cfg.DatabaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return config{}, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}

// redactDSN маскирует пароль в DSN для slog. Невалидный URL → "invalid".
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
