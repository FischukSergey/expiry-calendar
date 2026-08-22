package handler

import (
	"net/http"

	"github.com/swaggest/swgui/v5emb"
)

func (a *API) openAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(a.spec)
}

func (a *API) swaggerUI() http.Handler {
	return v5emb.New("Duekeep API", "/openapi.yaml", "/docs/")
}
