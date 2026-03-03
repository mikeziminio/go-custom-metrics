# Go Migrations Skill

## Overview
This skill provides comprehensive guidance for implementing database migrations in Go applications using golang-migrate/v4. It covers migration file creation, execution patterns, rollback procedures, and integration with Go applications following best practices.

## Core Concepts

### Migration Files
Migration files follow a strict naming convention:
- `{version}_{title}.up.{extension}`
- `{version}_{title}.down.{extension}`

Examples:
- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`
- `1500360784_initialize_schema.up.sql`
- `1500360784_initialize_schema.down.sql`

### Versioning Strategies
1. **Sequential integers**: 1, 2, 3...
2. **Timestamps**: Unix timestamps (1500360784) for better ordering
3. **Semantic versioning**: v1.0.0, v1.1.0, etc. (though less common for migrations)

## Installation

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/{driver}
go get -u github.com/golang-migrate/migrate/v4/source/{source}
```

Example for PostgreSQL:
```bash
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

## Migration File Creation

### Using CLI
```bash
# Create migration files with sequential numbering
migrate create -ext sql -dir db/migrations -seq create_users_table

# Create migration files with timestamp numbering
migrate create -ext sql -dir db/migrations -timestamp create_users_table
```

### Manual Creation
Each migration must have both up and down files:
```sql
-- db/migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(300) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- db/migrations/000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

## Best Practices for Migration Content

### Idempotency
Always make migrations idempotent using `IF NOT EXISTS`/`IF EXISTS` clauses:
```sql
-- Good
CREATE TABLE IF NOT EXISTS users (...);
DROP TABLE IF EXISTS users;

-- Bad
CREATE TABLE users (...); 
DROP TABLE users;
```

### Transactions
Wrap multi-statement migrations in transactions when supported by your database:
```sql
-- PostgreSQL example
BEGIN;
CREATE TYPE user_status AS ENUM ('active', 'inactive');
ALTER TABLE users ADD COLUMN status user_status DEFAULT 'active';
COMMIT;

-- Down migration
BEGIN;
ALTER TABLE users DROP COLUMN status;
DROP TYPE user_status;
COMMIT;
```

### Error Handling
Avoid complex logic in migrations. Keep them simple and focused on schema changes:
```sql
-- Good: Simple schema change
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;

-- Avoid: Complex business logic
UPDATE users SET last_login = NOW() WHERE last_login IS NULL;
```

## Go Application Integration

### Basic Setup
```go
import (
    "database/sql"
    "log"
    
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    m, err := migrate.New(
        "file://db/migrations",
        "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    
    if err := m.Up(); err != nil {
        log.Fatal(err)
    }
}
```

### With Existing Database Connection
```go
import (
    "database/sql"
    "log"
    
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    m, err := migrate.NewWithDatabaseInstance(
        "file://db/migrations",
        "postgres", driver)
    if err != nil {
        log.Fatal(err)
    }
    
    if err := m.Up(); err != nil {
        log.Fatal(err)
    }
}
```

### Migration Execution Patterns

#### Running All Migrations
```go
// Apply all pending migrations
err := m.Up()
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}
```

#### Running Specific Number of Migrations
```go
// Apply 2 migrations
err := m.Steps(2)
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}
```

#### Running Specific Version
```go
// Apply to version 3
err := m.Migrate(3)
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}
```

#### Rollback Operations
```go
// Rollback last migration
err := m.Down()
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}

// Rollback specific number of migrations
err := m.Steps(-2)
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}

// Rollback to specific version
err := m.Migrate(2) // Rollback to version 2
if err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}
```

### Graceful Shutdown Support
```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // Setup signal handling for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Start migration in a goroutine
    go func() {
        <-sigChan
        cancel()
    }()
    
    m, err := migrate.New("file://db/migrations", "postgres://...")
    if err != nil {
        log.Fatal(err)
    }
    
    // Set grace period for migration cancellation
    go func() {
        select {
        case <-ctx.Done():
            // Handle cancellation
            log.Println("Migration cancelled")
        }
    }()
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }
}
```

## Testing Migrations

### Unit Testing Migration Commands
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func TestMigrations(t *testing.T) {
    m, err := migrate.New("file://db/migrations", "sqlite3://:memory:")
    assert.NoError(t, err)
    
    err = m.Up()
    assert.NoError(t, err)
    
    // Verify database state
    // ... test assertions
    
    err = m.Down()
    assert.NoError(t, err)
}
```

### Integration Testing with Containerized Databases
Use docker-compose for testing migrations in isolated environments:
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"
```

## Common Migration Patterns

### Adding Columns
```sql
-- Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Down
ALTER TABLE users DROP COLUMN phone;
```

### Renaming Columns
```sql
-- Up
ALTER TABLE users RENAME COLUMN old_name TO new_name;

-- Down
ALTER TABLE users RENAME COLUMN new_name TO old_name;
```

### Creating Indexes
```sql
-- Up
CREATE INDEX idx_users_email ON users(email);

-- Down
DROP INDEX idx_users_email;
```

### Altering Data Types
```sql
-- Up
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500);

-- Down
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(300);
```

## Error Handling and Debugging

### Common Errors
1. **Dirty database version**: Migration failed partway through
2. **No change**: Migration already applied
3. **Version mismatch**: Migration version conflicts

### Recovery Procedures
```bash
# Force migration version (after fixing issues)
migrate -path db/migrations -database postgres://... force 3

# Check current migration version
migrate -path db/migrations -database postgres://... version

# Get migration status
migrate -path db/migrations -database postgres://... status
```

### Logging Migration Events
```go
import (
    "log"
    "github.com/golang-migrate/migrate/v4"
)

// Custom logger for migrations
type MigrationLogger struct{}

func (l MigrationLogger) Printf(format string, v ...interface{}) {
    log.Printf("[MIGRATION] "+format, v...)
}

func (l MigrationLogger) Println(v ...interface{}) {
    log.Println("[MIGRATION]", v...)
}

// Usage
m, err := migrate.New("file://db/migrations", "postgres://...")
m.Log = MigrationLogger{}
```

## Performance Considerations

### Large Schema Changes
For large tables or expensive operations:
1. Break migrations into smaller chunks
2. Use background jobs for data transformation
3. Implement gradual rollouts

### Migration Scripts Best Practices
- Minimize downtime during migrations
- Avoid long-running migrations
- Test migrations in staging environments
- Always backup before production migrations

## Security Considerations

### Migration File Permissions
Ensure migration files have appropriate permissions:
```bash
chmod 644 db/migrations/*.sql
```

### Secret Management
Never embed sensitive information in migration files. Use environment variables or configuration management:
```sql
-- Use placeholders or configuration-driven approach
-- Avoid hardcoded credentials in migration files
```

## Deployment Integration

### CI/CD Pipeline
Add migration steps to deployment pipeline:
```yaml
# Example GitHub Actions workflow
jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - name: Run migrations
        run: |
          migrate -path db/migrations \
                  -database postgres://user:pass@localhost:5432/db \
                  up
```

### Docker Deployment
In Dockerfile or entrypoint scripts:
```dockerfile
# Run migrations before starting application
RUN migrate -path /app/migrations \
            -database postgres://user:pass@db:5432/mydb \
            up
```

## Advanced Features

### Multiple Database Support
```go
// Support multiple databases with different drivers
import (
    _ "github.com/golang-migrate/migrate/v4/database/mysql"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)
```

### Embedded Migrations
Using go-bindata or pkger to embed migrations:
```go
import (
    _ "github.com/golang-migrate/migrate/v4/source/go_bindata"
)

m, err := migrate.New(
    "go_bindata://db/migrations",
    "postgres://...")
```

### Custom Migration Sources
Implement custom sources for special requirements:
```go
import (
    "github.com/golang-migrate/migrate/v4/source"
)

type CustomSource struct{}

// Implement source.Driver interface
var _ source.Driver = (*CustomSource)(nil)
```