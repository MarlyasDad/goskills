---
name: skill-creator
description: "Правила создания навыков goskills: структура каталога, YAML frontmatter, обязательное описание, references и проверка перед коммитом. Используется при добавлении или правке skills/*/SKILL.md."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: documentation
---

# Создание навыка

Навык отвечает на вопрос «**как оформить новый skill**». Содержание
конкретных навыков — в их собственных `SKILL.md`.

## Каталог

```schema
skills/
└── <skill-id>/
    ├── SKILL.md
    └── references/
        └── <файл-эталон>
```

- `<skill-id>` — латиница в нижнем регистре, слова через `-`:
  `project-structure`, `agents-md`, `skill-creator`.
- `name` во frontmatter совпадает с `<skill-id>`.
- В `references/` лежат готовые файлы, которые копируются в проект:
  `Makefile`, `.gitignore`, `.agentignore`, шаблон `AGENTS.md`.
- Пути в тексте — относительно навыка: `references/Makefile`.

## Frontmatter

```yaml
---
name: <skill-id>
description: "<что делает, когда применяется и чего не описывает>"
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: <scaffolding|architecture|configuration|testing|tooling|documentation>
---
```

Правила YAML:

- **`description` всегда в двойных кавычках.** Двоеточие с пробелом (`: `)
  внутри незакавыченного значения ломает YAML:
  `mapping values are not allowed in this context`.
- Кавычки внутри значения экранируются: `\"`.
- Если значение начинается с `*`, `&`, `!`, `-`, `{`, `[`, `#`, `>` или
  содержит `: `, ` #`, его тоже нужно цитировать.
- Перед коммитом проверять frontmatter парсером YAML, а не глазами.

Пример ошибки:

```yaml
# плохо: двоеточие после «проекта» превращает строку в mapping
description: Правила проекта: структура и конфиг
```

```yaml
# хорошо
description: "Правила проекта: структура и конфиг"
```

## Описание

- Что навык делает и **когда** его загружать.
- Границы: что им не покрывается и в каком навыке это лежит.
- Без общих слов вроде «полезный навык». Агент выбирает навык по описанию.
- Одно-два предложения, конкретные триггеры: «при создании Makefile»,
  «при добавлении адаптера», «при обновлении AGENTS.md».

## Тело навыка

1. Заголовок `# <Название>`.
2. Назначение и границы с другими навыками.
3. Правила и порядок действий.
4. Примеры или ссылка на `references/`.
5. Проверка/чек-лист.
6. Границы: «это здесь, а это — в другом навыке».

Не дублировать другой навык: ссылаться на него по id
(`skills/project-structure`). Один вопрос — один навык.

Если это **факт** о системе (схема, контракт, статус) — это спека или
`AGENTS.md`, а не навык. Навык описывает повторяемый процесс «как делать».

## Проверка

```bash
# frontmatter парсится как YAML
ruby -ryaml -e 'YAML.load_file(ARGV[0])' skills/<skill-id>/SKILL.md

# имя совпадает с каталогом
grep -n '^name:' skills/<skill-id>/SKILL.md
```

Чек-лист:

- [ ] Каталог и `name` совпадают.
- [ ] `description` в кавычках и без сломанного YAML.
- [ ] Есть назначение, правила, проверка и границы.
- [ ] `references/` содержит только эталонные файлы, не копии чужих навыков.
- [ ] Нет дублей правил из других навыков.
- [ ] README и CHANGELOG обновлены.

## Эталон

Каркас навыка — `references/SKILL.template.md`.

## Границы

- «Что писать в проектную память» — `skills/agents-md`.
- «Где лежит файл в Go-проекте» — `skills/project-structure`.
- «Как оформить коммит» — `skills/commit-conventions`.
