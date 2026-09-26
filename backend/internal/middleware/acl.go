package middleware

import (
	"net/http"

	"duekeep/internal/model"
)

// RequireAdmin — после Bearer. Свои записи пишут admin и administrator. Viewer получает 403, не 401.
func RequireAdmin(next http.Handler) http.Handler {
	return requireRole(next, `{"error":{"code":"forbidden","message":"admin only"}}`, canMutateOwn)
}

// RequireAdministrator — мутации общего справочника item_kinds.
func RequireAdministrator(next http.Handler) http.Handler {
	return requireRole(next, `{"error":{"code":"forbidden","message":"administrator only"}}`, func(role string) bool {
		return role == string(model.RoleAdministrator)
	})
}

func canMutateOwn(role string) bool {
	return role == string(model.RoleAdmin) || role == string(model.RoleAdministrator)
}

func requireRole(next http.Handler, forbidden string, allow func(string) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allow(Role(r.Context())) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(forbidden))
			return
		}
		next.ServeHTTP(w, r)
	})
}
