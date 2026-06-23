# CLAUDE.md

Руководство для AI-агентов в репозитории **AF-phone-observer** — микросервис наблюдения за экраном Android-устройств платформы **AF** (Android Farm).

> Репозиторий: [Progery222/AF-phone-observer](https://github.com/Progery222/AF-phone-observer) — автономный сервис; соседние сервисы AF — отдельные репозитории под тем же организационным аккаунтом.

## Назначение

**phone-observer** — съём состояния экрана для orchestrator, recovery-engine и аналитики: скриншоты (ADB screencap → MinIO), UI-dump через **uiautomator**, определение текущего package/activity. Не выполняет жесты — только read-only наблюдение.

## Стек

| Слой | Технология | Зачем |
|------|------------|-------|
| Язык | **Go 1.25+** | Параллельные dump/screencap, малый образ |
| API | **gRPC + protobuf** | Низкая latency для orchestrator |
| Android | **ADB**, **uiautomator** | `screencap`, `uiautomator dump`, `cat` XML |
| Storage | **MinIO** (S3-compatible) | Персистентные скриншоты `{serial}/{timestamp}.png` |
| Observability | **slog** (JSON stdout), HTTP health | Логи и probes |
| Инфра | Docker, Compose → K3s, Traefik | Горизонтальное масштабирование observer-реплик |

**Fallback:** если MinIO недоступен при старте — `NoopStorage` (локальный stub URL), сервис продолжает отдавать DumpUI/DetectState.

## Место в платформе AF

| Сервис | Репозиторий | gRPC-порт | Роль |
|--------|-------------|-----------|------|
| phone-orchestrator | `AF-phone-orchestrator` | — | Сценарии, координация |
| phone-connector | `AF-phone-connector` | `:50052` | ADB-сессии, forward |
| phone-action-executor | `AF-phone-action-executor` | `:50051` | Жесты |
| **phone-observer** | **`AF-phone-observer`** | **`:50053`** | Скриншоты, UI-dump |

```
Orchestrator ──gRPC──► Observer ──ADB──► uiautomator / screencap
                              │
                              └── MinIO (af-screenshots)
RecoveryEngine ◄── object_key / xml_dump (через orchestrator или NATS)
```

Observer stateless; скриншоты пишутся в MinIO, временные файлы — `SCREENSHOT_TMP_DIR` (по умолчанию OS temp).

## Архитектура

### Hexagonal (ports & adapters)

```
AF-phone-observer/
├── cmd/server/main.go           # Инициализация, graceful shutdown
├── internal/
│   ├── config/                  # env (MinIO, порты)
│   ├── domain/                  # ScreenState, Screenshot, ошибки
│   ├── port/                    # UIDumper, ScreenshotCapture, ObjectStorage, Logger
│   ├── adapter/
│   │   ├── handler/             # gRPC ObserverService, HTTP health
│   │   ├── driver/              # uiautomator.go, adb screencap
│   │   └── repository/          # minio_storage.go, NoopStorage
│   └── service/                 # ObserveService, ScreenshotService
├── proto/
│   ├── observer/v1/observer.proto
│   └── common/v1/phone.proto
├── deploy/Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

**Правила слоёв:**
- `domain` и `service` не импортируют gRPC, MinIO SDK, `os/exec`.
- Новый источник UI (Accessibility API, OCR) = новый driver, реализующий `port.UIDumper`.
- Protobuf — единый источник правды для API.

### gRPC API (`proto/observer/v1/observer.proto`)

| RPC | Назначение |
|-----|------------|
| `CaptureScreenshot` | ADB screencap → upload MinIO → `object_key` |
| `DumpUI` | uiautomator dump → XML string |
| `DetectState` | DumpUI + parse `package=` из hierarchy |

### Ключевые файлы

- `internal/domain/screen_state.go` — `ScreenState`, `Screenshot`
- `internal/port/observer.go` — `UIDumper`, `ScreenshotCapture`, `ObjectStorage`
- `internal/adapter/driver/uiautomator.go` — dump + screencap через adb
- `internal/adapter/repository/minio_storage.go` — upload PNG, auto-create bucket
- `internal/service/observe_service.go` — ObserveService, ScreenshotService
- `internal/adapter/handler/grpc.go` — gRPC-хендлер

### Обязательные паттерны

1. **Graceful shutdown** — 10s timeout; завершить текущий upload/dump.
2. **Structured logging** — JSON (`slog`), поля `service`, `serial`, `key`.
3. **Health / Ready** — HTTP `:9090`, `/health`, `/ready` (MinIO Ping или noop).
4. **Контекст** — ADB и MinIO через `context.Context`.
5. **Stub serial** — `"stub"` для тестов без устройства и MinIO.
6. **Ключ объекта** — `{serial}/{YYYYMMDD-HHMMSS}.png` в bucket `MINIO_BUCKET`.

## Команды

```bash
cd AF-phone-observer
go mod download
make proto          # protoc + plugins (см. Makefile)
make build
make run            # gRPC :50053 + health :9090
make test
make lint           # golangci-lint (whitelist, см. .golangci.yml)
make lint-fix       # golangci-lint --fix
make docker-build
```

### Переменные окружения (`internal/config/config.go`)

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `GRPC_ADDR` | `:50053` | gRPC-сервер |
| `HEALTH_ADDR` | `:9090` | HTTP health/ready |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO host:port |
| `MINIO_ACCESS_KEY` | `minioadmin` | Access key |
| `MINIO_SECRET_KEY` | `minioadmin` | Secret key |
| `MINIO_BUCKET` | `af-screenshots` | Bucket для PNG |
| `MINIO_USE_SSL` | `false` | HTTPS к MinIO |
| `SCREENSHOT_TMP_DIR` | OS temp | Локальный temp (расширение) |
| `LOG_LEVEL` | `info` | debug / info / warn / error |

## Локальная разработка

- **Windows:** в PowerShell — `Set-Location '…'; команда`, не `cd … && …`.
- Долгие процессы (`go run`, MinIO compose) — в фоне терминала Cursor.
- Нужны: `adb devices`, uiautomator на устройстве, локальный MinIO (или noop mode).
- Типичный compose: MinIO на `:9000`, bucket создаётся автоматически при старте.

## Git / командная работа (GitHub Flow)

Платформа AF — набор микросервисов; **у каждого разработчика свой сервис** (отдельный репозиторий). Внутри репозитория — **GitHub Flow**.

### Принципы

- **Никто не пушит напрямую в `main`.**
- Каждая задача — **отдельная ветка**; в `main` только через **Pull Request** с ревью.
- Короткие ветки, маленькие коммиты, ежедневная синхронизация с `main`.

### Защита ветки `main` (настраивает тимлид, один раз)

Settings → Branches → branch protection для `main`:

- Require pull request before merging
- Require approvals (1–2)
- Require status checks (CI / тесты)
- Require branch up to date before merging

### Ежедневный цикл

```bash
# 1. свежий main
git checkout main && git pull origin main

# 2. ветка под задачу
git checkout -b feature/async-minio-upload      # fix/, refactor/, chore/

# 3. маленькие коммиты (Conventional Commits)
git add . && git commit -m "feat: async MinIO upload for screenshots"

# 4. push + Pull Request
git push -u origin feature/async-minio-upload
```

Дальше на GitHub: открыть PR с описанием → ревью → зелёный CI + апрув → Merge → удалить ветку.

**Префиксы коммитов:** `feat:` `fix:` `refactor:` `docs:` `test:` `chore:`

### Синхронизация со свежим `main`

Раз в день подтягивай `main`, чтобы не копить конфликты:

```bash
git fetch origin
git rebase origin/main        # если в ветке работаешь только ты
# git merge origin/main       # если ветку делят несколько человек
git push --force-with-lease   # после rebase
```

- **rebase** — личная feature-ветка
- **merge** — общая ветка, над которой работают несколько человек

### Конфликты

Git помечает спорные места `<<<<<<<` `=======` `>>>>>>>`. Оставить корректный вариант, убрать маркеры:

```bash
git add <файл>
git rebase --continue          # или git commit при merge
```

Меньше конфликтов = короткие ветки + частая синхронизация.

### Координация в GitHub

| Инструмент | Назначение |
|------------|------------|
| **Issues** | задачи и баги |
| **Projects** | канбан: To Do → In Progress → Review → Done |
| **Assignees** | кто что делает |

### Чек-лист перед merge

- [ ] `git checkout main && git pull`
- [ ] новая ветка под задачу
- [ ] маленькие коммиты (Conventional Commits)
- [ ] раз в день: rebase на `main`
- [ ] push + PR с описанием
- [ ] зелёный CI + апрув → merge → удалить ветку

## Конвенции для агентов

- Минимальный diff; не менять protobuf без согласования с orchestrator.
- **Не пушить в `main`** — только feature-ветка + PR.
- Коммиты — **Conventional Commits** (`feat:`, `fix:`, …).
- Комментарии — только для неочевидной логики (parse package из XML, noop storage).
- Сообщения gRPC — на **русском**.
- Не коммитить MinIO credentials и `.env`.
- Скриншоты и dump — read-only; жесты только в executor.

## Связанные документы

- Codex / другие агенты: `AGENTS.md`
- Правила Cursor: `.cursor/rules/af.mdc`, `.cursorrules`
- Соседние сервисы: `Progery222/AF-phone-connector`, `Progery222/AF-phone-action-executor`, `Progery222/AF-phone-orchestrator`
