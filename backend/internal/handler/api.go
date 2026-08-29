package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"duekeep/internal/middleware"
	"duekeep/internal/sse"
)

// API — HTTP-вход приложения.
type API struct {
	health        HealthService
	auth          AuthService
	kinds         KindService
	categories    CategoryService
	items         ItemService
	overview      OverviewService
	notifications NotificationService
	push          PushService
	hub           *sse.Hub
	ssePing       time.Duration
	spec          []byte // сырой openapi.yaml, тот же duekeep.OpenAPISpec.
	jwtSecret     []byte
	cookieSecure  bool
	refreshTTL    time.Duration
}

// New собирает handlers.
func New(d Deps) *API {
	return &API{
		health:        d.Health,
		auth:          d.Auth,
		kinds:         d.Kinds,
		categories:    d.Categories,
		items:         d.Items,
		overview:      d.Overview,
		notifications: d.Notifications,
		push:          d.Push,
		hub:           cmpHub(d.Hub),
		ssePing:       d.SSEPing,
		spec:          d.Spec,
		jwtSecret:     d.JWTSecret,
		cookieSecure:  d.CookieSecure,
		refreshTTL:    d.RefreshTTL,
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
		r.With(middleware.BearerOrQuery(a.jwtSecret)).Get("/events", a.events)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Bearer(a.jwtSecret))
			r.Post("/auth/logout-all", a.logoutAll)
			r.Get("/me", a.me)
			r.Get("/kinds", a.listKinds)
			r.Get("/categories", a.listCategories)
			r.Get("/items", a.listItems)
			r.Get("/items/{id}", a.getItem)
			r.Get("/dashboard", a.dashboard)
			r.Get("/calendar", a.calendar)
			r.Get("/notifications", a.listNotifications)
			r.Post("/notifications/read-all", a.readAllNotifications)
			r.Post("/notifications/{id}/read", a.readNotification)
			r.Get("/push/vapid-public", a.vapidPublic)
			r.Post("/push/subscribe", a.pushSubscribe)
			r.Delete("/push/subscribe", a.pushUnsubscribe)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Post("/kinds", a.createKind)
				r.Patch("/kinds/{id}", a.patchKind)
				r.Delete("/kinds/{id}", a.deleteKind)
				r.Post("/categories", a.createCategory)
				r.Patch("/categories/{id}", a.patchCategory)
				r.Delete("/categories/{id}", a.deleteCategory)
				r.Post("/items", a.createItem)
				r.Post("/items/bulk", a.bulkItems)
				r.Patch("/items/{id}", a.patchItem)
				r.Delete("/items/{id}", a.deleteItem)
				r.Post("/items/{id}/renew", a.renewItem)
				r.Get("/audit", a.listAudit)
			})
		})
	})
	return r
}

// requestLog пишет method/path/status/ms. Query (access_token) не логируем.
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

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func cmpHub(h *sse.Hub) *sse.Hub {
	if h != nil {
		return h
	}
	return sse.NewHub()
}
