package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Limiter — порт ограничителя. Реализация (in-memory, Redis) живёт в infra.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// RateLimit проверяет лимит по ключу из запроса.
func RateLimit(limiter Limiter, key func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := limiter.Allow(r.Context(), key(r))
			if err != nil {
				http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
				return
			}
			if !allowed {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP возвращает IP клиента. X-Forwarded-For читается только при
// доверенном прокси, который перезаписывает заголовок.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			first, _, _ := strings.Cut(xff, ",")
			return strings.TrimSpace(first)
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
