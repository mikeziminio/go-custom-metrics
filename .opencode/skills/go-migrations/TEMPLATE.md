# Go Migrations Template

## Overview
Template for creating database migrations with golang-migrate/v4 in Go applications.

## Migration Structure
```
db/
└── migrations/
    ├── 000001_create_users_table.up.sql
    ├── 000001_create_users_table.down.sql
    ├── 000002_add_email_index.up.sql
    └── 000002_add_email_index.down.sql
```

## Sample Migration Files

### Basic Table Creation
```sql
-- 000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(300) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

### Adding Column with Default Value
```sql
-- 000002_add_phone_column.up.sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20) DEFAULT '';

-- 000002_add_phone_column.down.sql
ALTER TABLE users DROP COLUMN phone;
```

## Go Migration Runner
```go
package main

import (
    "log"
    
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    m, err := migrate.New(
        "file://db/migrations",
        "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }
    
    log.Println("Migrations applied successfully")
}
```

## Migration Best Practices Checklist
- [ ] Migration files follow naming convention: `{version}_{title}.up/down.{ext}`
- [ ] Migrations are idempotent using `IF NOT EXISTS`/`IF EXISTS`
- [ ] Multi-statement migrations wrapped in transactions where supported
- [ ] Each migration has both up and down variants
- [ ] Migration files are properly version controlled
- [ ] Migration tests are written and executed in CI
- [ ] Rollback procedures tested regularly
- [ ] Large migrations broken into smaller chunks
- [ ] Secrets not embedded in migration files