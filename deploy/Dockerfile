# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /phone-observer ./cmd/server

# Runtime needs adb/uiautomator: observer shells out to `adb` for
# screencap and `uiautomator dump`. distroless lacks adb, so DumpUI/
# CaptureScreen failed. Use alpine + android-tools.
FROM alpine:3.20
RUN apk add --no-cache android-tools ca-certificates
COPY --from=builder /phone-observer /usr/local/bin/phone-observer
EXPOSE 50053 9090
ENTRYPOINT ["/usr/local/bin/phone-observer"]
