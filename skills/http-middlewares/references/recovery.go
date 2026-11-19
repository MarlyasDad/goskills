package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"<module>/internal/transport/<appname>/http/responses"
)

// Recovery ловит панику, логирует стек и отдаёт 500.
func Recovery(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				// ErrAbortHandler — специальный сигнал net/http: соединение
				// нужно оборвать, а не отдавать 500.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				log.ErrorContext(r.Context(), "panic recovered",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
					slog.String("request_id", RequestIDFromContext(r.Context())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)

				// Если заголовки уже ушли, 500 записать нельзя. Лучше
				// оборвать соединение, чем дописать битый JSON.
				if sw, ok := w.(*statusWriter); ok && sw.status != 0 {
					panic(http.ErrAbortHandler)
				}

				if err := responses.RespondWithError(
					w,
					http.StatusInternalServerError,
					responses.CodeUnassigned,
					"internal server error",
				); err != nil {
					log.ErrorContext(r.Context(), "write recovery response", "error", err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
