package middleware

import (
	"context"
	"net/http"
	"strings"

	"duekeep/internal/service"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxRole
)

// Bearer требует валидный access. Без заголовка или с битым JWT — 401.
func Bearer(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(w)
				return
			}
			id, err := service.ParseAccess(secret, raw)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(withIdentity(r.Context(), id)))
		})
	}
}

// BearerOrQuery — access из Authorization или ?access_token= (EventSource без заголовков).
func BearerOrQuery(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				raw = strings.TrimSpace(r.URL.Query().Get("access_token"))
			}
			if raw == "" {
				writeUnauthorized(w)
				return
			}
			id, err := service.ParseAccess(secret, raw)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(withIdentity(r.Context(), id)))
		})
	}
}

// OptionalBearer кладёт identity, если access валиден. Иначе не трогает запрос
// (logout может идти только по refresh).
func OptionalBearer(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			id, err := service.ParseAccess(secret, raw)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(withIdentity(r.Context(), id)))
		})
	}
}

func withIdentity(ctx context.Context, id service.AccessIdentity) context.Context {
	ctx = context.WithValue(ctx, ctxUserID, id.UserID)
	return context.WithValue(ctx, ctxRole, id.Role)
}

// UserID — sub из контекста после Bearer.
func UserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

// Role — роль из access, не из БД.
func Role(ctx context.Context) string {
	v, _ := ctx.Value(ctxRole).(string)
	return v
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	tok := strings.TrimSpace(header[len(prefix):])
	return tok, tok != ""
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":{"code":"unauthorized","message":"missing or invalid access token"}}`))
}
