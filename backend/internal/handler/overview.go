package handler

import (
	"net/http"
	"strconv"

	"duekeep/internal/middleware"
)

func (a *API) dashboard(w http.ResponseWriter, r *http.Request) {
	if a.overview == nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal")
		return
	}
	out, err := a.overview.Dashboard(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) calendar(w http.ResponseWriter, r *http.Request) {
	if a.overview == nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal")
		return
	}
	q := r.URL.Query()
	year, err := strconv.Atoi(q.Get("year"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid year")
		return
	}
	month, err := strconv.Atoi(q.Get("month"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid month")
		return
	}
	out, err := a.overview.Calendar(r.Context(), year, month, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}
