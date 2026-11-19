---
name: http-controllers
description: "Правила HTTP-контроллеров на net/http: handler.go, ports.go, request.go, response.go, тонкий обработчик, DTO, коды и маппинг ошибок, тесты на httptest. Используется при добавлении HTTP-ручки. Общие правила адаптера — skills/transport, сервер — skills/http-server, middleware — skills/http-middlewares."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: architecture
---

# HTTP-контроллеры

Навык отвечает на вопрос «**как написать один HTTP-обработчик**». Общие
правила входящих адаптеров (роль, интерфейс операции, границы DTO и ошибок)
— в `skills/transport`. HTTP-сервер и роутинг — `skills/http-server`.
Middleware — `skills/http-middlewares`.

## Границы

- Бизнес-правила и оркестрация — `skills/services`, `internal/domain`.
- Интерфейсы репозиториев и адаптеров — `skills/architecture`.
- Контракт API (`openapi.yaml`) — в `api/`.
- Тесты — `skills/testing`.

## Каталог

```schema
internal/transport/<appname>/http/<group>/<resource>/
├── handler.go       # Handler, New и HTTP-методы
├── ports.go         # интерфейсы операций, которые нужны контроллеру
├── request.go       # Request DTO и toInput
├── response.go      # Response DTO и fromOutput
├── errors.go        # маппинг ошибок, если он не помещается в handler.go
└── handler_test.go
```

- `<group>` — `internal`, `external` или смысловая группа. `<resource>` —
  ресурс: `documents`, `appeals`.
- Пакет называется по последнему сегменту: `documents`, `appeals`. Если в
  router встречаются два `user` из разных групп, импортируй с алиасом:
  `internaluser`, `externaluser`.
- Один каталог — один ресурс или небольшая группа родственных операций.
  Не складывать все ручки приложения в один `handler.go`.
- Общие помощники ответа (`JSON`, `Error`) — `http/respond/`. Если
  контроллер один, можно оставить их в `handler.go`; при втором
  контроллере выноси в `respond`, чтобы формат ошибок не разъехался.

Готовые файлы — `references/ports.go`, `references/request.go`,
`references/response.go`, `references/handler.go`, `references/respond.go` и
`references/handler_test.go`. Скопировать в
`internal/transport/<appname>/http/documents/` (кроме `respond.go` — он
идёт в `http/respond/`) и заменить `<module>`/`<appname>`. Фрагменты ниже
объясняют правила; канонический код — в `references/`.

## ports.go

Интерфейс объявляет потребитель — контроллер. Метод принимает и возвращает
типы из `shared` или `domain`, а не `Request`/`Response` транспорта.

```go
// internal/transport/<appname>/http/documents/ports.go
package documents

import (
	"context"

	"<module>/internal/shared"
)

// Creator — операция, которую контроллер вызывает у сервиса.
type Creator interface {
	Create(ctx context.Context, input shared.CreateDocumentInput) (shared.CreateDocumentOutput, error)
}
```

Правила:

- 1–3 метода на интерфейс. Больше — раздели по операциям.
- Не импортировать `services`, `infra`, `app`, `config`.
- Если `Input`/`Output` нужны и сервису, и транспорту, они лежат в `shared`;
  подробности — `skills/services` и `skills/transport`.
- Интерфейс можно назвать по операции (`Creator`, `Getter`) или по
  зависимости (`DocumentService`). Главное — он описывает только то, что
  реально вызывает этот контроллер.

## request.go

`Request` — DTO транспорта. Здесь JSON-теги, разбор path/query и
форматная валидация. Бизнес-валидация — в сервисе или домене.

```go
// internal/transport/<appname>/http/documents/request.go
package documents

import (
	"fmt"

	"<module>/internal/shared"
)

type CreateRequest struct {
	Title string `json:"title" example:"Акт осмотра"`
	Kind  string `json:"kind" example:"inspection"`
}

func (r CreateRequest) toInput() (shared.CreateDocumentInput, error) {
	if r.Title == "" {
		return shared.CreateDocumentInput{}, fmt.Errorf("title is required")
	}
	if r.Kind == "" {
		return shared.CreateDocumentInput{}, fmt.Errorf("kind is required")
	}
	return shared.CreateDocumentInput{
		Title: r.Title,
		Kind:  r.Kind,
	}, nil
}
```

Правила:

- JSON-теги — только здесь. Доменные сущности не сериализуются напрямую.
- Path-параметр: `r.PathValue("id")`; проверь, что он не пустой, и разбери
  тип (`strconv`, `time.Parse`).

  ```go
  id := r.PathValue("id")
  if id == "" {
      h.writeError(w, r, fmt.Errorf("%w: id is required", domain.ErrInvalidInput))
      return
  }
  ```

- Query-параметры: `r.URL.Query().Get(...)`; отсутствие и пустая строка —
  разные случаи, если это важно контракту.
- `json:"-"` — для полей, которые заполняет сервер (ID пользователя из
  контекста, вычисленные права).
- Не добавлять в DTO методы с бизнес-логикой; только маппинг.

## response.go

`Response` — DTO транспорта. Никаких доменных сущностей и внутренних полей.

```go
// internal/transport/<appname>/http/documents/response.go
package documents

import "<module>/internal/shared"

type CreateResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func fromCreateOutput(out shared.CreateDocumentOutput) CreateResponse {
	return CreateResponse{
		ID:     out.ID,
		Status: out.Status,
	}
}
```

Правила:

- Имена JSON — по контракту API, а не по именам полей Go.
- Время — RFC3339 в строке; не отдавать `time.Time` с неявным форматом.
- Опциональные поля — `omitempty` или указатель, если нужно отличить
  «нет значения» от нуля.
- Не отдавать `error`, стек, SQL, имена таблиц и внутренние ID, если они не
  часть контракта.

## handler.go

Обработчик делает ровно четыре шага: разобрать запрос, провалидировать
формат, вызвать операцию, собрать ответ. Бизнес-логики и SQL здесь нет.

```go
// internal/transport/<appname>/http/documents/handler.go
package documents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"<module>/internal/domain"
	"<module>/internal/transport/<appname>/http/respond"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var errBodyTooLarge = errors.New("request body too large")

type Handler struct {
	log     *slog.Logger
	creator Creator
}

func New(log *slog.Logger, creator Creator) *Handler {
	return &Handler{log: log, creator: creator}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		if errors.Is(err, errBodyTooLarge) {
			h.writeError(w, r, err)
			return
		}
		h.writeError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	input, err := req.toInput()
	if err != nil {
		h.writeError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	out, err := h.creator.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := respond.JSON(w, http.StatusCreated, fromCreateOutput(out)); err != nil {
		h.log.ErrorContext(r.Context(), "write response", "error", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return errBodyTooLarge
		}
		return err
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code := http.StatusInternalServerError, "internal_error"

	switch {
	case errors.Is(err, errBodyTooLarge):
		status, code = http.StatusRequestEntityTooLarge, "payload_too_large"
	case errors.Is(err, domain.ErrInvalidInput):
		status, code = http.StatusBadRequest, "invalid_request"
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, context.Canceled):
		h.log.WarnContext(r.Context(), "request canceled", "error", err)
		return
	case errors.Is(err, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "timeout"
	}

	level := slog.LevelWarn
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	h.log.LogAttrs(r.Context(), level, "request failed",
		slog.Any("error", err),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.String("code", code),
	)

	if writeErr := respond.Error(w, status, code, publicMessage(status)); writeErr != nil {
		h.log.ErrorContext(r.Context(), "write error response", "error", writeErr)
	}
}

func publicMessage(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid request"
	case http.StatusRequestEntityTooLarge:
		return "request body too large"
	case http.StatusNotFound:
		return "not found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusGatewayTimeout:
		return "timeout"
	default:
		return "internal server error"
	}
}

```

Если операций несколько, заведи отдельный интерфейс в `ports.go` и
отдельный файл `get.go`, если `handler.go` разрастается. Второй обработчик
выглядит так же: разбор → вызов → ответ.

### Тело запроса

- `http.MaxBytesReader` обязателен: без лимита клиент пришлёт гигабайт.
  Превышение — `*http.MaxBytesError` и код 413, а не 400.
- `DisallowUnknownFields` — по умолчанию для API. Если нужна
  совместимость с лишними полями, отключи осознанно.
- Проверка второго JSON-объекта (`dec.Decode(&struct{}{})` и `io.EOF`)
  защищает от «двух тел в одном запросе».
- `Content-Type` проверяется, если контракт это требует: не все клиенты
  присылают корректный заголовок. Не делай это без нужды.
- `r.Context()` передаётся в сервис всегда. `context.Background()` в
  обработчике — ошибка.

### Ошибки и коды

| Ситуация | Код | Код в теле |
|---|---|---|
| Успех | 200/201/204 | — |
| Неверный JSON, путь, query, формат | 400 | `invalid_request` |
| Тело больше лимита | 413 | `payload_too_large` |
| Нет аутентификации | 401 | `unauthorized` |
| Нет прав | 403 | `forbidden` |
| Не найдено | 404 | `not_found` |
| Конфликт состояния | 409 | `conflict` |
| Семантическая валидация | 422 | `unprocessable_entity` |
| Слишком много запросов | 429 | `too_many_requests` |
| Внутренняя ошибка | 500 | `internal_error` |
| Зависимость недоступна | 503 | `unavailable` |
| Таймаут зависимости | 504 | `timeout` |

- Домен и сервис возвращают sentinel/типизированные ошибки, а не HTTP-коды.
- Маппинг — в одном месте (`errors.go` или `writeError`). Не сравнивай
  ошибки по тексту.
- Наружу — публичное сообщение, внутреннее — в лог.
- `context.Canceled` — клиент ушёл; писать ответ обычно некуда. Не
  превращай это в 500 в логах уровня error.
- `context.DeadlineExceeded` — 504, если это таймаут зависимости.
- 499 — не стандартный код; не используй без явного решения проекта.

### Ответ

```go
// internal/transport/<appname>/http/respond/respond.go
package respond

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func JSON(w http.ResponseWriter, status int, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		writeInternalError(w)
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return nil
	}

	_, err = w.Write(body)
	return err
}

func Error(w http.ResponseWriter, status int, code, message string) error {
	return JSON(w, status, ErrorBody{Error: message, Code: code})
}

func writeInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ErrorBody{
		Error: "internal server error",
		Code:  "internal_error",
	})
}
```

- `json.Marshal` до `WriteHeader`: если сериализация упала, ещё можно
  отдать 500.
- `204 No Content` — без тела.
- `Content-Type` — `application/json; charset=utf-8`.
- Форма ошибки одна на весь API. Если контракт уже другой — следуй
  контракту, но не смешивай `http.Error` и JSON в одном API.

## Логирование

- Access log — в middleware (`skills/http-middlewares`), не в каждом
  обработчике.
- В обработчике логируй только то, что middleware не знает: доменную
  ошибку, ID операции, код ошибки.
- Обязательно `logger.ErrorContext`/`WarnContext` с `r.Context()`, чтобы
  сохранить request ID и трейс.
- Не логируй тело запроса, `Authorization`, токены, cookie и персональные
  данные. Query-параметры тоже могут содержать секреты.

### Аудит и побочные эффекты

- Аудит, метрики бизнес-событий и уведомления — не бизнес-логика
  контроллера. Если они нужны, это порт сервиса или отдельный декоратор.
- Не передавай `*http.Request` в сервис: это тянет HTTP в прикладной слой.
  Передай явные данные — user ID, request ID, путь, результат операции.
- Если аудит должен срабатывать на каждый HTTP-запрос, это middleware;
  если только на успешную бизнес-операцию — сервис.
- Ошибка аудита не должна менять успешный ответ, если контракт не требует
  обратного; решение фиксируется в проекте.

## Тесты

```go
// internal/transport/<appname>/http/documents/handler_test.go
package documents

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"<module>/internal/domain"
	"<module>/internal/shared"

	"github.com/stretchr/testify/require"
)

type fakeCreator struct {
	out shared.CreateDocumentOutput
	err error
}

func (f fakeCreator) Create(context.Context, shared.CreateDocumentInput) (shared.CreateDocumentOutput, error) {
	return f.out, f.err
}

func TestCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "created",
			body:       `{"title":"Акт","kind":"inspection"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"1","status":"created"}`,
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid request","code":"invalid_request"}`,
		},
		{
			name:       "body too large",
			body:       `{"title":"` + strings.Repeat("a", maxBodyBytes) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   `{"error":"request body too large","code":"payload_too_large"}`,
		},
		{
			name:       "conflict",
			body:       `{"title":"Акт","kind":"inspection"}`,
			err:        domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":"conflict","code":"conflict"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := New(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				fakeCreator{
					out: shared.CreateDocumentOutput{ID: "1", Status: "created"},
					err: tt.err,
				},
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/documents", strings.NewReader(tt.body))

			h.Create(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
```

Правила:

- `httptest.NewRequest` + `httptest.NewRecorder`; реальный порт не нужен.
- Фейк или mockery-мок интерфейса операции; без БД и сети.
- Проверяй статус, `Content-Type`, тело через `require.JSONEq`,
  заголовки, которые важны контракту.
- Кейсы: успех, битый JSON, отсутствующий path/query, ошибка сервиса,
  таймаут/отмена контекста.
- Отдельный тест router — в `skills/http-server`; middleware — в
  `skills/http-middlewares`.

## Антипаттерны

- Бизнес-логика, SQL, транзакции и вызовы внешних API в обработчике.
- Импорт `services`, `infra`, `app`, `config` из контроллера.
- Доменная сущность с JSON-тегами вместо `Response`.
- Маппинг ошибок по строке или по HTTP-коду внутри сервиса.
- `http.Error` в одном месте и JSON в другом для одного API.
- `json.NewDecoder(r.Body).Decode` без лимита тела.
- `context.Background()` вместо `r.Context()`.
- `w.WriteHeader` после `w.Write`.
- Логирование тела, токенов и персональных данных.
- Передача `*http.Request` в сервис или домен.
- Паника на ожидаемой ошибке пользователя.
- `_ = err` и молчаливое проглатывание ошибки записи ответа.

## Чек-лист

- [ ] Обработчик только разбирает, вызывает и маппит.
- [ ] Интерфейс операции объявлен в `ports.go`.
- [ ] `Request`/`Response` не протекают в сервис.
- [ ] Тело ограничено `http.MaxBytesReader`.
- [ ] `r.Context()` доходит до сервиса.
- [ ] Ошибки маппятся в одном месте.
- [ ] Наружу не уходят внутренние детали.
- [ ] Формат ошибки един для API.
- [ ] Есть тест на успех, битый запрос и ошибку сервиса.

## Эталон

- `references/ports.go` — интерфейс операции.
- `references/request.go` — `Request` и `toInput`.
- `references/response.go` — `Response` и `fromOutput`.
- `references/handler.go` — тонкий обработчик, лимит тела, маппинг ошибок.
- `references/respond.go` — общий JSON-ответ и формат ошибки.
- `references/handler_test.go` — табличный тест на `httptest`.

## Границы

- «Общие правила входящих адаптеров и интерфейса операции» —
  `skills/transport`.
- «HTTP-сервер и роутинг» — `skills/http-server`.
- «Middleware, auth и rate limit» — `skills/http-middlewares`.
- «Сервис, который вызывает контроллер» — `skills/services`.
- «Как тестировать» — `skills/testing`.
