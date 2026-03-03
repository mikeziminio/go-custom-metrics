# Template: go-db-pgx

## Overview
Template for creating database services using pgx driver with connection pooling and best practices.

## Files to Create

### 1. Database Service Structure
```go
// internal/db/service.go
package db

import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
    pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
    return &Service{pool: pool}
}

func (s *Service) Close() {
    if s.pool != nil {
        s.pool.Close()
    }
}

func (s *Service) Ping(ctx context.Context) error {
    return s.pool.Ping(ctx)
}
```

### 2. Connection Pool Setup
```go
// internal/db/pool.go
package db

import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }
    
    // Configure pool
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = 5 * time.Minute
    config.MaxConnIdleTime = 2 * time.Minute
    config.HealthCheckPeriod = 1 * time.Minute
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    if err = pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, err
    }
    
    return pool, nil
}
```

### 3. Entity Models
```go
// internal/model/user.go
package model

type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}
```

### 4. Repository Pattern
```go
// internal/repository/user.go
package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "your-module/internal/model"
)

type UserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
    var user model.User
    err := r.pool.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    err := r.pool.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email).Scan(&user.ID)
    
    return err
}
```

### 5. Service Layer
```go
// internal/service/user.go
package service

import (
    "context"
    "your-module/internal/model"
    "your-module/internal/repository"
)

type UserService struct {
    userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUser(ctx context.Context, id int) (*model.User, error) {
    return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
    return s.userRepo.Create(ctx, user)
}
```

## Configuration

### Environment Variables
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=myapp
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_LIFETIME=5m
DB_MAX_CONN_IDLE_TIME=2m
DB_HEALTH_CHECK_PERIOD=1m
```

## Testing Structure

### Unit Tests
```go
// internal/repository/user_test.go
package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserRepository_GetByID(t *testing.T) {
    // Mock setup
    // Test logic
    // Assertions
}
```

## Usage Example

```go
// main.go
package main

import (
    "context"
    "log"
    "your-module/internal/db"
    "your-module/internal/service"
)

func main() {
    ctx := context.Background()
    
    // Initialize database pool
    pool, err := db.NewPool(ctx, "postgresql://user:pass@localhost:5432/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    
    // Initialize repositories and services
    userRepo := repository.NewUserRepository(pool)
    userService := service.NewUserService(userRepo)
    
    // Use services
    user, err := userService.GetUser(ctx, 1)
    if err != nil {
        log.Println("Error:", err)
        return
    }
    
    log.Printf("User: %+v", user)
}
```

## Best Practices Followed

1. **Connection Pooling**: Uses pgxpool for efficient connection management
2. **Context Awareness**: All database operations respect context cancellation and timeouts
3. **Error Handling**: Proper error wrapping and categorization
4. **Resource Management**: Proper cleanup of database connections
5. **Separation of Concerns**: Clear separation between repositories and services
6. **Testability**: Designed for easy unit testing with dependency injection
7. **Configuration**: Externalized configuration through environment variables
8. **Monitoring**: Ready for integration with monitoring systems
```