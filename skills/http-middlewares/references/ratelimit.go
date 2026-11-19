package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"<module>/internal/transport/<appname>/http/responses"
)

// Limiter — порт ограничителя. Реализация (in-memory, Redis) живёт в infra.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// RateLimit проверяет лимит по ключу из запроса.
func RateLimit(log *slog.Logger, limiter Limiter, key func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := limiter.Allow(r.Context(), key(r))
			if err != nil {
				if writeErr := responses.RespondWithError(
					w,
					http.StatusServiceUnavailable,
					responses.CodeUnassigned,
					"rate limiter unavailable",
				); writeErr != nil {
					log.ErrorContext(r.Context(), "write rate limit response", "error", writeErr)
				}
				return
			}
			if !allowed {
				w.Header().Set("Retry-After", "60")
				if writeErr := responses.RespondWithError(
					w,
					http.StatusTooManyRequests,
					responses.CodeUnassigned,
					"too many requests",
				); writeErr != nil {
					log.ErrorContext(r.Context(), "write rate limit response", "error", writeErr)
				}
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
