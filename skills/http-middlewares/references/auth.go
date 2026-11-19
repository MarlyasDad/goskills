package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"<module>/internal/transport/<appname>/http/responses"
)

// Principal — аутентифицированный субъект.
type Principal struct {
	Subject string
	Roles   []string
}

// Authenticator проверяет аутентификацию. Реализация живёт в infra,
// middleware не знает провайдера.
type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}

type principalKey struct{}

// Auth кладёт Principal в контекст или отвечает 401.
func Auth(log *slog.Logger, a Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := a.Authenticate(r.Context(), r)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				if writeErr := responses.RespondWithError(
					w,
					http.StatusUnauthorized,
					responses.CodeUnassigned,
					"unauthorized",
				); writeErr != nil {
					log.ErrorContext(r.Context(), "write auth response", "error", writeErr)
				}
				return
			}

			ctx := context.WithValue(r.Context(), principalKey{}, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
