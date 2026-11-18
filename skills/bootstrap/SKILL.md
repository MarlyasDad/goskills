---
name: bootstrap
description: Создание нового Go-проекта с нуля после подключения навыков goskills — инициализация модуля, минимальный каркас и обязательный набор каталогов. Используется на старте проекта. Раскладка — skills/project-structure, зависимости — skills/architecture, конфиг — skills/config.
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
`skills/architecture`, конфигурацию — `skills/config`, тесты —
`skills/testing`.

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
| Имя модуля | Путь импорта (`module`), не имя проекта. Кандидат — из git remote: `git@github.com:org/repo.git` → `github.com/org/repo`. Спросить, будет ли проект импортироваться другими; если нет — предложить короткое имя и получить явное согласие. Без remote — `go.work` или родительский `go.mod`; иначе **спросить пользователя** |
| Имя приложения `<appname>` | Спросить пользователя, если он не указал его явно. Выбирается по бинарю: короткое, без доменного префикса и без имени репозитория — `creator`, `api`. Это имя бинаря и одинаковый сегмент в `cmd/`, `internal/app/` и `internal/transport/` |
| Версия Go | явно указанная пользователем; иначе `.go-version`, `toolchain` или CI-конфиг; иначе системная из `go version`. Если системная ниже `1.26` — уточнить у пользователя |
| Данные разработчика | имя и email для локального git-конфига. Спросить вместе с остальными данными; если пользователь не указал — показать текущие `git config user.name` и `git config user.email` и подтвердить |

Правила:

- `module` и `<appname>` — разные вещи: модуль — путь импорта, `appname` —
  имя бинаря. Не выводить одно из другого.
- Не угадывать имя модуля по имени каталога или репозитория.
- Не подставлять `example.com/...` без явного согласия.
- Полный путь из git remote — дефолт. Короткое имя модуля допустимо, только
  если проект не импортируется другими, и только после явного согласия
  пользователя. Причина записывается в `AGENTS.md` или `README.md`.
- Не выводить `<appname>` из имени репозитория: у `vp-appeal-doc-creator`
  приложение — `creator`, а не `vp-appeal-doc-creator`.
- Версию Go уточнять вместе с модулем и именем приложения. Если пользователь её
  не указал — взять системную из `go version`, а не выдумывать. Если `go`
  недоступен или системная версия ниже `1.26` — спросить пользователя.
- Данные разработчика записывать в **локальный** конфиг проекта, а не в
  глобальный: имя и email могут отличаться от настроек ПК. Не придумывать их:
  если пользователь не указал и git-конфиг пуст — запросить.
- `<appname>` — латиница в нижнем регистре, без пробелов и дефисов: имя
  каталога, бинаря и Go-пакета.
- Если проект ещё не под git — выполнить `git init`. Remote нужен только для
  имени модуля.
- Задать все вопросы одним сообщением, до создания файлов. Каркас не
  создаётся, пока не определены модуль, приложение, версия Go и данные
  разработчика по правилам выше.

## Результат

```schema
.
├── go.mod
├── .gitignore
├── README.md
├── Makefile
├── AGENTS.md
├── configs/
│   └── .env.example
├── cmd/
│   └── <appname>/
│       └── main.go
└── internal/
    ├── app/
    │   └── <appname>/
    │       └── app.go
    ├── config/
    │   ├── env.go
    │   ├── config.go
    │   └── config_test.go
    └── platform/
        └── logger/
            └── logger.go
```

Это обязательный минимум. Остальные каталоги из `project-structure`
(`internal/domain`, `internal/services`, `internal/repository`,
`internal/infra`, `internal/transport`, `internal/shared`, `tests/`, `api/`)
создаются вместе с первым файлом и в каркас не входят: пустые каталоги не
создаются.

## Порядок шагов

1. **Git и разработчик.** Если проект ещё не под git — `git init`. Записать
   имя и email в локальный конфиг проекта:
   ```bash
   git config --local user.name "<name>"
   git config --local user.email "<email>"
   ```
   Проверить `git config --show-origin --get user.email`: источник должен быть
   `.git/config`. Глобальный конфиг не менять.
2. **Модуль и приложение.** Согласовать `module` (путь импорта) и
   `<appname>` (имя бинаря) — это разные значения. Определить версию Go по
   приоритету: пользователь → `.go-version`/`toolchain`/CI → `go version`.
   Выполнить `go mod init <module>`. Проверить, что `module` корректен, а
   `go` в `go.mod` соответствует выбранной версии и не ниже `1.26`.
   Зависимости на этом шаге не добавлять. Если модуль короткий (не
   импортируется) — зафиксировать причину в `README.md` или `AGENTS.md`.
3. **Корневые файлы.** `.gitignore`, `README.md`,
   `configs/.env.example`.
   При коротком имени модуля — записать причину выбора в `README.md`.
4. **Платформа.** `internal/platform/logger/logger.go` — создание логгера.
5. **Конфиг.** `internal/config/env.go` — чтение env;
   `internal/config/config.go` — одно зеркало конфиг-файла: `Config` и
   `Load()`; локальные конфиги адаптеров вкладываются, а не дублируются.
   `config_test.go` — дефолты и ошибка на неверном значении. Локальный файл
   (`file.go`, флаг `-config`) добавляется, когда нужен запуск вне k8s.
   Правила источников, композиции и валидации — `skills/config`.
6. **Приложение.** `internal/app/<appname>/app.go` — точка сборки сценариев.
   Пока сценариев нет — минимальный `App` с `Run(ctx)`.
7. **Точка входа.** `cmd/<appname>/main.go`: конфиг → логгер → приложение →
   запуск и graceful shutdown по сигналу.
8. **Команды.** `Makefile`: `build`, `run`, `test`, `test-race`, `lint`.
9. **AGENTS.md.** Обзор проекта, команды из `Makefile`, правило «тесты по
   `skills/testing`», границы. Если модуль короткий — причина выбора имени.
   Если файл уже есть — слить, а не перезаписать.
10. **Проверка.** `go build ./...`, `go vet ./...`, `go test ./...`,
    `make build`.

## Минимальная реализация

Код ниже — эталон каркаса. Плейсхолдеры `<module>` и `<appname>` заменить,
импорты привести к фактическому имени модуля.

### internal/config/env.go

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

### internal/config/config.go

```go
package config

import (
	"fmt"
	"time"
)

// Config — зеркало конфиг-файла, общее для всех приложений.
// Локальные конфиги адаптеров добавляются сюда вложенными полями.
type Config struct {
	Env             string
	ShutdownTimeout time.Duration
}

// Load читает env и разбирает все секции.
func Load() (*Config, error) {
	timeout, err := Duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("shutdown timeout: %w", err)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("shutdown timeout must be positive")
	}

	return &Config{
		Env:             String("APP_ENV", "local"),
		ShutdownTimeout: timeout,
	}, nil
}
```

### internal/config/config_test.go

```go
package config

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
	appconfig "<module>/internal/config"
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
configs/*.local.*
.env
.env.*
!configs/.env.example
coverage.out
.idea/
.DS_Store
```

### configs/.env.example

```dotenv
# local | staging | production
APP_ENV=local
# Таймаут graceful shutdown
SHUTDOWN_TIMEOUT=10s
```

### README.md

Минимум: назначение проекта, требования (Go), запуск (`make run`), переменные
окружения со ссылкой на `configs/.env.example`, команды `make`.

### AGENTS.md

Минимум:

- имя проекта и модуль;
- если имя модуля короткое — почему выбрано короткое и что проект не
  импортируется другими;
- команды из `Makefile` без придуманных;
- правило: раскладка — `skills/project-structure`, зависимости —
  `skills/architecture`, тесты — `skills/testing`;
- границы: не создавать пустые каталоги, не добавлять зависимости без
  необходимости, не коммитить `configs/*.local.*` и `.env`.

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
- `internal/config` — один пакет на модуль, без подкаталогов по приложениям.
- `internal/domain` и другие слои отсутствуют, пока в них нет файлов.
- В `.git/config` записаны имя и email разработчика, глобальный конфиг не
  изменён.

## Что не делать

- Не создавать пустые каталоги «на будущее».
- Не добавлять `Dockerfile`, `docker-compose.yml`, миграции и CI, пока нет
  реальной потребности.
- Не создавать `internal/domain`, `internal/services`, `internal/repository`,
  `internal/infra`, `internal/transport`, `tests/`, `api/` без первого файла.
- Не добавлять сторонние зависимости в `go.mod` на старте: только stdlib.
- Не дублировать настройки адаптеров в общем конфиге: вкладывать локальный
  `Config` (`postgres.Config` и т.п.).
- Не делить `internal/config` по приложениям: конфиг один на модуль.
- Не помещать загрузку env/файла в `platform`: это делает `internal/config`.
- Не выбирать короткое имя модуля молча и не выводить его из имени каталога:
  короткое имя — только для неимпортируемых проектов и с записанной причиной.
- Не выводить `<appname>` из имени репозитория или модуля: `vp-appeal-doc-creator`
  даёт приложение `creator`.
- Не помещать бизнес-логику и SQL в `cmd/`.
- Не заводить `pkg/`, `util/`, `common/`, `helpers/`.
- Не задавать имя и email разработчика через `git config --global`.
- Не коммитить `.env*`, `configs/*.local.*` и `bin/`; коммитится только
  `configs/.env.example`.
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
