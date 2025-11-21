# Hello World Go Application

A simple Go web application with PostgreSQL database and manual migration support.

## Prerequisites

- Go 1.19+
- Docker and Docker Compose
- PostgreSQL (if running without Docker)

## Quick Start

### Using Docker Compose (Recommended)

1. Start the application and database:
```bash
make docker-up
# or
docker-compose up --build
```

2. The application will be available at `http://localhost:8080/hello`

### Manual Setup

1. Set up the development environment:
```bash
make dev-setup
```

2. Set the database URL environment variable:
```bash
export DB_URL="postgres://user:password@localhost:5432/hello_world?sslmode=disable"
```

3. Run database migrations:
```bash
make migrate-up
```

4. Start the application:
```bash
make run
```

## Database Migrations

This project includes a custom migration system for managing database schema changes.

### Migration Commands

- **Apply migrations**: `make migrate-up`
- **Rollback migrations**: `make migrate-down`
- **Check migration status**: `make migrate-status`

### Manual Migration Usage

You can also run migrations directly:

```bash
# Apply migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Check status
go run cmd/migrate/main.go status
```

### How It Works

1. The migration system creates a `schema_migrations` table to track applied migrations
2. It reads the `db/schema.sql` file and applies it as migration `001_initial_schema`
3. Migrations are run in transactions to ensure consistency
4. The system prevents duplicate applications of the same migration

### Environment Variables

- `DB_URL` - PostgreSQL connection string (required)

Example:
```bash
export DB_URL="postgres://username:password@hostname:port/database?sslmode=disable"
```

## API Endpoints

- `GET /hello` - Returns "Hello, World!" message from the database

## Development

### Building

```bash
make build
```

### Running Tests

```bash
go test ./...
```

### Docker Commands

- Start services: `make docker-up`
- Stop services: `make docker-down`
- View logs: `make docker-logs`
- Clean up: `make clean`

## Project Structure

```
.
├── cmd/
│   └── migrate/
│       └── main.go          # Migration runner
├── db/
│   ├── generated/           # SQLC generated code
│   │   ├── db.go
│   │   ├── models.go
│   │   └── query.sql.go
│   └── schema.sql           # Database schema
├── sqlc/
│   └── query.sql           # SQL queries for SQLC
├── docker-compose.yaml     # Docker Compose configuration
├── Dockerfile             # Application Dockerfile
├── go.mod                 # Go module definition
├── main.go               # Main application
├── Makefile              # Build and migration commands
└── sqlc.yaml            # SQLC configuration
```

## Technology Stack

- **Language**: Go
- **Database**: PostgreSQL
- **SQL Generation**: SQLC
- **Containerization**: Docker
- **Database Driver**: pgx/v5