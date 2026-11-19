package http

import (
	"log/slog"
	"net/http"

	"<module>/internal/transport/<appname>/http/middleware"
)

// HandlerSet — бизнес-обработчики приложения. В каркасе пустой: есть только
// системные ручки. Поле и импорт контроллера добавляются вместе с первым
// бизнес-роутом.
type HandlerSet struct{}

// Deps — зависимости router.
type Deps struct {
	Config   Config
	Log      *slog.Logger
	Auth     middleware.Authenticator
	Handlers HandlerSet
}

// NewRouter собирает корневой mux и применяет общую цепочку middleware.
// Сервисы, репозитории и адаптеры здесь не создаются.
func NewRouter(d Deps) http.Handler {
	if d.Log == nil {
		d.Log = slog.Default()
	}

	mux := http.NewServeMux()

	registerHealthRoute(mux)
	registerOptionalRoutes(mux, d.Config)
	// registerDocumentRoutes(mux, d) — добавить вместе с первым контроллером.

	return middleware.Chain(
		mux,
		middleware.RequestID(),
		middleware.Logging(d.Log),
		middleware.CORS(middleware.CORSConfig{AllowedOrigins: d.Config.AllowedOrigins}),
		middleware.Recovery(d.Log),
	)
}

// registerHealthRoute — единственная обязательная системная ручка.
// Когда появится зависимость, добавь сюда проверку и отдельный роут, если
// нужен readiness.
func registerHealthRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// registerOptionalRoutes — заглушка без внешних зависимостей.
//
// Если проект использует Prometheus и Swagger, добавь зависимости:
//
//	go get github.com/prometheus/client_golang/prometheus/promhttp
//	go get github.com/swaggo/http-swagger
//
// и замени тело функции:
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
