# Makefile для Subscription Service

# Переменные
BINARY_NAME=subscription-service
DOCKER_IMAGE=subscription-service:latest
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')
MIGRATION_FILES=$(shell find ./migrations -name '*.sql')

.PHONY: all build test clean run docker-build docker-run migrate help

all: test build

# Сборка
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/server

# Запуск
run: build
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

# Тестирование
test:
	@echo "Running tests..."
	@go test -v ./... -cover

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Миграции
migrate:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go

# Линтинг
lint:
	@echo "Running linter..."
	@golangci-lint run

# тесты
test-unit:
	@echo "Running unit tests..."
	@go test ./internal/service ./internal/handler -v

test-repository:
	@echo "Running repository tests..."
	@go test ./internal/repository -v

test-integration:
	@echo "Running integration tests..."
	@echo "Note: These tests require running PostgreSQL"
	@go test ./... -tags=integration -v

test-quick:
	@echo "Быстрый тест API..."
	@powershell -File quick-test.ps1

# Запуск всех тестов
test-all: test test-api
	@echo "Все тесты пройдены! ✅"

# Форматирование
fmt:
	@echo "Formatting code..."
	@gofmt -w $(GO_FILES)

# Очистка
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# Docker
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE)

docker-compose-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up --build

docker-compose-down:
	@echo "Stopping services..."
	@docker-compose down

# База данных
db-shell:
	@echo "Connecting to PostgreSQL..."
	@docker exec -it subscription-postgres psql -U postgres -d subscriptions

db-logs:
	@echo "Showing PostgreSQL logs..."
	@docker logs subscription-postgres

# Генерация документации
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs

# Помощь
help:
	@echo "Available commands:"
	@echo "  make build           - Build the application"
	@echo "  make run             - Build and run the application"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make lint            - Run linter"
	@echo "  make fmt             - Format code"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Run Docker container"
	@echo "  make docker-compose-up   - Start with Docker Compose"
	@echo "  make docker-compose-down - Stop Docker Compose"
	@echo "  make db-shell        - Connect to PostgreSQL shell"
	@echo "  make db-logs         - Show PostgreSQL logs"
	@echo "  make swagger         - Generate Swagger docs"
	@echo "  make migrate         - Run database migrations"
	@echo "  make help            - Show this help"