---
name: config
description: "Правила конфигурации Go-проекта: env в k8s и локальный env-файл по флагу, общий Config на модуль, композиция из локальных Config адаптеров и платформы, валидация в конструкторе и секреты. Используется при добавлении настройки или адаптера. Раскладка — skills/project-structure, импорты — skills/architecture."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: configuration
---

# Конфигурация Go-проекта

Навык отвечает на вопросы «**откуда читать настройки**», «**как собрать
Config**» и «**где валидировать**». Вопрос «где лежит файл» — в
`skills/project-structure`; «что можно импортировать» — в
`skills/architecture`; «как тестировать» — в `skills/testing`.

## Главные правила

1. **Один конфиг на модуль.** Конфиг-файл и набор env общие для всех
   приложений. Деления `internal/config/<appname>` нет.
2. **Источник не смешивается.** В k8s — только env из ConfigMap/Secret. При
   локальном запуске — только env-файл, который загружается в окружение.
3. **Схема — рядом с владельцем.** Поля адаптера или платформенного пакета
   живут в его пакете (например, `logger.Config`); общий `Config` их
   вкладывает, а не дублирует.
4. **Валидация — в конструкторе.** `New` адаптера проверяет свой `Config`.
   `Load()` и DI не валидируют секции, которые приложение не использует.
5. **Секреты — только из env/Secret.** Не в git, не в примерах, не в логах.

## Источники

| Где запущено | Источник | Правило |
|---|---|---|
| k8s | env из ConfigMap/Secret | файл не используется |
| локально | env-файл по флагу `-config` | файл перетирает окружение |
| unit-тест | env через `t.Setenv` | файл не используется |

- Файл загружается в окружение **в лоб**: все его переменные выставляются
  через `os.Setenv`, уже существующие значения перетираются.
- Проверок «была ли переменная задана раньше» нет. Один запуск — один
  источник.
- После загрузки файла `Load()` читает готовое окружение и не знает, откуда
  пришло значение.
- В k8s не подкладывается env-файл: секции с секретами и env-переменные
  одновременно не смешиваются.
- Формат файла выбирает проект (env, JSON, YAML). Требование одно: после
  загрузки имена переменных совпадают с k8s.

## Слои конфигурации

| Слой | Что содержит | Чего не содержит |
|---|---|---|
| `configs/` | `.env.example`, локальные `.env.*` | кода |
| `internal/config/` | `Config`, `Load()`, `LoadFile()` | клиентов, бизнес-логики |
| `internal/infra/<adapter>/config.go` | локальный `Config`, `Validate()` | чтения env, импорта `config` |
| `internal/platform/<pkg>/` | локальный `Config` (например, `logger.Config`) | чтения env, импорта `config` |
| `internal/app/<appname>/app.go` | DI: адаптеры и сервисы из `Config` | дублирования полей |
| `cmd/<appname>/` | флаг `-config`, `Load()`, логгер, `app.New`, сигналы | создания адаптеров |

Подробная раскладка каталогов — в `skills/project-structure`; границы
импортов — в `skills/architecture`.

## Один Config на модуль

- `Config` — зеркало конфиг-файла: общие настройки плюс вложенные локальные
  конфиги адаптеров и платформы.
- `Load()` разбирает все секции, но не проверяет обязательные поля
  адаптеров. Если приложение не использует Postgres, `DB_DSN` может быть
  пустым — приложение всё равно запустится.
- Приложений может быть несколько: каждое берёт из общего `Config` свои
  секции. Отдельных конфигов на приложение нет.

```go
// internal/config/config.go — одно зеркало конфиг-файла на весь модуль
package config

import (
	"fmt"
	"time"

	"<module>/internal/infra/postgres"
)

type Config struct {
	Env             string
	ShutdownTimeout time.Duration
	Postgres        postgres.Config
}

// Load разбирает все секции. Обязательные поля проверяют конструкторы.
func Load() (*Config, error) {
	timeout, err := Duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("shutdown timeout: %w", err)
	}
	maxConns, err := Int("DB_MAX_CONNS", 10)
	if err != nil {
		return nil, fmt.Errorf("DB_MAX_CONNS: %w", err)
	}

	return &Config{
		Env:             String("APP_ENV", "local"),
		ShutdownTimeout: timeout,
		Postgres: postgres.Config{
			DSN:      String("DB_DSN", ""),
			MaxConns: maxConns,
		},
	}, nil
}
```

## Локальный Config адаптера

- Только поля и валидация. Env-теги, чтение `os.Getenv`, дефолты из env —
  не здесь.
- `New` принимает `Config` аргументом и сам вызывает `Validate()`.
- Адаптер не импортирует `internal/config`: направление обратное.

```go
// internal/infra/postgres/config.go
package postgres

import "fmt"

type Config struct {
	DSN      string
	MaxConns int
}

func (c Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("postgres: DSN is required")
	}
	if c.MaxConns <= 0 {
		return fmt.Errorf("postgres: MaxConns must be positive")
	}
	return nil
}
```

```go
// internal/infra/postgres/postgres.go
func New(cfg Config) (*Pool, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Pool{ /* ... */ }, nil
}
```

Если приложение не вызывает `postgres.New`, его конфиг не проверяется и не
мешает запуску.

## Локальный Config платформы

- Настройки платформенного пакета живут в нём же: например, `logger.Config`
  в `internal/platform/logger`.
- `internal/config` вкладывает его полем: `Logger logger.Config`. Это
  разрешённый импорт `config → platform`: только ради `Config`.
- Пакет платформы не импортирует `internal/config` и не читает env.
- `logger.Config` не валидируется: неверные значения заменяются дефолтами,
  чтобы логирование не роняло сервис. Это осознанное исключение; подробности —
  `skills/logger`.

## DI приложения

`internal/app/<appname>/app.go` — composition root приложения: создаёт
адаптеры и сервисы, запускает и останавливает приложение.

```go
// internal/app/creator/app.go
type App struct {
	svc *report.Service
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	repo := postgres.NewRepo(pg)
	svc := report.New(repo)

	return &App{svc: svc, log: log}, nil
}

func (a *App) Run(ctx context.Context) error {
	// запуск и останов по ctx
}
```

`cmd/<appname>/main.go` не создаёт адаптеры. Он делает только:

1. разбирает `-config` и вызывает `config.LoadFile` / `config.Load`;
2. создаёт логгер;
3. вызывает `app.New(cfg, log)`;
4. ловит сигналы ОС и вызывает `Run`/останов.

```go
// cmd/creator/main.go
cfg, err := config.Load()
log := logger.New(cfg.Env)

a, err := app.New(cfg, log)
if err != nil {
	return fmt.Errorf("new app: %w", err)
}

ctx, stop := signal.NotifyContext(context.Background(),
	os.Interrupt, syscall.SIGTERM)
defer stop()

if err := a.Run(ctx); err != nil {
	return fmt.Errorf("run app: %w", err)
}
```

## Локальный запуск

`cmd/<appname>` разбирает только флаг и отдаёт путь загрузчику:

```go
var configPath string
flag.StringVar(&configPath, "config", "", "path to config file (local run)")
flag.Parse()

if err := config.LoadFile(configPath); err != nil {
	return fmt.Errorf("load config file: %w", err)
}
cfg, err := config.Load()
```

- `configPath == ""` — только env, как в k8s;
- `LoadFile` выставляет все переменные из файла через `os.Setenv`;
- существующие env перетираются — это ожидаемо;
- `Load()` читает уже готовое окружение;
- путь и содержимое файла не логируются;
- сам файл с секретами (`configs/.env.local`) — в `.gitignore`.

```go
// internal/config/file.go — формат разбирает parseFile
func LoadFile(path string) error {
	if path == "" {
		return nil
	}
	values, err := parseFile(path)
	if err != nil {
		return err
	}
	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}
```

## Имена переменных

- Плоские, верхний регистр, `_`: `APP_ENV`, `SHUTDOWN_TIMEOUT`, `DB_DSN`,
  `KAFKA_BROKERS`.
- Имя отражает подсистему, а не приложение: общий `DB_DSN` для всех
  приложений модуля. Префикс приложения (`CREATOR_DB_DSN`) — только если
  приложения работают с разными значениями в одном окружении.
- Каждая переменная есть в `configs/.env.example` и описана в README.
- Секреты: `*_PASSWORD`, `*_TOKEN`, `*_KEY` — только из Secret/env, никогда
  из закоммиченного файла.

## Добавить пакет с конфигурацией

Пакету нужны настройки — добавляй их по всей цепочке сразу, не оставляя
«на потом»:

1. `configs/.env.example` — переменные с комментариями и дефолтами.
   Это единственный шаблон для всех приложений.
2. `internal/infra/<adapter>/config.go` — локальный `Config` и `Validate()`.
3. `internal/config/config.go` — поле в общем `Config` и разбор в `Load()`.
4. `internal/app/<appname>/app.go` — создать адаптер с локальным конфигом:
   `postgres.New(cfg.Postgres)`. Валидация уже внутри `New`.
5. `internal/config/config_test.go` и тест адаптера — дефолты и неверные
   значения.

Если пакет требует бинарь (`goose` для Postgres, `swag` для HTTP, `protoc`
для gRPC, `mockery` для моков) — обновить `.bin-deps` в `Makefile` в том же
изменении. Правила — `skills/makefile`.

Пример: появился Postgres.

```dotenv
# configs/.env.example
DB_DSN=postgres://user:pass@localhost:5432/app?sslmode=disable
DB_MAX_CONNS=10
```

```go
// internal/config/config.go
type Config struct {
	Env      string
	Postgres postgres.Config
}

func Load() (*Config, error) {
	maxConns, err := Int("DB_MAX_CONNS", 10)
	if err != nil {
		return nil, fmt.Errorf("DB_MAX_CONNS: %w", err)
	}

	return &Config{
		Env: String("APP_ENV", "local"),
		Postgres: postgres.Config{
			DSN:      String("DB_DSN", ""),
			MaxConns: maxConns,
		},
	}, nil
}
```

```go
// internal/app/creator/app.go — DI приложения
func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	pg, err := postgres.New(cfg.Postgres) // Validate внутри
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	repo := postgres.NewRepo(pg)
	svc := report.New(repo)

	return &App{svc: svc, log: log}, nil
}
```

Если приложение не использует Postgres, шаг 4 для него не выполняется:
`postgres.New` не вызывается, `DB_DSN` не проверяется.

## Тесты

- `internal/config/..._test.go`: дефолты, неверные значения.
- `internal/infra/<adapter>/..._test.go`: `New` отклоняет неверный `Config`.
- Загрузчик файла проверяется отдельно: файл перетирает уже заданную env.
- Env выставляется через `t.Setenv`, глобальное окружение не меняется.
- Тест загрузчика не зависит от реальных `configs/.env.*`: файл создаётся во
  временном каталоге.
- Подробно — в `skills/testing`.

## Антипаттерны

- `internal/config/<appname>` и отдельный конфиг на каждое приложение.
- Дублировать `DB_DSN`, `MaxConns`, `LOG_LEVEL` и другие поля адаптера или
  платформы в общем конфиге.
- Читать env в адаптерах или `platform`.
- Валидировать в `Load()` или DI секции, которые приложение не использует.
- Вызывать `cfg.Postgres.Validate()` в обход `New`.
- Создавать адаптеры в `cmd/<appname>` вместо `internal/app/<appname>/app.go`.
- Смешивать в k8s env-файл и env-переменные.
- Проверять «была ли переменная задана раньше» при загрузке файла.
- Загружать конфиг в `init()` или глобальную переменную.
- Коммитить реальный конфиг, печатать секреты в лог.
- Добавлять пакет с настройками, не обновив `configs/.env.example` и
  `internal/config`.
- Добавлять префикс приложения к env без реальной необходимости.

## Границы

- «Где лежит файл» — `skills/project-structure`.
- «Что можно импортировать» — `skills/architecture`.
- «Как тестировать» — `skills/testing`.
- «С чего начать проект» — `skills/bootstrap`.
