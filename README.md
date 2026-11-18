# goskills

## Подключение к проекту

### git clone
- клонируйте этот репозиторий в любой каталог на вашем ПК
- подключите к вашей среде разработки на примере OpenCode и macOS. 
  Отредактируйте файл ~/.config/opencode/opencode.jsonc следующим образом:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "skills": [
    "~/.config/opencode/skills",
    "~/Documents/GitHub/goskills/skills/"
  ]
}
```
- скиллы, подключенные через opencode.json, новый проект увидит автоматически. 
  OpenCode каждый раз сам обходит каталоги и показывает агенту список доступных навыков, 
  а агент уже решает, какой из них загрузить.
- обновление версии через `git pull`

### .agentignore и права OpenCode

OpenCode **не читает `.agentignore`** автоматически: такого поля в конфиге
нет. Файл остаётся переносимым списком для инструментов, которые его
поддерживают, и копируется в корень проекта.

Для OpenCode те же правила задаются через `permissions`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "permissions": [
    { "action": "read", "resource": "*.env", "effect": "deny" },
    { "action": "read", "resource": "*.env.*", "effect": "deny" },
    { "action": "read", "resource": "*.env.example", "effect": "allow" },
    { "action": "read", "resource": "*.key", "effect": "deny" },
    { "action": "read", "resource": "*.pem", "effect": "deny" },
    { "action": "read", "resource": "secrets/*", "effect": "deny" }
  ]
}
```

Правила применяются по последнему совпадению, поэтому `*.env.example` стоит
после `*.env.*`. При необходимости добавьте такие же правила для `edit`.

Подключить `skills/agentignore/references/.agentignore` через `instructions`
или `references` нельзя: это не включает игнорирование, а просто отправляет
текст агенту. Копия в корне нужна только инструментам, которые сами читают
`.agentignore`.

Если правил много, генерируйте `permissions` из `.agentignore` скриптом —
иначе придётся поддерживать два списка вручную.

## Создание проекта с нуля

После подключения навыков создайте каталог будущего проекта, откройте его в
OpenCode и попросите агента собрать каркас:

> Создай каркас Go-проекта по навыку `bootstrap`.
> Имя модуля: `github.com/org/repo`, имя приложения: `api`, версия Go: `1.26`.
> Разработчик: `Alexander Yudaev <yudaev.alex@gmail.com>`.

Версию Go можно не указывать — тогда агент возьмёт системную из `go version`
или уточнит её, если она ниже `1.26`. Имя и email разработчика агент запишет
в локальный git-конфиг проекта (`.git/config`), не меняя глобальный.

Имя модуля и имя приложения — разные значения. Модуль — путь импорта: если
проект импортируется другими, нужен полный путь из git remote; если никем не
импортируется — допустимо короткое имя, и причина фиксируется в `AGENTS.md`
или `README.md`. Имя приложения — имя бинаря, короткое (`creator`, `api`), не
выводится из имени репозитория. Правила — в навыке `project-structure`.

Агент загрузит `bootstrap`, `project-structure`, `architecture`, `config`,
`testing` и создаст минимальный работающий каркас:

- `go.mod` — имя модуля и версия Go, `testify` для тестов;
- локальный git-конфиг с именем и email разработчика;
- `cmd/<appname>/main.go` — bootstrap, сигналы ОС и graceful shutdown;
- `internal/app/<appname>/app.go` — DI приложения: адаптеры, сервисы, Run;
- `internal/config/` — общий конфиг модуля: `Config`, `Load()`, чтение env;
- `internal/platform/logger/` — примитивы;
- `configs/.env.example` — шаблон переменных окружения;
- `Makefile` — один на модуль: `build-<app>`, `run-<app>`, `build-all`, `test`, `test-race`, `lint`;
- `AGENTS.md`, `.gitignore`, `.agentignore`, `.mockery.yaml`, `README.md`.

Служебные файлы корня создаются по отдельным навыкам: `makefile`,
`gitignore`, `agentignore`.

Конфиг один на модуль: `internal/config` — общее зеркало конфиг-файла, без
деления по приложениям. В k8s значения приходят из env; локально —
`make run-<app>` с `-config configs/.env.local`: файл загружается в env и
перетирает уже заданные переменные, дальше `Load()` работает одинаково.

Это и есть необходимый минимум каталогов. Слои `internal/domain`,
`internal/services`, `internal/repository`, `internal/infra`,
`internal/transport`, `internal/shared`, а также `tests/` и `api/`
появляются вместе с первым файлом в них — пустые каталоги не создаются.
Дальнейшие шаги агент выполняет по навыкам `project-structure`,
`architecture`, `config` и `testing`.
