package middleware

import (
	"context"
	"net/http"
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
func Auth(a Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := a.Authenticate(r.Context(), r)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
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
