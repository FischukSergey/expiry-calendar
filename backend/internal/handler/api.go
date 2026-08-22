package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// API — HTTP-вход приложения.
type API struct {
	health HealthService
	spec   []byte
}

// New собирает handlers.
func New(health HealthService, spec []byte) *API {
	return &API{health: health, spec: spec}
}

// Router возвращает chi-роутер с /healthz и Swagger.
func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(requestLog)
	r.Get("/healthz", a.healthz)
	r.Get("/openapi.yaml", a.openAPISpec)
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/docs/", http.StatusFound)
	})
	ui := a.swaggerUI()
	r.Handle("/docs/", ui)
	r.Handle("/docs/*", ui)
	return r
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.InfoContext(r.Context(), "http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
