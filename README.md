# AF-phone-observer

Микросервис наблюдения за экраном Android: скриншоты, UI-dump (uiautomator), dumpsys.

- gRPC API: CaptureScreenshot, DumpUI, DetectState
- MinIO для хранения скриншотов
- Hexagonal architecture (Go 1.22+)
- Health/ready на `:9090`

## Запуск

```bash
go mod tidy
go test ./...
go run ./cmd/server
```

Env: `GRPC_ADDR` (`:50053`), `HEALTH_ADDR` (`:9090`), `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `LOG_LEVEL`.
