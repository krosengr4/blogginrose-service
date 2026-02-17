# rosenblog-service

REST API backend for the BlogginRose blogging platform — built with Go, PostgreSQL, and JWT authentication.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Features](#features)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)
- [Questions](#questions)

## Installation

**Prerequisites**: Go 1.25+, PostgreSQL 16, Docker & Docker Compose

```bash
git clone git@github.com:krosengr4/rosenblog-service.git
cd rosenblog-service
go mod download
```

## Quick Start

```bash
# 1. Start PostgreSQL (schema auto-loaded)
docker-compose up -d

# 2. Copy and configure environment
cp .env .env.local

# 3. Run the server
go run ./cmd/server/main.go
# Server available at http://localhost:8082
```

## Features

- **Blog post CRUD**: Create, read, update, and delete posts with auto-generated URL slugs
- **Full-text search**: Case-insensitive search across title, content, and tags
- **Tag management**: Aggregate unique tags across all posts using PostgreSQL GIN index
- **JWT authentication**: 24-hour token expiry with bcrypt password hashing (cost 14)
- **CORS middleware**: Configurable allowed origins with credentials support
- **Structured logging**: JSON request/response logs with timing via zerolog
- **Panic recovery**: Middleware catches panics and logs stack traces

## Usage

### Configuration

Copy `.env` and set the following variables:

```env
# Server
PORT=8082
FRONTEND_URL=http://localhost:3000
ALLOWED_ORIGINS=http://localhost:3000

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5444
POSTGRES_DB=blogginrose-db
POSTGRES_USER=postgres
POSTGRES_PASSWORD_FILE=postgres-password
POSTGRES_SSL_MODE=disable

# Secrets (files read from SECRETS_PATH directory)
SECRETS_PATH=/path/to/secrets
ADMIN_USERNAME=kros
ADMIN_PASSWORD_FILE=admin-password
JWT_SECRET_FILE=blogginrose-jwt

# Logging
LOG_LEVEL=info
```

Sensitive values (DB password, admin password, JWT secret) are read from files under `SECRETS_PATH`.

### API Endpoints

#### Public

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/posts` | Get all posts (ordered by date DESC) |
| `GET` | `/api/posts/{slug}` | Get a post by slug |
| `GET` | `/api/posts/search?q={query}` | Search posts (max 20 results) |
| `GET` | `/api/tags` | Get all unique tags |
| `POST` | `/api/login` | Log in and receive a JWT |

#### Admin (JWT Required)

All admin routes require `Authorization: Bearer <token>`.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/admin/posts` | Create a new post |
| `PUT` | `/api/admin/posts/{postID}` | Update a post |
| `DELETE` | `/api/admin/posts/{postID}` | Delete a post |

### Example Requests

**Login**
```bash
curl -X POST http://localhost:8082/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "kros", "password": "yourpassword"}'
```

**Create a post**
```bash
curl -X POST http://localhost:8082/api/admin/posts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "content": "# My first post\n\nContent here.",
    "tags": ["go", "backend"],
    "author": "kros"
  }'
```

**Search posts**
```bash
curl "http://localhost:8082/api/posts/search?q=golang"
```

### Project Structure

```
rosenblog-service/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/appconfig.go     # Config loading & validation
│   ├── handler/
│   │   ├── handlers.go         # Public handlers
│   │   └── admin.go            # Admin handlers
│   ├── middleware/
│   │   ├── auth.go             # JWT validation
│   │   ├── cors.go             # CORS headers
│   │   ├── logging.go          # Request logging
│   │   └── recovery.go         # Panic recovery
│   ├── model/models.go         # Post data model
│   └── repository/database.go  # Database access layer
├── schema.sql                  # PostgreSQL schema
├── docker-compose.yml          # Local dev setup
└── Dockerfile                  # Container image
```

### Building

```bash
# Binary
go build -o bin/server ./cmd/server/main.go

# Docker image
docker build -t rosenblog-service .
```

## Contributing

**Please contribute to this project:**

- [Submit Bugs and Request Features you'd like to see Implemented](https://github.com/krosengr4/rosenblog-service/issues)

## License

MIT

## Questions

- [Link to my GitHub Profile](https://github.com/krosengr4)

- For any additional questions, email me at rosenkev4@gmail.com
