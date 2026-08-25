package middleware

import (
	"net/http"

	"duekeep/internal/model"
)

// RequireAdmin — после Bearer. Viewer получает 403, не 401.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != string(model.RoleAdmin) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":"forbidden","message":"admin only"}}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
