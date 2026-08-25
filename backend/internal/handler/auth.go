package handler

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"duekeep/internal/middleware"
	"duekeep/internal/model"
)

type emailPasswordBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var body emailPasswordBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	pair, err := a.auth.Register(r.Context(), body.Email, body.Password, r.UserAgent())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	a.setRefreshCookie(w, pair.RefreshToken)
	writeBytes(w, http.StatusCreated, pair)
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var body emailPasswordBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	pair, err := a.auth.Login(r.Context(), body.Email, body.Password, r.UserAgent())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	a.setRefreshCookie(w, pair.RefreshToken)
	writeBytes(w, http.StatusOK, pair)
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	raw, err := a.refreshFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	pair, err := a.auth.Refresh(r.Context(), raw, r.UserAgent())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	a.setRefreshCookie(w, pair.RefreshToken)
	writeBytes(w, http.StatusOK, pair)
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	raw, err := a.refreshFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	userID := middleware.UserID(r.Context())
	if raw == "" && userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "access or refresh required")
		return
	}
	if err := a.auth.Logout(r.Context(), raw); err != nil {
		writeDomainError(w, err)
		return
	}
	a.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) logoutAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r.Context())
	if err := a.auth.LogoutAll(r.Context(), userID); err != nil {
		writeDomainError(w, err)
		return
	}
	a.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	user, err := a.auth.Me(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, user)
}

// refreshFromRequest: непустой body.refresh_token важнее cookie duekeep_refresh.
func (a *API) refreshFromRequest(r *http.Request) (string, error) {
	var body refreshBody
	if err := decodeJSON(r, &body); err != nil {
		return "", err
	}
	cookie, _ := r.Cookie(model.RefreshCookie)
	cookieVal := ""
	if cookie != nil {
		cookieVal = cookie.Value
	}
	return cmp.Or(strings.TrimSpace(body.RefreshToken), strings.TrimSpace(cookieVal)), nil
}

func (a *API) setRefreshCookie(w http.ResponseWriter, raw string) {
	c := refreshCookie(raw, int(a.refreshTTL.Seconds()), a.cookieSecure)
	http.SetCookie(w, c)
}

func (a *API) clearRefreshCookie(w http.ResponseWriter) {
	c := refreshCookie("", -1, a.cookieSecure)
	http.SetCookie(w, c)
}

// refreshCookie собирает duekeep_refresh. Secure=false на localhost, иначе браузер не сохранит.
func refreshCookie(value string, maxAge int, secure bool) *http.Cookie {
	return &http.Cookie{ //nolint:gosec // G124: Secure выключаем на HTTP localhost.
		Name:     model.RefreshCookie,
		Value:    value,
		Path:     model.RefreshCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dst)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func writeDomainError(w http.ResponseWriter, err error) {
	var val *model.ValidationError
	switch {
	case errors.As(err, &val):
		writeErrorDetails(w, http.StatusUnprocessableEntity, "validation_error", val.Msg, val.Fields)
	case errors.Is(err, model.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
	case errors.Is(err, model.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, model.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, model.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal")
	}
}
