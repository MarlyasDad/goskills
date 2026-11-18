# <Project>

<Одно предложение: что делает проект.>

## Модуль и приложения

- `module`: `<module>`
- приложения и бинари: `<app>` → `bin/<app>`
- короткое имя модуля: <полный путь из remote | причина, почему короткое>

## Команды

- `make build-<app>` — сборка
- `make run-<app>` — локальный запуск, `configs/.env.local`
- `make build-all`
- `make test`, `make test-race`, `make lint`
- `make .bin-deps` — Go-инструменты в `bin/`

## Конфигурация

- один `Config` на модуль: `internal/config`
- в k8s — env из ConfigMap/Secret
- локально — `configs/.env.local`
- секреты только из env/Secret, не в git

## Тесты

- `testify/require`
- `mockery`: моки генерируются, руками не пишутся
- правила — `skills/testing`

## Правила

- раскладка — `skills/project-structure`
- зависимости — `skills/architecture`
- конфиг — `skills/config`
- тесты — `skills/testing`
- `Makefile` — `skills/makefile`
- `.gitignore` — `skills/gitignore`
- `.agentignore` — `skills/agentignore`

## Решения и отклонения

- <YYYY-MM-DD>: <решение и причина>
