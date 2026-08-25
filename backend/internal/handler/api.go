package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"duekeep/internal/middleware"
)

// API — HTTP-вход приложения.
type API struct {
	health       HealthService
	auth         AuthService
	kinds        KindService
	categories   CategoryService
	spec         []byte // сырой openapi.yaml, тот же duekeep.OpenAPISpec.
	jwtSecret    []byte
	cookieSecure bool
	refreshTTL   time.Duration
}

// New собирает handlers.
func New(d Deps) *API {
	return &API{
		health:       d.Health,
		auth:         d.Auth,
		kinds:        d.Kinds,
		categories:   d.Categories,
		spec:         d.Spec,
		jwtSecret:    d.JWTSecret,
		cookieSecure: d.CookieSecure,
		refreshTTL:   d.RefreshTTL,
	}
}

// Router отдаёт chi-роутер: /healthz и /docs без auth, Bearer только на /me и logout-all.
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", a.register)
		r.Post("/auth/login", a.login)
		r.Post("/auth/refresh", a.refresh)
		r.With(middleware.OptionalBearer(a.jwtSecret)).Post("/auth/logout", a.logout)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Bearer(a.jwtSecret))
			r.Post("/auth/logout-all", a.logoutAll)
			r.Get("/me", a.me)
			r.Get("/kinds", a.listKinds)
			r.Get("/categories", a.listCategories)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Post("/kinds", a.createKind)
				r.Patch("/kinds/{id}", a.patchKind)
				r.Delete("/kinds/{id}", a.deleteKind)
				r.Post("/categories", a.createCategory)
				r.Patch("/categories/{id}", a.patchCategory)
				r.Delete("/categories/{id}", a.deleteCategory)
			})
		})
	})
	return r
}

// requestLog пишет method/path/status/ms. Секреты в query не ожидаем и не логируем.
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

// statusWriter запоминает код ответа: у ResponseWriter его не прочитать после записи.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
