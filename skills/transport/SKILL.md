---
name: transport
description: "Правила входящих адаптеров: HTTP, gRPC, Kafka и WebSocket, тонкие контроллеры, DTO, маппинг ошибок и интерфейсы операций. HTTP-часть — skills/http-server, http-controllers, http-middlewares; сборка — skills/composition-root. Используется при добавлении обработчика. Импорты — skills/architecture, сервисы — skills/services."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: architecture
---

# Транспорт и контроллеры

Навык отвечает на вопрос «**как принять запрос и вызвать сервис**».
HTTP-контроллеры — это часть транспорта, отдельный навык `controllers` не
нужен.

> Статус: HTTP-часть вынесена в `skills/http-server`,
> `skills/http-controllers` и `skills/http-middlewares`. Для gRPC, Kafka и
> WebSocket — базовые правила. TODO: OpenAPI, пагинация, idempotency keys,
> трейсинг запросов.

## Границы

- Какие импорты допустимы — `skills/architecture`.
- Как устроен сервис, который вызывается, — `skills/services`.
- Где лежит каталог — `skills/project-structure`.
- Как тестировать — `skills/testing`, `skills/integration-testing`.

## Каталог

```schema
internal/transport/<appname>/
├── http/
│   ├── config.go
│   ├── server.go
│   ├── router.go
│   ├── middleware/
│   └── <group>/<resource>/
├── grpc/
├── kafka/
└── websocket/
```

- Один подкаталог на приложение: у разных приложений не пересекаются роуты,
  консьюмеры и порты.
- Внутри — подкаталоги по протоколам. Создаются по необходимости.
- Контракты (`openapi.yaml`, `*.proto`) лежат в `api/`, не здесь.

## Роль

Транспорт — **driving/primary адаптер**:

1. разобрать запрос;
2. провалидировать вход;
3. вызвать операцию сервиса через интерфейс;
4. собрать ответ;
5. смаппить ошибку в код транспорта.

Бизнес-логики, SQL, транзакций и знания о БД здесь нет.

## Интерфейс операции

- `transport` объявляет интерфейс, который ему нужен: потребитель —
  контроллер.
- Реализует интерфейс сервис из `services/<appname>/<service>`; связывает
  `app.go`.
- `transport` не импортирует `app`, `services`, `infra`, `repository`,
  `config`, `cmd`.
- В сигнатурах — `context`, доменные типы или типы из `shared`, а не
  `Request`/`Response` транспорта. Иначе сервис начнёт импортировать
  транспорт.
- Интерфейс маленький: 1–3 метода.

```go
// internal/shared/user.go
type RegisterUserInput struct {
	Email string
}

type RegisterUserOutput struct {
	ID string
}
```

```go
// internal/transport/creator/http/handler.go
type RegisterUser interface {
	Register(ctx context.Context, input shared.RegisterUserInput) (shared.RegisterUserOutput, error)
}
```

```go
// internal/services/creator/user/register.go
func (s *Service) Register(
	ctx context.Context,
	input shared.RegisterUserInput,
) (shared.RegisterUserOutput, error) {
	// ...
}
```

Почему так:

- если транспорт возьмёт `user.RegisterInput` из сервиса — транспорт начнёт
  импортировать `services`;
- если сервис возьмёт `transport.RegisterRequest` — сервис начнёт
  импортировать `transport`;
- общий контракт в `shared` разрывает эту связь: `transport → shared` и
  `services → shared`, но не друг в друга.

Типы операции могут лежать в `domain`, если это доменные типы. `shared` — для
прикладных контрактов, которые нужны и транспорту, и сервису.

## DTO

- `Request`/`Response` — в `transport`.
- `Input`/`Output` — в сервисе; если нужны транспорту, вынести их в `shared`
  и импортировать обеими сторонами. Тогда `dto.go` в сервисе может содержать
  алиасы или не создаваться.
- JSON/XML/query-теги — только в DTO транспорта.
- Доменные сущности не сериализуются напрямую.
- Валидация формата — на транспорте; бизнес-валидация — в сервисе/домене.

## Ошибки

- Домен и сервис возвращают типизированные/`sentinel`-ошибки, а не HTTP-коды.
- Транспорт маппит их в HTTP/gRPC-коды в одном месте.
- Не сравнивать ошибки по тексту.
- Не отдавать наружу внутренние ошибки, SQL и стек.
- `context.Canceled` и `DeadlineExceeded` — отдельная обработка.

## HTTP

HTTP-часть разложена по отдельным навыкам:

- `skills/http-server` — `Config`, `server.go`, корневой router,
  health/readiness, metrics/docs, запуск и останов.
- `skills/http-controllers` — `handler.go`, `ports.go`, `request.go`,
  `response.go`, DTO, коды и маппинг ошибок.
- `skills/http-middlewares` — `Chain`, порядок, RequestID, Logging,
  Recovery, CORS, Auth, RateLimit.
- `skills/composition-root` — сборка сервера, сервисов и адаптеров в
  `app.go`, graceful shutdown.

Общие правила выше — интерфейс операции, DTO и ошибки — действуют и для
HTTP. Ниже перечислены только то, что не зависит от протокола:

- Коды: 200/201/204, 400, 401/403, 404, 409, 422, 429, 500, 503/504.
- `Content-Type`, лимит тела, таймауты — на транспорте.
- Health/readiness — отдельные роуты, без бизнес-логики.
- Auth и rate limit — middleware, а не код в каждом обработчике.

## gRPC

- Контракты — в `api/`, сгенерированный код не правится руками.
- Сервер — тонкий адаптер над тем же интерфейсом операции.
- Ошибки — gRPC status codes; не `http.Error`.
- Тест — `bufconn`.

## Kafka

- Consumer разбирает сообщение, вызывает операцию сервиса, коммитит offset.
- Идемпотентность и retry/DLQ — по правилам проекта.
- Не держать бизнес-логику в консьюмере.

## WebSocket

- Управление соединением и протоколом — здесь.
- Операции — через интерфейс сервиса.
- Не хранить доменное состояние в соединении.

## Тесты

- `httptest` для HTTP, `bufconn` для gRPC.
- Фейки интерфейса операции, без реальной БД и сети.
- Проверяются: разбор запроса, коды, маппинг ошибок, граничные случаи.
- HTTP-контроллеры — `skills/http-controllers`, middleware —
  `skills/http-middlewares`.
- Реальный брокер и полный стек — `skills/integration-testing`.
- Подробно — в `skills/testing`.

## Антипаттерны

- Бизнес-логика и SQL в контроллере.
- Импорт конкретного сервиса или `infra` из транспорта.
- Доменные сущности с JSON-тегами.
- Маппинг ошибок по строке.
- Один огромный handler на все роуты.
- Роуты и DTO разных приложений в одном пакете.
- Создание сервисов, репозиториев и адаптеров в `router.go`.
- Auth-middleware, импортирующий `infra` или конкретного провайдера
  identity.

## Чек-лист

- [ ] Обработчик только разбирает, вызывает и маппит.
- [ ] Интерфейс операции объявлен в транспорте.
- [ ] Сервис не импортирует транспорт.
- [ ] Ошибки маппятся в одном месте.
- [ ] DTO не протекают в домен.
- [ ] Есть тест на успех и на ошибку.
