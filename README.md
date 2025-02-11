# User Management API

## Overview
A robust, secure, and scalable User Management API built with Go, designed to provide comprehensive user authentication and management functionalities.

## 🌟 Features
- **User Registration**: Secure user account creation
- **Authentication**: JWT-based token authentication
- **Token Management**:
  - Access token generation
  - Refresh token mechanism
  - Token blacklisting
  - Secure logout
- **User Management**: CRUD operations for user profiles
- **Secure Password Handling**: Bcrypt password hashing
- **Logging**: Advanced structured logging with zerolog
- **Database**: SQLite-based persistent storage
- **Environment Configuration**: Flexible .env-based configuration

## 🛠 Tech Stack
- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: SQLite
- **Logging**: Zerolog
- **Authentication**: JWT
- **Environment**: godotenv

## 🔐 Security Features
- Bcrypt password hashing
- JWT token-based authentication
- Token blacklisting mechanism
- Refresh token support
- Environment-based configuration management
- Middleware-based route protection

## 🏗 Authentication Workflow
1. **Registration**: Create a new user account
2. **Login**: 
   - Receive access and refresh tokens
   - Tokens include user claims
3. **Token Refresh**: 
   - Use refresh token to generate new access token
4. **Logout**:
   - Blacklist current access token
   - Prevent further token usage

## 📦 Project Structure
```
.
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # Request middleware
│   └── repository/       # Data access layer
├── pkg/
│   ├── logger/           # Logging utility
│   └── utils/            # Utility functions
└── .env                  # Environment configuration
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- SQLite

### Installation
1. Clone the repository
2. Copy `.env.example` to `.env`
3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Configuration
Configure your environment variables in `.env`:
- `DB_PATH`: SQLite database path
- `SERVER_PORT`: API server port
- `JWT_SECRET`: Secret for token generation

### Running the Application
```bash
go run cmd/server/main.go
```

### Running Tests
```bash
go test ./...
```

## 🔍 API Endpoints

### Authentication
- `POST /auth/register`: Create new user account
- `POST /auth/login`: User login, returns JWT tokens
- `POST /auth/refresh`: Refresh access token
- `POST /auth/logout`: Invalidate current access token

### User Management (Protected)
- `GET /users`: List all users
- `GET /users/:id`: Get user by ID
- `PUT /users/:id`: Update user profile
- `DELETE /users/:id`: Delete user account

## 📝 Logging
The application uses zerolog for structured, high-performance logging:
- Supports multiple log levels (Debug, Info, Error)
- Configurable log outputs (console, file)
- Includes contextual information and stack traces

## 🔧 Token Blacklisting
- In-memory thread-safe token blacklist
- Tokens are invalidated upon logout
- Middleware prevents use of blacklisted tokens
- Automatic expiration of blacklisted tokens

## 🤝 Contributing
1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License
[Specify your license here]

## 💬 Contact
[Your contact information]