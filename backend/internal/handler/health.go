package handler

import (
	"context"
	"net/http"
)

// HealthService проверяет готовность зависимостей.
type HealthService interface {
	Check(ctx context.Context) error
}

// healthz: 200 при живой БД, иначе 503 code=internal (не 500 — процесс жив).
func (a *API) healthz(w http.ResponseWriter, r *http.Request) {
	if err := a.health.Check(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "internal", "database unavailable")
		return
	}
	writeHealthOK(w)
}
