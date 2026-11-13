
ARG GO_VERSION=1.26

# Сборка приложения
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /app/server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /app/migrate ./cmd/migrate

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate

COPY --from=builder /app/migrations /app/migrations

USER app

EXPOSE 8080

CMD ["/app/server"]
