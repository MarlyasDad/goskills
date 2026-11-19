# <Project>

<Одно предложение: что делает проект.>

## Требования

- Go <version>
- Docker — если нужен для локальной инфраструктуры

## Запуск

```bash
make run-<app>
```

Локальный конфиг — `configs/.env.local`, он не коммитится.

## Переменные окружения

Шаблон — `configs/.env.example`. В k8s значения приходят из ConfigMap/Secret.

## Команды

| Команда | Назначение |
|---|---|
| `make build-<app>` | сборка в `bin/<app>` |
| `make run-<app>` | локальный запуск |
| `make build-all` | сборка всех приложений |
| `make test` | unit-тесты |
| `make test-race` | тесты с race detector |
| `make lint` | `go vet` |
| `make mocks` | генерация моков |
| `make swag` | Swagger docs для HTTP |
| `make .bin-deps` | Go-инструменты в `bin/` |

Новая цель `Makefile` добавляется в эту таблицу.

## HTTP API

Ручек пока нет. Для каждой ручки добавь строку:

| Метод | Путь | Назначение |
|---|---|---|
| `<METHOD>` | `/<path>` | <назначение> |

Контракт API — в `api/`; README фиксирует метод, путь и назначение.
