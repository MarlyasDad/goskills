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

Агент загрузит `bootstrap`, `project-structure`, `architecture`, `testing` и
создаст минимальный работающий каркас:

- `go.mod` — имя модуля и версия Go;
- локальный git-конфиг с именем и email разработчика;
- `cmd/<appname>/main.go` — точка входа и graceful shutdown;
- `internal/app/<appname>/app.go` — прикладной слой;
- `internal/config/<appname>/config.go` — конфиг приложения;
- `internal/platform/config/`, `internal/platform/logger/` — примитивы;
- `config/.env.example` — шаблон переменных окружения;
- `Makefile` — команды `build`, `run`, `test`, `test-race`, `lint`;
- `AGENTS.md`, `.gitignore`, `README.md`.

Это и есть необходимый минимум каталогов. Слои `internal/domain`,
`internal/services`, `internal/repository`, `internal/infra`,
`internal/transport`, `internal/shared`, а также `tests/` и `api/`
появляются вместе с первым файлом в них — пустые каталоги не создаются.
Дальнейшие шаги агент выполняет по навыкам `project-structure` и
`architecture`.
