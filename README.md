# AF-phone-observer

Микросервис платформы **AF** для read-only наблюдения за экраном Android-устройства:
скриншоты через ADB `screencap`, UI dump через `uiautomator` и определение текущего состояния экрана.

## Что делает сервис

- поднимает gRPC-сервис на `:50053`;
- отдаёт health/ready endpoints на `:9090`;
- сохраняет PNG-скриншоты в MinIO;
- продолжает работать в noop-режиме, если MinIO недоступен при старте;
- не выполняет жесты, tap, swipe, ввод текста и ADB forward.

## Требования

- Go 1.25+;
- `make`;
- `adb` для работы с реальным Android-устройством;
- MinIO для хранения скриншотов, если нужен реальный upload;
- `golangci-lint` v2.4.0+ для `make lint` и `make lint-fix`;
- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` только для генерации protobuf.

## Быстрый старт

```bash
make deps
make check
make run
```

После запуска:

- gRPC слушает `GRPC_ADDR`, по умолчанию `:50053`;
- health endpoint доступен на `http://localhost:9090/health`;
- ready endpoint доступен на `http://localhost:9090/ready`.

## Makefile

| Команда | Что делает |
|---------|------------|
| `make deps` | скачивает Go-зависимости через `go mod download` |
| `make tidy` | синхронизирует `go.mod` и `go.sum` |
| `make fmt` | форматирует Go-код через `go fmt ./...` |
| `make vet` | запускает `go vet ./...` |
| `make test` | запускает unit-тесты |
| `make lint` | запускает `golangci-lint run ./...` |
| `make lint-fix` | запускает `golangci-lint run --fix ./...` |
| `make build` | проверяет сборку всех Go-пакетов без создания бинарника |
| `make build-bin` | собирает локальный бинарник `phone-observer` или `phone-observer.exe` |
| `make run` | запускает сервис из `./cmd/server` |
| `make check` | запускает `vet`, `test`, `build` |
| `make proto` | генерирует Go-код из файлов `proto/**/*.proto` |
| `make docker-build` | собирает Docker-образ `af-phone-observer:latest` |

Переменные можно переопределять при запуске:

```bash
make build-bin BINARY=observer-local
make docker-build IMAGE=registry.example.com/af-phone-observer:dev
```

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `GRPC_ADDR` | `:50053` | адрес gRPC-сервера |
| `HEALTH_ADDR` | `:9090` | адрес HTTP health/ready сервера |
| `MINIO_ENDPOINT` | `localhost:9000` | адрес MinIO |
| `MINIO_ACCESS_KEY` | `minioadmin` | access key для MinIO |
| `MINIO_SECRET_KEY` | `minioadmin` | secret key для MinIO |
| `MINIO_BUCKET` | `af-screenshots` | bucket для PNG-скриншотов |
| `MINIO_USE_SSL` | `false` | использовать HTTPS для MinIO |
| `SCREENSHOT_TMP_DIR` | OS temp | директория для временных файлов |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Пример локального запуска с нестандартными портами:

```bash
GRPC_ADDR=:50063 HEALTH_ADDR=:9091 make run
```

## Работа в команде

В репозитории используется GitHub Flow. Базовая ветка проекта сейчас — `master`.

1. Забери свежую базовую ветку:

   ```bash
   git checkout master
   git pull origin master
   ```

2. Создай отдельную ветку под задачу:

   ```bash
   git checkout -b feature/short-task-name
   ```

3. Делай небольшие коммиты в формате Conventional Commits:

   ```bash
   git commit -m "feat: add observer health checks"
   git commit -m "fix: handle empty ui dump"
   git commit -m "docs: describe team workflow"
   ```

4. Перед Pull Request запусти проверки:

   ```bash
   make fmt
   make check
   ```

5. Запушь feature-ветку и открой Pull Request в `master`:

   ```bash
   git push -u origin feature/short-task-name
   ```

6. В Pull Request кратко опиши, что изменилось, как проверялось и есть ли риски.

## Правила команды

- Не пушить напрямую в `master`; изменения попадают туда только через Pull Request.
- Держать ветки короткими и регулярно синхронизироваться с `origin/master`.
- Для личной ветки использовать `git rebase origin/master`; для общей ветки безопаснее `git merge origin/master`.
- Не коммитить `.env`, ключи, токены и MinIO credentials.
- Не менять protobuf без согласования с потребителями API.
- Не добавлять tap/swipe/text-логику в observer: действия принадлежат executor-сервису.
- Не управлять ADB connect/forward в observer: это зона connector-сервиса.
- Если меняется поведение сервиса, обновлять README, AGENTS.md или CLAUDE.md рядом с кодом.

## Синхронизация ветки

```bash
git fetch origin
git rebase origin/master
make check
```

После rebase пушить только свою feature-ветку:

```bash
git push --force-with-lease
```

`--force-with-lease` использовать только после rebase своей ветки. Для `master` прямой push запрещён.
