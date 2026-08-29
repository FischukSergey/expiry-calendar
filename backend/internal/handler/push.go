package handler

import (
	"net/http"

	"duekeep/internal/middleware"
	"duekeep/internal/model"
)

func (a *API) vapidPublic(w http.ResponseWriter, _ *http.Request) {
	key := ""
	if a.push != nil {
		key = a.push.PublicKey()
	}
	writeBytes(w, http.StatusOK, model.VAPIDPublic{PublicKey: key})
}

func (a *API) pushSubscribe(w http.ResponseWriter, r *http.Request) {
	if a.push == nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var body model.PushSubscribe
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	if err := a.push.Subscribe(r.Context(), middleware.UserID(r.Context()), body, r.UserAgent()); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) pushUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if a.push == nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var body model.PushUnsubscribe
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	if err := a.push.Unsubscribe(r.Context(), body.Endpoint); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
