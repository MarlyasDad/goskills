---
name: composition-root
description: "Правила composition root в app.go: порядок сборки адаптеров, репозиториев, сервисов, контроллеров и HTTP-сервера, запуск, ожидание и корректная остановка. Используется при подключении нового адаптера, сервиса или сервера. Границы слоёв — skills/architecture, конфиг — skills/config, main — skills/bootstrap."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: architecture
---

# Composition root

Навык отвечает на вопрос «**где и в каком порядке собирать приложение**» и
«**как запускать и останавливать его компоненты**». Архитектурные границы —
`skills/architecture`; источники настроек — `skills/config`; `main` и
сигналы ОС — `skills/bootstrap`; HTTP-сервер — `skills/http-server`.

## Главное правило

`internal/app/<appname>/app.go` — **единственное место**, где вызываются
конструкторы конкретных адаптеров, репозиториев, сервисов, контроллеров,
router и сервера.

```text
cmd/<appname>/main.go     конфиг, логгер, app.New, сигналы ОС
        │
        ▼
internal/app/<appname>/   composition root: собрать и запустить
        │
        ├── infra/        БД, клиенты, брокеры
        ├── repository/   порты и реализации доступа к данным
        ├── services/     use cases
        └── transport/    контроллеры, router, сервер
```

- `cmd` не создаёт адаптеры.
- `router.go` не создаёт сервисы и контроллеры.
- Контроллер не создаёт клиентов и репозитории.
- Сервис не создаёт адаптеры.
- Только `app.go` знает, какая реализация соответствует какому интерфейсу.

## Каталог

```schema
internal/app/<appname>/
└── app.go
```

Пока приложение небольшое, одного `app.go` достаточно. Если файл стал
нечитаемым, дели его по жизненному циклу (`http.go`, `consumers.go`) в том
же пакете. Не дели по слоям (`services.go`, `handlers.go`): так сборка
снова размазывается. Отдельный пакет под composition root не нужен.

Готовый файл — `references/app.go`. Скопировать в
`internal/app/<appname>/app.go`, заменить `<module>`/`<appname>` и
подставить фактические адаптеры. Фрагменты ниже объясняют порядок;
канонический код — в `references/`.

## Порядок сборки

Собирай зависимости снизу вверх:

| Шаг | Что создаётся | Откуда берётся `Config` |
|---|---|---|
| 1 | Платформа: логгер уже создан в `main`; метрики/трейсинг — если есть | `cfg.Logger`, `cfg.Metrics`, ... |
| 2 | Driven-адаптеры: пул БД, клиенты, producer | `cfg.Postgres`, `cfg.Kafka`, ... |
| 3 | Репозитории: обёртки над адаптерами | без своего конфига |
| 4 | Сервисы приложения: use cases | без своего конфига |
| 5 | Контроллеры: `New(log, svc)` | без своего конфига |
| 6 | Router: получает готовые контроллеры | `cfg.HTTP` |
| 7 | HTTP-сервер и другие серверы/consumer'ы | `cfg.HTTP`, `cfg.ShutdownTimeout` |

Правила:

- Конструктор адаптера принимает свой локальный `Config`, валидирует его и
  возвращает ошибку. `app.go` не вызывает `Validate` в обход `New`.
- `New` приложения возвращает ошибку, а не паникует. Если следующий шаг
  упал, уже созданные ресурсы нужно закрыть.
- Конструкторы не запускают горутины и не биндят порты. Всё, что имеет
  жизненный цикл, стартует в `Run`.
- `app.go` не содержит бизнес-логики: только создание, связывание,
  запуск и останов.

## Пример

```go
// internal/app/<appname>/app.go
package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"<module>/internal/config"
	"<module>/internal/infra/postgres"
	"<module>/internal/services/api/document"
	apihttp "<module>/internal/transport/api/http"
	"<module>/internal/transport/api/http/documents"
)

type App struct {
	log             *slog.Logger
	server          *http.Server
	shutdownTimeout time.Duration
	cleanup         func() error
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	repo := postgres.NewDocumentRepository(pg)
	svc := document.New(repo)
	handler := documents.New(log, svc)

	router := apihttp.NewRouter(apihttp.Deps{
		Config:   cfg.HTTP,
		Log:      log,
		Handlers: apihttp.HandlerSet{Documents: handler},
	})

	server, err := apihttp.NewServer(cfg.HTTP, log, router)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("http server: %w", err)
	}

	return &App{
		log:             log,
		server:          server,
		shutdownTimeout: cfg.ShutdownTimeout,
		cleanup: func() error {
			pg.Close()
			return nil
		},
	}, nil
}
```

Если `pg.Close()` возвращает ошибку, верни её из `cleanup`; если нет —
оберни в `func() error { pg.Close(); return nil }`, как выше.

## Запуск и останов

```go
// internal/app/<appname>/app.go
func (a *App) Run(ctx context.Context) (err error) {
	defer func() {
		if a.cleanup != nil {
			err = errors.Join(err, a.cleanup())
		}
	}()

	ln, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", a.server.Addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("http server started", "addr", ln.Addr().String())
		serveErr := a.server.Serve(ln)
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		_ = a.server.Close()
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := <-errCh; err != nil {
		return fmt.Errorf("serve http: %w", err)
	}
	return nil
}
```

Почему так:

- `net.Listen` до горутины: ошибка занятого порта возвращается сразу, а не
  теряется в логе.
- `Serve` в горутине, результат — в буферизованный канал.
- `select` ждёт либо `ctx.Done()` от сигнала ОС, либо падения сервера.
- `Shutdown` даёт активным запросам завершиться; `Close` — жёсткий fallback,
  если таймаут истёк. После `Close` обработчики могут ещё завершаться,
  поэтому общие ресурсы (БД, клиенты) закрываются только после остановки
  сервера.
- `ErrServerClosed` — нормальный результат после `Shutdown`/`Close`, а не
  ошибка.
- `Shutdown` не ждёт hijacked-соединения (WebSocket). Если они есть, их
  закрывает отдельный компонент, а не HTTP-сервер.
- `cleanup` вызывается всегда, в том числе при ошибке запуска.
- `errors.Join` сохраняет и ошибку `Run`, и ошибку закрытия ресурсов.

`main` передаёт сюда `signal.NotifyContext` и не занимается lifecycle сам;
полный `main.go` — `skills/bootstrap/references/main.go`.

## Несколько компонентов

Когда появляются consumer, relay, фоновые воркеры, у каждого свой запуск
и останов. Держи их в `Run` явно:

1. запусти компоненты в детерминированном порядке;
2. дождись либо `ctx.Done()`, либо первой фатальной ошибки;
3. вызывай `Stop` в обратном порядке;
4. дождись горутин и только потом закрывай общие ресурсы (БД, клиентов).

Если компонентов больше двух, введи маленький интерфейс:

```go
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

Не прячь запуск в `init()` и не делай глобальный реестр компонентов.
Порядок должен читаться в `Run`.

## Ресурсы и ошибки

- Каждый `New` возвращает `(T, error)`; `app.go` проверяет ошибку и
  добавляет контекст: `fmt.Errorf("postgres: %w", err)`.
- Если ресурс создан, но приложение не запустилось, он должен быть закрыт.
  Для одного-двух ресурсов хватает `defer` с флагом успеха или `cleanup` в
  `App`.
- Не логируй ошибку и не возвращай её одновременно на одном уровне:
  либо `log.Error` + `return nil`, либо `return fmt.Errorf(...)`.
- Не вызывай `log.Fatal`/`os.Exit` в `app.go`: это ломает тесты и
  `defer`.
- Не используй `panic` для ошибок конфигурации и запуска.

## Тесты

- Поведение обработчиков и сервисов тестируется отдельно
  (`skills/http-controllers`, `skills/testing`).
- `app.New` с реальной БД — интеграционный тест; unit-тестом он не
  становится от подмены env.
- Проверять wiring unit-тестом имеет смысл, только если `New` принимает
  фабрики/интерфейсы. По умолчанию это лишняя косвенность: wiring
  проверяется сборкой и интеграционным/e2e-тестом.
- Жизненный цикл `Run` можно проверить с `net.Listen("tcp",
  "127.0.0.1:0")`, но это уже интеграционный тест с реальным сокетом.

## Антипаттерны

- Создание адаптеров, сервисов или контроллеров в `cmd/main.go`.
- Создание сервисов и репозиториев в `router.go`.
- Создание клиентов и репозиториев в контроллере или сервисе.
- Глобальные переменные, `init()` с побочными эффектами, скрытый DI.
- Конструктор, который запускает сервер, биндит порт или стартует горутины.
- `http.ListenAndServe` без `Shutdown`.
- `log.Fatal`/`os.Exit` внутри `app`.
- Игнорирование ошибок конструкторов и закрытия ресурсов.
- Бизнес-логика, SQL или HTTP-коды в `app.go`.
- Общий `App` для нескольких бинарей.
- Запуск компонентов в произвольном порядке без остановки в обратном.

## Чек-лист

- [ ] Все конструкторы вызываются только в `app.go`.
- [ ] `cmd` делает конфиг, логгер, `app.New`, сигналы и `Run`.
- [ ] Router получает готовые контроллеры.
- [ ] Конструкторы не запускают горутины и не биндят порты.
- [ ] `Run` запускает, ждёт и останавливает компоненты.
- [ ] `Shutdown` с таймаутом, `Close` как fallback.
- [ ] `ErrServerClosed` не считается ошибкой.
- [ ] Ресурсы закрываются и при ошибке запуска.
- [ ] `app.go` не содержит бизнес-логики.
- [ ] Ошибки конструкторов обёрнуты и не потеряны.

## Эталон

`references/app.go` — полный `App`: `New`, сборка HTTP-сервера, `Run`,
`net.Listen`, `Serve`, `Shutdown`/`Close`, `cleanup`.

## Границы

- «Какие импорты допустимы и где границы слоёв» —
  `skills/architecture`.
- «Как собрать `Config` и где валидировать» — `skills/config`.
- «Bootstrap, `main.go` и сигналы ОС» — `skills/bootstrap`.
- «HTTP-сервер, router и Config» — `skills/http-server`.
- «Сервисы и use cases» — `skills/services`.
