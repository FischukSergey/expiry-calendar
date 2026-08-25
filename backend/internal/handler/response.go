package handler

import (
	"encoding/json"
	"net/http"

	"duekeep/internal/model"
)

// writeHealthOK — 200 {"status":"ok"} по api-sprint-1.
func writeHealthOK(w http.ResponseWriter) {
	writeBytes(w, http.StatusOK, model.HealthOK{Status: "ok"})
}

// writeError отдаёт конверт {"error":{"code","message"}} без details.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorDetails(w, status, code, message, nil)
}

// writeErrorDetails добавляет details (валидация полей).
func writeErrorDetails(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	writeBytes(w, status, model.ErrorResponse{
		Error: model.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// writeBytes кодирует JSON. Сбой marshal — сырой 500, без рекурсии в writeError.
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
