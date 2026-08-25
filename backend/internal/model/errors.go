package model

import "errors"

// Доменные ошибки. Хендлер мапит их на HTTP-коды контракта, SQL и net/http сюда не входят.
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation_error")
	ErrInternal     = errors.New("internal")
)

// ValidationError — 422 с details по полям.
type ValidationError struct {
	Msg     string
	Fields  map[string]any
	wrapped error
}

// Validation собирает 422. fields может быть nil.
func Validation(msg string, fields map[string]any) error {
	return &ValidationError{Msg: msg, Fields: fields, wrapped: ErrValidation}
}

func (e *ValidationError) Error() string { return e.Msg }

func (e *ValidationError) Unwrap() error { return e.wrapped }
