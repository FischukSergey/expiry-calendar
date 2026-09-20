package handler

import (
	"net/http"

	"github.com/swaggest/swgui/v5emb"
)

// openAPISpec отдаёт встроенную спеку. text/yaml — Swagger UI; no-store, чтобы PWA/браузер не подсунули старый index.html.
func (a *API) openAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(a.spec)
}

// swaggerUI — встроенный UI; спека с /openapi.yaml, сам UI с префикса /docs/.
func (a *API) swaggerUI() http.Handler {
	return v5emb.New("Duekeep API", "/openapi.yaml", "/docs/")
}
