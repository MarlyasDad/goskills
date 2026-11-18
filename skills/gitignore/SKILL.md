---
name: gitignore
description: Правила .gitignore Go-проекта: что не коммитить, секреты и локальные конфиги, артефакты сборки, проверка ignore-правил. Используется при создании .gitignore или добавлении файлов в репозиторий. Раскладка — skills/project-structure, конфиг — skills/config.
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: tooling
---

# .gitignore Go-проекта

Навык отвечает на вопрос «**что не должно попасть в git**». Где лежит файл —
в `skills/project-structure`; какие env-файлы бывают — в `skills/config`.

## Файл

- Имя — `.gitignore`, лежит в **корне проекта**, рядом с `go.mod`.
- Один файл на репозиторий. Вложенные `.gitignore` — только если правда нужны.
- Коммитится вместе с проектом.
- Синтаксис: `*`, `?`, `**`, `!`, комментарии `#`.

## Что игнорировать

| Категория | Примеры |
|---|---|
| Сборка | `bin/`, `*.exe`, `*.dll`, `*.so`, `*.dylib` |
| Тесты и покрытие | `*.test`, `*.out`, `coverage.*`, `*.coverprofile` |
| Локальный конфиг | `.env`, `.env.*`, `configs/.env.*` |
| Логи и временные файлы | `*.log`, `logs/`, `tmp/` |
| Локальная инфраструктура | `docker-compose.override.yml` |
| IDE и ОС | `.idea/`, `.vscode/`, `.DS_Store`, `Thumbs.db`, `*.swp` |

Не игнорировать:

- исходный код, `go.mod`, `go.sum`;
- `Makefile`, `README.md`, `AGENTS.md`, `.agentignore`;
- `configs/.env.example` — шаблон коммитится;
- `migrations/`, `api/`, `docs/`, тесты;
- lock-файлы инструментов, если проект их использует.

## Секреты

- `.gitignore` — не граница безопасности: файл мог попасть в историю раньше.
- Реальные секреты не коммитятся и не хранятся в репозитории.
- `configs/.env.example` содержит только имена и безопасные дефолты.
- Правила секретов дублируются в `.agentignore`.
- Для проверки истории — `gitleaks`, pre-commit.

## Порядок

1. Скопировать `references/.gitignore` в корень проекта.
2. Добавить артефакты конкретного проекта (бинарники, generated, dumps).
3. Проверить, что шаблон конфига не попал под ignore:

```bash
git check-ignore -v configs/.env.example   # пусто — файл коммитится
git check-ignore -v configs/.env.dev       # правило найдено — локальный скрыт
git status --ignored
```

4. `git add .` и глазами проверить `git status`: в индексе не должно быть
   `.env.*`, `bin/`, IDE-файлов.

## Эталон

Готовый файл — `references/.gitignore`. Скопировать в корень проекта и
дополнить под проект.

## Границы

- «Что не читает агент» — `skills/agentignore`.
- «Где лежит конфиг» — `skills/project-structure`, `skills/config`.
- «Какие команды у проекта» — `skills/makefile`.
