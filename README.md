# AF-phone-observer

Микросервис платформы **AF** для read-only наблюдения за экраном Android-устройства:
скриншоты через ADB `screencap`, UI dump через `uiautomator` и определение текущего состояния экрана.

## Что делает сервис

- поднимает gRPC-сервер на `:50053`; protobuf-контракт есть, но `ObserverService` пока не зарегистрирован в Go-handler;
- отдаёт health/ready endpoints на `:9090` при прямом запуске сервиса;
- сохраняет PNG-скриншоты в MinIO;
- продолжает работать в noop-режиме, если MinIO недоступен при старте;
- не выполняет жесты, tap, swipe, ввод текста и ADB forward.

## Функциональность, методы и связи

`AF-phone-observer` — read-only сервис наблюдения. Он получает данные с Android-устройства, приводит их к удобному API-ответу и отдаёт другим частям AF координаты, UI-дерево, скриншоты и признаки текущего экрана.

### Основной функционал

| Возможность | Что делает |
|-------------|------------|
| Скриншот | Получает PNG через ADB `exec-out screencap -p`, определяет размер изображения и сохраняет результат в MinIO или `NoopStorage` |
| UI dump | Запускает `uiautomator dump`, читает XML с телефона и парсит элементы: `type`, `text`, `resource_id`, `content_desc`, `hint`, `bounds`, `center` |
| Поиск элемента | Ищет один UI-элемент по `resource_id`, `text`, `content_desc`, `hint` или `type`; поддерживает `exact` и `contains` |
| Ожидание элемента | Повторяет UI dump до появления элемента или до timeout |
| Detect state | Определяет состояние экрана по UI dump, а в режимах `auto`/`vlm` может дополнять результат VLM-анализом скриншота |
| Cache | Хранит последний успешный screenshot/UI dump в памяти текущей observer-реплики и умеет очищать cache по `serial` |
| Очереди устройств | Для одного `serial` выполняет задачи последовательно через worker; разные `serial` обрабатываются параллельно |
| Health/ready | `/health` проверяет, что HTTP-сервер жив; `/ready` проверяет доступность storage через `ObjectStorage.Ping` |

### Публичные методы

Фактическая рабочая поверхность сейчас — HTTP API и CLI-команды из `cmd/*`.

| Метод | Endpoint | Назначение |
|-------|----------|------------|
| `GET` | `/health` | Проверить, что observer запущен |
| `GET` | `/ready` | Проверить готовность storage |
| `POST` | `/screenshot` | Сделать screenshot и сохранить его в storage |
| `POST` | `/dump-ui` | Получить свежий UI dump в `json` или `xml` |
| `POST` | `/find-element` | Найти элемент на текущем экране |
| `POST` | `/wait-for-element` | Дождаться появления элемента |
| `POST` | `/detect-state` | Определить состояние экрана в режиме `ui`, `auto` или `vlm` |
| `GET` | `/screen/{serial}` | Сделать свежий screenshot через worker конкретного телефона |
| `GET` | `/ui/{serial}` | Сделать свежий UI dump через worker конкретного телефона |
| `DELETE` | `/cache/{serial}` | Очистить in-memory cache для телефона |

В protobuf описан gRPC-сервис `ObserverService` с RPC `CaptureScreenshot`, `DumpUI`, `DetectState` в `proto/observer/v1/observer.proto`. На текущий момент `ObserverHandler.Register` оставлен как no-op, поэтому protobuf-контракт есть, но полноценная Go-регистрация gRPC API ещё не завершена.

### Ключевые функции и методы в коде

| Слой | Методы/функции | Роль |
|------|----------------|------|
| `internal/config` | `Load`, `env`, `envInt`, `parseLogLevel` | Загружает адреса, MinIO, очереди, timeouts, VLM-настройки и уровень логов из env |
| `internal/domain` | `ParseUIDump`, `FindElement`, `ValidateFindElementQuery`, `DetectScreenFromUIDump`, `MergeScreenDetections`, `NormalizeVLMState` | Парсит UI XML, ищет элементы и классифицирует состояние экрана |
| `internal/service` | `ObserveService.DumpUI`, `DumpUIDocument`, `DetectState` | Бизнес-логика UI-наблюдения поверх `port.UIDumper` |
| `internal/service` | `ScreenshotService.Capture`, `CaptureAndStore` | Получает screenshot через порт и сохраняет PNG в storage |
| `internal/service` | `ObservationDispatcher.Capture`, `DumpUI`, `FindElement`, `WaitForElement`, `DetectState`, `CurrentScreen`, `CurrentUI`, `ClearCache` | Ставит задачи в per-serial worker, управляет priority-очередями и cache |
| `internal/adapter/driver` | `UIAutomatorDriver.DumpUI`, `DetectState`, `Ping` | Работает с ADB `uiautomator` и проверяет ADB |
| `internal/adapter/driver` | `AdbScreenshotDriver.Capture` | Получает PNG через ADB или stub PNG для `serial=stub` |
| `internal/adapter/driver` | `CascadingScreenAnalyzer.Analyze` | Вызывает VLM backends по цепочке: VisionServer, Ollama, OpenAI-compatible API |
| `internal/adapter/repository` | `MinIOStorage.Upload`, `Ping`, `NoopStorage.Upload`, `Ping` | Загружает PNG в MinIO или возвращает noop-ссылку без реального upload |
| `internal/adapter/handler` | `HTTPHandler.Routes` и обработчики endpoint'ов | Принимает HTTP-запросы, валидирует параметры и маппит доменные ошибки в HTTP status |

### Связь с другими сервисами и инфраструктурой

| Компонент | Как связан |
|-----------|------------|
| AF orchestrator/clients | Вызывают observer, когда нужно понять, что сейчас на экране телефона, получить screenshot, UI dump, состояние экрана или координаты элемента |
| phone-executor | Использует данные observer для действий: observer только находит элементы и координаты, а tap/swipe/text выполняет executor |
| phone-connector | Отвечает за ADB connect/forward; observer не управляет подключением, а работает с уже доступным локальным `adb -s {serial}` |
| Android device | Источник данных: `screencap -p` для screenshot и `uiautomator dump` для UI XML |
| MinIO | Хранилище PNG-скриншотов; при недоступности на старте сервис использует `NoopStorage` |
| VisionServer/Ollama/OpenAI-compatible API | Опциональные VLM backends для анализа скриншотов в `detect-state` |

## Требования

- Go 1.25+;
- `make`;
- `adb` для работы с реальным Android-устройством;
- MinIO для хранения скриншотов, если нужен реальный upload;
- `make lint` сам запускает закреплённый `golangci-lint v2.4.0` через `go run`;
- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` только для генерации protobuf.

## Быстрый старт

```bash
make deps
make check
make run
```

После запуска:

- gRPC слушает `GRPC_ADDR`, по умолчанию `:50053`;
- `make run` использует dev HTTP endpoint `http://127.0.0.1:19090`;
- health endpoint доступен на `http://127.0.0.1:19090/health`;
- ready endpoint доступен на `http://127.0.0.1:19090/ready`.

Проверить подключённые телефоны:

```bash
make adb-devices
make adb-serial
```

Основной сценарий — реальный телефон, первый `device` из `adb devices`:

```bash
make phone-screen
make phone-ui
make phone-detect-state DETECT_MODE=ui
```

Все REST-команды из Makefile автоматически используют первый подключённый телефон через `PHONE_SERIAL`. По умолчанию они ходят на `OBSERVER_HTTP_URL=http://127.0.0.1:19090`, чтобы не цепляться за старый процесс на сервисном порту `9090`. Если локальный observer на `OBSERVER_HTTP_URL` не запущен, команда поднимет временный `cmd/server`, выполнит запрос и остановит его. Для обращения к уже запущенному или удалённому observer можно отключить это поведение через `OBSERVER_AUTO_START=false`.

Полный набор команд для реального телефона:

```bash
make phone-screenshot
make phone-dump-ui
make phone-screen
make phone-ui
make phone-clear-cache
make phone-find-element FIND_TEXT=OK
make phone-wait-for-element WAIT_TEXT=OK
make phone-detect-state DETECT_MODE=ui
```

Обычные команды тоже работают без `SERIAL=...`, потому `SERIAL` по умолчанию равен `PHONE_SERIAL`:

```bash
make screenshot
make dump-ui
make screen
make ui
make clear-cache
make detect-state DETECT_MODE=ui
```

Если нужно выбрать конкретный телефон, переопредели `PHONE_SERIAL` или `SERIAL`:

```bash
make phone-screen PHONE_SERIAL=R5GL2218DMR
make screen SERIAL=R5GL2218DMR
```

`GET /screen/{serial}` и `GET /ui/{serial}` делают свежий observation через общий worker телефона и после успешного результата обновляют in-memory cache observer. `DELETE /cache/{serial}` очищает только этот in-memory cache текущей observer-реплики; MinIO objects и файлы на телефоне не удаляются.

Найти элемент или дождаться элемента на реальном телефоне:

```bash
make phone-find-element FIND_TEXT=Войти
make phone-find-element FIND_TEXT=войти FIND_MATCH=contains
make phone-find-element FIND_RESOURCE_ID=com.app:id/login
make phone-wait-for-element WAIT_TEXT=Далее
make phone-wait-for-element WAIT_RESOURCE_ID=com.app:id/next WAIT_TIMEOUT_SEC=30 WAIT_CHECK_INTERVAL_MS=500
```

`DETECT_MODE=ui` читает только свежий `uiautomator dump`. Если Android не может отдать dump на динамическом экране, например на постоянно меняющемся видео, команда вернёт ошибку `could not get idle state` и не будет использовать старый XML. Для анализа текущего кадра нужен `DETECT_MODE=auto` или `DETECT_MODE=vlm` с настроенным VLM backend.

`DETECT_PLATFORM` — optional hint для classifier/VLM. Основной сценарий не привязан к конкретному приложению:

```bash
make phone-detect-state DETECT_MODE=ui DETECT_PLATFORM=android
make phone-detect-state DETECT_MODE=auto DETECT_PLATFORM=instagram
make phone-detect-state DETECT_MODE=vlm DETECT_PLATFORM=tiktok
```

## Локальная проверка без телефона

Для локальной проверки без реального телефона можно явно указать stub-драйвер:

```bash
make screenshot SERIAL=stub
make dump-ui SERIAL=stub
make screen SERIAL=stub
make ui SERIAL=stub
make clear-cache SERIAL=stub
make find-element SERIAL=stub FIND_TEXT=OK
make wait-for-element SERIAL=stub WAIT_TEXT=OK
make detect-state SERIAL=stub DETECT_MODE=ui
```

`SERIAL=stub` не смотрит на подключённый телефон и нужен только для тестов/локального smoke.

## Detect State / VLM setup

`POST /detect-state` работает в трёх режимах:

| Mode | Что делает | Что нужно установить |
|------|------------|----------------------|
| `ui` | использует только `uiautomator dump` и rule-based признаки | Go, `adb` в PATH, авторизованный Android-телефон |
| `auto` | сначала UI rules, затем VLM, если настроен backend | то же, плюс один из VLM backend ниже |
| `vlm` | требует screenshot и рабочий VLM backend | то же, плюс один из VLM backend ниже |

Если VLM не настроен, `DETECT_MODE=auto` всё равно вернёт результат по UI dump. `DETECT_MODE=vlm` без VLM backend вернёт `503`.

Локальный VLM через Ollama:

```bash
ollama pull qwen2.5vl:7b
ollama serve
```

Для `qwen2.5vl` нужен Ollama `0.7.0` или новее.

Переменные для observer:

```bash
VLM_BACKENDS=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_VLM_MODEL=qwen2.5vl:7b
```

Если уже поднят VisionServer из `server-144`:

```bash
VLM_BACKENDS=vision_server,ollama,openai
VISION_SERVER_URL=http://localhost:8000
```

OpenAI fallback:

```bash
VLM_BACKENDS=openai
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-5.4-mini
```

Полезные ссылки:

- Ollama install: https://ollama.com/download
- Ollama `qwen2.5vl`: https://ollama.com/library/qwen2.5vl
- OpenAI image input: https://platform.openai.com/docs/guides/images
- OpenAI models: https://platform.openai.com/docs/models

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
| `make adb-devices` | показывает вывод `adb devices` |
| `make adb-serial` | печатает первый serial со статусом `device` |
| `make phone-screenshot` | вызывает `POST /screenshot` для первого реального ADB-телефона |
| `make phone-dump-ui` | вызывает `POST /dump-ui` для первого реального ADB-телефона |
| `make phone-screen` | вызывает `GET /screen/{serial}` для первого реального ADB-телефона |
| `make phone-ui` | вызывает `GET /ui/{serial}` для первого реального ADB-телефона |
| `make phone-clear-cache` | вызывает `DELETE /cache/{serial}` для первого реального ADB-телефона |
| `make phone-find-element` | вызывает `POST /find-element`; нужен один `FIND_*` селектор |
| `make phone-wait-for-element` | вызывает `POST /wait-for-element`; нужен один `WAIT_*` селектор |
| `make phone-detect-state` | вызывает `POST /detect-state` для первого реального ADB-телефона |
| `make screenshot` | вызывает `POST /screenshot`; по умолчанию использует `PHONE_SERIAL` |
| `make dump-ui` | вызывает `POST /dump-ui`; по умолчанию использует `PHONE_SERIAL` |
| `make screen` | вызывает `GET /screen/{serial}`; по умолчанию использует `PHONE_SERIAL` |
| `make ui` | вызывает `GET /ui/{serial}`; по умолчанию использует `PHONE_SERIAL` |
| `make clear-cache` | вызывает `DELETE /cache/{serial}`; по умолчанию использует `PHONE_SERIAL` |
| `make find-element` | вызывает `POST /find-element`; нужен один `FIND_*` селектор |
| `make wait-for-element` | вызывает `POST /wait-for-element`; нужен один `WAIT_*` селектор |
| `make detect-state` | вызывает `POST /detect-state`; по умолчанию использует `PHONE_SERIAL` |
| `make check` | запускает `vet`, `test`, `build` |
| `make proto` | генерирует Go-код из файлов `proto/**/*.proto` |
| `make docker-build` | собирает Docker-образ `af-phone-observer:latest` |

Переменные можно переопределять при запуске:

```bash
make build-bin BINARY=observer-local
make docker-build IMAGE=registry.example.com/af-phone-observer:dev
make phone-screen
make phone-ui UI_FORMAT=xml
make phone-screenshot SCREENSHOT_PRIORITY=high
make phone-find-element FIND_TEXT=OK
make phone-wait-for-element WAIT_TEXT=OK
make phone-detect-state DETECT_MODE=ui
make phone-detect-state PHONE_SERIAL=R5GL2218DMR DETECT_MODE=ui
make screen SERIAL=R5GL2218DMR OBSERVER_HTTP_URL=http://127.0.0.1:19090
make screenshot SERIAL=stub OBSERVER_HTTP_URL=http://127.0.0.1:19090
make dump-ui SERIAL=stub DUMP_UI_FORMAT=xml OBSERVER_HTTP_URL=http://127.0.0.1:19090
make screen SERIAL=stub OBSERVER_HTTP_URL=http://127.0.0.1:19090
make ui SERIAL=stub UI_FORMAT=xml OBSERVER_HTTP_URL=http://127.0.0.1:19090
make clear-cache SERIAL=stub OBSERVER_HTTP_URL=http://127.0.0.1:19090
make find-element SERIAL=stub FIND_TEXT=OK OBSERVER_HTTP_URL=http://127.0.0.1:19090
make wait-for-element SERIAL=stub WAIT_TEXT=OK OBSERVER_HTTP_URL=http://127.0.0.1:19090
make detect-state SERIAL=stub DETECT_MODE=ui OBSERVER_HTTP_URL=http://127.0.0.1:19090
make dump-ui SERIAL=stub OBSERVER_AUTO_START=false
```

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `GRPC_ADDR` | `:50053` | адрес gRPC-сервера |
| `HEALTH_ADDR` | `:9090` | адрес HTTP health/ready сервера; `make run` по умолчанию переопределяет на `127.0.0.1:19090` |
| `MINIO_ENDPOINT` | `localhost:9000` | адрес MinIO |
| `MINIO_ACCESS_KEY` | `minioadmin` | access key для MinIO |
| `MINIO_SECRET_KEY` | `minioadmin` | secret key для MinIO |
| `MINIO_BUCKET` | `af-screenshots` | bucket для PNG-скриншотов |
| `MINIO_USE_SSL` | `false` | использовать HTTPS для MinIO |
| `SCREENSHOT_TMP_DIR` | OS temp | директория для временных файлов |
| `SCREENSHOT_TIMEOUT_SEC` | `10` | timeout по умолчанию для REST screenshot и `make screenshot` |
| `DUMP_UI_TIMEOUT_SEC` | `30` | timeout по умолчанию для REST dump-ui и `make dump-ui` |
| `SCREENSHOT_QUEUE_SIZE` | `32` | размер normal-очереди общего worker на один телефон |
| `SCREENSHOT_HIGH_QUEUE_SIZE` | `8` | размер high-очереди общего worker на один телефон |
| `VLM_BACKENDS` | пусто | список VLM backend через запятую: `vision_server`, `ollama`, `openai` |
| `VISION_SERVER_URL` | пусто | URL VisionServer из `server-144`, например `http://localhost:8000` |
| `OLLAMA_URL` | `http://localhost:11434` | URL локального Ollama |
| `OLLAMA_VLM_MODEL` | `qwen2.5vl:7b` | модель Ollama для screenshot-анализа |
| `OPENAI_API_KEY` | пусто | ключ OpenAI для fallback vision-анализа |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | base URL OpenAI-compatible API |
| `OPENAI_MODEL` | `gpt-5.4-mini` | модель OpenAI для image input |
| `VLM_TIMEOUT_SEC` | `20` | timeout одного VLM-анализа |
| `VLM_MAX_CONCURRENCY` | `2` | максимум одновременных VLM-запросов на observer |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `GOLANGCI_LINT_VERSION` | `v2.4.0` | версия `golangci-lint`, которую запускает `make lint` |

Общие переменные для REST-команд `make phone-*`, `make screenshot`, `make dump-ui`, `make screen`, `make ui`, `make clear-cache`, `make find-element`, `make wait-for-element`, `make detect-state`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `PHONE_SERIAL` | первый `adb devices` со статусом `device` | serial основного реального телефона для `make phone-*` |
| `SERIAL` | `$(PHONE_SERIAL)` | serial для обычных команд; переопредели на `stub` только для локальной проверки |
| `OBSERVER_HTTP_ADDR` | `127.0.0.1:19090` | dev HTTP address для `make run` и REST-команд |
| `OBSERVER_HTTP_URL` | `http://$(OBSERVER_HTTP_ADDR)` | HTTP-адрес observer для CLI-команд |
| `OBSERVER_AUTO_START` | `true` | для локального URL автоматически поднять временный observer, если он не запущен |

Переменные только для `make screenshot`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `SCREENSHOT_PRIORITY` | `normal` | `normal` или `high` |
| `SCREENSHOT_STORE_IN_MINIO` | `true` | должен оставаться `true`, прямой endpoint сейчас сохраняет PNG в MinIO |

Переменные только для `make dump-ui`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `DUMP_UI_FORMAT` | `json` | формат ответа: `json` или `xml` |
| `DUMP_UI_PRIORITY` | `normal` | `normal` или `high` |
| `DUMP_UI_TIMEOUT_SEC` | `30` | timeout запроса в секундах |

Переменные только для `make screen`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `SCREEN_PRIORITY` | `normal` | `normal` или `high` |
| `SCREEN_TIMEOUT_SEC` | `10` | timeout запроса в секундах |

Переменные только для `make ui`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `UI_FORMAT` | `json` | формат ответа: `json` или `xml` |
| `UI_PRIORITY` | `normal` | `normal` или `high` |
| `UI_TIMEOUT_SEC` | `30` | timeout запроса в секундах |

Переменные только для `make clear-cache`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `CACHE_PRIORITY` | `high` | `normal` или `high` |
| `CACHE_TIMEOUT_SEC` | `5` | timeout ожидания очереди в секундах |

Переменные только для `make find-element`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `FIND_TYPE` | пусто | короткий тип элемента, например `Button` |
| `FIND_TEXT` | пусто | точный или contains-поиск по `text` |
| `FIND_RESOURCE_ID` | пусто | точный поиск по `resource-id` |
| `FIND_CONTENT_DESC` | пусто | точный или contains-поиск по `content-desc` |
| `FIND_HINT` | пусто | точный или contains-поиск по `hint` |
| `FIND_MATCH` | `exact` | режим поиска: `exact` или `contains` |
| `FIND_PRIORITY` | `normal` | `normal` или `high` |
| `FIND_TIMEOUT_SEC` | `30` | timeout запроса в секундах |

Если элемент не найден, `POST /find-element` возвращает `404`: цель запроса не выполнена.

Переменные только для `make wait-for-element`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `WAIT_TYPE` | пусто | короткий тип элемента, например `Button` |
| `WAIT_TEXT` | пусто | точный или contains-поиск по `text` |
| `WAIT_RESOURCE_ID` | пусто | точный поиск по `resource-id` |
| `WAIT_CONTENT_DESC` | пусто | точный или contains-поиск по `content-desc` |
| `WAIT_HINT` | пусто | точный или contains-поиск по `hint` |
| `WAIT_MATCH` | `exact` | режим поиска: `exact` или `contains` |
| `WAIT_PRIORITY` | `normal` | `normal` или `high` |
| `WAIT_TIMEOUT_SEC` | `30` | сколько секунд ждать появления элемента |
| `WAIT_CHECK_INTERVAL_MS` | `500` | интервал проверки UI dump; минимум `100` |

Если элемент не появился за `WAIT_TIMEOUT_SEC`, `POST /wait-for-element` возвращает `408`.

Переменные только для `make detect-state`:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `DETECT_MODE` | `ui` | режим: `auto`, `ui` или `vlm` |
| `DETECT_PLATFORM` | `android` | подсказка для VLM: `instagram`, `tiktok`, `youtube`, `android` и т.п. |
| `DETECT_USE_SCREENSHOT` | `true` | делать screenshot для VLM-анализа |
| `DETECT_STORE_SCREENSHOT` | `false` | сохранять screenshot в MinIO и вернуть ссылку |
| `DETECT_PRIORITY` | `normal` | `normal` или `high` |
| `DETECT_TIMEOUT_SEC` | `30` | timeout запроса в секундах |

Если состояние не распознано, `POST /detect-state` возвращает `200` со `state="unknown"` и диагностикой `description`, `elements`, `matched_signals`, `backend_used`.

Пример локального запуска с нестандартными портами:

```bash
OBSERVER_HTTP_ADDR=127.0.0.1:19091 make run
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
