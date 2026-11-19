---
name: http-middlewares
description: "Правила HTTP-middleware на net/http: Chain, порядок применения, RequestID, Logging, Recovery, CORS, Auth, RateLimit и аккуратная обёртка ResponseWriter. Используется при добавлении сквозной логики HTTP. Сервер — skills/http-server, контроллеры — skills/http-controllers, сборка — skills/composition-root."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: architecture
---

# HTTP-middleware

Навык отвечает на вопрос «**как написать и в каком порядке применить
HTTP-middleware**». Сервер, роутинг и health — `skills/http-server`;
обработчики и DTO — `skills/http-controllers`.

## Границы

- Зависимости и слои — `skills/architecture`.
- Сервисы и операции — `skills/services`.
- Тесты — `skills/testing`.
- Конкретные настройки CORS/rate limit — в конфиге проекта, а не в навыке.

## Каталог

```schema
internal/transport/<appname>/http/middleware/
├── chain.go
├── requestid.go
├── logging.go
├── recovery.go
├── cors.go
├── auth.go
├── ratelimit.go
└── middleware_test.go
```

- Пакет называется `middleware` (единственное число), а не `middlewares`.
- Один файл — одна ответственность. Не складывать auth, rate limit и
  logging в один `middleware.go`.
- Middleware не знает про конкретный сервис, БД или identity-провайдер.
  Зависимости приходят интерфейсами и собираются в `app.go`.
- Middleware не содержит бизнес-логики и не ходит в БД напрямую.

Готовые файлы — `references/chain.go`, `references/requestid.go`,
`references/logging.go`, `references/recovery.go`, `references/cors.go`,
`references/auth.go`, `references/ratelimit.go` и
`references/middleware_test.go`. Скопировать в
`internal/transport/<appname>/http/middleware/` и заменить
`<module>`/`<appname>`. Фрагменты ниже объясняют правила; канонический код —
в `references/`.

## Chain

```go
// internal/transport/<appname>/http/middleware/chain.go
package middleware

import "net/http"

// Middleware — функция, которая оборачивает следующий обработчик.
type Middleware func(http.Handler) http.Handler

// Chain применяет middleware так, что первый в списке — самый внешний.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
```

`Chain` принимает `http.Handler`, а не `http.HandlerFunc`: `http.HandlerFunc`
реализует `http.Handler`, поэтому подходит и обычный обработчик, и метод
контроллера, обёрнутый в `http.HandlerFunc`.

## Порядок

Порядок — часть контракта. Если поменять местами Recovery и Logging,
логи перестанут видеть 500 от паники; если Auth поставить раньше CORS,
preflight начнёт получать 401.

Корневая цепочка (все роуты):

```text
RequestID -> Logging -> CORS -> Recovery -> mux
```

Цепочка защищённой группы (конкретные роуты):

```text
RateLimit -> Auth -> handler
```

Почему так:

1. **RequestID** — самый внешний: ID должен быть в контексте и в заголовке
   ответа даже для ошибок Recovery и CORS.
2. **Logging** — снаружи Recovery: он видит финальный статус 500, который
   выставил Recovery.
3. **CORS** — снаружи Recovery и Auth: заголовки CORS остаются на
   preflight и на ответах 401/500; preflight не должен требовать токен.
4. **Recovery** — внутри Logging, но снаружи прикладного обработчика:
   ловит панику и превращает её в 500.
5. **RateLimit** — на группе, до Auth, если считаем по IP; после Auth,
   если считаем по пользователю.
6. **Auth** — на группе: health, readiness и docs обычно публичные.

Не делай одну глобальную цепочку с Auth: тогда публичные роуты придётся
исключать по пути, а это хрупко.

## RequestID

```go
// internal/transport/<appname>/http/middleware/requestid.go
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey int

const requestIDKey contextKey = iota

const requestIDHeader = "X-Request-ID"

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestIDHeader)
			if !validRequestID(id) {
				id = newRequestID()
			}

			w.Header().Set(requestIDHeader, id)
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func validRequestID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}
```

- Не доверяй произвольному `X-Request-ID`: ограничь длину и символы, иначе
  получишь инъекцию в лог.
- Ключ контекста — собственный тип, не строка.
- Заголовок ответа выставляется до `next`, чтобы он был и на ошибке.

## Logging

```go
// internal/transport/<appname>/http/middleware/logging.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

// Unwrap нужен http.NewResponseController: без него не работают Flush,
// Hijack и SetReadDeadline через обёртку.
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func Logging(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r)

			if sw.status == 0 {
				sw.status = http.StatusOK
			}

			level := slog.LevelInfo
			switch {
			case sw.status >= http.StatusInternalServerError:
				level = slog.LevelError
			case sw.status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}

			log.LogAttrs(r.Context(), level, "http request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int("bytes", sw.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}
```

Правила:

- Не логируй `r.URL.RawQuery`: там бывают токены и персональные данные.
- Не логируй тело запроса и `Authorization`.
- `statusWriter` должен игнорировать повторный `WriteHeader`: иначе
  получишь `superfluous response.WriteHeader call`.
- `Unwrap` обязателен для `http.NewResponseController` (Flush, Hijack,
  deadlines). Альтернатива — реализовать `Flusher`/`Hijacker` вручную, но
  `Unwrap` короче и не теряет остальные методы.
- Access log — одна строка на запрос. Бизнес-события логирует сервис, а не
  middleware.

## Recovery

```go
// internal/transport/<appname>/http/middleware/recovery.go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"<module>/internal/transport/<appname>/http/responses"
)

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
					http.StatusText(http.StatusInternalServerError),
				); err != nil {
					log.ErrorContext(r.Context(), "write recovery response", "error", err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
```

- Стек — только в лог, не в ответ.
- Recovery не глотает `http.ErrAbortHandler`.
- Recovery должен стоять внутри Logging, иначе Logging не увидит 500.
- Паника — не штатный способ вернуть 400. Ожидаемые ошибки идут через
  `error`.

## CORS

```go
// internal/transport/<appname>/http/middleware/cors.go
package middleware

import "net/http"

type CORSConfig struct {
	AllowedOrigins []string
}

func CORS(cfg CORSConfig) Middleware {
	if len(cfg.AllowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	allowAll := false
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			allowAll = true
		}
		allowed[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Add("Vary", "Origin")
				if allowAll || contains(allowed, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func contains(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}
```

- `AllowedOrigins` пустой — CORS выключен.
- `*` — только для публичного API без credentials. С credentials `*`
  запрещён спецификацией.
- `Vary: Origin` обязателен, иначе кеши отдадут ответ чужому origin.
- Preflight (`OPTIONS`) обрабатывается до Auth: иначе браузер получит 401.
- Не отражай `Origin` без проверки: это превращает CORS в «разрешено всем».

## Auth

Authn (кто ты) и authz (что можно) — разные вещи. Middleware проверяет
аутентификацию и кладёт principal в контекст; проверка ролей — в
обработчике или отдельном middleware.

```go
// internal/transport/<appname>/http/middleware/auth.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"<module>/internal/transport/<appname>/http/responses"
)

type Principal struct {
	Subject string
	Roles   []string
}

type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}

type principalKey struct{}

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
```

- Реализация `Authenticator` живёт в `infra` (OIDC/JWT/внешний сервис) и
  передаётся в `app.go`. Middleware не импортирует `infra` и не знает
  провайдера.
- В контекст кладётся principal, а не сырой токен.
- 401 — не аутентифицирован; 403 — аутентифицирован, но нет прав. Не
  смешивай.
- Не логируй токен и заголовок `Authorization`.
- Auth не должен ходить в БД на каждый запрос, если это можно заменить
  проверкой подписи/кешем; решение фиксируется в проекте.
- Auth, RateLimit и Recovery отвечают через `responses.RespondWithError`,
  чтобы формат ошибок совпадал с контроллерами. Код ошибки — из
  корпоративного справочника; до присвоения — `responses.CodeUnassigned`.
- Конструкторы `responses` возвращают ошибку записи: логируй её, не
  проглатывай через `_ =`.
- Если middleware отвечает ошибкой, логируй тот же `code`, что ушёл в ответ:
  лог-коллектор сортирует по нему, HTTP-статуса недостаточно.
  Правила — `skills/http-controllers`.

## RateLimit

```go
// internal/transport/<appname>/http/middleware/ratelimit.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"<module>/internal/transport/<appname>/http/responses"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

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
```

- Лимитер — порт; реализация (in-memory, Redis) живёт в `infra`.
- Ключ: IP до Auth или `Principal.Subject` после Auth. Выбор зависит от
  того, защищаемся мы от нагрузки или от конкретного пользователя.
- `Retry-After` — в секундах или HTTP-дате.
- Не доверяй `X-Forwarded-For`/`X-Real-IP` без настройки trusted proxy:
  клиент подделает заголовок. Если приложение за прокси, прокси должен
  перезаписывать заголовок.
- IP из `r.RemoteAddr` извлекай через `net.SplitHostPort`, а не
  `LastIndex(":")`: IPv6 содержит двоеточия.

```go
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
```

## Timeout и размер тела

- Общие таймауты сервера (`ReadTimeout`, `WriteTimeout`,
  `ReadHeaderTimeout`, `IdleTimeout`) задаются в `Config` сервера, а не
  middleware.
- `http.TimeoutHandler` умеет вернуть 503 по таймауту, но он плохо
  сочетается со streaming/SSE/WebSocket и уже отправленными заголовками.
  Для обычного JSON API таймаут запроса лучше держать на уровне сервера и
  `context` сервиса.
- Лимит тела — `http.MaxBytesReader` в обработчике, а не глобальный
  middleware: у загрузки файлов и JSON разные лимиты.
- Не ставь `time.Sleep` в middleware «для нагрузки» и не буферизуй ответ
  целиком ради логов.

## Trailing slash

Не переписывай `r.URL.Path` и `r.RequestURI` в middleware. Это ломает
роутинг, подписанные URL и логи, а `strings.Contains` для исключений
срабатывает на `/docs/evil`.

Если редирект действительно нужен:

```go
func RedirectTrailingSlash() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
				target := strings.TrimSuffix(r.URL.Path, "/")
				if r.URL.RawQuery != "" {
					target += "?" + r.URL.RawQuery
				}
				http.Redirect(w, r, target, http.StatusPermanentRedirect)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

Лучше — сразу задать канонические пути в `router.go` и не добавлять этот
middleware.

## Тесты

```go
func TestRecovery(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))

	h := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		}),
		RequestID(),
		Logging(log),
		Recovery(log),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, buf.String(), "panic recovered")
}
```

Проверяй:

- порядок: RequestID виден в логе, Logging видит 500 после Recovery;
- `statusWriter` не ломает `Flush`/`Hijack` (через
  `http.NewResponseController`);
- CORS preflight без токена и с разрешённым/запрещённым origin;
- Auth: 401 без principal, principal в контексте с валидным;
- RateLimit: 429 и `Retry-After`;
- RequestID: входящий валидный сохраняется, невалидный заменяется;
- Recovery не логирует стек в ответ.

## Антипаттерны

- Auth-middleware импортирует `infra`/провайдера identity и лезет в БД.
- Один middleware делает auth, rate limit и logging.
- CORS `*` вместе с credentials.
- Доверие `X-Forwarded-For` без trusted proxy.
- Переписывание `r.URL.Path`/`r.RequestURI` ради слэша.
- `strings.Contains` для исключений путей.
- Обёртка `ResponseWriter` без `Unwrap`, ломающая `Flush`/`Hijack`.
- Логирование тела, query, `Authorization`, cookie.
- Recovery, который глотает `http.ErrAbortHandler`.
- Auth на всём mux с исключениями по пути.
- 499 или самодельные HTTP-коды без решения проекта.
- `time.Sleep` и буферизация ответа в middleware.

## Чек-лист

- [ ] `Chain` применяет первый middleware как самый внешний.
- [ ] Порядок root-цепочки: RequestID → Logging → CORS → Recovery.
- [ ] Auth и rate limit — на группе, не на всём mux.
- [ ] Auth зависит от интерфейса, а не от `infra`.
- [ ] CORS учитывает `Vary: Origin` и preflight.
- [ ] `statusWriter` имеет `Unwrap`.
- [ ] Recovery не отдаёт стек наружу и не глотает `ErrAbortHandler`.
- [ ] В логи не попадают секреты и тело запроса.
- [ ] IP берётся через `net.SplitHostPort`, proxy-заголовки — только при
      доверенном прокси.
- [ ] Есть тесты на порядок, 401, 429 и панику.

## Эталон

- `references/chain.go` — `Middleware` и `Chain`.
- `references/requestid.go` — `X-Request-ID` и контекст.
- `references/logging.go` — `statusWriter`, `Unwrap`, access log.
- `references/recovery.go` — паника, стек, `ErrAbortHandler`.
- `references/cors.go` — origin, preflight, `Vary`.
- `references/auth.go` — `Authenticator`, `Principal`, 401.
- `references/ratelimit.go` — `Limiter`, 429, `ClientIP`.
- `references/middleware_test.go` — тесты порядка, 401, 429, паники.

## Границы

- «HTTP-сервер и подключение цепочки» — `skills/http-server`.
- «Обработчик, DTO и маппинг ошибок» — `skills/http-controllers`.
- «Сборка зависимостей middleware в `app.go`» — `skills/composition-root`.
- «Границы импортов» — `skills/architecture`.
