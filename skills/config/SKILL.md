---
name: config
description: Правила конфигурации Go-проекта: env в k8s и локальный файл по флагу, приоритет источников, общий Config на модуль, композиция из локальных Config адаптеров, валидация и секреты. Используется при добавлении настройки или адаптера. Раскладка — skills/project-structure, импорты — skills/architecture.
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: configuration
---

# Конфигурация Go-проекта

Навык отвечает на вопросы «**откуда читать настройки**», «**как собрать
Config**» и «**что и когда валидировать**». Вопрос «где лежит файл» — в
`skills/project-structure`; «что можно импортировать» — в
`skills/architecture`; «как тестировать» — в `skills/testing`.

## Главные правила

1. **Один конфиг на модуль.** Конфиг-файл и набор env общие для всех
   приложений. Деления `internal/config/<appname>` нет.
2. **Env — единый runtime-интерфейс.** В k8s значения приходят из
   ConfigMap/Secret; локально файл по флагу `-config` загружается в env.
   Дальше код работает одинаково.
3. **Схема — рядом с владельцем.** Поля и валидация адаптера живут в его
   пакете; общий `Config` их вкладывает, а не дублирует.
4. **Секреты — только из env/Secret.** Не в git, не в примерах, не в логах.

## Источники

| Где запущено | Источник | Кто выставляет |
|---|---|---|
| k8s | env из ConfigMap/Secret | манифесты |
| локально | файл по `-config` | `cmd`, до `config.Load()` |
| unit-тест | env | `t.Setenv` |

Приоритет значений — от сильного к слабому:

1. **Непустой env до `LoadFile`** — shell, k8s, `t.Setenv`. Файлом не
   перетирается.
2. **Значение из файла** — заполняет переменные, которых не было или которые
   были пустыми до загрузки.
3. **Дефолт в `Load()`** — если значение не пришло ни из env, ни из файла.

После `LoadFile` процесс видит уже объединённое окружение: `Load()` не
различает, откуда пришло значение. «Не перетираются» относится к моменту
**до** загрузки файла, а не к чтению после неё.

Пустая переменная (`KEY=`) считается незаданной: её может заполнить файл,
иначе сработает дефолт. Это согласовано с `env.go`, где пустое значение даёт
fallback.

- Секреты локально можно передать обычным env, не файлом.
- Формат файла выбирает проект (env, JSON, YAML). Требование одно: после
  загрузки имена переменных совпадают с k8s.

## Слои конфигурации

| Слой | Что содержит | Чего не содержит |
|---|---|---|
| `configs/` | `config.example.env`, локальные файлы | кода |
| `internal/config/` | `Config`, `Load()`, чтение env и файла | клиентов, бизнес-логики |
| `internal/infra/<adapter>/config.go` | локальный `Config`, `Validate()` | чтения env, импорта `config` |
| `cmd/<appname>/` | флаг `-config`, вызов `Load()`, валидация используемых секций, DI | дублирования полей |

Подробная раскладка каталогов — в `skills/project-structure`.

## Один Config на модуль

- `Config` — зеркало конфиг-файла: общие настройки плюс вложенные локальные
  конфиги адаптеров.
- `Load()` разбирает все секции, но **не требует** настройки неиспользуемых
  адаптеров. Приложение валидирует только то, что подключает.
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

// Load разбирает все секции. Обязательные поля проверяет потребитель.
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

Если приложению нужен свой набор, оно не заводит второй `Config`, а берёт
подмножество:

```go
cfg, err := config.Load()
if err != nil {
	return err
}
if err := cfg.Postgres.Validate(); err != nil {
	return fmt.Errorf("postgres config: %w", err)
}
repo := postgres.New(cfg.Postgres)
```

## Локальный Config адаптера

- Только поля и валидация. Env-теги, чтение `os.Getenv`, дефолты из env —
  не здесь.
- `New`/`NewClient` принимает `Config` аргументом.
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
- `LoadFile` только выставляет env и возвращает ошибку; `Load()` читает env;
- файл не перетирает **непустые** env, заданные **до** `LoadFile`;
- пустую переменную файл заполняет — она считается незаданной;
- путь и содержимое файла не логируются;
- сам файл с секретами — `configs/*.local.*` в `.gitignore`.

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
		if os.Getenv(key) != "" {
			continue // env, заданный до загрузки, сильнее файла
		}
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
- Каждая переменная есть в `configs/config.example.env` и описана в README.
- Секреты: `*_PASSWORD`, `*_TOKEN`, `*_KEY` — только из Secret/env, никогда
  из закоммиченного файла.

## Добавить настройку

1. Поле — в локальном `Config` владельца (адаптер или `internal/config`).
2. Значение — в `internal/config/Load()`.
3. Переменная — в `configs/config.example.env` и README.
4. Валидация — в `Validate()` владельца, вызов там, где настройка
   используется.
5. Тест — дефолт, неверное значение, при необходимости приоритет файла и env.
6. Если появился новый адаптер — локальный `Config`, `Validate()`, вложение
   в общий `Config`, передача в конструктор из `cmd`.

## Тесты

- `internal/config/..._test.go`: дефолты, неверные значения, приоритет
  источников.
- Env выставляется через `t.Setenv`, глобальное окружение не меняется.
- Тест загрузчика файла не должен зависеть от реальных `configs/*.local.*`:
  файл создаётся во временном каталоге.
- Подробно — в `skills/testing`.

## Антипаттерны

- `internal/config/<appname>` и отдельный конфиг на каждое приложение.
- Дублировать `DB_DSN`, `MaxConns` и другие поля адаптера в общем конфиге.
- Читать env в адаптерах, `platform` или `app`.
- Валидировать в `Load()` секции, которые приложение не использует.
- Загружать конфиг в `init()` или глобальную переменную.
- Коммитить реальный конфиг, печатать секреты в лог.
- Перетирать непустые env, заданные до загрузки, локальным файлом.
- Добавлять префикс приложения к env без реальной необходимости.

## Границы

- «Где лежит файл» — `skills/project-structure`.
- «Что можно импортировать» — `skills/architecture`.
- «Как тестировать» — `skills/testing`.
- «С чего начать проект» — `skills/bootstrap`.
