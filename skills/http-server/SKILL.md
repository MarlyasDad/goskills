---
name: http-server
description: "Правила HTTP-сервера на net/http — Config, server.go, корневой router, health/metrics/docs и запуск без бизнес-логики. Используется при поднятии HTTP API и изменении настроек сервера. Контроллеры — skills/http-controllers, middleware — skills/http-middlewares, сборка и останов — skills/composition-root."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: architecture
---

# HTTP-сервер

Навык отвечает на вопрос «**как поднять HTTP-сервер и собрать корневой
router**». Как писать обработчик — `skills/http-controllers`; как устроены
middleware — `skills/http-middlewares`; кто и когда запускает сервер —
`skills/composition-root`.

## Границы

- Настройки и их источники — `skills/config`.
- Где физически лежит файл — `skills/project-structure`.
- Какие импорты допустимы — `skills/architecture`.
- Тесты обработчиков и middleware — `skills/testing` и профильные навыки.
- Контракт API (`openapi.yaml`) — в `api/`, не здесь.

## Каталог

```schema
internal/transport/<appname>/http/
├── config.go        # Config сервера и Validate
├── server.go        # NewServer: собрать *http.Server
├── router.go        # корневой router: системные и прикладные роуты
├── middleware/      # Cross-cutting HTTP-обвязка
└── <group>/<resource>/
    ├── handler.go
    ├── ports.go
    ├── request.go
    └── response.go
```

`<group>` — граница доступа (`internal`, `external`) или смысловая группа
(`documents`, `appeals`). `<resource>` — ресурс или операция внутри группы.

Готовые файлы — `references/config.go`, `references/server.go`,
`references/router.go` и `references/config_test.go`. Скопировать в
`internal/transport/<appname>/http/` и заменить `<module>`/`<appname>`.
Фрагменты ниже объясняют правила; канонический код — в `references/`.

Почему сервер лежит в `internal/transport/<appname>/http/`, а не в
`internal/transport/http/`:

- `transport` уже привязан к приложению: у каждого бинаря свои роуты, порты
  и конфиг сервера;
- общий `internal/transport/http/` пришлось бы делить между приложениями,
  а это смешивает их жизненные циклы;
- если приложений несколько и у каждого есть HTTP, у каждого будет свой
  `http`-пакет. Общий код выносится в `platform` только тогда, когда он
  действительно не знает о приложении.

## Config

`Config` — локальный конфиг HTTP-сервера. Он живёт рядом с сервером, не
читает env и не импортирует `internal/config`. Общий `Config` модуля
вкладывает его полем.

```go
// internal/transport/<appname>/http/config.go
package http

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Таймауты по умолчанию. Ноль в Config означает «не задан»: NewServer
// подставляет значение отсюда, а не config.Load.
const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
)

// Config — настройки HTTP-сервера. Значения приходят из internal/config.
// Таймауты необязательны: 0 — значение по умолчанию.
type Config struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	AllowedOrigins    []string
	Metrics           bool
	Swagger           bool
}

// Addr собирает адрес через net.JoinHostPort: это корректно и для IPv6.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// withDefaults подставляет таймауты по умолчанию. Отрицательные значения
// не исправляет: это ошибка конфигурации, её ловит Validate.
func (c Config) withDefaults() Config {
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = defaultIdleTimeout
	}
	return c
}

func (c Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("http: port %d out of range", c.Port)
	}
	if c.ReadHeaderTimeout < 0 {
		return fmt.Errorf("http: read header timeout must not be negative")
	}
	if c.ReadTimeout < 0 {
		return fmt.Errorf("http: read timeout must not be negative")
	}
	if c.WriteTimeout < 0 {
		return fmt.Errorf("http: write timeout must not be negative")
	}
	if c.IdleTimeout < 0 {
		return fmt.Errorf("http: idle timeout must not be negative")
	}
	if c.MaxHeaderBytes < 0 {
		return fmt.Errorf("http: max header bytes must not be negative")
	}
	return nil
}
```

Правила:

- `Host` может быть пустым — это значит «все интерфейсы». В проде лучше
  задавать явно.
- Таймауты необязательны: `0` — «не задан», `NewServer` подставит
  `5s/15s/30s/60s` из констант пакета. Отрицательное значение — ошибка
  `Validate`, а не «ещё один дефолт». Дефолты живут в HTTP-пакете, не в
  `config.Load`: иначе они продублируются в двух местах.
- Для SSE/WebSocket `WriteTimeout` управляется отдельно: такие соединения
  не должны жить под общим `WriteTimeout` обычного JSON API.
- `AllowedOrigins` пустой — CORS выключен. `AllowedOrigins` с `*` — только
  для публичного API без credentials.
- `Metrics` и `Swagger` — флаги стенда. На проде docs обычно выключены.
- `Validate` вызывает `NewServer`, а не `Load()`: приложение, которое не
  поднимает HTTP, не обязано заполнять эти поля.

```go
// internal/config/config.go — фрагмент
type Config struct {
	ShutdownTimeout time.Duration
	HTTP            apihttp.Config
}

// Load разбирает env и заполняет HTTP:
// HTTP_HOST, HTTP_PORT, HTTP_READ_HEADER_TIMEOUT, HTTP_READ_TIMEOUT,
// HTTP_WRITE_TIMEOUT, HTTP_IDLE_TIMEOUT, HTTP_MAX_HEADER_BYTES,
// HTTP_CORS_ALLOWED_ORIGINS, HTTP_METRICS, HTTP_SWAGGER.
//
// Если таймаут не задан, передай 0: NewServer подставит дефолт.
```

Импорт `config → transport/<appname>/http` допустим только ради `Config`.
Обратного импорта нет: HTTP-пакет не знает про `internal/config`.

## server.go

`server.go` только собирает `*http.Server`. Он не вызывает
`ListenAndServe`, не биндит порт и не запускает горутины: это делает
`app.go` в `skills/composition-root`.

```go
// internal/transport/<appname>/http/server.go
package http

import (
	"errors"
	"log/slog"
	"net/http"
)

func NewServer(cfg Config, log *slog.Logger, handler http.Handler) (*http.Server, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, errors.New("http: handler is nil")
	}
	if log == nil {
		return nil, errors.New("http: logger is nil")
	}

	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}, nil
}
```

Почему так:

- `ReadHeaderTimeout` защищает от медленных заголовков (Slowloris).
- `ReadTimeout`, `WriteTimeout` и `IdleTimeout` ограничивают время запроса
  и простоя. Без них один клиент может держать соединение сколько угодно,
  поэтому нулевые значения заменяются дефолтами до `Validate`.
- `withDefaults` не трогает отрицательные значения: `-1s` — это ошибка
  конфигурации, а не повод тихо поставить дефолт.
- `net.JoinHostPort` корректно собирает адрес для IPv6. `fmt.Sprintf("%s:%d")`
  ломается на `::1`.
- `ErrorLog` переводит служебные ошибки `http.Server` в общий `slog`, а не
  в стандартный `log`.
- Конструктор возвращает ошибку: неверный `Config` — это ошибка запуска, а
  не паника.

## router.go

`router.go` — единственное место, где перечислены пути. Он получает готовые
обработчики и применяет к ним middleware. Сервисы, репозитории и адаптеры
здесь **не создаются**: это `app.go`.

```go
// internal/transport/<appname>/http/router.go
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

// registerOptionalRoutes подключает метрики и Swagger только если проект
// реально использует эти зависимости. Пример с зависимостями — ниже.
func registerOptionalRoutes(_ *http.ServeMux, _ Config) {
	// Метрики и Swagger подключаются только вместе с их зависимостями.
}
```

Если проекту нужны метрики и Swagger, добавь зависимости и замени заглушку:

```go
import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func registerOptionalRoutes(mux *http.ServeMux, cfg Config) {
	if cfg.Metrics {
		mux.Handle("GET /metrics", promhttp.Handler())
	}
	if cfg.Swagger {
		mux.Handle("GET /docs/", httpSwagger.WrapHandler)
	}
}
```

Пока зависимости не добавлены, `Metrics` и `Swagger` должны быть `false`;
заглушка не тянет лишние пакеты в `go.mod`.

Правила маршрутизации:

- `http.ServeMux` с Go 1.22+ умеет методы и path-параметры:
  `"GET /v1/documents/{id}"`, `r.PathValue("id")`. Отдельный роутер-фреймворк
  не нужен, пока не появилась реальная причина.
- Пути канонические: без лишнего `/` в конце. Не переписывай `r.URL.Path` в
  middleware, чтобы «исправить» слэш: используй явный редирект или сразу
  правильный путь.
- Группируй роуты по смыслу (`/v1/documents`, `/internal/...`,
  `/external/...`). Один `router.go` остаётся точкой входа, вспомогательные
  `register...` функции — в нём же.
- Auth, rate limit и прочие требования конкретной группы вешай на группу,
  а не на весь сервер: health, readiness и docs обычно публичные.
- `mux.Handle("/", ...)` для SPA или заглушки ставится последним и не должен
  перехватывать `/metrics`, `/docs/`, `/healthz`.

## Health, metrics, docs

- `/healthz` — liveness: процесс жив, зависимости не проверяются.
- `/readyz` — readiness: проверяет зависимости через переданную функцию
  `Ready`. Проверка идёт через порт (например, `Ping` репозитория), а не
  прямым SQL в router.
- `/metrics` — Prometheus. Если метрики не должны быть публичными, вынеси
  их на отдельный admin-порт или закрой на ingress. Один общий mux — самый
  простой вариант, но не всегда безопасный.
- `/docs/` — Swagger UI. Включается флагом `Swagger`; на проде обычно
  выключен. Сгенерированные docs не правятся руками.
- Health-ответы могут быть текстовыми; это не бизнес-API. Если проект
  требует JSON и там, используй общий `respond` из
  `skills/http-controllers`.

## Запуск и останов

`server.go` не запускает сервер. Полный жизненный цикл — в
`skills/composition-root`:

- `net.Listen` до горутины, чтобы ошибка занятого порта вернулась сразу;
- `server.Serve(ln)` в горутине;
- ожидание `ctx.Done()` или ошибки сервера;
- `server.Shutdown(ctx)` с таймаутом;
- `server.Close()` как fallback, если `Shutdown` не уложился;
- ожидание горутины `Serve` и освобождение остальных ресурсов.

Не вызывай `http.ListenAndServe` напрямую: он не даёт управления
жизненным циклом и не позволяет корректно остановиться.

## Тесты

- `Config.Validate` — table-driven: неверный порт, отрицательные таймауты,
  отрицательный `MaxHeaderBytes`. Нулевые таймауты — валидны.
- `Config.withDefaults` — нули заменяются дефолтами, явные значения
  сохраняются.
- `NewServer` — проверь `Addr`, применение дефолтов и ошибку на `nil`
  handler.
- `NewRouter` — `httptest`: метод, путь, path-параметр, 404/405, публичный
  health и защищённый роут.
- Реальный порт и `Serve`/`Shutdown` — интеграционный тест, не unit.
- Подробно — `skills/testing`.

## Антипаттерны

- `http.ListenAndServe` в `main`, `NewServer` или `init`.
- Сервер без `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` и
  `IdleTimeout`; нули, которые нигде не заменяются дефолтами.
- `fmt.Sprintf("%s:%d", host, port)` вместо `net.JoinHostPort`.
- `http.DefaultServeMux` и глобальные обработчики: их не видно в router.
- Создание сервисов, репозиториев и адаптеров в `router.go`.
- Бизнес-логика в `router.go` или health-обработчике.
- Swagger и метрики всегда включены, без флага стенда.
- Переписывание `r.URL.Path`/`r.RequestURI` в middleware ради слэша.
- Роутер-фреймворк «на всякий случай», когда хватает `http.ServeMux`.
- Один общий mux, где `/metrics` и `/docs/` открыты наружу без решения
  по безопасности.

## Чек-лист

- [ ] `Config` лежит рядом с сервером и не читает env.
- [ ] `Validate` вызывается в `NewServer`, а не в `Load`.
- [ ] Таймауты необязательны: `0` заменяется дефолтом, отрицательные
      значения отклоняются.
- [ ] `Addr` собирается через `net.JoinHostPort`.
- [ ] `server.go` не запускает сервер и не биндит порт.
- [ ] `router.go` получает готовые обработчики, а не создаёт их.
- [ ] Публичные и защищённые группы разделены.
- [ ] Health/readiness, metrics и docs подключены осознанно.
- [ ] Пути канонические, без переписывания URL в middleware.
- [ ] Есть тесты на `Config`, `NewServer` и маршруты.

## Эталон

- `references/config.go` — `Config`, `Addr`, `withDefaults`, `Validate`.
- `references/server.go` — `NewServer` без запуска и биндинга порта.
- `references/router.go` — корневой router, health, auth на группе,
  заглушка для metrics/Swagger.
- `references/config_test.go` — тесты `Addr`, `withDefaults`, `Validate` и
  применения дефолтов в `NewServer`.

## Границы

- «Как написать обработчик, DTO и маппинг ошибок» — `skills/http-controllers`.
- «Как устроены middleware и их порядок» — `skills/http-middlewares`.
- «Кто собирает сервер, запускает и останавливает» —
  `skills/composition-root`.
- «Как читать env и собирать `Config` модуля» — `skills/config`.
- «Где лежит каталог» — `skills/project-structure`.
