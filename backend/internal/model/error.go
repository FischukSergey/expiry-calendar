package model

// APIError — поля ошибки из api-sprint-1.md.
type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorResponse — JSON-конверт ошибки.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// HealthOK — успешный ответ GET /healthz.
type HealthOK struct {
	Status string `json:"status"`
}
