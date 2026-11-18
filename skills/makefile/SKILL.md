---
name: makefile
description: Правила Makefile Go-проекта: один файл на модуль, цели с именем приложения build-<app> и run-<app>, локальный конфиг, тесты, lint и инструменты в bin/. Используется при создании или изменении Makefile. Раскладка — skills/project-structure, тесты — skills/testing.
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: tooling
---

# Makefile Go-проекта

Навык отвечает на вопрос «**какие команды есть у проекта и как они
называются**». Раскладка каталогов — в `skills/project-structure`; тесты — в
`skills/testing`; конфиг — в `skills/config`.

## Файл

- Имя — `Makefile`, лежит в **корне модуля**, рядом с `go.mod`.
- **Один `Makefile` на модуль/workspace**, а не на каждое приложение.
- Цели сборки и запуска содержат имя приложения: `build-<app>`,
  `run-<app>`. По имени цели сразу понятно, что запускается.
- `make` без аргументов показывает `help`, а не собирает случайное
  приложение.

## Обязательные цели

| Цель | Что делает |
|---|---|
| `help` | список приложений и команд, цель по умолчанию |
| `build-<app>` | `go build -o bin/<app> ./cmd/<app>` |
| `run-<app>` | `go run ./cmd/<app> -config configs/.env.local` |
| `build-all` | собрать все приложения из `APPS` |
| `test` | `go test ./...` |
| `test-race` | `go test -race -shuffle=on -count=1 ./...` |
| `lint` | `go vet ./...`; если есть `golangci-lint` — он |

Дополнительно: `fmt`, `tidy`, `.bin-deps`, `clean`.

## Переменные

```make
APPS := creator api
LOCAL_BIN := $(CURDIR)/bin
```

- `APPS` — список приложений модуля. Из него строятся `build-all` и `help`.
- `LOCAL_BIN` — `bin/` в корне проекта. Туда кладутся бинари приложений и
  Go-инструменты: `protoc`-плагины, `mockery`, `swag`, `goose` и прочее.
  `GOBIN=$(LOCAL_BIN) go install ...` не засоряет глобальный `GOPATH`, а
  `bin/` уже в `.gitignore`.
- Отдельных `APP`, `GO`, `BIN`, `ARGS` нет: приложение задаётся в имени цели,
  компилятор — `go` из `PATH`, локальный конфиг вшит в `run-<app>`.

## Бинарные зависимости

Go-инструменты ставятся в `bin/`, а не в глобальный `GOPATH`:

```make
.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
	GOBIN=$(LOCAL_BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	GOBIN=$(LOCAL_BIN) go install github.com/vektra/mockery/v2@latest
```

Цель точечная (с точкой), в обычный `make` не входит. Запускается явно:
`make .bin-deps`.

Правило: **добавил пакет или инструмент — обнови `.bin-deps` в том же
изменении**. Не оставляй установку «на потом».

| Что появилось в проекте | Что добавить |
|---|---|
| Postgres | `goose`; `sqlc`, если генерация запросов |
| HTTP-сервер | `swag` для OpenAPI |
| gRPC | `protoc` (системный) + `protoc-gen-go`, `protoc-gen-go-grpc` |
| Моки | `mockery` или `minimock` — выбрать один |
| Миграции | `goose` |

- `testify` — библиотека, а не бинарь: идёт в `go.mod`, не в `.bin-deps`.
- Если импортируемый модуль требует бинарник для генерации, Makefile
  меняется вместе с подключением модуля.
- Часть инструментов нужна **сразу при bootstrap**: например, `mockery`,
  если пользователь не хочет писать моки руками, и `testify` как библиотека
  проверок. Их не откладывают до первого теста.
- Версии инструментов фиксируются вместо `@latest`: воспроизводимая сборка
  важнее «свежести».

## Версия Go

- В `Makefile` **нет** версии Go и нет переменной вроде `GO_VERSION`.
- Версия живёт в `go.mod` (`go`/`toolchain`) и обновляется там.
- Команды вызывают `go` из `PATH`. Подменить бинарь при необходимости можно
  разово: `make test GO=/path/to/go`, но версия в `Makefile` не фиксируется.

## Правила

- Агент берёт команды из `Makefile`, а не придумывает свои.
- `build-<app>` кладёт артефакт в `bin/<app>`; `bin/` не коммитится.
- `run-<app>` всегда использует `configs/.env.local`; в k8s конфиг приходит
  из env, Makefile там не участвует.
- `test` не ходит в сеть. Интеграционные тесты — отдельная цель, если нужны
  внешние сервисы (`test-integration`).
- `lint` не форматирует код; форматирование — `fmt`.
- Секреты и деплой в `Makefile` не хранятся.
- Не делать отдельный `Makefile` в `cmd/<app>/`: цели с именем приложения
  решают ту же задачу в одном файле.

## Эталон

Готовый файл — `references/Makefile`. Скопировать в корень проекта, заменить
`APPS` на приложения модуля и зафиксировать версии инструментов вместо
`@latest`.

## Проверка

```bash
make help
make build-creator
make run-creator
make test
make lint
```

Ожидания:

- `make help` показывает список приложений;
- бинарь появился в `bin/<app>`;
- `make run-<app>` передаёт `-config configs/.env.local`;
- `make test` проходит без сети;
- в `Makefile` нет версии Go и команд, которых нет в проекте.
