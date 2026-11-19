package http

import (
	"context"
	"log/slog"
	"net/http"

	"<module>/internal/transport/<appname>/http/documents"
	"<module>/internal/transport/<appname>/http/middleware"
)

// HandlerSet — обработчики, которые собрал composition root.
type HandlerSet struct {
	Documents *documents.Handler
}

// Deps — зависимости router. Здесь только то, что нужно для маршрутизации.
type Deps struct {
	Config   Config
	Log      *slog.Logger
	Auth     middleware.Authenticator
	Ready    func(context.Context) error
	Handlers HandlerSet
}

// NewRouter собирает корневой mux и применяет общую цепочку middleware.
// Сервисы, репозитории и адаптеры здесь не создаются.
func NewRouter(d Deps) http.Handler {
	if d.Log == nil {
		d.Log = slog.Default()
	}

	mux := http.NewServeMux()

	registerSystemRoutes(mux, d.Ready)
	registerDocumentRoutes(mux, d)
	registerOptionalRoutes(mux, d.Config)

	return middleware.Chain(
		mux,
		middleware.RequestID(),
		middleware.Logging(d.Log),
		middleware.CORS(middleware.CORSConfig{AllowedOrigins: d.Config.AllowedOrigins}),
		middleware.Recovery(d.Log),
	)
}

func registerSystemRoutes(mux *http.ServeMux, ready func(context.Context) error) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready != nil {
			if err := ready(r.Context()); err != nil {
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func registerDocumentRoutes(mux *http.ServeMux, d Deps) {
	if d.Handlers.Documents == nil {
		return
	}

	create := http.Handler(http.HandlerFunc(d.Handlers.Documents.Create))
	if d.Auth != nil {
		create = middleware.Chain(create, middleware.Auth(d.Auth))
	}

	mux.Handle("POST /v1/documents", create)
}

// registerOptionalRoutes — заглушка без внешних зависимостей.
//
// Если проект использует Prometheus и Swagger, добавь зависимости и замени
// тело функции:
//
//	import (
//		"github.com/prometheus/client_golang/prometheus/promhttp"
//		httpSwagger "github.com/swaggo/http-swagger"
//	)
//
//	func registerOptionalRoutes(mux *http.ServeMux, cfg Config) {
//		if cfg.Metrics {
//			mux.Handle("GET /metrics", promhttp.Handler())
//		}
//		if cfg.Swagger {
//			mux.Handle("GET /docs/", httpSwagger.WrapHandler)
//		}
//	}
func registerOptionalRoutes(_ *http.ServeMux, _ Config) {
	// Метрики и Swagger подключаются только вместе с их зависимостями.
}
