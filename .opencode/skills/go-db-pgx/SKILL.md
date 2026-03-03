---
name: go-db-pgx
description: Skill for working with PostgreSQL databases using pgx driver with best practices
---

# Comprehensive PostgreSQL Database Skill with pgx Driver

## Overview
This skill provides comprehensive guidance for working with PostgreSQL databases using the pgx driver (github.com/jackc/pgx/v5). It covers connection management, query execution, transaction handling, and proper error handling patterns aligned with modern Go practices.

## Description
The go-db-pgx skill implements best practices for PostgreSQL database interactions in Go using the pgx library. It emphasizes connection pooling, proper error handling, transaction safety, and efficient resource management while following modern Go 1.26+ conventions.

## Roles
- senior go backend engineer
- database architect
- devops engineer

## Dependencies
- go 1.26+
- github.com/jackc/pgx/v5
- github.com/jackc/pgx/v5/pgxpool
- context
- database/sql/driver

## Core Concepts and Best Practices

### 1. Connection Management with Pooling
```go
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Initialize connection pool with recommended settings
func NewDBPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, fmt.Errorf("failed to parse connection string: %w", err)
    }

    // Configure pool settings
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = 5 * time.Minute
    config.MaxConnIdleTime = 2 * time.Minute
    config.HealthCheckPeriod = 1 * time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    // Test the connection
    if err = pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return pool, nil
}
```

### 2. Query Execution Patterns
```go
// Basic query execution
func GetUser(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
    var user User
    err := pool.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("user not found: %w", ErrUserNotFound)
        }
        return nil, fmt.Errorf("failed to query user: %w", err)
    }
    
    return &user, nil
}

// Query with multiple rows
func ListUsers(ctx context.Context, pool *pgxpool.Pool, limit, offset int) ([]*User, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, email FROM users ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            return nil, fmt.Errorf("failed to scan user row: %w", err)
        }
        users = append(users, &user)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("failed to iterate rows: %w", err)
    }

    return users, nil
}
```

### 3. Transaction Handling
```go
// Transaction with proper error handling
func CreateUserWithProfile(ctx context.Context, pool *pgxpool.Pool, user *User, profile *UserProfile) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
                // Log rollback error but don't override original error
                log.Printf("failed to rollback transaction: %v", rollbackErr)
            }
        } else {
            // Commit on success
            if commitErr := tx.Commit(ctx); commitErr != nil {
                err = fmt.Errorf("failed to commit transaction: %w", commitErr)
            }
        }
    }()

    // Insert user
    err = tx.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email).Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }

    // Insert profile
    err = tx.QueryRow(ctx, "INSERT INTO profiles (user_id, bio) VALUES ($1, $2) RETURNING id",
        user.ID, profile.Bio).Scan(&profile.ID)
    if err != nil {
        return fmt.Errorf("failed to insert profile: %w", err)
    }

    return nil
}

// Using savepoint for partial rollbacks
func UpdateUserWithSavepoint(ctx context.Context, pool *pgxpool.Pool, user *User) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
                log.Printf("failed to rollback transaction: %v", rollbackErr)
            }
        } else {
            if commitErr := tx.Commit(ctx); commitErr != nil {
                err = fmt.Errorf("failed to commit transaction: %w", commitErr)
            }
        }
    }()

    // Savepoint for partial rollback
    savepoint, err := tx.Prepare(ctx, "savepoint1", "SAVEPOINT sp1")
    if err != nil {
        return fmt.Errorf("failed to prepare savepoint: %w", err)
    }
    defer savepoint.Release(ctx)

    // Update user
    _, err = tx.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }

    // If something goes wrong here, we can rollback to savepoint
    // But still continue with the rest of the logic...

    return nil
}
```

### 4. Error Handling Patterns
```go
// Define custom errors for better error categorization
var (
    ErrUserNotFound = errors.New("user not found")
    ErrDatabase     = errors.New("database error")
    ErrValidation   = errors.New("validation error")
)

// Generic error handler that wraps database errors
func handleDBError(err error, operation string) error {
    if err == nil {
        return nil
    }
    
    switch {
    case errors.Is(err, pgx.ErrNoRows):
        return fmt.Errorf("%s: %w", operation, ErrUserNotFound)
    case errors.Is(err, context.DeadlineExceeded):
        return fmt.Errorf("%s: %w", operation, context.DeadlineExceeded)
    case errors.Is(err, context.Canceled):
        return fmt.Errorf("%s: %w", operation, context.Canceled)
    default:
        return fmt.Errorf("%s: %w", operation, fmt.Errorf("%w", ErrDatabase))
    }
}

// Example usage of error handling
func FindUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, error) {
    var user User
    err := pool.QueryRow(ctx, "SELECT id, name, email FROM users WHERE email = $1", email).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return nil, handleDBError(err, "FindUserByEmail")
    }
    
    return &user, nil
}
```

### 5. Context and Timeout Management
```go
// Using context with timeouts for database operations
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    // Set a reasonable timeout for database operations
    return context.WithTimeout(ctx, timeout)
}

// Example with timeout
func GetUserWithTimeout(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
    ctx, cancel := WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var user User
    err := pool.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("user not found: %w", ErrUserNotFound)
        }
        return nil, fmt.Errorf("failed to query user: %w", err)
    }
    
    return &user, nil
}
```

### 6. Resource Management and Cleanup
```go
// Proper cleanup of resources
type DBService struct {
    pool *pgxpool.Pool
}

func NewDBService(connString string) (*DBService, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pool, err := NewDBPool(ctx, connString)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize database service: %w", err)
    }

    return &DBService{pool: pool}, nil
}

func (s *DBService) Close() {
    if s.pool != nil {
        s.pool.Close()
    }
}

// Graceful shutdown
func (s *DBService) Shutdown(ctx context.Context) error {
    if s.pool != nil {
        s.pool.Close()
    }
    return nil
}
```

### 7. Async Operations and Batch Processing
```go
// Batch processing with goroutines and errgroup
func BatchInsertUsers(ctx context.Context, pool *pgxpool.Pool, users []*User) error {
    const batchSize = 100
    
    var eg errgroup.Group
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        batch := users[i:end]
        eg.Go(func() error {
            return insertBatch(ctx, pool, batch)
        })
    }
    
    return eg.Wait()
}

func insertBatch(ctx context.Context, pool *pgxpool.Pool, batch []*User) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
                log.Printf("failed to rollback transaction: %v", rollbackErr)
            }
        } else {
            if commitErr := tx.Commit(ctx); commitErr != nil {
                err = fmt.Errorf("failed to commit transaction: %w", commitErr)
            }
        }
    }()
    
    query := "INSERT INTO users (name, email) VALUES "
    args := make([]interface{}, 0, len(batch)*2)
    
    for i, user := range batch {
        if i > 0 {
            query += ", "
        }
        query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
        args = append(args, user.Name, user.Email)
    }
    
    _, err = tx.Exec(ctx, query, args...)
    if err != nil {
        return fmt.Errorf("failed to insert batch: %w", err)
    }
    
    return nil
}
```

## Configuration Management

```go
// Connection configuration with environment variables
type DBConfig struct {
    Host             string
    Port             int
    User             string
    Password         string
    Database         string
    MaxConns         int32
    MinConns         int32
    MaxConnLifetime  time.Duration
    MaxConnIdleTime  time.Duration
    HealthCheckPeriod time.Duration
}

func LoadDBConfig() (*DBConfig, error) {
    return &DBConfig{
        Host:             getEnv("DB_HOST", "localhost"),
        Port:             getEnvAsInt("DB_PORT", 5432),
        User:             getEnv("DB_USER", "postgres"),
        Password:         getEnv("DB_PASSWORD", ""),
        Database:         getEnv("DB_NAME", "appdb"),
        MaxConns:         getEnvAsInt32("DB_MAX_CONNS", 25),
        MinConns:         getEnvAsInt32("DB_MIN_CONNS", 5),
        MaxConnLifetime:  getEnvAsDuration("DB_MAX_CONN_LIFETIME", 5*time.Minute),
        MaxConnIdleTime:  getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 2*time.Minute),
        HealthCheckPeriod: getEnvAsDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if i, err := strconv.Atoi(value); err == nil {
            return i
        }
    }
    return defaultValue
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
    if value := os.Getenv(key); value != "" {
        if i, err := strconv.Atoi(value); err == nil {
            return int32(i)
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if d, err := time.ParseDuration(value); err == nil {
            return d
        }
    }
    return defaultValue
}
```

## Monitoring and Observability

```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics for database operations
var (
    dbQueryCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_queries_total",
        Help: "Total number of database queries",
    }, []string{"operation", "success"})

    dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "db_query_duration_seconds",
        Help: "Database query duration in seconds",
    }, []string{"operation"})
)

// Wrapper function for query tracking
func TrackQuery(ctx context.Context, pool *pgxpool.Pool, operation string, query string, args ...interface{}) (pgx.Row, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        dbQueryDuration.WithLabelValues(operation).Observe(duration)
    }()
    
    row := pool.QueryRow(ctx, query, args...)
    
    // Record counter
    if err := row.Err(); err != nil {
        dbQueryCount.WithLabelValues(operation, "false").Inc()
        return row, err
    }
    
    dbQueryCount.WithLabelValues(operation, "true").Inc()
    return row, nil
}
```

## Taskfile Integration

Add the following to your root Taskfile.yml:

```yaml
# Database tasks
db-migrate:
  desc: Run database migrations
  cmds:
    - echo "Running database migrations..."
    - go run cmd/migrate/main.go
    - echo "Database migrations completed"

db-reset:
  desc: Reset database (DANGEROUS!)
  cmds:
    - echo "Resetting database..."
    - go run cmd/reset/main.go
    - echo "Database reset completed"

db-connect:
  desc: Connect to database
  cmds:
    - echo "Connecting to database..."
    - psql -h localhost -p 5432 -U postgres -d appdb

db-test:
  desc: Run database tests
  cmds:
    - echo "Running database tests..."
    - go test -v ./internal/db/...
    - echo "Database tests completed"
```

## Testing Guidelines

### Unit Tests for Database Functions
```go
func TestGetUser(t *testing.T) {
    ctx := context.Background()
    
    // Setup mock database
    pool, cleanup := setupTestDB(t)
    defer cleanup()
    
    // Create test data
    _, err := pool.Exec(ctx, "INSERT INTO users (id, name, email) VALUES (1, 'John Doe', 'john@example.com')")
    require.NoError(t, err)
    
    // Test the function
    user, err := GetUser(ctx, pool, 1)
    require.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)
}

func TestGetUserNotFound(t *testing.T) {
    ctx := context.Background()
    
    pool, cleanup := setupTestDB(t)
    defer cleanup()
    
    // Test non-existent user
    user, err := GetUser(ctx, pool, 999)
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.True(t, errors.Is(err, ErrUserNotFound))
}
```

### Integration Tests
```go
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }
    
    // Use real database connection
    connString := "postgresql://user:password@localhost:5432/testdb"
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    pool, err := NewDBPool(ctx, connString)
    require.NoError(t, err)
    defer pool.Close()
    
    // Perform actual integration tests
    // ...
}
```

## Best Practices Checklist

### ✅ Database Connection Management
- Use connection pools with proper configuration
- Implement health checks
- Set appropriate timeouts
- Handle connection failures gracefully
- Close connections properly

### ✅ Query Execution
- Use prepared statements where appropriate
- Handle pgx.ErrNoRows correctly
- Properly scan query results
- Use context-aware queries
- Log and monitor slow queries

### ✅ Transaction Handling
- Always use explicit transactions
- Properly handle rollback scenarios
- Avoid nested transactions
- Use savepoints when needed
- Handle transaction timeouts

### ✅ Error Handling
- Wrap database errors with meaningful context
- Distinguish between different error types
- Return custom errors for business logic
- Handle context deadlines and cancellations
- Log errors appropriately

### ✅ Performance Optimization
- Use connection pooling effectively
- Batch similar operations
- Index database tables appropriately
- Monitor query performance
- Avoid N+1 query problems

### ❌ Avoid These Patterns
- Don't use raw SQL strings without validation
- Avoid direct connection creation in production
- Don't ignore errors during query execution
- Avoid long-running transactions
- Don't store passwords in plain text
- Avoid SQL injection vulnerabilities

## Migration Guide

When migrating existing database code to this style:
1. Replace direct pgx connections with connection pools
2. Add proper context propagation to all queries
3. Implement error wrapping with %w formatting
4. Add connection health checks
5. Configure proper pool settings
6. Ensure all transactions follow the proper template
7. Add error categorization with custom errors
8. Implement logging and monitoring

Base directory for this skill: file:///Users/mdzimin/sources/ai/.opencode/skills/go-db-pgx
Relative paths in this skill (e.g., scripts/, reference/) are relative to this base directory.
Note: file list is sampled.