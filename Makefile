run:
	go run ./cmd/server/main.go

migrate-up:
	go run ./cmd/migrate/main.go -direction=up

migrate-down:
	go run ./cmd/migrate/main.go -direction=down

migrate-force:
	go run ./cmd/migrate -direction=force -version=1

integration-test:
	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/comments-service?sslmode=disable" \
    go test -v -cover ./internal/repository/postgresRepository/...