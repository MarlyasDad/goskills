---
name: commit-conventions
description: Use when creating git commits. Follows Conventional Commits format.
license: MIT
compatibility: opencode
metadata:
  audience: developers
  workflow: git
---

# Правила оформления коммитов

# Commit Conventions

Применяется ко всем коммитам во всех проектах.

## Формат сообщения

```
<тип>[необязательный контекст][!]: <описание>

[необязательное тело]

[необязательная(ые) сноска(и)]
```

## Типы

| Тип       | Назначение                                   | Semver |
|-----------|----------------------------------------------|--------|
| `feat`    | Новый функционал                             | MINOR  |
| `fix`     | Исправление бага                             | PATCH  |
| `docs`    | Изменение документации/specs                 | —      |
| `style`   | Форматирование, без изменения логики         | —      |
| `refactor`| Рефакторинг без изменения поведения          | —      |
| `perf`    | Оптимизация производительности               | —      |
| `test`    | Добавление/правка тестов                     | —      |
| `build`   | Сборка, зависимости, Dockerfile              | —      |
| `ci`      | CI/CD, пайплайны                             | —      |
| `chore`   | Прочие рутинные задачи                       | —      |
| `revert`  | Отмена изменений (со сноской `Refs: <hash>`) | —      |

Критическое изменение, нарушающее обратную совместимость, — `BREAKING CHANGE` (MAJOR), помечается либо сноской, либо `!` после типа/контекста.

## Контекст (рекомендованный)

Контекст — короткий идентификатор модуля или слоя, которого касается изменение.
Набор контекстов задаётся проектом; ниже — типичные примеры:

| Контекст       | Область                                |
|----------------|----------------------------------------|
| `api`          | HTTP/RPC транспорт, хендлеры           |
| `db`           | Репозитории, миграции, запросы         |
| `core`         | Доменная логика                        |
| `web`          | Фронтенд                               |
| `ui`           | Компоненты и стили                     |
| `infra`        | Config, cache, queue, logger           |

## Правила

1. Тип — существительное в нижнем регистре.
2. После типа/контекста — двоеточие и пробел, затем краткое описание в повелительном наклонении (`add`, `fix`, не `added`/`fixed`).
3. Описание — краткое изложение изменения (до ~72 символов), без точки в конце.
4. Тело отделяется от описания пустой строкой, произвольной формы.
5. Сноски — по формату git trailer (`Token: value` или `Token #value`), токен через `-` вместо пробелов (кроме `BREAKING CHANGE`).
6. `BREAKING CHANGE` — только прописными (или `BREAKING-CHANGE` как токен сноски).
7. Один логический коммит = одно изменение; не смешивать несвязанные правки.

## Примеры

```bash
feat(api): add pagination to orders list

docs: add deployment guide

fix(core): handle websocket reconnect on token expiry

refactor(db): extract converters to separate file

feat!(api): drop support for legacy /v1/subscriptions

BREAKING CHANGE: endpoint /v1/subscriptions removed in favor of /v2/orders
```

## Антипримеры

```bash
fix: bug fixes — too vague

update code — not descriptive enough

wip: working on query stuff — not atomic
```

## Отмена изменений

```bash
revert: let us never again speak of the noodle incident

Refs: 676104e, a215868
```

## Автоматизация

- Changelog генерируется по типам `feat`/`fix`/`BREAKING CHANGE`.
- Версия повышается автоматически: `fix` → PATCH, `feat` → MINOR, `BREAKING CHANGE` → MAJOR.
- Типы вне списка (`docs`, `style`, и т.д.) не влияют на версию.