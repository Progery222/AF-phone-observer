# AGENTS.md

Инструкции для Codex и других AI-агентов в репозитории **AF-phone-observer**.

> Дублирует ключевые решения из `CLAUDE.md` в формате, удобном для автономных агентов. При расхождении приоритет у **кода** и **protobuf**; затем `CLAUDE.md`.

## Контекст проекта

Автономный Go-микросервис платформы **AF** (Android Farm): скриншоты, UI-dump (uiautomator), определение состояния экрана. Репозиторий: **Progery222/AF-phone-observer**.

Read-only наблюдение; жесты и ADB forward — в connector и executor.

## Стек (кратко)

- Go 1.22+, gRPC/protobuf
- ADB: `screencap -p`, `uiautomator dump`
- MinIO (minio-go v7) для PNG
- JSON-логи (`slog`), HTTP health `:9090`

## Структура (hexagonal)

```
cmd/server/main.go
internal/config/
internal/domain/          # ScreenState, Screenshot, errors
internal/port/            # UIDumper, ScreenshotCapture, ObjectStorage, Logger
internal/adapter/handler/ # gRPC, health
internal/adapter/driver/  # uiautomator.go (dump + screencap)
internal/adapter/repository/ # minio_storage.go, NoopStorage
internal/service/         # ObserveService, ScreenshotService
proto/observer/v1/observer.proto
proto/common/v1/phone.proto
deploy/Dockerfile
Makefile, go.mod
```

**Hexagonal:** `service` зависит от `port`, `adapter` реализует `port`. `domain` не зависит ни от чего внешнего.

## phone-observer

**Роль:** gRPC `:50053` — CaptureScreenshot, DumpUI, DetectState.

**Ключевые файлы:**
- `internal/domain/screen_state.go` — ScreenState, Screenshot
- `internal/port/observer.go` — UIDumper, ScreenshotCapture, ObjectStorage
- `internal/adapter/driver/uiautomator.go` — adb uiautomator + screencap
- `internal/adapter/repository/minio_storage.go` — upload, bucket auto-create
- `internal/service/observe_service.go` — бизнес-логика
- `proto/observer/v1/observer.proto`

**Поток скриншота:**
1. `AdbScreenshotDriver.Capture` → PNG bytes
2. `ScreenshotService.CaptureAndStore` → key `{serial}/{timestamp}.png`
3. `MinIOStorage.Upload` → return `{bucket}/{key}`

**UI dump:**
1. `adb shell uiautomator dump /sdcard/window_dump.xml`
2. `adb shell cat /sdcard/window_dump.xml`
3. `DetectState` — parse `package="..."` из XML

**Fallback:** MinIO недоступен → `NoopStorage`, warn в лог, DumpUI работает.

## Обязательные практики

| Практика | Где |
|----------|-----|
| Graceful shutdown 10s | `cmd/server/main.go` |
| JSON logs (slog) | stdout, поля `service`, `serial`, `key` |
| Health `:9090/health`, Ready `:9090/ready` | MinIO Ping |
| CommandContext | все ADB-вызовы |
| Stub serial `"stub"` | driver no-op для unit-тестов |

## Команды

```bash
cd AF-phone-observer
go mod download
make build && make test
make run
```

Env: `GRPC_ADDR` (`:50053`), `HEALTH_ADDR`, `MINIO_*`, `SCREENSHOT_TMP_DIR`, `LOG_LEVEL`.

## Git / GitHub Flow

Каждый микросервис AF — **отдельный репозиторий** под Progery222; у разработчика свой сервис. Модель ветвления — **GitHub Flow**.

| Правило | Детали |
|---------|--------|
| Запрет прямого push в `main` | только через PR с ревью |
| Ветка на задачу | `feature/`, `fix/`, `refactor/`, `chore/` |
| Коммиты | Conventional Commits: `feat:` `fix:` `refactor:` `docs:` `test:` `chore:` |
| Синхронизация | раз в день: `git fetch` + `rebase origin/main` (личная ветка) или `merge` (общая) |
| После rebase | `git push --force-with-lease` |
| Merge | зелёный CI + 1–2 апрува → merge → удалить ветку |

**Ежедневный цикл:**

```bash
git checkout main && git pull origin main
git checkout -b feature/my-task
git add . && git commit -m "feat: описание"
git push -u origin feature/my-task
# → открыть PR на GitHub
```

**Конфликты:** убрать маркеры `<<<<<<<` / `=======` / `>>>>>>>`, `git add`, `git rebase --continue`.

**Координация:** Issues (задачи), Projects (канбан: To Do → In Progress → Review → Done), Assignees.

**Защита `main` (тимлид):** require PR, approvals, status checks, branch up to date.

## Что не делать

- **Не пушить в `main`** и не создавать коммиты без явной просьбы пользователя.
- Не выполнять tap/swipe/text — это executor (`:50051`).
- Не управлять ADB connect/forward — это connector (`:50052`).
- Не вызывать adb или MinIO из `service/` — только через `port.*`.
- Не коммитить `.env` с MinIO credentials.

## Язык

Сообщения пользователю и тексты ошибок API — **русский**, если не указано иное.
