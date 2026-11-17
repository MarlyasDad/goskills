---
name: bootstrap
description: Создание нового Go-проекта с нуля после подключения навыков goskills — инициализация модуля, минимальный каркас и обязательный набор каталогов. Используется на старте проекта. Раскладка — skills/project-structure, зависимости — skills/architecture.
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: scaffolding
---

# Создание проекта с нуля

Навык задаёт **порядок первых шагов** нового Go-проекта и минимальный
результат, после которого проект собирается, запускается и готов к развитию.
Раскладку каталогов определяет `skills/project-structure`, зависимости —
`skills/architecture`, тесты — `skills/testing`.

## Когда применять

- Каталог проекта пуст или содержит только `.git`.
- `go.mod` отсутствует.
- Навыки goskills уже подключены к OpenCode (см. README репозитория goskills).

Если `go.mod` уже есть — каркас не пересоздаётся. Дальше работают
`project-structure` и `architecture`.

## Входные данные

До создания файлов выяснить:

| Что | Как определить |
|---|---|
| Имя модуля | git remote → путь из него (`git@github.com:org/repo.git` → `github.com/org/repo`); иначе `go.work` или родительский `go.mod`; иначе **спросить пользователя** |
| Имя приложения `<appname>` | спросить пользователя, если он не указал его явно. Это имя бинаря и одинаковый сегмент в `cmd/`, `internal/app/`, `internal/config/` |
| Версия Go | `.go-version`, CI-конфиг; иначе минимальная поддерживаемая — `1.26`; иную версию брать только по указанию пользователя |

Правила:

- Не угадывать имя модуля по имени каталога.
- Не подставлять `example.com/...` без явного согласия.
- `<appname>` — латиница в нижнем регистре, без пробелов и дефисов: имя
  каталога, бинаря и Go-пакета.
- Если проект ещё не под git — выполнить `git init`. Remote нужен только для
  имени модуля.
- Задать все вопросы одним сообщением, до создания файлов. Каркас не
  создаётся, пока не определены модуль и приложение.

## Результат

```schema
.
├── go.mod
├── .gitignore
├── README.md
├── Makefile
├── AGENTS.md
├── config/
│   └── .env.example
├── cmd/
│   └── <appname>/
│       └── main.go
└── internal/
    ├── app/
    │   └── <appname>/
    │       └── app.go
    ├── config/
    │   └── <appname>/
    │       ├── config.go
    │       └── config_test.go
    └── platform/
        ├── config/
        │   └── env.go
        └── logger/
            └── logger.go
```

Это обязательный минимум. Остальные каталоги из `project-structure`
(`internal/domain`, `internal/services`, `internal/repository`,
`internal/infra`, `internal/transport`, `internal/shared`, `tests/`, `api/`)
создаются вместе с первым файлом и в каркас не входят: пустые каталоги не
создаются.

## Порядок шагов

1. **Модуль.** `go mod init <module>`. Проверить `module` и `go` в `go.mod`.
   Зависимости на этом шаге не добавлять.
2. **Корневые файлы.** `.gitignore`, `README.md`, `config/.env.example`.
3. **Платформа.** `internal/platform/config/env.go` — чтение env;
   `internal/platform/logger/logger.go` — создание логгера.
4. **Конфиг приложения.** `internal/config/<appname>/config.go` — структура,
   `Load()`, валидация; `config_test.go` — дефолты и ошибка на неверном
   значении.
5. **Приложение.** `internal/app/<appname>/app.go` — точка сборки сценариев.
   Пока сценариев нет — минимальный `App` с `Run(ctx)`.
6. **Точка входа.** `cmd/<appname>/main.go`: конфиг → логгер → приложение →
   запуск и graceful shutdown по сигналу.
7. **Команды.** `Makefile`: `build`, `run`, `test`, `test-race`, `lint`.
8. **AGENTS.md.** Обзор проекта, команды из `Makefile`, правило «тесты по
   `skills/testing`», границы. Если файл уже есть — слить, а не перезаписать.
9. **Проверка.** `go build ./...`, `go vet ./...`, `go test ./...`,
   `make build`.

## Минимальная реализация

Код ниже — эталон каркаса. Плейсхолдеры `<module>` и `<appname>` заменить,
импорты привести к фактическому имени модуля.

### internal/platform/config/env.go

```go
package config

import (
	"os"
	"strconv"
	"time"
)

// String возвращает значение переменной окружения или fallback.
func String(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// Duration разбирает длительность из env; при ошибке возвращает её.
func Duration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}
	return d, nil
}

// Int разбирает целое из env; при ошибке возвращает её.
func Int(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return n, nil
}
```

### internal/platform/logger/logger.go

```go
package logger

import (
	"log/slog"
	"os"
)

// New создаёт JSON-логгер: debug вне production, info в production.
func New(env string) *slog.Logger {
	level := slog.LevelDebug
	if env == "production" {
		level = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
```

### internal/config/<appname>/config.go

```go
package <appname>

import (
	"fmt"
	"time"

	"<module>/internal/platform/config"
)

// Config — типизированные настройки приложения.
type Config struct {
	Env             string
	ShutdownTimeout time.Duration
}

// Load читает и валидирует настройки.
func Load() (*Config, error) {
	timeout, err := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("shutdown timeout: %w", err)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("shutdown timeout must be positive")
	}

	return &Config{
		Env:             config.String("APP_ENV", "local"),
		ShutdownTimeout: timeout,
	}, nil
}
```

### internal/config/<appname>/config_test.go

```go
package <appname>

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "local" {
		t.Errorf("Env = %q, want %q", cfg.Env, "local")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 10*time.Second)
	}
}

func TestLoad_RejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "soon")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
```

### internal/app/<appname>/app.go

```go
package <appname>

import "context"

// App — точка сборки сценариев приложения.
type App struct{}

func New() *App {
	return &App{}
}

// Run работает до отмены контекста. Сценарии добавляются сюда.
func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
```

### cmd/<appname>/main.go

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"<module>/internal/app/<appname>"
	appconfig "<module>/internal/config/<appname>"
	"<module>/internal/platform/logger"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.New().Run(ctx); err != nil {
		log.Error("run app", "error", err)
		os.Exit(1)
	}

	log.Info("stopped", "timeout", cfg.ShutdownTimeout)
}
```

### Makefile

```make
APP := <appname>

.PHONY: build run test test-race lint

build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

test-race:
	go test -race -shuffle=on -count=1 ./...

lint:
	go vet ./...
```

`bin/` в git не коммитится.

### .gitignore

```gitignore
bin/
.env
.env.*
!.env.example
coverage.out
.idea/
.DS_Store
```

### config/.env.example

```dotenv
# local | staging | production
APP_ENV=local
# Таймаут graceful shutdown
SHUTDOWN_TIMEOUT=10s
```

### README.md

Минимум: назначение проекта, требования (Go), запуск (`make run`), переменные
окружения со ссылкой на `config/.env.example`, команды `make`.

### AGENTS.md

Минимум:

- имя проекта и модуль;
- команды из `Makefile` без придуманных;
- правило: раскладка — `skills/project-structure`, зависимости —
  `skills/architecture`, тесты — `skills/testing`;
- границы: не создавать пустые каталоги, не добавлять зависимости без
  необходимости, не коммитить `.env`.

Если `AGENTS.md` уже существует — дополнить, не перезаписывать. Расширенный
шаблон и слияние — в `skills/agents-md`.

## Проверка

```bash
go build ./...
go vet ./...
go test ./...
make build
```

Ожидания:

- `go.mod` содержит только `module` и `go`, без зависимостей.
- `go build` создаёт бинарь, `make run` запускается и завершается по Ctrl+C.
- `go test` проходит; тест не ходит в сеть и не использует `time.Sleep`.
- `internal/platform` не импортирует пакеты проекта.
- `internal/domain` и другие слои отсутствуют, пока в них нет файлов.

## Что не делать

- Не создавать пустые каталоги «на будущее».
- Не добавлять `Dockerfile`, `docker-compose.yml`, миграции и CI, пока нет
  реальной потребности.
- Не создавать `internal/domain`, `internal/services`, `internal/repository`,
  `internal/infra`, `internal/transport`, `tests/`, `api/` без первого файла.
- Не добавлять сторонние зависимости в `go.mod` на старте: только stdlib.
- Не помещать бизнес-логику и SQL в `cmd/`.
- Не заводить `pkg/`, `util/`, `common/`, `helpers/`.
- Не коммитить `.env*` (кроме `.env.example`) и `bin/`.
- Не оставлять проект в нерабочем состоянии: каждый шаг заканчивается
  собираемым кодом.

## Передача дальше

После каркаса:

- первая сущность → `internal/domain/<entity>/`;
- первый порт и адаптер → `internal/repository/`, `internal/infra/<adapter>/`;
- первый входящий адаптер → `internal/transport/<appname>/`;
- первый интеграционный тест → `tests/` с тегом `integration`;
- коммиты — по `skills/commit-conventions`, запись — по `skills/changelog`,
  тесты — по `skills/testing`.
