package handler

import (
	"net/http"

	"github.com/swaggest/swgui/v5emb"
)

// openAPISpec отдаёт встроенную спеку. Content-Type application/yaml, не download-файл.
func (a *API) openAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(a.spec)
}

// swaggerUI — встроенный UI; спека с /openapi.yaml, сам UI с префикса /docs/.
func (a *API) swaggerUI() http.Handler {
	return v5emb.New("Duekeep API", "/openapi.yaml", "/docs/")
}
