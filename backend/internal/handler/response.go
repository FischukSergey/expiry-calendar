package handler

import (
	"encoding/json"
	"net/http"

	"duekeep/internal/model"
)

func writeHealthOK(w http.ResponseWriter) {
	writeBytes(w, http.StatusOK, model.HealthOK{Status: "ok"})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeBytes(w, status, model.ErrorResponse{
		Error: model.APIError{
			Code:    code,
			Message: message,
		},
	})
}

func writeBytes(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"error":{"code":"internal","message":"encode failed"}}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
