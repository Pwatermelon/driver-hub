# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /out/driver-hub ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/driver-hub /app/driver-hub
COPY migrations /app/migrations

ENV MIGRATIONS_PATH=/app/migrations/001_init.sql
ENV HTTP_ADDR=:8080
EXPOSE 8080
CMD ["/app/driver-hub"]
