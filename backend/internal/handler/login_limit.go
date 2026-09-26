package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	loginFailLimit  = 8
	loginFailWindow = 15 * time.Minute
)

// loginLimit считает неудачные POST /auth/login на пару IP+email. На процесс, без Redis.
type loginLimit struct {
	mu    sync.Mutex
	fails map[string][]time.Time
}

func newLoginLimit() *loginLimit {
	return &loginLimit{fails: map[string][]time.Time{}}
}

func (l *loginLimit) blocked(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fails[key] = recentFails(l.fails[key], now)
	return len(l.fails[key]) >= loginFailLimit
}

func (l *loginLimit) fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fails[key] = append(recentFails(l.fails[key], now), now)
}

func (l *loginLimit) success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}

func recentFails(in []time.Time, now time.Time) []time.Time {
	cut := now.Add(-loginFailWindow)
	out := in[:0]
	for _, ts := range in {
		if ts.After(cut) {
			out = append(out, ts)
		}
	}
	return out
}

func loginKey(r *http.Request, email string) string {
	return clientIP(r) + "\n" + strings.ToLower(strings.TrimSpace(email))
}

func clientIP(r *http.Request) string {
	if fwd := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); fwd != "" {
		host, _, _ := strings.Cut(fwd, ",")
		return strings.TrimSpace(host)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
