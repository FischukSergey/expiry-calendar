package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"duekeep/internal/handler"
)

type fakeHealth struct {
	err error
}

func (f fakeHealth) Check(context.Context) error {
	return f.err
}

func TestHealthzOK(t *testing.T) {
	t.Parallel()
	api := handler.New(fakeHealth{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body: %+v", body)
	}
}

func TestHealthzUnavailable(t *testing.T) {
	t.Parallel()
	api := handler.New(fakeHealth{err: errors.New("down")}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d", rec.Code)
	}
}
