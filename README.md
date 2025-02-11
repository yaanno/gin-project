# User Management API

## Prerequisites
- Go 1.21+
- PostgreSQL

## Setup
1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your configuration
3. Install dependencies: `go mod tidy`
4. Run migrations: `go run cmd/server/main.go`

## Environment Variables
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSLMODE`: SSL mode for database connection
- `SERVER_PORT`: API server port
- `JWT_SECRET`: Secret for access token generation
- `JWT_REFRESH_SECRET`: Secret for refresh token generation

## Features
- User registration
- JWT Authentication
- Token refresh
- CRUD operations for users