---
name: logger
description: "Правила логгера Go-проекта на log/slog: internal/platform/logger, logger.Config, уровни, JSON и text, общие атрибуты сервиса, передача логгера в приложение и тесты. Используется при добавлении логирования, новой настройки логов или разборе шумных логов. Откуда брать настройки — skills/config, где лежит файл — skills/project-structure."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: configuration
---

# Логгер

Навык отвечает на вопросы «**где создать логгер**», «**как его настроить**» и
«**как писать сообщения**». Вопрос «откуда берутся значения» — в
`skills/config`; «в каком каталоге лежит пакет» — в
`skills/project-structure`; «как тестировать» — в `skills/testing`.

## Главные правила

1. **Только `log/slog` из стандартной библиотеки.** Сторонние логгеры не
   добавляются: `slog` уже даёт уровни, JSON, `context` и общие атрибуты.
2. **Один логгер на приложение.** Создаётся в `cmd/<appname>/main.go` и
   передаётся аргументом в `app.New`, дальше — в сервисы и адаптеры.
3. **`platform/logger` не читает env и не импортирует проект.** Настройки
   приходят через `logger.Config`; env разбирает `internal/config`.
4. **По умолчанию — JSON в stdout.** Text включается локально для чтения
   глазами. Контейнер сам собирает stdout, файлы логов приложение не пишет.
5. **Общие атрибуты — один раз на handler.** `service`, `version`, `env`
   добавляются в `New`, а не в каждом сообщении.
6. **Сообщение — стабильное, данные — в атрибутах.** Сообщение ищут в логах,
   поэтому оно не должно меняться от запроса к запросу.
7. **В логи не попадают секреты и персональные данные.** Токены, пароли,
   `Authorization`, тела запросов целиком и PII — не логируются.

## Где лежит

| Что | Куда |
|---|---|
| Создание логгера | `internal/platform/logger/logger.go` |
| Тест | `internal/platform/logger/logger_test.go` |
| Настройки в общем конфиге | `internal/config/config.go` |
| Переменные окружения | `configs/.env.example`, README |

Готовые файлы — `references/logger.go` и `references/logger_test.go` —
копируются в `internal/platform/logger/`.

Пакет `internal/platform/logger` — инфраструктурный примитив. Он не знает ни
о домене, ни о конфиге приложения.

## Config

```go
type Config struct {
	Level       string // debug | info | warn | warning | error
	Format      string // json | text
	Output      string // stdout | stderr
	ServiceName string
	Version     string
	Environment string
}
```

| Поле | Значения | По умолчанию в `New` |
|---|---|---|
| `Level` | `debug`, `info`, `warn`/`warning`, `error` | `info` |
| `Format` | `json`, `text` | `json` |
| `Output` | `stdout`, `stderr` | `stdout` |
| `ServiceName` | имя приложения, например `api` | пусто |
| `Version` | версия продукта или сборки, например `1.4.2` или `dev` | пусто |
| `Environment` | `local`, `staging`, `production` | пусто |

`New` сам подставляет только уровень, формат и поток. `SERVICE_*` приходят
из конфига: пустое значение останется пустым атрибутом, поэтому в k8s его
задают явно. Дефолты приложения — в `internal/config`.

Неверное значение не ломает запуск: уровень становится `info`, формат —
`json`, поток — `stdout`. Для логгера это осознанное исключение из правила
«валидировать в `New`»: логирование не должно быть причиной падения
сервиса. Если проект требует строгой проверки, она делается в
`internal/config.Load()` вместе с остальными настройками.

## Создание

`logger.New` собирает handler, выбирает поток и добавляет общие атрибуты.
`AddSource` включается только на `debug`: путь к файлу и строка дороги и
раздувают каждый лог.

Полная реализация — `references/logger.go`. Ключевое:

```go
func New(cfg Config) *slog.Logger {
	return newLogger(cfg, parseOutput(cfg.Output))
}
```

`newLogger(cfg, output io.Writer)` — точка подмены для тестов: тест пишет в
`bytes.Buffer`, а не перехватывает `os.Stdout`.

## Настройки в конфиге

`internal/config` вкладывает `logger.Config`, как локальный конфиг адаптера:
поля живут в пакете логгера, общий `Config` их не дублирует. Это
разрешённый импорт `config → platform/logger`, но только ради `Config`:
сам `platform/logger` не импортирует `config` и не читает env.

```go
// internal/config/config.go
import "<module>/internal/platform/logger"

type Config struct {
	Env             string
	ShutdownTimeout time.Duration
	Logger          logger.Config
}

func Load() (*Config, error) {
	env := String("APP_ENV", "local")

	return &Config{
		Env: env,
		Logger: logger.Config{
			Level:       String("LOG_LEVEL", "info"),
			Format:      String("LOG_FORMAT", "json"),
			Output:      String("LOG_OUTPUT", "stdout"),
			ServiceName: String("SERVICE_NAME", "<appname>"),
			Version:     String("SERVICE_VERSION", "dev"),
			Environment: String("SERVICE_ENVIRONMENT", ""),
		},
	}, nil
}
```

```dotenv
# configs/.env.example
SERVICE_NAME=goskills
SERVICE_VERSION=1.0.0
SERVICE_ENVIRONMENT=dev
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_OUTPUT=stdout
```

| Переменная | Что означает | Значения |
|---|---|---|
| `SERVICE_NAME` | имя сервиса в общих атрибутах | `api`, `worker` |
| `SERVICE_VERSION` | версия продукта или сборки | `1.0.0`, `dev` |
| `SERVICE_ENVIRONMENT` | стенд, на котором работает сервис | `dev`, `staging`, `production` |
| `LOG_LEVEL` | порог логирования | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | формат вывода | `json`, `text` |
| `LOG_OUTPUT` | поток вывода | `stdout`, `stderr` |

`SERVICE_*` — метаданные сервиса, `LOG_*` — настройки логгера. Это разные
группы, и связывать их не нужно: `SERVICE_ENVIRONMENT` — только атрибут
`env` в логе, `APP_ENV` (если он есть) — поведение приложения. Если оба
означают один и тот же стенд, задавайте одинаковые значения в
`configs/.env.example`.

`Environment` по умолчанию пустой, а не `local`: пустой атрибут `env`
виден в логе и сигналит, что переменную не задали. Дефолт `local`
замаскировал бы стенд под локальную машину.

### Почему `SERVICE_ENVIRONMENT` — строка, а не enum

- **Это метаданные, а не поведение.** По `env` в логе нет ветвлений;
  `SERVICE_ENVIRONMENT` — только атрибут, `APP_ENV` — поведение приложения.
  Типизация метаданных ничего не защищает.
- **Логгер не должен падать на неизвестном значении.** Enum либо
  валидируется и роняет запуск, либо молча мапится — и тогда защиты нет.
  Правило логгера: неверный уровень/формат откатывается к дефолту.
- **Набор стендов открыт и принадлежит эксплуатации.** `dev`, `preprod`,
  `sandbox`, `dev2` появляются без изменения кода. Enum заставил бы править
  Go ради нового стенда, хотя значение всё равно приходит строкой из
  ConfigMap.

Имена переменных — плоские, без префикса приложения: настройки логов общие
для всех приложений модуля. Правила источников и `Load()` —
`skills/config`.

## Подключение в main

Полный файл — `skills/bootstrap/references/main.go`; ниже — только
логгерная часть.

```go
// cmd/<appname>/main.go, внутри run()
log := logger.New(cfg.Logger)
slog.SetDefault(log)
```

- `slog.SetDefault` нужен только для сторонних библиотек, которые пишут
  через глобальный `slog`.
- Свой код получает `*slog.Logger` аргументом и не обращается к
  `slog.Default()`.
- До создания логгера ошибку загрузки конфига печатают через
  `slog.Error`/`fmt.Fprintln(os.Stderr, ...)` и завершают процесс.

## Использование

```go
// internal/services/api/user/service.go
type Service struct {
	repo Repository
	log  *slog.Logger
}

func New(repo Repository, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log.With("component", "user"),
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	s.log.InfoContext(ctx, "register user", "user_id", req.UserID)

	user, err := s.repo.Create(ctx, req.UserID)
	if err != nil {
		// Ошибку логирует граница операции — хендлер или main.
		return fmt.Errorf("create user: %w", err)
	}

	s.log.DebugContext(ctx, "user created", "user_id", user.ID)
	return nil
}
```

- `ctx`-варианты (`InfoContext`, `DebugContext`, `ErrorContext`) сохраняют
  `trace_id` и `request_id`, если они лежат в контексте.
- Ошибка возвращается наверх как есть, с `%w`. Логирует её один раз граница
  операции, а не каждый слой по пути.
- `With("component", ...)` — дочерний логгер с постоянным атрибутом. Так
  видно, какой сервис или адаптер пишет сообщение.
- `With` создаётся один раз в конструкторе, а не на каждый вызов.

## Что и на каком уровне писать

| Уровень | Когда | Пример |
|---|---|---|
| `debug` | диагностика для разработчика, подробности ветвлений | `cache miss`, `query plan` |
| `info` | значимое событие жизненного цикла | `started`, `stopped`, `user registered` |
| `warn` | неожиданно, но работа продолжается | `retry attempt`, `fallback to default` |
| `error` | операция не выполнена, нужно внимание | `create order`, `send message` |

Правила:

- Сообщение — короткое, в нижнем регистре, без точки: `create order`,
  `load config`. Не подставлять в него данные: `"create order 123"` плохо,
  `"create order", "order_id", 123` хорошо.
- Ключи атрибутов — `snake_case`, одинаковые во всём проекте: `user_id`,
  `order_id`, `request_id`, `trace_id`, `duration_ms`.
- Ошибка передаётся атрибутом `"error", err`, а не склеивается в сообщение.
- Ошибку логируют **один раз**, на границе операции (хендлер, consumer,
  `main`). Внутри — только возвращают с `%w`, иначе один сбой даёт несколько
  одинаковых записей.
- Ожидаемая бизнес-ошибка (валидация, «не найдено») — не `error`, а `info`
  или `warn` на границе.
- `Fatal` в `slog` нет. Завершение процесса — только в `main`, через
  `os.Exit` после записи ошибки.
- Не логировать в цикле на каждый элемент, если это не `debug` и не
  ограниченный объём.

## Формат и поток

- **production / k8s** — JSON в stdout: так лог читает сборщик, а не
  человек. Одна строка — одно событие.
- **local** — `LOG_FORMAT=text`: удобно читать в терминале. Набор полей тот
  же.
- `stderr` — только если платформа разделяет потоки. По умолчанию всё в
  `stdout`.
- Файлы, ротация и несколько назначений — задача сборщика логов, а не
  приложения. `io.MultiWriter` добавляется, только если это явное
  требование.

## Тесты

Готовый пример — `references/logger_test.go`. Проверяется поведение, а не
внутренности; исключение — чистые функции и тестовые швы:

- `parseLevel` — все уровни, регистр, `warning`, неизвестное значение;
- JSON содержит `service`, `version`, `env`, `msg` и атрибуты события;
- уровень фильтрует: `info` не печатает `debug`;
- `debug` добавляет `source`, остальные уровни — нет;
- `text` содержит `level=INFO` и общие атрибуты;
- неизвестный формат даёт JSON, неизвестный уровень — `info`.

Тест пишет в `bytes.Buffer` через `newLogger`, не перехватывает `os.Stdout`
и не зависит от env. `parseLevel` — чистая функция, `newLogger` — шов с
подменой writer; их тесты не нарушают правило про внутренности
(`skills/testing`).

## Добавить настройку логгера

1. Поле в `logger.Config` и разбор в `New` (или в `newLogger`).
2. `internal/config.Load()` — чтение env с дефолтом.
3. `configs/.env.example` и README — новая переменная с комментарием.
4. `logger_test.go` — поведение при новом значении и при неизвестном.
5. Если поле меняет формат вывода — проверить, что JSON остаётся одной
   строкой на событие.

## Антипаттерны

- Сторонний логгер вместо `log/slog`.
- Глобальная переменная `var log = ...` или создание логгера в `init()`.
- Чтение env внутри `platform/logger`.
- Импорт `internal/config` из `platform/logger`.
- Создание логгера на каждый запрос или в каждом методе.
- `fmt.Printf`, `println`, `log.Print` рядом со `slog`.
- `log.Fatal`/`os.Exit` в сервисе, адаптере или библиотеке.
- Логировать ошибку на каждом уровне вызова и потом ещё в `main`.
- Склеивать данные в сообщение вместо атрибутов.
- Разные имена одного и того же поля: `userID`, `user_id`, `uid`.
- Токены, пароли, `Authorization`, cookie, полные тела запросов и PII в
  логах.
- `text` в production и `json` в локальной разработке «по привычке».
- `AddSource` на всех уровнях.
- `debug`-логи на каждый элемент большого цикла.
- Писать логи в файл внутри приложения.

## Проверка

```bash
go build ./...
go vet ./...
go test ./...
```

- Логгер создаётся один раз в `cmd/<appname>/main.go`.
- `go list -f '{{join .Imports "\n"}}' ./internal/platform/logger` не
  содержит пакетов проекта — только стандартную библиотеку.
- В JSON-логе есть `service`, `version`, `env`.
- `LOG_LEVEL=info` не печатает `debug`; `LOG_LEVEL=debug` добавляет
  `source`.
- В логах нет секретов и PII.
- Все переменные логгера есть в `configs/.env.example`.
- `go test ./internal/platform/logger` проходит без сети и `time.Sleep`.

## Границы

- «Откуда брать настройки» — `skills/config`.
- «В каком каталоге лежит пакет» — `skills/project-structure`.
- «Что можно импортировать» — `skills/architecture`.
- «Как тестировать» — `skills/testing`.
- «Полный `main.go` и порядок запуска» — `skills/bootstrap`.
- «Что не коммитить и что не читает агент» — `skills/gitignore`,
  `skills/agentignore`.
