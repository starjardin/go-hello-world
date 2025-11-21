# Makefile for database migrations and common tasks

.PHONY: migrate-up migrate-down migrate-status build run clean help

# Database migration commands
migrate-up:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back database migrations..."
	go run cmd/migrate/main.go down

migrate-status:
	@echo "Checking migration status..."
	go run cmd/migrate/main.go status

# Application commands
build:
	@echo "Building application..."
	go build -o bin/hello-world .

run:
	@echo "Running application..."
	go run main.go

# Docker commands
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Development commands
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	@echo "Development environment ready!"

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	docker-compose down --volumes --remove-orphans

# Help
help:
	@echo "Available commands:"
	@echo "  migrate-up      - Apply database migrations"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  migrate-status  - Check migration status"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application locally"
	@echo "  docker-up       - Start services with Docker Compose"
	@echo "  docker-down     - Stop Docker services"
	@echo "  docker-logs     - Show Docker logs"
	@echo "  dev-setup       - Set up development environment"
	@echo "  clean           - Clean up build artifacts and Docker volumes"
	@echo "  help            - Show this help message"