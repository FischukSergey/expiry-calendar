package handler

import (
	"fmt"
	"net/http"
	"time"

	"duekeep/internal/sse"
)

const defaultSSEPing = 15 * time.Second

func (a *API) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal", "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	writeSSE(w, flusher, sse.EventPing, []byte("{}"))

	id, ch := a.hub.Subscribe()
	defer a.hub.Unsubscribe(id)

	pingEvery := a.ssePing
	if pingEvery <= 0 {
		pingEvery = defaultSSEPing
	}
	tick := time.NewTicker(pingEvery)
	defer tick.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, flusher, ev.Name, ev.Data)
		case <-tick.C:
			writeSSE(w, flusher, sse.EventPing, []byte("{}"))
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, name string, data []byte) {
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, data)
	flusher.Flush()
}
