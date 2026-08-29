package handler

import "net/http"

func (a *API) listNotifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := queryPage(q.Get("page"), q.Get("per_page"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out, err := a.notifications.List(r.Context(), q.Get("unread") == "true", page)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) readNotification(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.notifications.MarkRead(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) readAllNotifications(w http.ResponseWriter, r *http.Request) {
	if err := a.notifications.MarkAllRead(r.Context()); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
