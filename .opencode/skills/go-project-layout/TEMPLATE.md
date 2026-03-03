# Go Project Layout Template

This template demonstrates a complete implementation of the recommended Go project layout.

## Basic Structure

```
my-go-project/
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── .golangci.yml
├── Makefile
├── Dockerfile
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── model/
│   │   ├── user.go
│   │   ├── order.go
│   │   └── withdrawal.go
│   ├── config/
│   │   └── config.go
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
│   │   └── models/
│   │       ├── user.go
│   │       └── order.go
│   ├── clients/
│   │   └── http_client.go
│   └── log/
│       └── logger.go
├── pkg/
│   └── utilities/
│       └── helpers.go
└── test/
    ├── unit/
    └── integration/
```

## Implementation Examples

### 1. Module Definition (`go.mod`)
```go
module github.com/mycompany/my-go-project

go 1.26.0

require (
    github.com/go-chi/chi/v5 v5.2.3
    go.uber.org/zap v1.27.1
    github.com/jackc/pgx/v5 v5.8.0
)
```

### 2. Main Application Entry Point (`cmd/myapp/main.go`)
```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/mycompany/my-go-project/internal/server"
    "github.com/mycompany/my-go-project/internal/config"
    "github.com/mycompany/my-go-project/internal/log"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config", err)
    }

    srv := server.New(cfg)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Graceful shutdown handling
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-c
        cancel()
    }()

    if err := srv.Run(ctx); err != nil {
        log.Fatal("Server failed", err)
    }
}
```

### 3. Domain Model (`internal/model/user.go`)
```go
package model

import (
    "errors"
)

var (
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrUserNotFound      = errors.New("user not found")
    ErrWrongPassword     = errors.New("wrong password")
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

### 4. Configuration (`internal/config/config.go`)
```go
package config

import (
    "time"
    
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    ServerAddr    string        `split_words:"true" default:":8080"`
    DatabaseURL   string        `split_words:"true" required:"true"`
    JWTSecret     string        `split_words:"true" required:"true"`
    LogLevel      string        `split_words:"true" default:"info"`
    ShutdownTimeout time.Duration `split_words:"true" default:"30s"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### 5. HTTP Handler (`internal/server/handlers/user_handler.go`)
```go
package handlers

import (
    "net/http"
    
    "github.com/mycompany/my-go-project/internal/model"
    "github.com/mycompany/my-go-project/internal/server/middleware"
)

type UserHandler struct {
    // Dependencies
}

func NewUserHandler() *UserHandler {
    return &UserHandler{}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Implementation here
    // Use model.User type
}
```

### 6. Server Initialization (`internal/server/server.go`)
```go
package server

import (
    "context"
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/mycompany/my-go-project/internal/config"
    "github.com/mycompany/my-go-project/internal/log"
    "github.com/mycompany/my-go-project/internal/server/handlers"
)

type Server struct {
    cfg    *config.Config
    router *chi.Mux
    logger log.Logger
}

func New(cfg *config.Config) *Server {
    router := chi.NewRouter()
    
    // Middleware
    router.Use(middleware.RequestID)
    router.Use(middleware.RealIP)
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)
    
    s := &Server{
        cfg:    cfg,
        router: router,
        logger: log.New(),
    }
    
    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    // Setup your routes here
    // Use handlers to manage logic
}

func (s *Server) Run(ctx context.Context) error {
    server := &http.Server{
        Addr:    s.cfg.ServerAddr,
        Handler: s.router,
    }
    
    go func() {
        <-ctx.Done()
        s.logger.Info("Shutting down server...")
        
        shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
        defer cancel()
        
        if err := server.Shutdown(shutdownCtx); err != nil {
            s.logger.Error("Server shutdown error", err)
        }
    }()
    
    s.logger.Info("Starting server on", s.cfg.ServerAddr)
    return server.ListenAndServe()
}
```

## Usage Instructions

1. **Clone this template** for new projects
2. **Update module name** in `go.mod`
3. **Customize structure** based on project needs
4. **Add required dependencies** to `go.mod`
5. **Implement business logic** in appropriate packages
6. **Ensure all models are defined** in `internal/model/`
7. **Run `go mod tidy`** to resolve dependencies

## Development Guidelines

1. **Always use `internal/model/`** for shared domain entities
2. **Keep `internal/` packages private** - they should not be importable externally
3. **Use `pkg/` for public libraries** that can be used by other projects
4. **Organize tests** in `test/unit/` and `test/integration/` directories
5. **Follow Go naming conventions** consistently
6. **Document exported functions and types** with comments
7. **Implement proper error handling** with custom error types where appropriate