# AF-phone-observer

╨Ь╨╕╨║╤А╨╛╤Б╨╡╤А╨▓╨╕╤Б ╨┐╨╗╨░╤В╤Д╨╛╤А╨╝╤Л **AF** ╨┤╨╗╤П read-only ╨╜╨░╨▒╨╗╤О╨┤╨╡╨╜╨╕╤П ╨╖╨░ ╤Н╨║╤А╨░╨╜╨╛╨╝ Android-╤Г╤Б╤В╤А╨╛╨╣╤Б╤В╨▓╨░:
╤Б╨║╤А╨╕╨╜╤И╨╛╤В╤Л ╤З╨╡╤А╨╡╨╖ ADB `screencap`, UI dump ╤З╨╡╤А╨╡╨╖ `uiautomator` ╨╕ ╨╛╨┐╤А╨╡╨┤╨╡╨╗╨╡╨╜╨╕╨╡ ╤В╨╡╨║╤Г╤Й╨╡╨│╨╛ ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╤П ╤Н╨║╤А╨░╨╜╨░.

## ╨з╤В╨╛ ╨┤╨╡╨╗╨░╨╡╤В ╤Б╨╡╤А╨▓╨╕╤Б

- ╨┐╨╛╨┤╨╜╨╕╨╝╨░╨╡╤В gRPC-╤Б╨╡╤А╨▓╨╡╤А ╨╜╨░ `:50053`; protobuf-╨║╨╛╨╜╤В╤А╨░╨║╤В ╨╡╤Б╤В╤М, ╨╜╨╛ `ObserverService` ╨┐╨╛╨║╨░ ╨╜╨╡ ╨╖╨░╤А╨╡╨│╨╕╤Б╤В╤А╨╕╤А╨╛╨▓╨░╨╜ ╨▓ Go-handler;
- ╨╛╤В╨┤╨░╤С╤В health/ready endpoints ╨╜╨░ `:9090` ╨┐╤А╨╕ ╨┐╤А╤П╨╝╨╛╨╝ ╨╖╨░╨┐╤Г╤Б╨║╨╡ ╤Б╨╡╤А╨▓╨╕╤Б╨░;
- ╤Б╨╛╤Е╤А╨░╨╜╤П╨╡╤В PNG-╤Б╨║╤А╨╕╨╜╤И╨╛╤В╤Л ╨▓ MinIO;
- ╨┐╤А╨╛╨┤╨╛╨╗╨╢╨░╨╡╤В ╤А╨░╨▒╨╛╤В╨░╤В╤М ╨▓ noop-╤А╨╡╨╢╨╕╨╝╨╡, ╨╡╤Б╨╗╨╕ MinIO ╨╜╨╡╨┤╨╛╤Б╤В╤Г╨┐╨╡╨╜ ╨┐╤А╨╕ ╤Б╤В╨░╤А╤В╨╡;
- ╨╜╨╡ ╨▓╤Л╨┐╨╛╨╗╨╜╤П╨╡╤В ╨╢╨╡╤Б╤В╤Л, tap, swipe, ╨▓╨▓╨╛╨┤ ╤В╨╡╨║╤Б╤В╨░ ╨╕ ADB forward.

## ╨д╤Г╨╜╨║╤Ж╨╕╨╛╨╜╨░╨╗╤М╨╜╨╛╤Б╤В╤М, ╨╝╨╡╤В╨╛╨┤╤Л ╨╕ ╤Б╨▓╤П╨╖╨╕

`AF-phone-observer` тАФ read-only ╤Б╨╡╤А╨▓╨╕╤Б ╨╜╨░╨▒╨╗╤О╨┤╨╡╨╜╨╕╤П. ╨Ю╨╜ ╨┐╨╛╨╗╤Г╤З╨░╨╡╤В ╨┤╨░╨╜╨╜╤Л╨╡ ╤Б Android-╤Г╤Б╤В╤А╨╛╨╣╤Б╤В╨▓╨░, ╨┐╤А╨╕╨▓╨╛╨┤╨╕╤В ╨╕╤Е ╨║ ╤Г╨┤╨╛╨▒╨╜╨╛╨╝╤Г API-╨╛╤В╨▓╨╡╤В╤Г ╨╕ ╨╛╤В╨┤╨░╤С╤В ╨┤╤А╤Г╨│╨╕╨╝ ╤З╨░╤Б╤В╤П╨╝ AF ╨║╨╛╨╛╤А╨┤╨╕╨╜╨░╤В╤Л, UI-╨┤╨╡╤А╨╡╨▓╨╛, ╤Б╨║╤А╨╕╨╜╤И╨╛╤В╤Л ╨╕ ╨┐╤А╨╕╨╖╨╜╨░╨║╨╕ ╤В╨╡╨║╤Г╤Й╨╡╨│╨╛ ╤Н╨║╤А╨░╨╜╨░.

### ╨Ю╤Б╨╜╨╛╨▓╨╜╨╛╨╣ ╤Д╤Г╨╜╨║╤Ж╨╕╨╛╨╜╨░╨╗

| ╨Т╨╛╨╖╨╝╨╛╨╢╨╜╨╛╤Б╤В╤М | ╨з╤В╨╛ ╨┤╨╡╨╗╨░╨╡╤В |
|-------------|------------|
| ╨б╨║╤А╨╕╨╜╤И╨╛╤В | ╨Я╨╛╨╗╤Г╤З╨░╨╡╤В PNG ╤З╨╡╤А╨╡╨╖ ADB `exec-out screencap -p`, ╨╛╨┐╤А╨╡╨┤╨╡╨╗╤П╨╡╤В ╤А╨░╨╖╨╝╨╡╤А ╨╕╨╖╨╛╨▒╤А╨░╨╢╨╡╨╜╨╕╤П ╨╕ ╤Б╨╛╤Е╤А╨░╨╜╤П╨╡╤В ╤А╨╡╨╖╤Г╨╗╤М╤В╨░╤В ╨▓ MinIO ╨╕╨╗╨╕ `NoopStorage` |
| UI dump | ╨Ч╨░╨┐╤Г╤Б╨║╨░╨╡╤В `uiautomator dump`, ╤З╨╕╤В╨░╨╡╤В XML ╤Б ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ ╨╕ ╨┐╨░╤А╤Б╨╕╤В ╤Н╨╗╨╡╨╝╨╡╨╜╤В╤Л: `type`, `text`, `resource_id`, `content_desc`, `hint`, `bounds`, `center` |
| ╨Я╨╛╨╕╤Б╨║ ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ | ╨Ш╤Й╨╡╤В ╨╛╨┤╨╕╨╜ UI-╤Н╨╗╨╡╨╝╨╡╨╜╤В ╨┐╨╛ `resource_id`, `text`, `content_desc`, `hint` ╨╕╨╗╨╕ `type`; ╨┐╨╛╨┤╨┤╨╡╤А╨╢╨╕╨▓╨░╨╡╤В `exact` ╨╕ `contains` |
| ╨Ю╨╢╨╕╨┤╨░╨╜╨╕╨╡ ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ | ╨Я╨╛╨▓╤В╨╛╤А╤П╨╡╤В UI dump ╨┤╨╛ ╨┐╨╛╤П╨▓╨╗╨╡╨╜╨╕╤П ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ ╨╕╨╗╨╕ ╨┤╨╛ timeout |
| Detect state | ╨Ю╨┐╤А╨╡╨┤╨╡╨╗╤П╨╡╤В ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╤Н╨║╤А╨░╨╜╨░ ╨┐╨╛ UI dump, ╨░ ╨▓ ╤А╨╡╨╢╨╕╨╝╨░╤Е `auto`/`vlm` ╨╝╨╛╨╢╨╡╤В ╨┤╨╛╨┐╨╛╨╗╨╜╤П╤В╤М ╤А╨╡╨╖╤Г╨╗╤М╤В╨░╤В VLM-╨░╨╜╨░╨╗╨╕╨╖╨╛╨╝ ╤Б╨║╤А╨╕╨╜╤И╨╛╤В╨░ |
| Cache | ╨е╤А╨░╨╜╨╕╤В ╨┐╨╛╤Б╨╗╨╡╨┤╨╜╨╕╨╣ ╤Г╤Б╨┐╨╡╤И╨╜╤Л╨╣ screenshot/UI dump ╨▓ ╨┐╨░╨╝╤П╤В╨╕ ╤В╨╡╨║╤Г╤Й╨╡╨╣ observer-╤А╨╡╨┐╨╗╨╕╨║╨╕ ╨╕ ╤Г╨╝╨╡╨╡╤В ╨╛╤З╨╕╤Й╨░╤В╤М cache ╨┐╨╛ `serial` |
| ╨Ю╤З╨╡╤А╨╡╨┤╨╕ ╤Г╤Б╤В╤А╨╛╨╣╤Б╤В╨▓ | ╨Ф╨╗╤П ╨╛╨┤╨╜╨╛╨│╨╛ `serial` ╨▓╤Л╨┐╨╛╨╗╨╜╤П╨╡╤В ╨╖╨░╨┤╨░╤З╨╕ ╨┐╨╛╤Б╨╗╨╡╨┤╨╛╨▓╨░╤В╨╡╨╗╤М╨╜╨╛ ╤З╨╡╤А╨╡╨╖ worker; ╤А╨░╨╖╨╜╤Л╨╡ `serial` ╨╛╨▒╤А╨░╨▒╨░╤В╤Л╨▓╨░╤О╤В╤Б╤П ╨┐╨░╤А╨░╨╗╨╗╨╡╨╗╤М╨╜╨╛ |
| Health/ready | `/health` ╨┐╤А╨╛╨▓╨╡╤А╤П╨╡╤В, ╤З╤В╨╛ HTTP-╤Б╨╡╤А╨▓╨╡╤А ╨╢╨╕╨▓; `/ready` ╨┐╤А╨╛╨▓╨╡╤А╤П╨╡╤В ╨┤╨╛╤Б╤В╤Г╨┐╨╜╨╛╤Б╤В╤М storage ╤З╨╡╤А╨╡╨╖ `ObjectStorage.Ping` |

### ╨Я╤Г╨▒╨╗╨╕╤З╨╜╤Л╨╡ ╨╝╨╡╤В╨╛╨┤╤Л

╨д╨░╨║╤В╨╕╤З╨╡╤Б╨║╨░╤П ╤А╨░╨▒╨╛╤З╨░╤П ╨┐╨╛╨▓╨╡╤А╤Е╨╜╨╛╤Б╤В╤М ╤Б╨╡╨╣╤З╨░╤Б тАФ HTTP API ╨╕ CLI-╨║╨╛╨╝╨░╨╜╨┤╤Л ╨╕╨╖ `cmd/*`.

| ╨Ь╨╡╤В╨╛╨┤ | Endpoint | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|-------|----------|------------|
| `GET` | `/health` | ╨Я╤А╨╛╨▓╨╡╤А╨╕╤В╤М, ╤З╤В╨╛ observer ╨╖╨░╨┐╤Г╤Й╨╡╨╜ |
| `GET` | `/ready` | ╨Я╤А╨╛╨▓╨╡╤А╨╕╤В╤М ╨│╨╛╤В╨╛╨▓╨╜╨╛╤Б╤В╤М storage |
| `POST` | `/screenshot` | ╨б╨┤╨╡╨╗╨░╤В╤М screenshot ╨╕ ╤Б╨╛╤Е╤А╨░╨╜╨╕╤В╤М ╨╡╨│╨╛ ╨▓ storage |
| `POST` | `/dump-ui` | ╨Я╨╛╨╗╤Г╤З╨╕╤В╤М ╤Б╨▓╨╡╨╢╨╕╨╣ UI dump ╨▓ `json` ╨╕╨╗╨╕ `xml` |
| `POST` | `/find-element` | ╨Э╨░╨╣╤В╨╕ ╤Н╨╗╨╡╨╝╨╡╨╜╤В ╨╜╨░ ╤В╨╡╨║╤Г╤Й╨╡╨╝ ╤Н╨║╤А╨░╨╜╨╡ |
| `POST` | `/wait-for-element` | ╨Ф╨╛╨╢╨┤╨░╤В╤М╤Б╤П ╨┐╨╛╤П╨▓╨╗╨╡╨╜╨╕╤П ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ |
| `POST` | `/detect-state` | ╨Ю╨┐╤А╨╡╨┤╨╡╨╗╨╕╤В╤М ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╤Н╨║╤А╨░╨╜╨░ ╨▓ ╤А╨╡╨╢╨╕╨╝╨╡ `ui`, `auto` ╨╕╨╗╨╕ `vlm` |
| `GET` | `/screen/{serial}` | ╨б╨┤╨╡╨╗╨░╤В╤М ╤Б╨▓╨╡╨╢╨╕╨╣ screenshot ╤З╨╡╤А╨╡╨╖ worker ╨║╨╛╨╜╨║╤А╨╡╤В╨╜╨╛╨│╨╛ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `GET` | `/ui/{serial}` | ╨б╨┤╨╡╨╗╨░╤В╤М ╤Б╨▓╨╡╨╢╨╕╨╣ UI dump ╤З╨╡╤А╨╡╨╖ worker ╨║╨╛╨╜╨║╤А╨╡╤В╨╜╨╛╨│╨╛ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `DELETE` | `/cache/{serial}` | ╨Ю╤З╨╕╤Б╤В╨╕╤В╤М in-memory cache ╨┤╨╗╤П ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |

╨Т protobuf ╨╛╨┐╨╕╤Б╨░╨╜ gRPC-╤Б╨╡╤А╨▓╨╕╤Б `ObserverService` ╤Б RPC `CaptureScreenshot`, `DumpUI`, `DetectState` ╨▓ `proto/observer/v1/observer.proto`. ╨Э╨░ ╤В╨╡╨║╤Г╤Й╨╕╨╣ ╨╝╨╛╨╝╨╡╨╜╤В `ObserverHandler.Register` ╨╛╤Б╤В╨░╨▓╨╗╨╡╨╜ ╨║╨░╨║ no-op, ╨┐╨╛╤Н╤В╨╛╨╝╤Г protobuf-╨║╨╛╨╜╤В╤А╨░╨║╤В ╨╡╤Б╤В╤М, ╨╜╨╛ ╨┐╨╛╨╗╨╜╨╛╤Ж╨╡╨╜╨╜╨░╤П Go-╤А╨╡╨│╨╕╤Б╤В╤А╨░╤Ж╨╕╤П gRPC API ╨╡╤Й╤С ╨╜╨╡ ╨╖╨░╨▓╨╡╤А╤И╨╡╨╜╨░.

### ╨Ъ╨╗╤О╤З╨╡╨▓╤Л╨╡ ╤Д╤Г╨╜╨║╤Ж╨╕╨╕ ╨╕ ╨╝╨╡╤В╨╛╨┤╤Л ╨▓ ╨║╨╛╨┤╨╡

| ╨б╨╗╨╛╨╣ | ╨Ь╨╡╤В╨╛╨┤╤Л/╤Д╤Г╨╜╨║╤Ж╨╕╨╕ | ╨а╨╛╨╗╤М |
|------|----------------|------|
| `internal/config` | `Load`, `env`, `envInt`, `parseLogLevel` | ╨Ч╨░╨│╤А╤Г╨╢╨░╨╡╤В ╨░╨┤╤А╨╡╤Б╨░, MinIO, ╨╛╤З╨╡╤А╨╡╨┤╨╕, timeouts, VLM-╨╜╨░╤Б╤В╤А╨╛╨╣╨║╨╕ ╨╕ ╤Г╤А╨╛╨▓╨╡╨╜╤М ╨╗╨╛╨│╨╛╨▓ ╨╕╨╖ env |
| `internal/domain` | `ParseUIDump`, `FindElement`, `ValidateFindElementQuery`, `DetectScreenFromUIDump`, `MergeScreenDetections`, `NormalizeVLMState` | ╨Я╨░╤А╤Б╨╕╤В UI XML, ╨╕╤Й╨╡╤В ╤Н╨╗╨╡╨╝╨╡╨╜╤В╤Л ╨╕ ╨║╨╗╨░╤Б╤Б╨╕╤Д╨╕╤Ж╨╕╤А╤Г╨╡╤В ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╤Н╨║╤А╨░╨╜╨░ |
| `internal/service` | `ObserveService.DumpUI`, `DumpUIDocument`, `DetectState` | ╨С╨╕╨╖╨╜╨╡╤Б-╨╗╨╛╨│╨╕╨║╨░ UI-╨╜╨░╨▒╨╗╤О╨┤╨╡╨╜╨╕╤П ╨┐╨╛╨▓╨╡╤А╤Е `port.UIDumper` |
| `internal/service` | `ScreenshotService.Capture`, `CaptureAndStore` | ╨Я╨╛╨╗╤Г╤З╨░╨╡╤В screenshot ╤З╨╡╤А╨╡╨╖ ╨┐╨╛╤А╤В ╨╕ ╤Б╨╛╤Е╤А╨░╨╜╤П╨╡╤В PNG ╨▓ storage |
| `internal/service` | `ObservationDispatcher.Capture`, `DumpUI`, `FindElement`, `WaitForElement`, `DetectState`, `CurrentScreen`, `CurrentUI`, `ClearCache` | ╨б╤В╨░╨▓╨╕╤В ╨╖╨░╨┤╨░╤З╨╕ ╨▓ per-serial worker, ╤Г╨┐╤А╨░╨▓╨╗╤П╨╡╤В priority-╨╛╤З╨╡╤А╨╡╨┤╤П╨╝╨╕ ╨╕ cache |
| `internal/adapter/driver` | `UIAutomatorDriver.DumpUI`, `DetectState`, `Ping` | ╨а╨░╨▒╨╛╤В╨░╨╡╤В ╤Б ADB `uiautomator` ╨╕ ╨┐╤А╨╛╨▓╨╡╤А╤П╨╡╤В ADB |
| `internal/adapter/driver` | `AdbScreenshotDriver.Capture` | ╨Я╨╛╨╗╤Г╤З╨░╨╡╤В PNG ╤З╨╡╤А╨╡╨╖ ADB ╨╕╨╗╨╕ stub PNG ╨┤╨╗╤П `serial=stub` |
| `internal/adapter/driver` | `CascadingScreenAnalyzer.Analyze` | ╨Т╤Л╨╖╤Л╨▓╨░╨╡╤В VLM backends ╨┐╨╛ ╤Ж╨╡╨┐╨╛╤З╨║╨╡: VisionServer, Ollama, OpenAI-compatible API |
| `internal/adapter/repository` | `MinIOStorage.Upload`, `Ping`, `NoopStorage.Upload`, `Ping` | ╨Ч╨░╨│╤А╤Г╨╢╨░╨╡╤В PNG ╨▓ MinIO ╨╕╨╗╨╕ ╨▓╨╛╨╖╨▓╤А╨░╤Й╨░╨╡╤В noop-╤Б╤Б╤Л╨╗╨║╤Г ╨▒╨╡╨╖ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ upload |
| `internal/adapter/handler` | `HTTPHandler.Routes` ╨╕ ╨╛╨▒╤А╨░╨▒╨╛╤В╤З╨╕╨║╨╕ endpoint'╨╛╨▓ | ╨Я╤А╨╕╨╜╨╕╨╝╨░╨╡╤В HTTP-╨╖╨░╨┐╤А╨╛╤Б╤Л, ╨▓╨░╨╗╨╕╨┤╨╕╤А╤Г╨╡╤В ╨┐╨░╤А╨░╨╝╨╡╤В╤А╤Л ╨╕ ╨╝╨░╨┐╨┐╨╕╤В ╨┤╨╛╨╝╨╡╨╜╨╜╤Л╨╡ ╨╛╤И╨╕╨▒╨║╨╕ ╨▓ HTTP status |

### ╨б╨▓╤П╨╖╤М ╤Б ╨┤╤А╤Г╨│╨╕╨╝╨╕ ╤Б╨╡╤А╨▓╨╕╤Б╨░╨╝╨╕ ╨╕ ╨╕╨╜╤Д╤А╨░╤Б╤В╤А╤Г╨║╤В╤Г╤А╨╛╨╣

| ╨Ъ╨╛╨╝╨┐╨╛╨╜╨╡╨╜╤В | ╨Ъ╨░╨║ ╤Б╨▓╤П╨╖╨░╨╜ |
|-----------|------------|
| AF orchestrator/clients | ╨Т╤Л╨╖╤Л╨▓╨░╤О╤В observer, ╨║╨╛╨│╨┤╨░ ╨╜╤Г╨╢╨╜╨╛ ╨┐╨╛╨╜╤П╤В╤М, ╤З╤В╨╛ ╤Б╨╡╨╣╤З╨░╤Б ╨╜╨░ ╤Н╨║╤А╨░╨╜╨╡ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░, ╨┐╨╛╨╗╤Г╤З╨╕╤В╤М screenshot, UI dump, ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╤Н╨║╤А╨░╨╜╨░ ╨╕╨╗╨╕ ╨║╨╛╨╛╤А╨┤╨╕╨╜╨░╤В╤Л ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ |
| phone-executor | ╨Ш╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В ╨┤╨░╨╜╨╜╤Л╨╡ observer ╨┤╨╗╤П ╨┤╨╡╨╣╤Б╤В╨▓╨╕╨╣: observer ╤В╨╛╨╗╤М╨║╨╛ ╨╜╨░╤Е╨╛╨┤╨╕╤В ╤Н╨╗╨╡╨╝╨╡╨╜╤В╤Л ╨╕ ╨║╨╛╨╛╤А╨┤╨╕╨╜╨░╤В╤Л, ╨░ tap/swipe/text ╨▓╤Л╨┐╨╛╨╗╨╜╤П╨╡╤В executor |
| phone-connector | ╨Ю╤В╨▓╨╡╤З╨░╨╡╤В ╨╖╨░ ADB connect/forward; observer ╨╜╨╡ ╤Г╨┐╤А╨░╨▓╨╗╤П╨╡╤В ╨┐╨╛╨┤╨║╨╗╤О╤З╨╡╨╜╨╕╨╡╨╝, ╨░ ╤А╨░╨▒╨╛╤В╨░╨╡╤В ╤Б ╤Г╨╢╨╡ ╨┤╨╛╤Б╤В╤Г╨┐╨╜╤Л╨╝ ╨╗╨╛╨║╨░╨╗╤М╨╜╤Л╨╝ `adb -s {serial}` |
| Android device | ╨Ш╤Б╤В╨╛╤З╨╜╨╕╨║ ╨┤╨░╨╜╨╜╤Л╤Е: `screencap -p` ╨┤╨╗╤П screenshot ╨╕ `uiautomator dump` ╨┤╨╗╤П UI XML |
| MinIO | ╨е╤А╨░╨╜╨╕╨╗╨╕╤Й╨╡ PNG-╤Б╨║╤А╨╕╨╜╤И╨╛╤В╨╛╨▓; ╨┐╤А╨╕ ╨╜╨╡╨┤╨╛╤Б╤В╤Г╨┐╨╜╨╛╤Б╤В╨╕ ╨╜╨░ ╤Б╤В╨░╤А╤В╨╡ ╤Б╨╡╤А╨▓╨╕╤Б ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `NoopStorage` |
| VisionServer/Ollama/OpenAI-compatible API | ╨Ю╨┐╤Ж╨╕╨╛╨╜╨░╨╗╤М╨╜╤Л╨╡ VLM backends ╨┤╨╗╤П ╨░╨╜╨░╨╗╨╕╨╖╨░ ╤Б╨║╤А╨╕╨╜╤И╨╛╤В╨╛╨▓ ╨▓ `detect-state` |

## ╨в╤А╨╡╨▒╨╛╨▓╨░╨╜╨╕╤П

- Go 1.25+;
- `make`;
- `adb` ╨┤╨╗╤П ╤А╨░╨▒╨╛╤В╤Л ╤Б ╤А╨╡╨░╨╗╤М╨╜╤Л╨╝ Android-╤Г╤Б╤В╤А╨╛╨╣╤Б╤В╨▓╨╛╨╝;
- MinIO ╨┤╨╗╤П ╤Е╤А╨░╨╜╨╡╨╜╨╕╤П ╤Б╨║╤А╨╕╨╜╤И╨╛╤В╨╛╨▓, ╨╡╤Б╨╗╨╕ ╨╜╤Г╨╢╨╡╨╜ ╤А╨╡╨░╨╗╤М╨╜╤Л╨╣ upload;
- `make lint` ╤Б╨░╨╝ ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В ╨╖╨░╨║╤А╨╡╨┐╨╗╤С╨╜╨╜╤Л╨╣ `golangci-lint v2.4.0` ╤З╨╡╤А╨╡╨╖ `go run`;
- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П ╨│╨╡╨╜╨╡╤А╨░╤Ж╨╕╨╕ protobuf.

## ╨С╤Л╤Б╤В╤А╤Л╨╣ ╤Б╤В╨░╤А╤В

```bash
make deps
make check
make run
```

╨Я╨╛╤Б╨╗╨╡ ╨╖╨░╨┐╤Г╤Б╨║╨░:

- gRPC ╤Б╨╗╤Г╤И╨░╨╡╤В `GRPC_ADDR`, ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О `:50053`;
- `make run` ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В dev HTTP endpoint `http://127.0.0.1:19090`;
- health endpoint ╨┤╨╛╤Б╤В╤Г╨┐╨╡╨╜ ╨╜╨░ `http://127.0.0.1:19090/health`;
- ready endpoint ╨┤╨╛╤Б╤В╤Г╨┐╨╡╨╜ ╨╜╨░ `http://127.0.0.1:19090/ready`.

╨Я╤А╨╛╨▓╨╡╤А╨╕╤В╤М ╨┐╨╛╨┤╨║╨╗╤О╤З╤С╨╜╨╜╤Л╨╡ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╤Л:

```bash
make adb-devices
make adb-serial
```

╨Ю╤Б╨╜╨╛╨▓╨╜╨╛╨╣ ╤Б╤Ж╨╡╨╜╨░╤А╨╕╨╣ тАФ ╤А╨╡╨░╨╗╤М╨╜╤Л╨╣ ╤В╨╡╨╗╨╡╤Д╨╛╨╜, ╨┐╨╡╤А╨▓╤Л╨╣ `device` ╨╕╨╖ `adb devices`:

```bash
make phone-screen
make phone-ui
make phone-detect-state DETECT_MODE=ui
```

╨Т╤Б╨╡ REST-╨║╨╛╨╝╨░╨╜╨┤╤Л ╨╕╨╖ Makefile ╨░╨▓╤В╨╛╨╝╨░╤В╨╕╤З╨╡╤Б╨║╨╕ ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╤О╤В ╨┐╨╡╤А╨▓╤Л╨╣ ╨┐╨╛╨┤╨║╨╗╤О╤З╤С╨╜╨╜╤Л╨╣ ╤В╨╡╨╗╨╡╤Д╨╛╨╜ ╤З╨╡╤А╨╡╨╖ `PHONE_SERIAL`. ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╛╨╜╨╕ ╤Е╨╛╨┤╤П╤В ╨╜╨░ `OBSERVER_HTTP_URL=http://127.0.0.1:19090`, ╤З╤В╨╛╨▒╤Л ╨╜╨╡ ╤Ж╨╡╨┐╨╗╤П╤В╤М╤Б╤П ╨╖╨░ ╤Б╤В╨░╤А╤Л╨╣ ╨┐╤А╨╛╤Ж╨╡╤Б╤Б ╨╜╨░ ╤Б╨╡╤А╨▓╨╕╤Б╨╜╨╛╨╝ ╨┐╨╛╤А╤В╤Г `9090`. ╨Х╤Б╨╗╨╕ ╨╗╨╛╨║╨░╨╗╤М╨╜╤Л╨╣ observer ╨╜╨░ `OBSERVER_HTTP_URL` ╨╜╨╡ ╨╖╨░╨┐╤Г╤Й╨╡╨╜, ╨║╨╛╨╝╨░╨╜╨┤╨░ ╨┐╨╛╨┤╨╜╨╕╨╝╨╡╤В ╨▓╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╣ `cmd/server`, ╨▓╤Л╨┐╨╛╨╗╨╜╨╕╤В ╨╖╨░╨┐╤А╨╛╤Б ╨╕ ╨╛╤Б╤В╨░╨╜╨╛╨▓╨╕╤В ╨╡╨│╨╛. ╨Ф╨╗╤П ╨╛╨▒╤А╨░╤Й╨╡╨╜╨╕╤П ╨║ ╤Г╨╢╨╡ ╨╖╨░╨┐╤Г╤Й╨╡╨╜╨╜╨╛╨╝╤Г ╨╕╨╗╨╕ ╤Г╨┤╨░╨╗╤С╨╜╨╜╨╛╨╝╤Г observer ╨╝╨╛╨╢╨╜╨╛ ╨╛╤В╨║╨╗╤О╤З╨╕╤В╤М ╤Н╤В╨╛ ╨┐╨╛╨▓╨╡╨┤╨╡╨╜╨╕╨╡ ╤З╨╡╤А╨╡╨╖ `OBSERVER_AUTO_START=false`.

╨Я╨╛╨╗╨╜╤Л╨╣ ╨╜╨░╨▒╨╛╤А ╨║╨╛╨╝╨░╨╜╨┤ ╨┤╨╗╤П ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░:

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

╨Ю╨▒╤Л╤З╨╜╤Л╨╡ ╨║╨╛╨╝╨░╨╜╨┤╤Л ╤В╨╛╨╢╨╡ ╤А╨░╨▒╨╛╤В╨░╤О╤В ╨▒╨╡╨╖ `SERIAL=...`, ╨┐╨╛╤В╨╛╨╝╤Г `SERIAL` ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╤А╨░╨▓╨╡╨╜ `PHONE_SERIAL`:

```bash
make screenshot
make dump-ui
make screen
make ui
make clear-cache
make detect-state DETECT_MODE=ui
```

╨Х╤Б╨╗╨╕ ╨╜╤Г╨╢╨╜╨╛ ╨▓╤Л╨▒╤А╨░╤В╤М ╨║╨╛╨╜╨║╤А╨╡╤В╨╜╤Л╨╣ ╤В╨╡╨╗╨╡╤Д╨╛╨╜, ╨┐╨╡╤А╨╡╨╛╨┐╤А╨╡╨┤╨╡╨╗╨╕ `PHONE_SERIAL` ╨╕╨╗╨╕ `SERIAL`:

```bash
make phone-screen PHONE_SERIAL=R5GL2218DMR
make screen SERIAL=R5GL2218DMR
```

`GET /screen/{serial}` ╨╕ `GET /ui/{serial}` ╨┤╨╡╨╗╨░╤О╤В ╤Б╨▓╨╡╨╢╨╕╨╣ observation ╤З╨╡╤А╨╡╨╖ ╨╛╨▒╤Й╨╕╨╣ worker ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ ╨╕ ╨┐╨╛╤Б╨╗╨╡ ╤Г╤Б╨┐╨╡╤И╨╜╨╛╨│╨╛ ╤А╨╡╨╖╤Г╨╗╤М╤В╨░╤В╨░ ╨╛╨▒╨╜╨╛╨▓╨╗╤П╤О╤В in-memory cache observer. `DELETE /cache/{serial}` ╨╛╤З╨╕╤Й╨░╨╡╤В ╤В╨╛╨╗╤М╨║╨╛ ╤Н╤В╨╛╤В in-memory cache ╤В╨╡╨║╤Г╤Й╨╡╨╣ observer-╤А╨╡╨┐╨╗╨╕╨║╨╕; MinIO objects ╨╕ ╤Д╨░╨╣╨╗╤Л ╨╜╨░ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨╡ ╨╜╨╡ ╤Г╨┤╨░╨╗╤П╤О╤В╤Б╤П.

╨Э╨░╨╣╤В╨╕ ╤Н╨╗╨╡╨╝╨╡╨╜╤В ╨╕╨╗╨╕ ╨┤╨╛╨╢╨┤╨░╤В╤М╤Б╤П ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ ╨╜╨░ ╤А╨╡╨░╨╗╤М╨╜╨╛╨╝ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨╡:

```bash
make phone-find-element FIND_TEXT=╨Т╨╛╨╣╤В╨╕
make phone-find-element FIND_TEXT=╨▓╨╛╨╣╤В╨╕ FIND_MATCH=contains
make phone-find-element FIND_RESOURCE_ID=com.app:id/login
make phone-wait-for-element WAIT_TEXT=╨Ф╨░╨╗╨╡╨╡
make phone-wait-for-element WAIT_RESOURCE_ID=com.app:id/next WAIT_TIMEOUT_SEC=30 WAIT_CHECK_INTERVAL_MS=500
```

`DETECT_MODE=ui` ╤З╨╕╤В╨░╨╡╤В ╤В╨╛╨╗╤М╨║╨╛ ╤Б╨▓╨╡╨╢╨╕╨╣ `uiautomator dump`. ╨Х╤Б╨╗╨╕ Android ╨╜╨╡ ╨╝╨╛╨╢╨╡╤В ╨╛╤В╨┤╨░╤В╤М dump ╨╜╨░ ╨┤╨╕╨╜╨░╨╝╨╕╤З╨╡╤Б╨║╨╛╨╝ ╤Н╨║╤А╨░╨╜╨╡, ╨╜╨░╨┐╤А╨╕╨╝╨╡╤А ╨╜╨░ ╨┐╨╛╤Б╤В╨╛╤П╨╜╨╜╨╛ ╨╝╨╡╨╜╤П╤О╤Й╨╡╨╝╤Б╤П ╨▓╨╕╨┤╨╡╨╛, ╨║╨╛╨╝╨░╨╜╨┤╨░ ╨▓╨╡╤А╨╜╤С╤В ╨╛╤И╨╕╨▒╨║╤Г `could not get idle state` ╨╕ ╨╜╨╡ ╨▒╤Г╨┤╨╡╤В ╨╕╤Б╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╤М ╤Б╤В╨░╤А╤Л╨╣ XML. ╨Ф╨╗╤П ╨░╨╜╨░╨╗╨╕╨╖╨░ ╤В╨╡╨║╤Г╤Й╨╡╨│╨╛ ╨║╨░╨┤╤А╨░ ╨╜╤Г╨╢╨╡╨╜ `DETECT_MODE=auto` ╨╕╨╗╨╕ `DETECT_MODE=vlm` ╤Б ╨╜╨░╤Б╤В╤А╨╛╨╡╨╜╨╜╤Л╨╝ VLM backend.

`DETECT_PLATFORM` тАФ optional hint ╨┤╨╗╤П classifier/VLM. ╨Ю╤Б╨╜╨╛╨▓╨╜╨╛╨╣ ╤Б╤Ж╨╡╨╜╨░╤А╨╕╨╣ ╨╜╨╡ ╨┐╤А╨╕╨▓╤П╨╖╨░╨╜ ╨║ ╨║╨╛╨╜╨║╤А╨╡╤В╨╜╨╛╨╝╤Г ╨┐╤А╨╕╨╗╨╛╨╢╨╡╨╜╨╕╤О:

```bash
make phone-detect-state DETECT_MODE=ui DETECT_PLATFORM=android
make phone-detect-state DETECT_MODE=auto DETECT_PLATFORM=instagram
make phone-detect-state DETECT_MODE=vlm DETECT_PLATFORM=tiktok
```

## ╨Ы╨╛╨║╨░╨╗╤М╨╜╨░╤П ╨┐╤А╨╛╨▓╨╡╤А╨║╨░ ╨▒╨╡╨╖ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░

╨Ф╨╗╤П ╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨╣ ╨┐╤А╨╛╨▓╨╡╤А╨║╨╕ ╨▒╨╡╨╖ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ ╨╝╨╛╨╢╨╜╨╛ ╤П╨▓╨╜╨╛ ╤Г╨║╨░╨╖╨░╤В╤М stub-╨┤╤А╨░╨╣╨▓╨╡╤А:

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

`SERIAL=stub` ╨╜╨╡ ╤Б╨╝╨╛╤В╤А╨╕╤В ╨╜╨░ ╨┐╨╛╨┤╨║╨╗╤О╤З╤С╨╜╨╜╤Л╨╣ ╤В╨╡╨╗╨╡╤Д╨╛╨╜ ╨╕ ╨╜╤Г╨╢╨╡╨╜ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П ╤В╨╡╤Б╤В╨╛╨▓/╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨│╨╛ smoke.

## Detect State / VLM setup

`POST /detect-state` ╤А╨░╨▒╨╛╤В╨░╨╡╤В ╨▓ ╤В╤А╤С╤Е ╤А╨╡╨╢╨╕╨╝╨░╤Е:

| Mode | ╨з╤В╨╛ ╨┤╨╡╨╗╨░╨╡╤В | ╨з╤В╨╛ ╨╜╤Г╨╢╨╜╨╛ ╤Г╤Б╤В╨░╨╜╨╛╨▓╨╕╤В╤М |
|------|------------|----------------------|
| `ui` | ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В ╤В╨╛╨╗╤М╨║╨╛ `uiautomator dump` ╨╕ rule-based ╨┐╤А╨╕╨╖╨╜╨░╨║╨╕ | Go, `adb` ╨▓ PATH, ╨░╨▓╤В╨╛╤А╨╕╨╖╨╛╨▓╨░╨╜╨╜╤Л╨╣ Android-╤В╨╡╨╗╨╡╤Д╨╛╨╜ |
| `auto` | ╤Б╨╜╨░╤З╨░╨╗╨░ UI rules, ╨╖╨░╤В╨╡╨╝ VLM, ╨╡╤Б╨╗╨╕ ╨╜╨░╤Б╤В╤А╨╛╨╡╨╜ backend | ╤В╨╛ ╨╢╨╡, ╨┐╨╗╤О╤Б ╨╛╨┤╨╕╨╜ ╨╕╨╖ VLM backend ╨╜╨╕╨╢╨╡ |
| `vlm` | ╤В╤А╨╡╨▒╤Г╨╡╤В screenshot ╨╕ ╤А╨░╨▒╨╛╤З╨╕╨╣ VLM backend | ╤В╨╛ ╨╢╨╡, ╨┐╨╗╤О╤Б ╨╛╨┤╨╕╨╜ ╨╕╨╖ VLM backend ╨╜╨╕╨╢╨╡ |

╨Х╤Б╨╗╨╕ VLM ╨╜╨╡ ╨╜╨░╤Б╤В╤А╨╛╨╡╨╜, `DETECT_MODE=auto` ╨▓╤Б╤С ╤А╨░╨▓╨╜╨╛ ╨▓╨╡╤А╨╜╤С╤В ╤А╨╡╨╖╤Г╨╗╤М╤В╨░╤В ╨┐╨╛ UI dump. `DETECT_MODE=vlm` ╨▒╨╡╨╖ VLM backend ╨▓╨╡╤А╨╜╤С╤В `503`.

╨Ы╨╛╨║╨░╨╗╤М╨╜╤Л╨╣ VLM ╤З╨╡╤А╨╡╨╖ Ollama:

```bash
ollama pull qwen2.5vl:7b
ollama serve
```

╨Ф╨╗╤П `qwen2.5vl` ╨╜╤Г╨╢╨╡╨╜ Ollama `0.7.0` ╨╕╨╗╨╕ ╨╜╨╛╨▓╨╡╨╡.

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╨┤╨╗╤П observer:

```bash
VLM_BACKENDS=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_VLM_MODEL=qwen2.5vl:7b
```

╨Х╤Б╨╗╨╕ ╤Г╨╢╨╡ ╨┐╨╛╨┤╨╜╤П╤В VisionServer ╨╕╨╖ `server-144`:

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

╨Я╨╛╨╗╨╡╨╖╨╜╤Л╨╡ ╤Б╤Б╤Л╨╗╨║╨╕:

- Ollama install: https://ollama.com/download
- Ollama `qwen2.5vl`: https://ollama.com/library/qwen2.5vl
- OpenAI image input: https://platform.openai.com/docs/guides/images
- OpenAI models: https://platform.openai.com/docs/models

## Makefile

| ╨Ъ╨╛╨╝╨░╨╜╨┤╨░ | ╨з╤В╨╛ ╨┤╨╡╨╗╨░╨╡╤В |
|---------|------------|
| `make deps` | ╤Б╨║╨░╤З╨╕╨▓╨░╨╡╤В Go-╨╖╨░╨▓╨╕╤Б╨╕╨╝╨╛╤Б╤В╨╕ ╤З╨╡╤А╨╡╨╖ `go mod download` |
| `make tidy` | ╤Б╨╕╨╜╤Е╤А╨╛╨╜╨╕╨╖╨╕╤А╤Г╨╡╤В `go.mod` ╨╕ `go.sum` |
| `make fmt` | ╤Д╨╛╤А╨╝╨░╤В╨╕╤А╤Г╨╡╤В Go-╨║╨╛╨┤ ╤З╨╡╤А╨╡╨╖ `go fmt ./...` |
| `make vet` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В `go vet ./...` |
| `make test` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В unit-╤В╨╡╤Б╤В╤Л |
| `make lint` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В `golangci-lint run ./...` |
| `make lint-fix` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В `golangci-lint run --fix ./...` |
| `make build` | ╨┐╤А╨╛╨▓╨╡╤А╤П╨╡╤В ╤Б╨▒╨╛╤А╨║╤Г ╨▓╤Б╨╡╤Е Go-╨┐╨░╨║╨╡╤В╨╛╨▓ ╨▒╨╡╨╖ ╤Б╨╛╨╖╨┤╨░╨╜╨╕╤П ╨▒╨╕╨╜╨░╤А╨╜╨╕╨║╨░ |
| `make build-bin` | ╤Б╨╛╨▒╨╕╤А╨░╨╡╤В ╨╗╨╛╨║╨░╨╗╤М╨╜╤Л╨╣ ╨▒╨╕╨╜╨░╤А╨╜╨╕╨║ `phone-observer` ╨╕╨╗╨╕ `phone-observer.exe` |
| `make run` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В ╤Б╨╡╤А╨▓╨╕╤Б ╨╕╨╖ `./cmd/server` |
| `make adb-devices` | ╨┐╨╛╨║╨░╨╖╤Л╨▓╨░╨╡╤В ╨▓╤Л╨▓╨╛╨┤ `adb devices` |
| `make adb-serial` | ╨┐╨╡╤З╨░╤В╨░╨╡╤В ╨┐╨╡╤А╨▓╤Л╨╣ serial ╤Б╨╛ ╤Б╤В╨░╤В╤Г╤Б╨╛╨╝ `device` |
| `make phone-screenshot` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /screenshot` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make phone-dump-ui` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /dump-ui` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make phone-screen` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `GET /screen/{serial}` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make phone-ui` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `GET /ui/{serial}` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make phone-clear-cache` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `DELETE /cache/{serial}` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make phone-find-element` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /find-element`; ╨╜╤Г╨╢╨╡╨╜ ╨╛╨┤╨╕╨╜ `FIND_*` ╤Б╨╡╨╗╨╡╨║╤В╨╛╤А |
| `make phone-wait-for-element` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /wait-for-element`; ╨╜╤Г╨╢╨╡╨╜ ╨╛╨┤╨╕╨╜ `WAIT_*` ╤Б╨╡╨╗╨╡╨║╤В╨╛╤А |
| `make phone-detect-state` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /detect-state` ╨┤╨╗╤П ╨┐╨╡╤А╨▓╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ADB-╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ |
| `make screenshot` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /screenshot`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make dump-ui` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /dump-ui`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make screen` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `GET /screen/{serial}`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make ui` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `GET /ui/{serial}`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make clear-cache` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `DELETE /cache/{serial}`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make find-element` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /find-element`; ╨╜╤Г╨╢╨╡╨╜ ╨╛╨┤╨╕╨╜ `FIND_*` ╤Б╨╡╨╗╨╡╨║╤В╨╛╤А |
| `make wait-for-element` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /wait-for-element`; ╨╜╤Г╨╢╨╡╨╜ ╨╛╨┤╨╕╨╜ `WAIT_*` ╤Б╨╡╨╗╨╡╨║╤В╨╛╤А |
| `make detect-state` | ╨▓╤Л╨╖╤Л╨▓╨░╨╡╤В `POST /detect-state`; ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В `PHONE_SERIAL` |
| `make check` | ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В `vet`, `test`, `build` |
| `make proto` | ╨│╨╡╨╜╨╡╤А╨╕╤А╤Г╨╡╤В Go-╨║╨╛╨┤ ╨╕╨╖ ╤Д╨░╨╣╨╗╨╛╨▓ `proto/**/*.proto` |
| `make docker-build` | ╤Б╨╛╨▒╨╕╤А╨░╨╡╤В Docker-╨╛╨▒╤А╨░╨╖ `af-phone-observer:latest` |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╨╝╨╛╨╢╨╜╨╛ ╨┐╨╡╤А╨╡╨╛╨┐╤А╨╡╨┤╨╡╨╗╤П╤В╤М ╨┐╤А╨╕ ╨╖╨░╨┐╤Г╤Б╨║╨╡:

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

## ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╨╛╨║╤А╤Г╨╢╨╡╨╜╨╕╤П

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `GRPC_ADDR` | `:50053` | ╨░╨┤╤А╨╡╤Б gRPC-╤Б╨╡╤А╨▓╨╡╤А╨░ |
| `HEALTH_ADDR` | `:9090` | ╨░╨┤╤А╨╡╤Б HTTP health/ready ╤Б╨╡╤А╨▓╨╡╤А╨░; `make run` ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨┐╨╡╤А╨╡╨╛╨┐╤А╨╡╨┤╨╡╨╗╤П╨╡╤В ╨╜╨░ `127.0.0.1:19090` |
| `MINIO_ENDPOINT` | `localhost:9000` | ╨░╨┤╤А╨╡╤Б MinIO |
| `MINIO_ACCESS_KEY` | `minioadmin` | access key ╨┤╨╗╤П MinIO |
| `MINIO_SECRET_KEY` | `minioadmin` | secret key ╨┤╨╗╤П MinIO |
| `MINIO_BUCKET` | `af-screenshots` | bucket ╨┤╨╗╤П PNG-╤Б╨║╤А╨╕╨╜╤И╨╛╤В╨╛╨▓ |
| `MINIO_USE_SSL` | `false` | ╨╕╤Б╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╤М HTTPS ╨┤╨╗╤П MinIO |
| `SCREENSHOT_TMP_DIR` | OS temp | ╨┤╨╕╤А╨╡╨║╤В╨╛╤А╨╕╤П ╨┤╨╗╤П ╨▓╤А╨╡╨╝╨╡╨╜╨╜╤Л╤Е ╤Д╨░╨╣╨╗╨╛╨▓ |
| `SCREENSHOT_TIMEOUT_SEC` | `10` | timeout ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨┤╨╗╤П REST screenshot ╨╕ `make screenshot` |
| `DUMP_UI_TIMEOUT_SEC` | `30` | timeout ╨┐╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О ╨┤╨╗╤П REST dump-ui ╨╕ `make dump-ui` |
| `SCREENSHOT_QUEUE_SIZE` | `32` | ╤А╨░╨╖╨╝╨╡╤А normal-╨╛╤З╨╡╤А╨╡╨┤╨╕ ╨╛╨▒╤Й╨╡╨│╨╛ worker ╨╜╨░ ╨╛╨┤╨╕╨╜ ╤В╨╡╨╗╨╡╤Д╨╛╨╜ |
| `SCREENSHOT_HIGH_QUEUE_SIZE` | `8` | ╤А╨░╨╖╨╝╨╡╤А high-╨╛╤З╨╡╤А╨╡╨┤╨╕ ╨╛╨▒╤Й╨╡╨│╨╛ worker ╨╜╨░ ╨╛╨┤╨╕╨╜ ╤В╨╡╨╗╨╡╤Д╨╛╨╜ |
| `VLM_BACKENDS` | ╨┐╤Г╤Б╤В╨╛ | ╤Б╨┐╨╕╤Б╨╛╨║ VLM backend ╤З╨╡╤А╨╡╨╖ ╨╖╨░╨┐╤П╤В╤Г╤О: `vision_server`, `ollama`, `openai` |
| `VISION_SERVER_URL` | ╨┐╤Г╤Б╤В╨╛ | URL VisionServer ╨╕╨╖ `server-144`, ╨╜╨░╨┐╤А╨╕╨╝╨╡╤А `http://localhost:8000` |
| `OLLAMA_URL` | `http://localhost:11434` | URL ╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨│╨╛ Ollama |
| `OLLAMA_VLM_MODEL` | `qwen2.5vl:7b` | ╨╝╨╛╨┤╨╡╨╗╤М Ollama ╨┤╨╗╤П screenshot-╨░╨╜╨░╨╗╨╕╨╖╨░ |
| `OPENAI_API_KEY` | ╨┐╤Г╤Б╤В╨╛ | ╨║╨╗╤О╤З OpenAI ╨┤╨╗╤П fallback vision-╨░╨╜╨░╨╗╨╕╨╖╨░ |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | base URL OpenAI-compatible API |
| `OPENAI_MODEL` | `gpt-5.4-mini` | ╨╝╨╛╨┤╨╡╨╗╤М OpenAI ╨┤╨╗╤П image input |
| `VLM_TIMEOUT_SEC` | `20` | timeout ╨╛╨┤╨╜╨╛╨│╨╛ VLM-╨░╨╜╨░╨╗╨╕╨╖╨░ |
| `VLM_MAX_CONCURRENCY` | `2` | ╨╝╨░╨║╤Б╨╕╨╝╤Г╨╝ ╨╛╨┤╨╜╨╛╨▓╤А╨╡╨╝╨╡╨╜╨╜╤Л╤Е VLM-╨╖╨░╨┐╤А╨╛╤Б╨╛╨▓ ╨╜╨░ observer |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `GOLANGCI_LINT_VERSION` | `v2.4.0` | ╨▓╨╡╤А╤Б╨╕╤П `golangci-lint`, ╨║╨╛╤В╨╛╤А╤Г╤О ╨╖╨░╨┐╤Г╤Б╨║╨░╨╡╤В `make lint` |

╨Ю╨▒╤Й╨╕╨╡ ╨┐╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╨┤╨╗╤П REST-╨║╨╛╨╝╨░╨╜╨┤ `make phone-*`, `make screenshot`, `make dump-ui`, `make screen`, `make ui`, `make clear-cache`, `make find-element`, `make wait-for-element`, `make detect-state`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `PHONE_SERIAL` | ╨┐╨╡╤А╨▓╤Л╨╣ `adb devices` ╤Б╨╛ ╤Б╤В╨░╤В╤Г╤Б╨╛╨╝ `device` | serial ╨╛╤Б╨╜╨╛╨▓╨╜╨╛╨│╨╛ ╤А╨╡╨░╨╗╤М╨╜╨╛╨│╨╛ ╤В╨╡╨╗╨╡╤Д╨╛╨╜╨░ ╨┤╨╗╤П `make phone-*` |
| `SERIAL` | `$(PHONE_SERIAL)` | serial ╨┤╨╗╤П ╨╛╨▒╤Л╤З╨╜╤Л╤Е ╨║╨╛╨╝╨░╨╜╨┤; ╨┐╨╡╤А╨╡╨╛╨┐╤А╨╡╨┤╨╡╨╗╨╕ ╨╜╨░ `stub` ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П ╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨╣ ╨┐╤А╨╛╨▓╨╡╤А╨║╨╕ |
| `OBSERVER_HTTP_ADDR` | `127.0.0.1:19090` | dev HTTP address ╨┤╨╗╤П `make run` ╨╕ REST-╨║╨╛╨╝╨░╨╜╨┤ |
| `OBSERVER_HTTP_URL` | `http://$(OBSERVER_HTTP_ADDR)` | HTTP-╨░╨┤╤А╨╡╤Б observer ╨┤╨╗╤П CLI-╨║╨╛╨╝╨░╨╜╨┤ |
| `OBSERVER_AUTO_START` | `true` | ╨┤╨╗╤П ╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨│╨╛ URL ╨░╨▓╤В╨╛╨╝╨░╤В╨╕╤З╨╡╤Б╨║╨╕ ╨┐╨╛╨┤╨╜╤П╤В╤М ╨▓╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╣ observer, ╨╡╤Б╨╗╨╕ ╨╛╨╜ ╨╜╨╡ ╨╖╨░╨┐╤Г╤Й╨╡╨╜ |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make screenshot`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `SCREENSHOT_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `SCREENSHOT_STORE_IN_MINIO` | `true` | ╨┤╨╛╨╗╨╢╨╡╨╜ ╨╛╤Б╤В╨░╨▓╨░╤В╤М╤Б╤П `true`, ╨┐╤А╤П╨╝╨╛╨╣ endpoint ╤Б╨╡╨╣╤З╨░╤Б ╤Б╨╛╤Е╤А╨░╨╜╤П╨╡╤В PNG ╨▓ MinIO |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make dump-ui`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `DUMP_UI_FORMAT` | `json` | ╤Д╨╛╤А╨╝╨░╤В ╨╛╤В╨▓╨╡╤В╨░: `json` ╨╕╨╗╨╕ `xml` |
| `DUMP_UI_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `DUMP_UI_TIMEOUT_SEC` | `30` | timeout ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make screen`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `SCREEN_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `SCREEN_TIMEOUT_SEC` | `10` | timeout ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make ui`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `UI_FORMAT` | `json` | ╤Д╨╛╤А╨╝╨░╤В ╨╛╤В╨▓╨╡╤В╨░: `json` ╨╕╨╗╨╕ `xml` |
| `UI_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `UI_TIMEOUT_SEC` | `30` | timeout ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make clear-cache`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `CACHE_PRIORITY` | `high` | `normal` ╨╕╨╗╨╕ `high` |
| `CACHE_TIMEOUT_SEC` | `5` | timeout ╨╛╨╢╨╕╨┤╨░╨╜╨╕╤П ╨╛╤З╨╡╤А╨╡╨┤╨╕ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make find-element`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `FIND_TYPE` | ╨┐╤Г╤Б╤В╨╛ | ╨║╨╛╤А╨╛╤В╨║╨╕╨╣ ╤В╨╕╨┐ ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░, ╨╜╨░╨┐╤А╨╕╨╝╨╡╤А `Button` |
| `FIND_TEXT` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `text` |
| `FIND_RESOURCE_ID` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `resource-id` |
| `FIND_CONTENT_DESC` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `content-desc` |
| `FIND_HINT` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `hint` |
| `FIND_MATCH` | `exact` | ╤А╨╡╨╢╨╕╨╝ ╨┐╨╛╨╕╤Б╨║╨░: `exact` ╨╕╨╗╨╕ `contains` |
| `FIND_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `FIND_TIMEOUT_SEC` | `30` | timeout ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Х╤Б╨╗╨╕ ╤Н╨╗╨╡╨╝╨╡╨╜╤В ╨╜╨╡ ╨╜╨░╨╣╨┤╨╡╨╜, `POST /find-element` ╨▓╨╛╨╖╨▓╤А╨░╤Й╨░╨╡╤В `404`: ╤Ж╨╡╨╗╤М ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨╜╨╡ ╨▓╤Л╨┐╨╛╨╗╨╜╨╡╨╜╨░.

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make wait-for-element`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `WAIT_TYPE` | ╨┐╤Г╤Б╤В╨╛ | ╨║╨╛╤А╨╛╤В╨║╨╕╨╣ ╤В╨╕╨┐ ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░, ╨╜╨░╨┐╤А╨╕╨╝╨╡╤А `Button` |
| `WAIT_TEXT` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `text` |
| `WAIT_RESOURCE_ID` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `resource-id` |
| `WAIT_CONTENT_DESC` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `content-desc` |
| `WAIT_HINT` | ╨┐╤Г╤Б╤В╨╛ | ╤В╨╛╤З╨╜╤Л╨╣ ╨╕╨╗╨╕ contains-╨┐╨╛╨╕╤Б╨║ ╨┐╨╛ `hint` |
| `WAIT_MATCH` | `exact` | ╤А╨╡╨╢╨╕╨╝ ╨┐╨╛╨╕╤Б╨║╨░: `exact` ╨╕╨╗╨╕ `contains` |
| `WAIT_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `WAIT_TIMEOUT_SEC` | `30` | ╤Б╨║╨╛╨╗╤М╨║╨╛ ╤Б╨╡╨║╤Г╨╜╨┤ ╨╢╨┤╨░╤В╤М ╨┐╨╛╤П╨▓╨╗╨╡╨╜╨╕╤П ╤Н╨╗╨╡╨╝╨╡╨╜╤В╨░ |
| `WAIT_CHECK_INTERVAL_MS` | `500` | ╨╕╨╜╤В╨╡╤А╨▓╨░╨╗ ╨┐╤А╨╛╨▓╨╡╤А╨║╨╕ UI dump; ╨╝╨╕╨╜╨╕╨╝╤Г╨╝ `100` |

╨Х╤Б╨╗╨╕ ╤Н╨╗╨╡╨╝╨╡╨╜╤В ╨╜╨╡ ╨┐╨╛╤П╨▓╨╕╨╗╤Б╤П ╨╖╨░ `WAIT_TIMEOUT_SEC`, `POST /wait-for-element` ╨▓╨╛╨╖╨▓╤А╨░╤Й╨░╨╡╤В `408`.

╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╤Л╨╡ ╤В╨╛╨╗╤М╨║╨╛ ╨┤╨╗╤П `make detect-state`:

| ╨Я╨╡╤А╨╡╨╝╨╡╨╜╨╜╨░╤П | ╨Я╨╛ ╤Г╨╝╨╛╨╗╤З╨░╨╜╨╕╤О | ╨Э╨░╨╖╨╜╨░╤З╨╡╨╜╨╕╨╡ |
|------------|--------------|------------|
| `DETECT_MODE` | `ui` | ╤А╨╡╨╢╨╕╨╝: `auto`, `ui` ╨╕╨╗╨╕ `vlm` |
| `DETECT_PLATFORM` | `android` | ╨┐╨╛╨┤╤Б╨║╨░╨╖╨║╨░ ╨┤╨╗╤П VLM: `instagram`, `tiktok`, `youtube`, `android` ╨╕ ╤В.╨┐. |
| `DETECT_USE_SCREENSHOT` | `true` | ╨┤╨╡╨╗╨░╤В╤М screenshot ╨┤╨╗╤П VLM-╨░╨╜╨░╨╗╨╕╨╖╨░ |
| `DETECT_STORE_SCREENSHOT` | `false` | ╤Б╨╛╤Е╤А╨░╨╜╤П╤В╤М screenshot ╨▓ MinIO ╨╕ ╨▓╨╡╤А╨╜╤Г╤В╤М ╤Б╤Б╤Л╨╗╨║╤Г |
| `DETECT_PRIORITY` | `normal` | `normal` ╨╕╨╗╨╕ `high` |
| `DETECT_TIMEOUT_SEC` | `30` | timeout ╨╖╨░╨┐╤А╨╛╤Б╨░ ╨▓ ╤Б╨╡╨║╤Г╨╜╨┤╨░╤Е |

╨Х╤Б╨╗╨╕ ╤Б╨╛╤Б╤В╨╛╤П╨╜╨╕╨╡ ╨╜╨╡ ╤А╨░╤Б╨┐╨╛╨╖╨╜╨░╨╜╨╛, `POST /detect-state` ╨▓╨╛╨╖╨▓╤А╨░╤Й╨░╨╡╤В `200` ╤Б╨╛ `state="unknown"` ╨╕ ╨┤╨╕╨░╨│╨╜╨╛╤Б╤В╨╕╨║╨╛╨╣ `description`, `elements`, `matched_signals`, `backend_used`.

╨Я╤А╨╕╨╝╨╡╤А ╨╗╨╛╨║╨░╨╗╤М╨╜╨╛╨│╨╛ ╨╖╨░╨┐╤Г╤Б╨║╨░ ╤Б ╨╜╨╡╤Б╤В╨░╨╜╨┤╨░╤А╤В╨╜╤Л╨╝╨╕ ╨┐╨╛╤А╤В╨░╨╝╨╕:

```bash
OBSERVER_HTTP_ADDR=127.0.0.1:19091 make run
```

## ╨а╨░╨▒╨╛╤В╨░ ╨▓ ╨║╨╛╨╝╨░╨╜╨┤╨╡

╨Т ╤А╨╡╨┐╨╛╨╖╨╕╤В╨╛╤А╨╕╨╕ ╨╕╤Б╨┐╨╛╨╗╤М╨╖╤Г╨╡╤В╤Б╤П GitHub Flow. ╨С╨░╨╖╨╛╨▓╨░╤П ╨▓╨╡╤В╨║╨░ ╨┐╤А╨╛╨╡╨║╤В╨░ ╤Б╨╡╨╣╤З╨░╤Б тАФ `master`.

1. ╨Ч╨░╨▒╨╡╤А╨╕ ╤Б╨▓╨╡╨╢╤Г╤О ╨▒╨░╨╖╨╛╨▓╤Г╤О ╨▓╨╡╤В╨║╤Г:

   ```bash
   git checkout master
   git pull origin master
   ```

2. ╨б╨╛╨╖╨┤╨░╨╣ ╨╛╤В╨┤╨╡╨╗╤М╨╜╤Г╤О ╨▓╨╡╤В╨║╤Г ╨┐╨╛╨┤ ╨╖╨░╨┤╨░╤З╤Г:

   ```bash
   git checkout -b feature/short-task-name
   ```

3. ╨Ф╨╡╨╗╨░╨╣ ╨╜╨╡╨▒╨╛╨╗╤М╤И╨╕╨╡ ╨║╨╛╨╝╨╝╨╕╤В╤Л ╨▓ ╤Д╨╛╤А╨╝╨░╤В╨╡ Conventional Commits:

   ```bash
   git commit -m "feat: add observer health checks"
   git commit -m "fix: handle empty ui dump"
   git commit -m "docs: describe team workflow"
   ```

4. ╨Я╨╡╤А╨╡╨┤ Pull Request ╨╖╨░╨┐╤Г╤Б╤В╨╕ ╨┐╤А╨╛╨▓╨╡╤А╨║╨╕:

   ```bash
   make fmt
   make check
   ```

5. ╨Ч╨░╨┐╤Г╤И╤М feature-╨▓╨╡╤В╨║╤Г ╨╕ ╨╛╤В╨║╤А╨╛╨╣ Pull Request ╨▓ `master`:

   ```bash
   git push -u origin feature/short-task-name
   ```

6. ╨Т Pull Request ╨║╤А╨░╤В╨║╨╛ ╨╛╨┐╨╕╤И╨╕, ╤З╤В╨╛ ╨╕╨╖╨╝╨╡╨╜╨╕╨╗╨╛╤Б╤М, ╨║╨░╨║ ╨┐╤А╨╛╨▓╨╡╤А╤П╨╗╨╛╤Б╤М ╨╕ ╨╡╤Б╤В╤М ╨╗╨╕ ╤А╨╕╤Б╨║╨╕.

## ╨Я╤А╨░╨▓╨╕╨╗╨░ ╨║╨╛╨╝╨░╨╜╨┤╤Л

- ╨Э╨╡ ╨┐╤Г╤И╨╕╤В╤М ╨╜╨░╨┐╤А╤П╨╝╤Г╤О ╨▓ `master`; ╨╕╨╖╨╝╨╡╨╜╨╡╨╜╨╕╤П ╨┐╨╛╨┐╨░╨┤╨░╤О╤В ╤В╤Г╨┤╨░ ╤В╨╛╨╗╤М╨║╨╛ ╤З╨╡╤А╨╡╨╖ Pull Request.
- ╨Ф╨╡╤А╨╢╨░╤В╤М ╨▓╨╡╤В╨║╨╕ ╨║╨╛╤А╨╛╤В╨║╨╕╨╝╨╕ ╨╕ ╤А╨╡╨│╤Г╨╗╤П╤А╨╜╨╛ ╤Б╨╕╨╜╤Е╤А╨╛╨╜╨╕╨╖╨╕╤А╨╛╨▓╨░╤В╤М╤Б╤П ╤Б `origin/master`.
- ╨Ф╨╗╤П ╨╗╨╕╤З╨╜╨╛╨╣ ╨▓╨╡╤В╨║╨╕ ╨╕╤Б╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╤М `git rebase origin/master`; ╨┤╨╗╤П ╨╛╨▒╤Й╨╡╨╣ ╨▓╨╡╤В╨║╨╕ ╨▒╨╡╨╖╨╛╨┐╨░╤Б╨╜╨╡╨╡ `git merge origin/master`.
- ╨Э╨╡ ╨║╨╛╨╝╨╝╨╕╤В╨╕╤В╤М `.env`, ╨║╨╗╤О╤З╨╕, ╤В╨╛╨║╨╡╨╜╤Л ╨╕ MinIO credentials.
- ╨Э╨╡ ╨╝╨╡╨╜╤П╤В╤М protobuf ╨▒╨╡╨╖ ╤Б╨╛╨│╨╗╨░╤Б╨╛╨▓╨░╨╜╨╕╤П ╤Б ╨┐╨╛╤В╤А╨╡╨▒╨╕╤В╨╡╨╗╤П╨╝╨╕ API.
- ╨Э╨╡ ╨┤╨╛╨▒╨░╨▓╨╗╤П╤В╤М tap/swipe/text-╨╗╨╛╨│╨╕╨║╤Г ╨▓ observer: ╨┤╨╡╨╣╤Б╤В╨▓╨╕╤П ╨┐╤А╨╕╨╜╨░╨┤╨╗╨╡╨╢╨░╤В executor-╤Б╨╡╤А╨▓╨╕╤Б╤Г.
- ╨Э╨╡ ╤Г╨┐╤А╨░╨▓╨╗╤П╤В╤М ADB connect/forward ╨▓ observer: ╤Н╤В╨╛ ╨╖╨╛╨╜╨░ connector-╤Б╨╡╤А╨▓╨╕╤Б╨░.
- ╨Х╤Б╨╗╨╕ ╨╝╨╡╨╜╤П╨╡╤В╤Б╤П ╨┐╨╛╨▓╨╡╨┤╨╡╨╜╨╕╨╡ ╤Б╨╡╤А╨▓╨╕╤Б╨░, ╨╛╨▒╨╜╨╛╨▓╨╗╤П╤В╤М README, AGENTS.md ╨╕╨╗╨╕ CLAUDE.md ╤А╤П╨┤╨╛╨╝ ╤Б ╨║╨╛╨┤╨╛╨╝.

## ╨б╨╕╨╜╤Е╤А╨╛╨╜╨╕╨╖╨░╤Ж╨╕╤П ╨▓╨╡╤В╨║╨╕

```bash
git fetch origin
git rebase origin/master
make check
```

╨Я╨╛╤Б╨╗╨╡ rebase ╨┐╤Г╤И╨╕╤В╤М ╤В╨╛╨╗╤М╨║╨╛ ╤Б╨▓╨╛╤О feature-╨▓╨╡╤В╨║╤Г:

```bash
git push --force-with-lease
```

`--force-with-lease` ╨╕╤Б╨┐╨╛╨╗╤М╨╖╨╛╨▓╨░╤В╤М ╤В╨╛╨╗╤М╨║╨╛ ╨┐╨╛╤Б╨╗╨╡ rebase ╤Б╨▓╨╛╨╡╨╣ ╨▓╨╡╤В╨║╨╕. ╨Ф╨╗╤П `master` ╨┐╤А╤П╨╝╨╛╨╣ push ╨╖╨░╨┐╤А╨╡╤Й╤С╨╜.
