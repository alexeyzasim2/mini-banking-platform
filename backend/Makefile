install:
	go mod download

run:
	go run cmd/main.go

test:
	go test -v ./...

# Start test database container
test-db-up:
	cd .. && docker-compose up -d postgres_test

# Apply migrations to test database
test-db-migrate:
	DB_HOST=localhost DB_PORT=55432 DB_USER=postgres DB_PASSWORD=Kd8mPq2Ln5Rv9Xt3Yw7Bz DB_NAME=banking_platform_test \
		goose -dir migrations postgres "postgres://postgres:Kd8mPq2Ln5Rv9Xt3Yw7Bz@localhost:55432/banking_platform_test?sslmode=disable" up

# Run service tests against test database
test-service:
	DB_HOST=localhost DB_PORT=55432 DB_USER=postgres DB_PASSWORD=Kd8mPq2Ln5Rv9Xt3Yw7Bz DB_NAME=banking_platform_test go test -v ./internal/service/...

# Stop test database container
test-db-down:
	cd .. && docker-compose stop postgres_test

# Remove test database container and volume (clean slate)
test-db-clean:
	cd .. && docker-compose down postgres_test -v

migrate-up:
	goose -dir migrations postgres "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

