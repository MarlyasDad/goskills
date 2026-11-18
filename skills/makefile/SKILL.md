---
name: makefile
description: Правила Makefile Go-проекта: обязательные команды build, run, test, test-race, lint, имя приложения и бинаря, локальный запуск с конфигом. Используется при создании или изменении Makefile. Раскладка — skills/project-structure, тесты — skills/testing.
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: tooling
---

# Makefile Go-проекта

Навык отвечает на вопрос «**какие команды есть у проекта и как они
называются**». Раскладка каталогов — в `skills/project-structure`; что и как
тестировать — в `skills/testing`; конфиг — в `skills/config`.

## Файл

- Имя — `Makefile`, лежит в **корне проекта**, рядом с `go.mod`.
- Один файл на модуль, а не на приложение.
- `make` без цели показывает список команд или собирает приложение — но не
  запускает тесты и не деплоит.

## Обязательные цели

| Цель | Что делает |
|---|---|
| `build` | `go build -o bin/<appname> ./cmd/<appname>` |
| `run` | `go run ./cmd/<appname>`, аргументы — через `ARGS` |
| `test` | `go test ./...` |
| `test-race` | `go test -race -shuffle=on -count=1 ./...` |
| `lint` | `go vet ./...`; если в проекте есть `golangci-lint` — он |

Дополнительно, по необходимости: `fmt`, `tidy`, `clean`.

## Переменные

```make
APP ?= <appname>
GO  ?= go
BIN := bin/$(APP)
```

- `APP` — имя приложения и бинаря. Совпадает с `cmd/<appname>/` и
  `bin/<appname>`.
- Приложений несколько — `make build APP=api`, один `Makefile` на модуль.
- `GO` — позволяет подменить компилятор, не хардкодить путь.
- `ARGS` — аргументы запуска: `make run ARGS="-config configs/.env.dev"`.

## Правила

- Агент берёт команды из `Makefile`, а не придумывает свои.
- `build` кладёт артефакт в `bin/`; `bin/` не коммитится.
- `test` не ходит в сеть. Интеграционные тесты — отдельная цель, если нужны
  внешние сервисы (`test-integration`).
- `lint` не форматирует код; форматирование — `fmt`.
- Секреты и деплой в `Makefile` не хранятся.
- Цели не дублируют `go`-команды из CI: CI вызывает `make`.

## Эталон

Готовый файл — `references/Makefile`. Скопировать в корень проекта и заменить
`<appname>`.

## Проверка

```bash
make build
make test
make lint
```

Ожидания:

- бинарь появился в `bin/<appname>`;
- `make run ARGS="-config ..."` передаёт флаг приложению;
- `make test` проходит без сети;
- в `Makefile` нет команд, которых нет в проекте.
