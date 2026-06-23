GO ?= go
PROTOC ?= protoc
DOCKER ?= docker
GOLANGCI_LINT ?= golangci-lint

MODULE := github.com/mobilefarm/af/phone-observer
MAIN := ./cmd/server
BINARY ?= phone-observer
IMAGE ?= af-phone-observer:latest
PROTO_FILES := proto/common/v1/phone.proto proto/observer/v1/observer.proto

.PHONY: deps tidy fmt vet test lint lint-fix build build-bin run check proto docker-build

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

run:
	$(GO) run $(MAIN)

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
