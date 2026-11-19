---
name: bootstrap
description: "Создание нового Go-проекта с нуля после подключения навыков goskills — инициализация модуля, минимальный каркас и обязательный набор каталогов. Используется на старте проекта. Раскладка — skills/project-structure, зависимости — skills/architecture, конфиг — skills/config."
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
`skills/architecture`, конфигурацию — `skills/config`, unit-тесты —
`skills/testing`, интеграционные — `skills/integration-testing`, e2e —
`skills/e2e-testing`.

## Когда применять

- Каталог проекта пуст или содержит только `.git`.
- `go.mod` отсутствует.
- Навыки goskills уже подключены к OpenCode (см. README репозитория goskills).

Если `go.mod` уже есть — каркас не пересоздаётся. Привести проект по
`skills/project-structure`, раздел «Проект создан без навыков»; границы
импортов — `skills/architecture`.

## Входные данные

До создания файлов выяснить:

| Что | Как определить |
|---|---|
| Имя модуля | Путь импорта (`module`), не имя проекта. Кандидат — из git remote: `git@github.com:org/repo.git` → `github.com/org/repo`. Спросить, будет ли проект импортироваться другими; если нет — предложить короткое имя и получить явное согласие. Без remote — `go.work` или родительский `go.mod`; иначе **спросить пользователя** |
| Имя приложения `<appname>` | Спросить пользователя, если он не указал его явно. Выбирается по бинарю: короткое, без доменного префикса и без имени репозитория — `creator`, `api`. Это имя бинаря и одинаковый сегмент в `cmd/`, `internal/app/` и `internal/transport/` |
| Версия Go | явно указанная пользователем; иначе `.go-version`, `toolchain` или CI-конфиг; иначе системная из `go version`. Если системная ниже `1.26` — уточнить у пользователя |
| Тестовый стек | `mockery` v3 + `testify` по умолчанию. `stdlib` и ручные моки — только если пользователь явно попросил. `mockery` → `.bin-deps`, `testify` → `go.mod`, `.mockery.yaml` в корень |
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
- `mockery` v3 + `testify` — тестовый стек по умолчанию, а не поздняя правка:
  `mockery` идёт в `.bin-deps`, `testify` — в `go.mod`, моки генерируются,
  а не пишутся руками. `stdlib` и ручные проверки — только по явному
  запросу пользователя.
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
├── .agentignore
├── README.md
├── Makefile
├── .mockery.yaml
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
    │   ├── file.go
    │   ├── config.go
    │   └── config_test.go
    └── platform/
        └── logger/
            ├── logger.go
            └── logger_test.go
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
3. **Корневые файлы.** `.gitignore`, `.agentignore`, `README.md`,
   `configs/.env.example`.
   При коротком имени модуля — записать причину выбора в `README.md`.
4. **Платформа.** `internal/platform/logger/` — создание логгера по
   `skills/logger`: `references/logger.go` и `references/logger_test.go`
   копируются в пакет. Логгер — `slog`, настройки приходят через
   `logger.Config`, env читает `internal/config`.
5. **Конфиг.** `internal/config/env.go` — чтение env;
   `internal/config/config.go` — одно зеркало конфиг-файла: `Config` и
   `Load()`; локальные конфиги адаптеров и платформы вкладываются, а не
   дублируются.
   `config_test.go` — дефолты и ошибка на неверном значении.
   `internal/config/file.go` — по `skills/config/references/file.go`:
   локальный env-файл по `-config`, пустой путь ничего не делает.
   По умолчанию — `go get github.com/stretchr/testify`, тест конфига
   пишется на `require`. `stdlib` без testify — только по явному запросу.
   Правила источников, композиции и валидации — `skills/config`.
6. **Приложение.** `internal/app/<appname>/app.go` — DI приложения: создаёт
   адаптеры, сервисы и транспорт из `Config`, запускает и останавливает.
   Пока зависимостей нет — минимальный `App`, `New(cfg, log)` и `Run(ctx)`.
   Импорт `infra`, `transport`, `config`, `platform` — только в `app.go`.
7. **Точка входа.** `cmd/<appname>/main.go`: скопировать
   `references/main.go` и заменить `<module>`/`<appname>`. Конфиг → логгер →
   `app.New` → сигналы ОС → `Run` и останов. Адаптеры здесь не создаются.
8. **Команды.** `Makefile` по `skills/makefile`: `build-<app>`, `run-<app>`,
   `build-all`, `test`, `test-race`, `lint`, `help`, `.bin-deps`, `mocks`.
   `mockery` ставится через `.bin-deps`; `.mockery.yaml` — в корне по
   `references/.mockery.yaml`.
9. **AGENTS.md.** По `skills/agents-md`: взять `references/AGENTS.md`,
   записать модуль, приложения, короткое имя и причину, команды, конфиг,
   тестовый стек и границы. Если файл уже есть — слить, а не перезаписать.
10. **Проверка.** `go build ./...`, `go vet ./...`, `go test ./...`,
    `make build-all`.

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

### internal/platform/logger/

Создаётся по `skills/logger`: скопировать `skills/logger/references/logger.go`
и `skills/logger/references/logger_test.go` в `internal/platform/logger/`.
В каркасе — JSON в stdout, уровень `info`, общие атрибуты `service`,
`version`, `env`. `platform/logger` не читает env и не импортирует проект.

### internal/config/config.go

```go
package config

import (
	"fmt"
	"time"

	"<module>/internal/platform/logger"
)

// Config — зеркало конфиг-файла, общее для всех приложений.
// Локальные конфиги адаптеров и платформы добавляются сюда вложенными полями.
type Config struct {
	Env             string
	ShutdownTimeout time.Duration
	Logger          logger.Config
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

	env := String("APP_ENV", "local")

	return &Config{
		Env:             env,
		ShutdownTimeout: timeout,
		Logger: logger.Config{
			Level:       String("LOG_LEVEL", "info"),
			Format:      String("LOG_FORMAT", "json"),
			Output:      String("LOG_OUTPUT", "stdout"),
			ServiceName: String("SERVICE_NAME", "<appname>"),
			Version:     String("SERVICE_VERSION", "dev"),
			Environment: String("SERVICE_ENVIRONMENT", "local"),
		},
	}, nil
}
```

### internal/config/config_test.go

По умолчанию — `testify/require`. Если пользователь явно отказался от
`testify`, заменить `require` на `if`/`t.Fatalf`.

```go
package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "local", cfg.Env)
	require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
}

func TestLoad_RejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "soon")

	_, err := Load()
	require.Error(t, err)
}
```

### internal/app/<appname>/app.go

```go
package <appname>

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"<module>/internal/config"
)

// App — DI приложения: адаптеры, сервисы, запуск и останов.
type App struct {
	log             *slog.Logger
	shutdownTimeout time.Duration
	wg              sync.WaitGroup
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	return &App{
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}, nil
}

// Run работает до отмены контекста и завершает компоненты с таймаутом.
func (a *App) Run(ctx context.Context) error {
	// здесь запускаются компоненты и добавляются в a.wg

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	return a.stop(shutdownCtx)
}

// stop останавливает компоненты и ждёт их горутины, но не дольше shutdownCtx.
func (a *App) stop(ctx context.Context) error {
	// если компоненту нужен явный Stop, вызови его здесь до ожидания
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stop: %w", ctx.Err())
	}
}
```

`Run` — точка graceful shutdown: после `<-ctx.Done()` он останавливает
компоненты с таймаутом и ждёт их горутины. `WaitGroup` принадлежит `App`
или компоненту, а не `main`: `main` вызывает `Run(ctx)` и не делает
`wg.Wait()`.

Порядок зависит от компонента: если он выходит по `ctx`, достаточно
дождаться; если его будит `Stop`, вызови `Stop` до `wg.Wait`.

### cmd/<appname>/main.go

Готовый файл — `references/main.go`. Скопировать в `cmd/<appname>/main.go` и
заменить `<module>` и `<appname>`.

Структура:

1. `main` — разбирает `-config` (пустой по умолчанию), вызывает `run(path)`
   и логирует ошибку один раз.
2. `run(path string) error` — загружает файл и env, создаёт логгер, собирает
   приложение и запускает `Run`.
3. `signal.NotifyContext` даёт контекст для `app.Run`; отдельный обработчик
   сигналов и `WaitGroup` в `main` не нужны.
4. Ошибки возвращаются с контекстом (`load config: %w`), логируются один раз
   в `main`.

Логи: сообщение стабильное и строчное, данные — в атрибутах. Ошибку
логирует один раз `main`: `slog.Error("service failed", "error", err)`.
Имя сервиса уже в `logger.Config`, поэтому префикс вида `RD-Hub:` и
`fmt.Sprintf` в сообщении не нужны.

Базовый вариант — `Run(ctx)`: приложение само ждёт отмены контекста и
завершает работу. Отдельные `Start`/`Stop` добавляй только если они
действительно нужны: иначе сигналы и остановка расползаются по `main`.

`-config` по умолчанию пустой: в k8s файла нет. Локально `make run-<app>`
передаёт `configs/.env.local`.

### Makefile

Создаётся по `skills/makefile`: `skills/makefile/references/Makefile` →
корень проекта. Один файл на модуль, заменить `APPS` на приложения.
Обязательный минимум: `build-<app>`, `run-<app>`, `build-all`, `test`,
`test-race`, `lint`, `help`. `run-<app>` передаёт
`-config configs/.env.local`; `bin/` не коммитится.

### .mockery.yaml

Создаётся по `references/.mockery.yaml`: mockery v3, шаблон `testify`, моки
рядом с интерфейсами. По умолчанию рекурсивно обходит `<module>/internal` и
запускается даже на пустом каркасе. Заменить `<module>`; когда моки нужны не
везде — сузить `packages` до конкретных import path. Версия фиксируется в
`Makefile` (`MOCKERY_VERSION`). Генерация — `make mocks`.

### .gitignore

Создаётся по `skills/gitignore`: `skills/gitignore/references/.gitignore` →
корень проекта. Минимум: `bin/`, `configs/.env.*`, `.env*`, IDE/OS;
`configs/.env.example` коммитится.

### .agentignore

Создаётся по `skills/agentignore`: `skills/agentignore/references/.agentignore`
→ корень проекта. Закрывает секреты, выгрузки, артефакты и IDE-файлы от
чтения агентом. Правила секретов дублируются в `.gitignore`.

### configs/.env.example

```dotenv
# local | staging | production
APP_ENV=local
# Таймаут graceful shutdown
SHUTDOWN_TIMEOUT=10s
# debug | info | warn | error
LOG_LEVEL=info
# json | text
LOG_FORMAT=json
# stdout | stderr
LOG_OUTPUT=stdout
# Имя сервиса, версия продукта и стенд в общих атрибутах лога
SERVICE_NAME=<appname>
SERVICE_VERSION=dev
SERVICE_ENVIRONMENT=local
```

### README.md

Минимум: назначение проекта, требования (Go), запуск (`make run-<app>`),
переменные окружения со ссылкой на `configs/.env.example`, команды `make`.

### AGENTS.md

Создаётся по `skills/agents-md`: `references/AGENTS.md` → корень проекта.
Записать выбор пользователя и факты проекта:

- `module`, приложения, бинари;
- короткое имя модуля и причину, если оно короткое;
- команды из `Makefile`;
- конфиг: один `Config`, env в k8s, локально `configs/.env.local`;
- тестовый стек: `mockery` v3 + `testify`, если не отказались явно;
- границы: не создавать пустые каталоги, не коммитить `configs/.env.*` и
  `.env`.

Полный текст навыков не дублировать — только ссылки. Если `AGENTS.md` уже
существует — дополнить, не перезаписывать.

## Проверка

```bash
go build ./...
go vet ./...
go test ./...
make build-all
```

Ожидания:

- `go.mod` содержит `module`, `go` и `testify`; лишних зависимостей нет.
- `make build-all` создаёт бинарь, `make run-<appname>` запускается и
  завершается по Ctrl+C.
- `go test` проходит; тест не ходит в сеть и не использует `time.Sleep`.
- `internal/platform` не импортирует пакеты проекта.
- `internal/config` — один пакет на модуль, без подкаталогов по приложениям.
- `cmd/<appname>` содержит только конфиг, логгер, `app.New`, сигналы,
  запуск и останов; адаптеров там нет.
- `cmd/<appname>/main.go` собран по `references/main.go`: без `MustLoad`,
  `panic` и отдельного обработчика сигналов.
- `.gitignore`, `.agentignore` и `Makefile` лежат в корне проекта.
- `internal/domain` и другие слои отсутствуют, пока в них нет файлов.
- В `.git/config` записаны имя и email разработчика, глобальный конфиг не
  изменён.

## Что не делать

- Не создавать пустые каталоги «на будущее».
- Не добавлять `Dockerfile`, `docker-compose.yml`, миграции и CI, пока нет
  реальной потребности.
- Не оставлять `Dockerfile` с `COPY . .` без `.dockerignore`: контекст
  потянет `.git` и локальные `.env`.
- Не создавать `internal/domain`, `internal/services`, `internal/repository`,
  `internal/infra`, `internal/transport`, `tests/`, `api/` без первого файла.
- Не добавлять сторонние зависимости в `go.mod` на старте: только stdlib и
  `testify` по умолчанию. `stdlib` без testify — только по явному запросу.
- Не откладывать `mockery`/`testify`: они входят в каркас по умолчанию, моки
  генерируются, руками не пишутся.
- Не дублировать настройки адаптеров в общем конфиге: вкладывать локальный
  `Config` (`postgres.Config` и т.п.).
- Не делить `internal/config` по приложениям: конфиг один на модуль.
- Не помещать загрузку env/файла в `platform`: это делает `internal/config`.
- Не выбирать короткое имя модуля молча и не выводить его из имени каталога:
  короткое имя — только для неимпортируемых проектов и с записанной причиной.
- Не выводить `<appname>` из имени репозитория или модуля: `vp-appeal-doc-creator`
  даёт приложение `creator`.
- Не помещать бизнес-логику и SQL в `cmd/`.
- Не создавать адаптеры и сервисы в `cmd/` — это `internal/app/<appname>/app.go`.
- Не валидировать секции конфига в `app.go` или `Load()` — это делает `New`
  адаптера.
- Не заводить `pkg/`, `util/`, `common/`, `helpers/`.
- Не задавать имя и email разработчика через `git config --global`.
- Не коммитить `.env*`, `configs/.env.*` и `bin/`; коммитится только
  `configs/.env.example`.
- Не оставлять проект в нерабочем состоянии: каждый шаг заканчивается
  собираемым кодом.

## Передача дальше

После каркаса:

- первая сущность → `internal/domain/<entity>/`;
- первая операция → `internal/services/<appname>/<service>/`;
- первый порт и адаптер → `internal/repository/`, `internal/infra/<adapter>/`;
- первый входящий адаптер → `internal/transport/<appname>/`;
- первый интеграционный тест → `tests/integration/` по
  `skills/integration-testing`;
- e2e — по `skills/e2e-testing`;
- коммиты — по `skills/commit-conventions`, запись — по `skills/changelog`,
  unit-тесты — по `skills/testing`.
