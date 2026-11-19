---
name: project-structure
description: "Правила инициализации Go-проекта и физической раскладки файлов по каталогам. Используется при создании проекта с нуля, добавлении нового модуля или файла. Не описывает правила зависимостей между слоями — это в skills/architecture."
license: MIT
compatibility: opencode
metadata:
  audience: developer
  workflow: scaffolding
---

# Структура Go-проекта

Навык отвечает на вопрос «**где лежит файл**» и «**как инициализировать проект**».
Вопрос «**что может импортировать**» — в `skills/architecture`; «**как
читать и собирать конфиг**» — в `skills/config`.

## Имя модуля и имя приложения

Это две разные сущности. Одну нельзя выводить из другой.

### module — путь импорта

`module` в `go.mod` — путь импорта, а не имя проекта.

- Если модуль импортируется другими проектами — только полный путь:
  `github.com/org/repo`. Короткое имя снаружи не соберётся.
- Если модуль не импортируется никем — допустимо короткое имя:
  `billing-service` или `creator`. Это осознанное исключение, а не
  дефолт. Причина фиксируется в `AGENTS.md` или `README.md`: например,
  «модуль не переиспользуется, короткое имя выбрано осознанно».
- Git remote при коротком имени остаётся: коротким становится только путь
  импорта, репозиторий не переименовывается.
- Вопрос «будет ли проект импортироваться другими» задаётся при
  инициализации, а не решается молча.

Последствия короткого имени:

- первый элемент без точки (`creator`, `billing-service`) Go трактует
  как локальный или стандартный путь: `go get` такого модуля снаружи не
  работает;
- при переиспользовании в `go.mod` потребителя нужен `replace`;
- переименование позже — правка всех импортов внутри проекта.

### appname — имя приложения

`<appname>` — имя бинаря, а не имя проекта.

- Выбирается по бинарю: короткое, одно слово — `creator`, `api`, `worker`.
- Без доменного префикса и без имени репозитория: `billing-service`
  как `appname` не годится.
- Имя проекта в `appname` не обязано попадать.
- Одинаков в `cmd/<appname>/`, `internal/app/<appname>/`,
  `internal/services/<appname>/`, `internal/transport/<appname>/` и
  совпадает с именем файла в `bin/` (`Makefile`: `APP := <appname>`).
- Если приложений несколько — у каждого свой `appname`.

### Пример

| Репозиторий | module | appname | бинарь |
|---|---|---|---|
| `billing-service` | `billing-service` — не импортируется | `billing` | `bin/billing` |
| `github.com/org/orders-api` | `github.com/org/orders-api` — импортируется | `api` | `bin/api` |

## Инициализация проекта

Выполняется, если в корне проекта отсутствует `go.mod`. Если `go.mod` есть —
раздел пропускается, проект считается инициализированным.

### Шаг 1. Определить имя модуля и имя приложения

Правила — в разделе «Имя модуля и имя приложения».

Имя модуля:

1. Спросить пользователя, будет ли проект импортироваться другими.
2. Если импортируется — `module` = полный путь из git remote:
   `git@github.com:org/repo.git` → `github.com/org/repo`. Если remote нет —
   взять `go.work` или родительский `go.mod` (для вложенных модулей); иначе —
   спросить пользователя.
3. Если не импортируется — предложить короткое имя и получить явное
   согласие. Причину записать в `AGENTS.md` или `README.md`.

Не угадывать имя модуля по имени локального каталога и не использовать
`example.com/...` по умолчанию.

Имя приложения:

1. Спросить пользователя, если он не указал его явно.
2. Предложить короткое имя по бинарю; имя репозитория и имя модуля в него
   не переносить.

Запрещено:
- Угадывать имя модуля по имени локального каталога.
- Использовать заглушку вроде `example.com/project` без явного согласия.
- Выбирать короткое имя модуля молча или без записи причины.
- Выводить `appname` из имени репозитория.

### Шаг 2. Определить версию Go

- Минимальная поддерживаемая версия: `1.26`.
- Если пользователь явно указал другую версию — использовать её.
- Если в проекте уже есть `.go-version`, `toolchain` или CI-конфиг
  с версией Go — взять версию оттуда, не переопределять.
- Если версия нигде не задана — взять системную из `go version`.
- Если системная версия ниже `1.26` или `go` недоступен — уточнить у
  пользователя, не занижать версию молча.

### Шаг 3. Выполнить инициализацию

```bash
go mod init <module>
```

Проверка: go.mod создан, содержит корректные module и go.

### Шаг 4. Перейти к структуре

После успешного go mod init — перейти к разделу «Структура каталогов»
этого же навыка. Не создавать каталоги до инициализации модуля.

Что не делать:
- Не добавлять зависимости в go.mod на этом шаге — только module и go.
- Не создавать main.go или другие файлы до определения структуры.
- Не выполнять go mod tidy на пустом модуле — нечего tidy.

## Структура каталогов

Базовая раскладка:

> Дерево показывает полный набор каталогов, а не обязательный минимум. В
> проекте создаются только нужные; пустые каталоги не создаются. `appname` —
> плейсхолдер: он одинаковый в `cmd/`, `internal/app/`, `internal/services/`
> и `internal/transport/` и совпадает с именем бинаря. Выбор `appname` — в
> разделе «Имя модуля и имя приложения».

```schema
.
├── go.mod
├── Dockerfile
├── .dockerignore
├── Makefile
├── .mockery.yaml
├── README.md
├── .gitignore
├── .agentignore
├── api/
├── bin/
├── configs/
│   └── .env.example
├── cmd/
│   └── appname/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── appname/
│   │       └── app.go
│   ├── config/
│   │   ├── env.go
│   │   ├── file.go
│   │   ├── config.go
│   │   └── config_test.go
│   ├── domain/
│   ├── infra/
│   │   ├── http/
│   │   ├── grpc/
│   │   ├── kafka/
│   │   └── minio/
│   ├── platform/
│   │   ├── logger/
│   │   └── tracing/
│   ├── repository/
│   ├── services/
│   │   └── appname/
│   │       └── user/
│   │           ├── ports.go
│   │           ├── service.go
│   │           ├── dto.go
│   │           ├── register.go
│   │           └── unregister.go
│   ├── shared/
│   └── transport/
│       └── appname/
│           ├── http/
│           │   ├── config.go
│           │   ├── server.go
│           │   ├── router.go
│           │   ├── middleware/
│           │   ├── responses/
│           │   ├── docs/          # generated by swag
│           │   └── <group>/
│           │       └── <resource>/
│           ├── grpc/
│           ├── kafka/
│           └── websocket/
└── tests/
```

## Назначение каталогов

### Корень проекта

- `go.mod` — имя модуля и версия Go. Единственный источник имени модуля.
- `Makefile` — один на модуль. Цели с именем приложения: `build-<app>`,
  `run-<app>`, плюс `build-all`, `test`, `test-race`, `lint`. Бинарные
  зависимости — `.bin-deps`. Агент берёт команды отсюда, а не придумывает
  свои. Правила — `skills/makefile`.
- `.mockery.yaml` — конфиг mockery v3; эталон —
  `skills/bootstrap/references/.mockery.yaml`.
- `Dockerfile` — сборка образа приложения.
- `.dockerignore` — что не попадает в контекст сборки Docker: `.git`,
  локальные `.env`, артефакты. Правила — `skills/dockerignore`.
- `README.md` — как запустить проект, команды, переменные окружения и
  публичные ручки; раздел `HTTP API` заполняется вместе с первой ручкой.
  Шаблон — `skills/bootstrap/references/README.md`.
- `AGENTS.md` — проектная память: модуль, приложения, команды, решения и
  отклонения. Правила — `skills/agents-md`.
- `.gitignore` — что не коммитить. Минимум: `bin/`, `configs/.env.*`, `.env*`,
  IDE/OS; `configs/.env.example` коммитится. Правила — `skills/gitignore`.
- `.agentignore` — что агент не читает: секреты, выгрузки, артефакты.
  Правила — `skills/agentignore`.

### cmd/<appname>/

- `main.go` — bootstrap: прочитать конфиг, создать логгер, собрать приложение
  через `app.New`, поймать сигналы ОС, запустить и корректно остановить.
  Полный пример — `skills/bootstrap`.
- Адаптеры, сервисы и транспорт не создаёт — это делает `app.go`.
- Один подкаталог — одно приложение (один бинарь). Несколько приложений —
  несколько подкаталогов.
- Только запуск: бизнес-логики, запросов и SQL здесь нет.

### internal/domain/

- Сущности, value objects, доменные ошибки, инварианты.
- `<entity-name>/` — отдельный пакет на агрегат, когда у сущности появляется
  поведение. Пока пакет маленький — один файл.
- Не место для JSON/SQL-тегов, HTTP-кодов и сериализации: это детали
  транспорта и хранения.

### internal/services/<appname>/<service>/

- Каталог на сервис; файлы: `ports.go`, `service.go`, `dto.go`,
  `<method>.go`, тест рядом.
- Путь и имена файлов — здесь; как проектировать сервис, use cases, DTO и
  саги — `skills/services`.
- Не складывать все сервисы приложения в один пакет.

### internal/repository/

- Порты доступа к данным: интерфейсы репозиториев и хранилищ, нужные
  сервисам.
- Один файл на агрегат: `order.go`, `user.go`.
- Реализации репозиториев здесь не лежат — они в `internal/infra/<store>/`.

### internal/app/<appname>/

- `app.go` — composition root приложения: создаёт адаптеры, сервисы и
  транспорт из `Config`, связывает их, запускает и останавливает. Правила —
  `skills/composition-root`.
- При росте допустимы соседние файлы того же пакета по жизненному циклу
  (`http.go`, `consumers.go`). Отдельных пакетов и деления по слоям нет.
- Импортирует `domain`, `services`, `repository`, `infra`, `transport`,
  `config`, `platform`.
- Бизнес-логики и сценариев здесь нет: они в `domain` и
  `services/<appname>/<service>`.

### internal/infra/

- Исходящие адаптеры — то, что приложение вызывает само.
- `http/` — HTTP-клиенты внешних API; `grpc/` — gRPC-клиенты;
  `kafka/` — producer; `minio/` — объектное хранилище.
- Один подкаталог — один внешний сервис или технология, самодостаточный.
- Реализации репозиториев для БД — `internal/infra/postgres/`,
  `internal/infra/redis/` и т.п. (создаются по необходимости).
- `config.go` рядом с адаптером — локальный `Config`: поля и валидация.
  Конструктор (`New`, `NewClient`) принимает его аргументом, вызывает
  `Validate()` и возвращает ошибку; адаптер не читает env и не импортирует
  конфиг приложения. Правила — `skills/config`.
- Входящих хендлеров и бизнес-правил здесь нет.

### internal/transport/<appname>/

- Входящие адаптеры — то, что вызывает внешний мир.
- `http/` — HTTP-сервер, роутинг, middleware и контроллеры; `grpc/` —
  gRPC-серверы; `kafka/` — consumer; `websocket/` — соединения.
- Подкаталог по приложению: у разных приложений не пересекаются роуты,
  консьюмеры и порты. Общего `internal/transport/http/` в обход `<appname>`
  нет.
- В `http/` лежат `config.go` (локальный `Config` сервера), `server.go`
  (сборка `*http.Server` без запуска), `router.go` (корневой router),
  `middleware/`, `responses/` (JSON, ошибки, валидация, файлы), `docs/`
  (сгенерированный Swagger) и каталоги контроллеров `<group>/<resource>/`.
- Группу контроллеров не называй `internal`: это зарезервированное имя Go,
  пакет снаружи `.../http` не сможет его импортировать. Используй
  `internalapi`/`externalapi`.
- Слой тонкий: разобрать запрос → вызвать операцию сервиса → собрать ответ.
- Не импортирует `app`: операция вызывается через интерфейс, объявленный
  здесь; реализация — в `services/<appname>/<service>`, связывает `app.go`.
- Структура, DTO, коды и маппинг ошибок — `skills/transport`; HTTP-сервер —
  `skills/http-server`, контроллеры — `skills/http-controllers`, middleware —
  `skills/http-middlewares`, сборка — `skills/composition-root`.

### internal/platform/

- Инфраструктурные примитивы, не связанные с предметной областью.
- `logger/` — создание и настройка логгера на `log/slog`. Правила —
  `skills/logger`.
- `tracing/` — инициализация трейсинга, провайдеры, экспортёры.
- По необходимости сюда же: `metrics/`, `shutdown/`, `clock/`.
- Настройки приходят аргументами конструктора. `platform` не читает env и не
  импортирует пакеты проекта.

### internal/config/

- Один пакет на модуль, без деления на `<appname>`.
- `config.go` — общий `Config` и `Load()`; `env.go` — чтение env; `file.go` —
  загрузка env-файла по флагу `-config`. Входит в каркас всегда: пустой путь —
  no-op, поэтому в k8s он не мешает.
- Источники, композиция, валидация и секреты — `skills/config`.

### internal/shared/

- Общие типы прикладного уровня, нужные нескольким приложениям: пагинация,
  фильтры, ошибки слоя приложения, общие DTO.
- Жёсткое правило: код, нужный одному пакету, живёт в этом пакете; доменные
  типы — в `domain/`. `shared/` не растёт «на всякий случай».

### api/

- Исходные контракты: `openapi.yaml`, `*.proto`, JSON-схемы.
- Сгенерированный код здесь не хранится: он кладётся в пакет, который его
  использует, и не правится руками. Пример: `swag init` пишет docs в
  `internal/transport/<appname>/http/docs`, а не в `api/`.

### configs/

- `configs/.env.example` — шаблон всех переменных с комментариями;
  коммитится. Один файл на модуль.
- Локально путь передаётся флагом `-config` и загружается в env; в k8s
  значения приходят из ConfigMap/Secret, без файла.
- Реальные файлы с секретами (`configs/.env.local`) не коммитятся.
- Правила источников — `skills/config`.

### tests/

- `tests/integration/` — интеграционные тесты с реальными сервисами
  (БД, Kafka, MinIO). Правила — `skills/integration-testing`.
- `tests/e2e/` — критические сценарии через весь стек. Правила —
  `skills/e2e-testing`.
- Unit-тесты — `_test.go` рядом с тестируемым кодом.
- Фикстуры — `testdata/` рядом с тестом.
- `tests/` не повторяет структуру `internal/`: группируется по сценарию.

### bin/

- Артефакты сборки (`go build -o bin/...`). Не коммитится.

## Несколько приложений в одном модуле

- Имя `<appname>` одинаковое в `cmd/<appname>/`, `internal/app/<appname>/`,
  `internal/services/<appname>/` и `internal/transport/<appname>/`.
- Общий код приложений — `internal/domain/`, `internal/repository/`,
  `internal/infra/`, `internal/platform/`, `internal/shared/`,
  `internal/config/`. Сервисы приложения — `internal/services/<appname>/<service>/`.
- Конфиг общий: один `internal/config` и один конфиг-файл на модуль. Каждое
  приложение берёт из `Config` только свои секции и валидирует только их;
  локальные конфиги адаптеров и платформы (`postgres.Config`,
  `logger.Config` и т.п.) вкладываются, а не копируются.
- `services/<a>` не используется в `<b>`: общее правило выносится в `domain`
  или `shared`, а не импортируется из чужого приложения.
- В `app/` только DI: общего кода и сервисов между приложениями нет.

## Куда класть новый файл

| Что создаётся | Куда класть |
|---|---|
| Новая сущность или агрегат | `internal/domain/<entity-name>/` |
| Правило над несколькими сущностями | `internal/services/<appname>/<service>/` |
| Новый сервис | `internal/services/<appname>/<service>/` |
| Новая операция | `internal/services/<appname>/<service>/<method>.go` |
| Интерфейс репозитория | `internal/repository/<entity>.go` |
| Реализация репозитория (БД, кэш) | `internal/infra/<store>/` |
| Локальный конфиг адаптера | `internal/infra/<adapter>/config.go` |
| Локальный конфиг HTTP-сервера | `internal/transport/<appname>/http/config.go` |
| Сборка конфига модуля | `internal/config/config.go` |
| Клиент внешнего HTTP/gRPC API | `internal/infra/http/`, `internal/infra/grpc/` |
| Producer в Kafka | `internal/infra/kafka/` |
| Consumer Kafka | `internal/transport/<appname>/kafka/` |
| HTTP-сервер, router, HTTP Config | `internal/transport/<appname>/http/` |
| HTTP-middleware | `internal/transport/<appname>/http/middleware/` |
| HTTP-ответы: JSON, ошибки, валидация, файлы | `internal/transport/<appname>/http/responses/` |
| Swagger docs (generated) | `internal/transport/<appname>/http/docs/` |
| HTTP-контроллер | `internal/transport/<appname>/http/<group>/<resource>/` |
| gRPC-сервер | `internal/transport/<appname>/grpc/` |
| WebSocket | `internal/transport/<appname>/websocket/` |
| Новая env-переменная | `internal/config/config.go` + `configs/.env.example` |
| Локальный источник конфига | `internal/config/file.go` |
| Логгер, трейсинг, метрики | `internal/platform/<name>/` |
| DI приложения, создание адаптеров | `internal/app/<appname>/app.go` |
| Bootstrap, сигналы ОС | `cmd/<appname>/main.go` |
| Контракт API | `api/` |
| Интеграционный тест | `tests/integration/` |
| E2E-тест | `tests/e2e/` |
| Unit-тест | рядом с кодом, `<file>_test.go` |
| Общий прикладной тип для 2+ приложений | `internal/shared/` |
| Код для внешних потребителей | `pkg/` — только если правда нужно |
| Makefile, `.gitignore`, `.agentignore` | корень проекта; правила — `skills/makefile`, `skills/gitignore`, `skills/agentignore` |

## Что не создавать

- Пустые каталоги «на будущее» — создаются при появлении первого файла.
- `pkg/` без реальной необходимости — по умолчанию его нет.
- `util/`, `common/`, `helpers/`, `misc/` — это свалка; код кладётся туда,
  где он используется, или в домен.
- Дубли настроек: поля `postgres.Config`, `kafka.Config` не повторяются в
  конфиге каждого приложения — сборка их вкладывает.
- `internal/config/<appname>/` — конфиг один на модуль, деление по
  приложениям не нужно.
- Бизнес-логику в `cmd/`, `internal/app/` и `internal/transport/`: в `app`
  только DI, операции — в `services/<appname>/<service>/`.
- Не заводить отдельный пакет под use case: сценарий реализуется сервисом;
  правила — `skills/services`.
- Создание адаптеров и сервисов в `cmd/` — это `app.go`.
- Валидацию секций конфига в DI или `Load()` — это `New` адаптера.
- SQL, HTTP-коды и JSON-теги транспорта в `internal/domain/`.
- Входящие хендлеры в `internal/infra/` и исходящие клиенты в
  `internal/transport/` — это разные направления.
- `internal/transport/http/` в обход `<appname>`: HTTP-сервер и роуты
  принадлежат приложению.
- Создание сервисов, репозиториев и адаптеров в `router.go` — это `app.go`.
- HTTP-сервер, который стартует из конструктора или `router.go`, — запуск
  и останов только в `app.go`.
- Все тесты подряд в `tests/` — там только интеграционные и e2e.
- `bin/` в git.

## Каталоги, создаваемые по необходимости

Базовая раскладка не обязана включать всё сразу. Добавляются при появлении
первого файла:

- `internal/infra/postgres/`, `internal/infra/redis/` — работа с БД и кэшем.
- `internal/platform/metrics/`, `internal/platform/shutdown/` — метрики и
  graceful shutdown.
- `internal/testkit/` — общие тестовые двойники, контрактные наборы и
  хелперы; прод-код его не импортирует.
- `migrations/` — SQL-миграции.
- `docs/` — ADR, схемы, описание проекта.
- `docker-compose.yml` — локальная инфраструктура для разработки и тестов.
- `pkg/` — только код для внешних потребителей.

## Порядок создания каркаса

1. `go mod init <module>` — см. раздел «Инициализация проекта».
2. Корневые файлы: `.gitignore`, `README.md`.
3. `internal/platform/logger/` — логгер по `skills/logger`; `internal/config`
   будет вкладывать `logger.Config`, поэтому логгер создаётся первым.
4. `internal/config/config.go`, `env.go`, `file.go` + `configs/.env.example` —
   общий `Config`, `Load()` и загрузка локального env-файла.
5. `internal/app/<appname>/app.go` — минимальный DI и `Run(ctx)`.
6. `cmd/<appname>/main.go` — конфиг, логгер, `app.New`, сигналы, запуск и
   останов.
7. `internal/domain/` — при появлении первой сущности.
8. `internal/repository/`, `internal/infra/`,
   `internal/services/<appname>/<service>/`, `internal/transport/` — при
   появлении первого порта, операции и адаптера.
9. `Makefile`, `Dockerfile`, `.dockerignore` — когда появляются команды
   сборки и деплой.
10. `tests/integration/` — по `skills/integration-testing`;
    `tests/e2e/` — по `skills/e2e-testing`.

Пустые каталоги на любом шаге не создаются.

### Чек-лист согласованности имён

После переименования модуля или приложения проверить:

- В коде не осталось старого пути импорта:
  ```bash
  grep -rn "<префикс старого модуля>" --include=*.go .
  ```
  Вывод пустой.
- `<appname>` одинаков в `cmd/<appname>/`, `internal/app/<appname>/`,
  `internal/services/<appname>/`, `internal/transport/<appname>/`.
- Имена env совпадают в `internal/config/config.go` и
  `configs/.env.example`.
- Имя бинаря совпадает везде, где оно задаётся:
  - `Makefile` — `go build -o bin/<appname>`;
  - `Dockerfile` — `go build -o`, `COPY` собранного файла, `ENTRYPOINT`;
  - `swag init -g` — путь к `main.go` приложения;
  - CI и k8s-манифесты — имя контейнера и артефакта.
- `go build ./...` проходит.

## Проект скопирован или перенесён из другого репозитория

Относится к двум случаям: перенесён весь проект и вставлен фрагмент кода из
другого репозитория. В обоих случаях чужое имя модуля и приложения — главный
источник ошибок.

### Проект целиком

1. `go.mod` — заменить `module` на имя текущего проекта; версия `go` должна
   быть не ниже `1.26`. Старый путь не оставлять «на время».
2. Импорты — заменить все `<старый module>/...` на `<новый module>/...`:
   grep по старому префиксу должен дать пустой вывод.
3. `<appname>` — переименовать каталоги `cmd/`, `internal/app/`,
   `internal/services/`, `internal/transport/`; имена пакетов и бинаря
   привести к новому `appname`.
4. Сборка — `Makefile` (`APP`, `go build -o`), `Dockerfile` (`-o`, `COPY`,
   `ENTRYPOINT`), CI и k8s-манифесты.
5. `swag init -g` и OpenAPI — путь к `main.go`, `title`/`host`/`basePath` в
   спецификации, каталог сгенерированных docs.
6. env-переменные — имена и префиксы в `internal/config/` и локальных
   конфигах адаптеров, `configs/.env.example`, README.
7. README — назначение, имя проекта, модуль, имя бинаря, команды.
8. CI — пути к `cmd/<appname>`, имена артефактов и образов.
9. `go build ./...`, `go vet ./...`, `go test ./...`.

### Фрагмент кода

Вставленный код не переносит чужие имена:

- импорты — только `<module>/...` текущего проекта; пакеты старого модуля
  переписать или скопировать в подходящий слой;
- `package` — имя по месту, а не из исходного репозитория;
- `appname` — если код ссылается на каталоги приложения, использовать
  текущий `appname`;
- env-переменные — префиксы текущего проекта, а не исходного;
- сгенерированный код (swagger, proto) — перегенерировать, не копировать;
- тесты — обновить импорты и фикстуры; чужой `testdata` не тянуть.

После вставки действует тот же чек-лист: grep по старому префиксу модуля
даёт пустой вывод, `go build ./...` проходит.

## Проект создан без навыков

Относится к существующему Go-проекту, который уже собирается, но вырос без
goskills. `bootstrap` его не пересоздаёт: приводим к каркасу по шагам, не
ломая сборку.

1. **Модуль и приложение.** Проверить `module` в `go.mod`; выбрать
   `<appname>`, если его нет. `module` — только префикс импортов, в именах
   каталогов он не используется. `<appname>` — короткое имя бинаря, не
   репозиторий и не модуль: для `module github.com/longusername/longreponame`
   каталоги будут `cmd/api/`, `internal/app/api/`. Один `<appname>` в
   `cmd/<appname>/`, `internal/app/<appname>/`,
   `internal/services/<appname>/`, `internal/transport/<appname>/`.
2. **Конфиг.** Завести `configs/.env.example`; локальный
   `configs/.env.local` — в `.gitignore`. Один `internal/config` на модуль:
   `Config`, `Load()`, `LoadFile()`, вложенные `logger.Config` и локальные
   `Config` адаптеров. Убрать чтение env из `main` и адаптеров.
3. **Makefile.** Привести цели: `build-<app>`, `run-<app>`, `build-all`,
   `test`, `test-race`, `lint`, `help`, `.bin-deps`, `mocks`;
   `test-integration` и `test-e2e` — когда появятся такие тесты.
   `run-<app>` передаёт `-config configs/.env.local`; версия Go — только в
   `go.mod`.
4. **Корневые файлы.** `.gitignore`, `.agentignore`, `.dockerignore` (если
   есть Docker), `AGENTS.md`, `.mockery.yaml`, `README.md`. Шаблоны —
   `skills/*/references/`.
5. **Слои и импорты.** Убрать `util/`, `common/`, `helpers/`, пустые
   каталоги и мёртвый код. Чужие импорты заменить на `<module>/...`;
   `main.go` оставить только bootstrap; исходящие адаптеры — в `infra`,
   входящие — в `transport`. Границы — `skills/architecture`.
6. **Моки.** `.mockery.yaml` и `make mocks`; `mockery` — в `.bin-deps` с
   зафиксированной версией.
7. **Проверка.** `go mod tidy`, `make .bin-deps`, `make mocks`,
   `go build ./...`, `go vet ./...`, `make test`, `make test-race`,
   `make build-all`; grep по старому префиксу модуля пуст.
8. **AGENTS.md.** Записать оставшиеся осознанные отклонения и их причину.

## Границы

- «Где лежит файл» — этот навык.
- «Что может импортировать» — `skills/architecture`.
- «Как читать и собирать конфиг» — `skills/config`.
- «Какие команды у проекта» — `skills/makefile`.
- «Что не коммитить» — `skills/gitignore`; «что не читает агент» —
  `skills/agentignore`.
- Если правило отвечает на вопрос «какой каталог» — оно здесь.
- Если «какая зависимость допустима» — оно в `skills/architecture`.
- Если «откуда взялось значение» — оно в `skills/config`. Не дублировать.
