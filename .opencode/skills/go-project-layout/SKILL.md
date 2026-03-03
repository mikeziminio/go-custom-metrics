---
name: go-project-layout
description: Skill for creating clean Go project structure with proper package organization and internal/model folder usage
---

# Go Project Layout Skill

This skill defines a comprehensive and standardized Go project layout following the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) guidelines, incorporating best practices from the Go community and real-world examples such as the go-loyalty-system project.

## Project Structure Overview

```
project-name/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .gitignore
├── .golangci.yml
├── .editorconfig
├── .gitattributes
├── README.md
├── LICENSE
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── cmd/
│   ├── application-name/
│   │   └── main.go
│   └── another-application/
│       └── main.go
├── internal/
│   ├── model/                 # Shared models - create only when needed
│   │   ├── user.go
│   │   ├── order.go
│   │   └── withdrawal.go
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── server/
│   │   ├── handlers/
│   │   │   ├── user_handler.go
│   │   │   └── order_handler.go
│   │   ├── middleware/
│   │   │   ├── auth_middleware.go
│   │   │   └── logging_middleware.go
│   │   ├── routes/
│   │   │   ├── routes.go
│   │   │   └── user_routes.go
│   │   └── server.go
│   ├── db/
│   │   ├── database.go
│   │   ├── migrations/        # Database migrations - create only when needed
│   │   │   ├── 000001_init.up.sql
│   │   │   └── 000001_init.down.sql
│   │   └── models/
│   │       ├── user.go
│   │       └── order.go
│   ├── clients/
│   │   ├── http_client.go
│   │   └── external_service/
│   │       ├── service_client.go
│   │       └── service_client_test.go
│   └── log/
│       └── logger.go
├── pkg/
│   ├── utils/
│   │   ├── helpers.go
│   │   └── validators.go
│   └── api/
│       ├── v1/
│       │   ├── client.go
│       │   └── types.go
│       └── v2/
│           ├── client.go
│           └── types.go
├── docs/
│   ├── architecture.md
│   └── api.md
├── scripts/
│   └── build.sh
├── test/
│   ├── integration/
│   └── unit/
├── assets/
│   └── static/
└── examples/
    └── usage.go
```

## Key Directories and Their Purpose

### 1. Root Directory Files
- `go.mod` - Go module definition
- `go.sum` - Go module checksums
- `README.md` - Project documentation
- `.gitignore` - Git ignore patterns
- `LICENSE` - Project license
- `Makefile` - Build automation
- `Dockerfile` - Container build configuration
- `docker-compose.yml` - Multi-container docker setup
- `.golangci.yml` - Linter configuration
- `.editorconfig` - Editor configuration
- `.gitattributes` - Git attributes

### 2. `cmd/` Directory
Contains the main applications in the project. Each executable should have its own subdirectory:
- `cmd/application-name/` - Main entry point for the application
- `main.go` - Entry point with `func main()`

### 3. `internal/` Directory (Most Important)
This directory contains private application and library code that should not be imported by external projects. **All directories inside `internal/` should be created progressively as they are needed during project development**:

#### `internal/model/`
This folder is essential for sharing data structures across packages. All domain models should be defined here:
- `user.go` - User model and related constants/errors
- `order.go` - Order model and related constants/errors
- `withdrawal.go` - Withdrawal model and related constants/errors
- Shared business logic types that are used throughout the application

> ⚠️ **Note**: Only create directories in `internal/` when they are actually needed. Don't create empty directories that serve no purpose yet. Create them progressively as your project evolves.

#### `internal/config/`
Configuration loading and management:
- `config.go` - Configuration structure and loading logic
- `config_test.go` - Configuration tests

#### `internal/server/`
HTTP server implementation:
- `handlers/` - HTTP endpoint handlers
- `middleware/` - HTTP request/response middleware
- `routes/` - Route definitions and routing logic
- `server.go` - Server initialization and startup

#### `internal/db/`
Database-related code:
- `database.go` - Database connection management
- `migrations/` - Database migration files (create only when needed)
- `models/` - Database-specific models (separate from domain models)

#### `internal/clients/`
External service clients:
- HTTP clients, API clients, and service connectors
- Example: `clients/accrual/service_client.go`

#### `internal/log/`
Logging utility setup and configuration

### 4. `pkg/` Directory
Contains public libraries that can be imported by external projects:
- `pkg/utils/` - Utility functions
- `pkg/api/` - Public APIs for external consumption

### 5. `docs/` Directory
Documentation files:
- Architecture diagrams
- API documentation
- Technical specifications

### 6. `scripts/` Directory
Build scripts, deployment scripts, and utilities:
- `build.sh` - Build automation
- Deployment scripts

### 7. `test/` Directory
Test files organized by type:
- `test/integration/` - Integration tests
- `test/unit/` - Unit tests

### 8. `examples/` Directory
Example usage files demonstrating how to use the project

### 9. `assets/` Directory
Static assets such as:
- CSS, JavaScript, images
- Templates
- Static files served by the application

## Special Requirements

### The Internal Model Folder (Required)
The `internal/model/` directory is **essential** for any Go project using this layout because:

1. **Shared Data Models**: All domain entities are defined here to ensure consistency across the application
2. **Prevents Circular Dependencies**: Prevents circular imports between packages by providing a single source of truth for models
3. **Business Logic Separation**: Keeps business domain models separate from infrastructure concerns
4. **Reusability**: Models defined here can be used by handlers, services, database layers, etc.

> ⚠️ **Note**: Only create directories in `internal/` when they are actually needed. Don't create empty directories that serve no purpose yet.

**Example structure for internal/model:**
```go
// internal/model/user.go
package model

import (
    "errors"
    "time"
)

var (
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrUserNotFound      = errors.New("user not found")
)

type User struct {
    ID       int
    Login    string
    Password string
    Token    string
    Balance  float64
    Withdrawn float64
}

type Withdrawal struct {
    OrderID     string    `json:"order"`
    Sum         float64   `json:"sum"`
    ProcessedAt time.Time `json:"processed_at"`
}
```

### Database Migrations Directory
The `internal/db/migrations/` directory should be created when database migrations are needed:
- Contains SQL migration files for database schema changes
- Follows timestamp or sequential numbering convention
- Includes both up and down migration files

### Module Structure Compliance
- Uses Go modules (`go.mod`)
- Follows Go 1.26+ standards
- Uses semantic versioning for packages
- Proper dependency management

## Best Practices

### 1. Package Organization
- Use descriptive package names
- Keep packages small and focused
- Group related functionality together
- Follow the principle of "one responsibility per package"

### 2. Code Separation
- **internal/** - Private application code, not importable by external projects
- **pkg/** - Public libraries that can be imported by external projects
- **cmd/** - Executable applications

### 3. Import Guidelines
- Use full import paths for external dependencies
- Import groups separated by blank lines
- Organize imports alphabetically within each group

### 4. Testing Strategy
- Separate unit tests from integration tests
- Use table-driven tests where appropriate
- Include both happy path and edge case tests
- Test error conditions explicitly

### 5. Documentation Standards
- Document exported functions and types with comments
- Maintain a README.md with setup instructions
- Document configuration options
- Include examples where appropriate

## Implementation Notes

When implementing this structure:

1. **Always place models in `internal/model/`** - This is critical for preventing circular dependencies
2. **Follow Go naming conventions** - Use camelCase for variables and PascalCase for exported identifiers
3. **Use meaningful directory and file names** - Names should clearly indicate their purpose
4. **Keep internal packages private** - They should not be importable from outside the module
5. **Maintain clear separation of concerns** - Each package should have a single well-defined purpose
6. **Use consistent error handling patterns** - Define error types and constants in appropriate packages
7. **Implement proper logging** - Use the internal/log package for all logging needs

This structure ensures maintainable, scalable, and testable Go applications while adhering to industry best practices.