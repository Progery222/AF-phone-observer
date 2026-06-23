# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /phone-observer ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /phone-observer /phone-observer
EXPOSE 50053 9090
USER nonroot:nonroot
ENTRYPOINT ["/phone-observer"]
