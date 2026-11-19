# <Project>

<Одно предложение: что делает проект.>

## Стек

- язык: <Go 1.26+>
- БД: <...>
- фронтенд: <... | нет>
- инфраструктура: <Docker, k8s, ...>
- логи/трейсы: <slog, OpenTelemetry>

## Модуль и приложения

- `module`: `<module>`
- приложения и бинари: `<app>` → `bin/<app>`
- короткое имя модуля: <полный путь из remote | причина, почему короткое>

## Команды

- `make build-<app>` — сборка
- `make run-<app>` — локальный запуск, `configs/.env.local`
- `make build-all`
- `make test`, `make test-race`, `make lint`
- `make test-integration`, `make test-e2e` — если есть такие тесты
- `make .bin-deps` — Go-инструменты в `bin/`

## Workflow

1. Уточнить задачу.
2. Сделать минимальный рабочий slice.
3. Обновить артефакты: `api/`, `migrations/`, `docs/specs/`, `CHANGELOG.md`.
4. Прогнать проверки из `Makefile`: `test`, `test-race`, `lint`, `build-all`.
5. Если нашёл противоречие между навыками — сообщить явно: какой навык что
   требует и где конфликт. Не записывать его как отклонение проекта, пока
   правило не согласовано. Текст issue — `skills/skill-issue`.
6. Спросить, нужен ли коммит. Без ответа — не коммитить.
7. Вывести итог: шаги и изменения.

## Конфигурация

- один `Config` на модуль: `internal/config`
- в k8s — env из ConfigMap/Secret
- локально — `configs/.env.local`
- секреты только из env/Secret, не в git

## Тесты

- unit: `testify/require`, `mockery`; правила — `skills/testing`
- интеграционные — `skills/integration-testing`, добавляются по согласованию
- e2e — `skills/e2e-testing`, владелец — тестировщик

## Документация

- спеки — источник истины о системе: `docs/specs/`
- индекс — `docs/specs/README.md`; читать 1–2 релевантных, не все
- спека — факты, навык — процесс
- правила — `skills/agents-md`

## Скиллы и правила

- раскладка — `skills/project-structure`
- зависимости — `skills/architecture`
- конфиг — `skills/config`
- unit-тесты — `skills/testing`
- интеграционные — `skills/integration-testing`
- e2e — `skills/e2e-testing`
- логгер — `skills/logger`
- `Makefile` — `skills/makefile`
- `.gitignore` — `skills/gitignore`
- `.agentignore` — `skills/agentignore`
- `.dockerignore` — `skills/dockerignore`
- issue о навыках — `skills/skill-issue`

## Безопасность

- никаких секретов в коде и репозитории
- реальные `.env` не коммитятся
- доступ к секретам — через env/Secret

## Решения и отклонения

- <YYYY-MM-DD>: <решение и причина>
- конфликт навыков — не отклонение проекта: сначала сообщить явно, после
  согласования — записать.
