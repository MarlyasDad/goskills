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
internal/transport/<appname>/http/
├── middleware/      # рядом с контроллерами
├── responses/       # JSON, ошибки, валидация, файлы
└── <group>/<resource>/
    ├── handler.go       # Handler, New и HTTP-методы
    ├── ports.go         # интерфейсы операций, которые нужны контроллеру
    ├── request.go       # Request DTO, validate и toInput
    ├── response.go      # Response DTO и fromOutput
    ├── errors.go        # маппинг ошибок, если он не помещается в handler.go
    └── handler_test.go
```

- `<group>` — `internalapi`, `externalapi` или смысловая группа. `<resource>`
  — ресурс: `documents`, `appeals`.
- Не называй группу `internal`: это зарезервированное имя Go, и пакет
  снаружи `.../http` не сможет его импортировать. Для внутреннего API —
  `internalapi`, для внешнего — `externalapi`.
- Пакет называется по последнему сегменту: `documents`, `appeals`. Если в
  router встречаются два `user` из разных групп, импортируй с алиасом:
  `internaluser`, `externaluser`.
- Один каталог — один ресурс или небольшая группа родственных операций.
  Не складывать все ручки приложения в один `handler.go`.
- Общие конструкторы ответов — `http/responses/`, рядом с
  `http/middleware/`: `RespondWithJSON`, `RespondWithText`,
  `RespondWithError`, `RespondWithValidationErrors`, `RespondWithFile`.
  Формат ошибок должен быть один на весь API.
- Если контроллер один, можно оставить помощники в `handler.go`; при втором
  контроллере выноси их в `responses`, чтобы формат не разъехался.

Готовые файлы — `references/ports.go`, `references/request.go`,
`references/response.go`, `references/handler.go`, `references/responses.go`,
`references/file.go` и тесты `references/handler_test.go`,
`references/file_test.go`. Скопировать контроллер в
`internal/transport/<appname>/http/documents/`, а `responses.go` и `file.go`
— в `http/responses/`; заменить `<module>`/`<appname>`. Фрагменты ниже
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
	"<module>/internal/shared"
	"<module>/internal/transport/<appname>/http/responses"
)

type CreateRequest struct {
	Title string `json:"title" example:"Акт осмотра"`
	Kind  string `json:"kind" example:"inspection"`
}

// validate проверяет формат полей. Бизнес-валидация — в сервисе или домене.
func (r CreateRequest) validate() responses.ValidationErrors {
	var errs responses.ValidationErrors

	if r.Title == "" {
		errs = append(errs, responses.ValidationError{
			Message: "title: value is required",
			Code:    responses.CodeUnassigned,
		})
	}
	if r.Kind == "" {
		errs = append(errs, responses.ValidationError{
			Message: "kind: value is required",
			Code:    responses.CodeUnassigned,
		})
	}

	return errs
}

func (r CreateRequest) toInput() shared.CreateDocumentInput {
	return shared.CreateDocumentInput{
		Title: r.Title,
		Kind:  r.Kind,
	}
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
- В 400 имя поля уже внутри `Message` (`"title: value is required"`).
  `ValidationError.Field` имеет `omitempty` и не заполняется; иначе ответ
  разойдётся с контрактом `{errors:[{message, code}]}`.
- Не добавлять в DTO бизнес-логику; только маппинг и форматную валидацию.

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
	"<module>/internal/transport/<appname>/http/responses"
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

	if errs := req.validate(); len(errs) > 0 {
		if err := responses.RespondWithValidationErrors(w, errs); err != nil {
			h.log.ErrorContext(r.Context(), "write validation response", "error", err)
		}
		return
	}

	out, err := h.creator.Create(r.Context(), req.toInput())
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := responses.RespondWithJSON(w, http.StatusCreated, fromCreateOutput(out)); err != nil {
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
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, errBodyTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled):
		h.log.WarnContext(r.Context(), "request canceled", "error", err)
		return
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
	}

	// TODO: код из корпоративного справочника. До присвоения — заглушка,
	// чтобы не выдумывать значения.
	code := responses.CodeUnassigned

	level := slog.LevelWarn
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	h.log.LogAttrs(r.Context(), level, "request failed",
		slog.Any("error", err),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.Int("code", code),
	)

	if err := responses.RespondWithError(w, status, code, publicMessage(status)); err != nil {
		h.log.ErrorContext(r.Context(), "write error response", "error", err)
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

| Ситуация | HTTP |
|---|---|
| Успех | 200/201/204 |
| Неверный JSON, путь, query, формат | 400 |
| Тело больше лимита | 413 |
| Нет аутентификации | 401 |
| Нет прав | 403 |
| Не найдено | 404 |
| Конфликт состояния | 409 |
| Семантическая валидация | 422 |
| Слишком много запросов | 429 |
| Внутренняя ошибка | 500 |
| Зависимость недоступна | 503 |
| Таймаут зависимости | 504 |

- HTTP-статус выбирает транспорт.
- `Error.Code` — код из **корпоративного справочника ошибок**, не HTTP-статус.
  Значения не выдумываются: пока эксплуатация не присвоила код, используй
  `responses.CodeUnassigned` (`0`). Лог-коллектор сортирует по коду, поэтому
  присвоенные значения должны быть стабильными.
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
// internal/transport/<appname>/http/responses/responses.go
package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CodeUnassigned — код ещё не присвоен эксплуатацией. Это не код из
// корпоративного справочника, а заглушка: значения кодов не выдумываются.
const CodeUnassigned = 0

type Error struct {
	Message string `json:"message"`
	// Code — код из корпоративного справочника; 0 — ещё не присвоен.
	Code int `json:"code"`
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func RespondWithJSON(w http.ResponseWriter, status int, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		if writeErr := RespondWithText(w, http.StatusInternalServerError, "internal server error"); writeErr != nil {
			return fmt.Errorf("marshal response: %w; write fallback: %v", err, writeErr)
		}
		return fmt.Errorf("marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return nil
	}

	_, err = w.Write(body)
	return err
}

func RespondWithText(w http.ResponseWriter, status int, message string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(message))
	return err
}

func RespondWithError(w http.ResponseWriter, status, code int, message string) error {
	return RespondWithJSON(w, status, ErrorResponse{
		Errors: []Error{{Message: message, Code: code}},
	})
}
```

Полные типы `ValidationError`, `ValidationErrors` и
`RespondWithValidationErrors` — в `references/responses.go`. Конструктор
валидации всегда отвечает 400 и отдаёт список полей:

```go
type ValidationError struct {
	// Field не входит в контракт 400: имя поля уже внутри Message.
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	// Code — код из корпоративного справочника; 0 — ещё не присвоен.
	Code int `json:"code"`
}

type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

func RespondWithValidationErrors(w http.ResponseWriter, errs []ValidationError) error {
	return RespondWithJSON(w, http.StatusBadRequest, ValidationErrorResponse{Errors: errs})
}
```

- `json.Marshal` до `WriteHeader`: если сериализация упала, ещё можно
  отдать 500.
- `204 No Content` — без тела.
- Конструкторы возвращают ошибку записи; вызывающий логирует её. Не
  превращай `_, _ = w.Write(...)` в молчаливое проглатывание.
- `Content-Type` — `application/json; charset=utf-8`.
- Форма ошибки одна на весь API:
  `{"errors":[{"message":"...","code":0}]}`. Валидация:
  `{"errors":[{"message":"title: ...","code":0}]}`.
- `0` здесь — `responses.CodeUnassigned`, пока эксплуатация не присвоила
  код из корпоративного справочника.
- Если контракт уже другой — следуй контракту, но не смешивай `http.Error`
  и JSON в одном API.
- Коды приложения (`Code`) — не HTTP-статусы. HTTP-статус выбирает
  обработчик, а код ошибки — из корпоративного справочника и стабилен для
  лог-коллектора.

### Файл

Для `.docx` и других бинарных ответов — `RespondWithFile`, а не JSON и не
ручная сборка заголовков в handler.

```go
docx, err := h.service.Generate(r.Context(), req.toInput())
if err != nil {
	h.writeError(w, r, err)
	return
}

if err := responses.RespondWithFile(
	w,
	http.StatusOK,
	"act.docx",
	responses.ContentTypeDOCX,
	bytes.NewReader(docx),
); err != nil {
	h.log.ErrorContext(r.Context(), "write file response", "error", err)
}
```

Правила:

- `Content-Type` для docx — `responses.ContentTypeDOCX`
  (`application/vnd.openxmlformats-officedocument.wordprocessingml.document`).
  Если тип неизвестен — `application/octet-stream`.
- `Content-Disposition: attachment; filename=...` ставит конструктор; имя
  очищается от путей и управляющих символов. Не собирай заголовок вручную
  из пользовательской строки.
- `RespondWithFile` пишет потоком через `io.Copy`: не буферизуй большой файл
  в памяти и не превращай `[]byte` в строку.
- `Content-Length` ставится, если размер можно узнать (`bytes.Buffer`,
  `bytes.Reader`, `io.Seeker`). Иначе ответ идёт chunked — это нормально.
- Ошибка после `WriteHeader` означает, что заголовки и часть тела уже ушли:
  статус не изменить. Верни/залогируй ошибку, не пытайся писать второй ответ.
- Для `Range`, `If-Modified-Since` или файла на диске/S3 используй
  `http.ServeContent` с `io.ReadSeeker`, а не `RespondWithFile`.
- `io.ReadCloser` (например, файл из S3) закрывай в handler после
  `RespondWithFile`.

## Логирование

- Access log — в middleware (`skills/http-middlewares`), не в каждом
  обработчике.
- В обработчике логируй только то, что middleware не знает: доменную
  ошибку, ID операции, код ошибки.
- В лог ошибки добавляй `code` — по нему сортирует лог-коллектор.
  HTTP-статуса для этого недостаточно; до присвоения кода — `CodeUnassigned`.
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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"<module>/internal/domain"
	"<module>/internal/shared"
	"<module>/internal/transport/<appname>/http/responses"

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
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"invalid request","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "body too large",
			body:       `{"title":"` + strings.Repeat("a", maxBodyBytes) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"request body too large","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "missing title",
			body:       `{"kind":"inspection"}`,
			wantStatus: http.StatusBadRequest,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"title: value is required","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "conflict",
			body:       `{"title":"Акт","kind":"inspection"}`,
			err:        domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"conflict","code":%d}]}`,
				responses.CodeUnassigned,
			),
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
- Кейсы: успех, битый JSON, ошибка валидации, отсутствующий path/query,
  ошибка сервиса, таймаут/отмена контекста.
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
- `references/request.go` — `Request`, `validate` и `toInput`.
- `references/response.go` — `Response` и `fromOutput`.
- `references/handler.go` — тонкий обработчик, лимит тела, маппинг ошибок.
- `references/responses.go` — JSON, текст, ошибки и валидационные ответы.
- `references/file.go` — `RespondWithFile`, docx MIME, Content-Disposition,
  потоковая запись.
- `references/handler_test.go` — табличный тест на `httptest`.
- `references/file_test.go` — тесты файлового ответа, размера и ошибки
  записи.

## Границы

- «Общие правила входящих адаптеров и интерфейса операции» —
  `skills/transport`.
- «HTTP-сервер и роутинг» — `skills/http-server`.
- «Middleware, auth и rate limit» — `skills/http-middlewares`.
- «Сервис, который вызывает контроллер» — `skills/services`.
- «Как тестировать» — `skills/testing`.
