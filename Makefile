GO ?= go
PROTOC ?= protoc
DOCKER ?= docker
GOLANGCI_LINT_VERSION ?= v2.4.0
GOLANGCI_LINT ?= $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

MODULE := github.com/mobilefarm/af/phone-observer
MAIN := ./cmd/server
BINARY ?= phone-observer
IMAGE ?= af-phone-observer:latest
PROTO_FILES := proto/common/v1/phone.proto proto/observer/v1/observer.proto
OBSERVER_HTTP_ADDR ?= 127.0.0.1:19090
OBSERVER_HTTP_URL ?= http://$(OBSERVER_HTTP_ADDR)
OBSERVER_AUTO_START ?= true
HEALTH_ADDR ?= $(OBSERVER_HTTP_ADDR)
SCREENSHOT_PRIORITY ?= normal
SCREENSHOT_TIMEOUT_SEC ?= 10
SCREENSHOT_STORE_IN_MINIO ?= true
SCREEN_PRIORITY ?= normal
SCREEN_TIMEOUT_SEC ?= 10
DUMP_UI_FORMAT ?= json
DUMP_UI_PRIORITY ?= normal
DUMP_UI_TIMEOUT_SEC ?= 30
UI_FORMAT ?= json
UI_PRIORITY ?= normal
UI_TIMEOUT_SEC ?= 30
CACHE_PRIORITY ?= high
CACHE_TIMEOUT_SEC ?= 5
FIND_TYPE ?=
FIND_TEXT ?=
FIND_RESOURCE_ID ?=
FIND_CONTENT_DESC ?=
FIND_HINT ?=
FIND_MATCH ?= exact
FIND_PRIORITY ?= normal
FIND_TIMEOUT_SEC ?= 30
WAIT_TYPE ?=
WAIT_TEXT ?=
WAIT_RESOURCE_ID ?=
WAIT_CONTENT_DESC ?=
WAIT_HINT ?=
WAIT_MATCH ?= exact
WAIT_PRIORITY ?= normal
WAIT_TIMEOUT_SEC ?= 30
WAIT_CHECK_INTERVAL_MS ?= 500
DETECT_MODE ?= ui
DETECT_PLATFORM ?= android
DETECT_USE_SCREENSHOT ?= true
DETECT_STORE_SCREENSHOT ?= false
DETECT_PRIORITY ?= normal
DETECT_TIMEOUT_SEC ?= 30
PHONE_SERIAL ?= $(shell $(GO) run ./cmd/adbserial -mode=first 2>/dev/null)
SERIAL ?= $(PHONE_SERIAL)

.PHONY: deps tidy fmt vet test lint lint-fix build build-bin run adb-devices adb-serial phone-screenshot phone-dump-ui phone-screen phone-ui phone-clear-cache phone-find-element phone-wait-for-element phone-detect-state screenshot dump-ui screen ui clear-cache find-element wait-for-element detect-state check proto docker-build

deps:
	$(GO) mod download

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

build:
	$(GO) build ./...

build-bin:
	$(GO) build -o $(BINARY)$(shell $(GO) env GOEXE) $(MAIN)

run: export HEALTH_ADDR := $(HEALTH_ADDR)
run:
	$(GO) run $(MAIN)

adb-devices:
	adb devices

adb-serial:
	$(GO) run ./cmd/adbserial -mode=first

phone-screenshot:
	$(MAKE) screenshot SERIAL="$(PHONE_SERIAL)"

phone-dump-ui:
	$(MAKE) dump-ui SERIAL="$(PHONE_SERIAL)"

phone-screen:
	$(MAKE) screen SERIAL="$(PHONE_SERIAL)"

phone-ui:
	$(MAKE) ui SERIAL="$(PHONE_SERIAL)"

phone-clear-cache:
	$(MAKE) clear-cache SERIAL="$(PHONE_SERIAL)"

phone-find-element:
	$(MAKE) find-element SERIAL="$(PHONE_SERIAL)"

phone-wait-for-element:
	$(MAKE) wait-for-element SERIAL="$(PHONE_SERIAL)"

phone-detect-state:
	$(GO) run ./cmd/detectstate -url="$(OBSERVER_HTTP_URL)" -serial="$(PHONE_SERIAL)" -mode="$(DETECT_MODE)" -platform="$(DETECT_PLATFORM)" -use-screenshot="$(DETECT_USE_SCREENSHOT)" -store-screenshot="$(DETECT_STORE_SCREENSHOT)" -priority="$(DETECT_PRIORITY)" -timeout-sec="$(DETECT_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

screenshot:
	$(GO) run ./cmd/screenshot -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -priority="$(SCREENSHOT_PRIORITY)" -timeout-sec="$(SCREENSHOT_TIMEOUT_SEC)" -store-in-minio="$(SCREENSHOT_STORE_IN_MINIO)" -auto-start="$(OBSERVER_AUTO_START)"

dump-ui:
	$(GO) run ./cmd/dumpui -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -format="$(DUMP_UI_FORMAT)" -priority="$(DUMP_UI_PRIORITY)" -timeout-sec="$(DUMP_UI_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

screen:
	$(GO) run ./cmd/screen -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -priority="$(SCREEN_PRIORITY)" -timeout-sec="$(SCREEN_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

ui:
	$(GO) run ./cmd/ui -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -format="$(UI_FORMAT)" -priority="$(UI_PRIORITY)" -timeout-sec="$(UI_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

clear-cache:
	$(GO) run ./cmd/clearcache -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -priority="$(CACHE_PRIORITY)" -timeout-sec="$(CACHE_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

find-element:
	$(GO) run ./cmd/findelement -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -type="$(FIND_TYPE)" -text="$(FIND_TEXT)" -resource-id="$(FIND_RESOURCE_ID)" -content-desc="$(FIND_CONTENT_DESC)" -hint="$(FIND_HINT)" -match="$(FIND_MATCH)" -priority="$(FIND_PRIORITY)" -timeout-sec="$(FIND_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

wait-for-element:
	$(GO) run ./cmd/waitforelement -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -type="$(WAIT_TYPE)" -text="$(WAIT_TEXT)" -resource-id="$(WAIT_RESOURCE_ID)" -content-desc="$(WAIT_CONTENT_DESC)" -hint="$(WAIT_HINT)" -match="$(WAIT_MATCH)" -priority="$(WAIT_PRIORITY)" -timeout-sec="$(WAIT_TIMEOUT_SEC)" -check-interval-ms="$(WAIT_CHECK_INTERVAL_MS)" -auto-start="$(OBSERVER_AUTO_START)"

detect-state:
	$(GO) run ./cmd/detectstate -url="$(OBSERVER_HTTP_URL)" -serial="$(SERIAL)" -mode="$(DETECT_MODE)" -platform="$(DETECT_PLATFORM)" -use-screenshot="$(DETECT_USE_SCREENSHOT)" -store-screenshot="$(DETECT_STORE_SCREENSHOT)" -priority="$(DETECT_PRIORITY)" -timeout-sec="$(DETECT_TIMEOUT_SEC)" -auto-start="$(OBSERVER_AUTO_START)"

check: vet test build

lint:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

proto:
	$(PROTOC) -I proto \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

docker-build:
	$(DOCKER) build -f deploy/Dockerfile -t $(IMAGE) .
