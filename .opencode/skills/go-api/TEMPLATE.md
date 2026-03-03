# Go API Template

This template demonstrates a complete implementation of an API layer using the principles described in the go-api skill.

## Project Structure
```
cmd/
└── myapp/
    └── main.go

internal/
├── api/
│   ├── api.go
│   ├── user.go
│   └── auth.go
├── model/
│   ├── user.go
│   └── order.go
├── repository/
│   ├── user_repository.go
│   └── order_repository.go
└── service/
    └── user_service.go
```

## API Implementation

### main.go
```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/mikeziminio/myapp/internal/api"
    "github.com/mikeziminio/myapp/internal/repository"
    "github.com/mikeziminio/myapp/internal/service"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    // Initialize repositories
    userRepo := repository.NewUserRepository()
    orderRepo := repository.NewOrderRepository()

    // Initialize services
    userService := service.NewUserService(userRepo)

    // Initialize API
    api := api.NewAPI(":8080", userService, orderRepo, logger)

    // Register routers
    api.RegisterRouters()

    // Start server
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        logger.Info("Starting server", zap.String("address", ":8080"))
        api.Run(ctx)
    }()

    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    logger.Info("Shutting down server...")
}
```

### api/api.go
```go
package api

import (
    "context"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"
)

// UserRepository interface defines the repository contract for user operations
type UserRepository interface {
    Register(ctx context.Context, login string, password string) (*User, error)
    AuthByLogin(ctx context.Context, login string, password string) (*User, error)
    AuthByToken(ctx context.Context, token string) (*User, error)
    // Add other required methods here
}

// OrderFetcher interface defines the repository contract for order operations
type OrderFetcher interface {
    Order(ctx context.Context, orderID string) (*Order, error)
}

// User represents a user entity
type User struct {
    ID       int    `json:"id"`
    Login    string `json:"login"`
    Token    string `json:"token"`
    Balance  float64 `json:"balance"`
    Withdrawn float64 `json:"withdrawn"`
}

// Order represents an order entity
type Order struct {
    ID        string  `json:"id"`
    Accrual   float64 `json:"accrual"`
    Status    string  `json:"status"`
}

// API struct holds the API dependencies and server configuration
type API struct {
    logger         *zap.Logger
    httpServer     *http.Server
    router         *chi.Mux
    userRepository UserRepository
    orderFetcher   OrderFetcher
}

// NewAPI creates a new API instance with the provided dependencies
func NewAPI(
    address string,
    userRepository UserRepository,
    orderFetcher OrderFetcher,
    logger *zap.Logger,
) *API {
    r := chi.NewRouter()

    httpServer := &http.Server{
        Addr:              address,
        Handler:           r,
        ReadTimeout:       2 * time.Second,
        ReadHeaderTimeout: 1 * time.Second,
    }

    return &API{
        logger:         logger,
        httpServer:     httpServer,
        router:         r,
        userRepository: userRepository,
        orderFetcher:   orderFetcher,
    }
}

// RegisterRouters registers all HTTP routes with the router
func (a *API) RegisterRouters() {
    a.router.Use(middleware.StripSlashes)

    // Public routes
    a.router.Post("/api/user/register", a.Register)
    a.router.Post("/api/user/login", a.Login)

    // Protected routes
    a.router.With(a.authMiddlewareHandler).
        Get("/api/user/balance", a.Balance)
    a.router.With(a.authMiddlewareHandler).
        Get("/api/user/withdrawals", a.Withdrawals)
    a.router.With(a.authMiddlewareHandler).
        Post("/api/user/balance/withdraw", a.AddWithdrawal)
    a.router.With(a.authMiddlewareHandler).
        Get("/api/user/orders", a.Orders)
    a.router.With(a.authMiddlewareHandler).
        Post("/api/user/orders", a.AddOrder)
}

// Run starts the HTTP server
func (a *API) Run(ctx context.Context) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    go func() {
        a.logger.Info("Start listening", zap.String("addr", a.httpServer.Addr))
        err := a.httpServer.ListenAndServe()
        if err != nil {
            a.logger.Error("Stop listening", zap.String("addr", a.httpServer.Addr), zap.Error(err))
            cancel()
        }
    }()

    ctx, cancel = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    <-ctx.Done()
}

// standardError logs an error and sends an HTTP error response
func (a *API) standardError(w http.ResponseWriter, logText string, err error, code int) {
    a.logger.Error(logText, zap.Error(err))
    http.Error(w, http.StatusText(code), code)
}

// customError logs a custom error and sends an HTTP error response
func (a *API) customError(w http.ResponseWriter, errText string, err error, code int) {
    a.logger.Error(errText, zap.Error(err))
    http.Error(w, errText, code)
}
```

### api/user.go
```go
package api

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Register - user registration endpoint
// POST /api/user/register
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        a.standardError(w, "Failed to read body", err, http.StatusInternalServerError)
        return
    }

    var data struct {
        Login    string `json:"login"`
        Password string `json:"password"`
    }

    err = json.Unmarshal(body, &data)
    if err != nil {
        var se *json.SyntaxError
        if errors.As(err, &se) {
            a.customError(w, "JSON syntax error", se, http.StatusBadRequest)
        } else {
            a.standardError(w, "Unexpected JSON error", err, http.StatusInternalServerError)
        }
        return
    }

    u, err := a.userRepository.Register(ctx, data.Login, data.Password)
    if err != nil {
        if errors.Is(err, ErrUserAlreadyExists) {
            a.customError(w, fmt.Sprintf("User %s already exists", data.Login), err, http.StatusConflict)
        } else {
            a.standardError(w, "Unexpected error during registration", err, http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", u.Token))
}

// Login - user authentication endpoint
// POST /api/user/login
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        a.standardError(w, "Failed to read body", err, http.StatusInternalServerError)
        return
    }

    var data struct {
        Login    string `json:"login"`
        Password string `json:"password"`
    }

    err = json.Unmarshal(body, &data)
    if err != nil {
        var se *json.SyntaxError
        if errors.As(err, &se) {
            a.customError(w, "JSON syntax error", se, http.StatusBadRequest)
        } else {
            a.standardError(w, "Unexpected JSON error", err, http.StatusInternalServerError)
        }
        return
    }

    u, err := a.userRepository.AuthByLogin(ctx, data.Login, data.Password)
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            a.customError(w, fmt.Sprintf("User %s not found", data.Login), err, http.StatusUnauthorized)
        case errors.Is(err, ErrWrongPassword):
            a.customError(w, "Wrong password", err, http.StatusUnauthorized)
        default:
            a.standardError(w, "Unexpected error during authentication", err, http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", u.Token))
}

// Balance - get user balance endpoint
// GET /api/user/balance
func (a *API) Balance(w http.ResponseWriter, r *http.Request) {
    u, ok := a.userFromContext(w, r)
    if !ok {
        return
    }

    type resData struct {
        Current   float64 `json:"current"`
        Withdrawn float64 `json:"withdrawn"`
    }

    data, err := json.Marshal(resData{
        Current:   u.Balance,
        Withdrawn: u.Withdrawn,
    })
    if err != nil {
        a.standardError(w, "Failed to marshal JSON", err, http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write(data)
}

// Withdrawals - get user withdrawals endpoint
// GET /api/user/withdrawals
func (a *API) Withdrawals(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    u, ok := a.userFromContext(w, r)
    if !ok {
        return
    }
    
    // Call repository method to fetch withdrawals
    // ...

    if len(ws) == 0 {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    type resDataItem struct {
        Order       string    `json:"order"`
        Sum         float64   `json:"sum"`
        ProcessedAt time.Time `json:"processed_at"`
    }

    var data []resDataItem

    for _, w := range ws {
        data = append(data, resDataItem{
            Order:       w.OrderID,
            Sum:         w.Sum,
            ProcessedAt: w.ProcessedAt,
        })
    }

    b, err := json.Marshal(data)
    if err != nil {
        a.standardError(w, "Failed to JSON marshal", err, http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write(b)
}

// AddWithdrawal - request withdrawal endpoint
// POST /api/user/balance/withdraw
func (a *API) AddWithdrawal(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    u, ok := a.userFromContext(w, r)
    if !ok {
        return
    }

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        a.standardError(w, "Failed to read body", err, http.StatusInternalServerError)
        return
    }

    var data struct {
        OrderID string  `json:"order"`
        Sum     float64 `json:"sum"`
    }

    err = json.Unmarshal(body, &data)
    if err != nil {
        var se *json.SyntaxError
        if errors.As(err, &se) {
            a.customError(w, "JSON syntax error", se, http.StatusBadRequest)
        } else {
            a.standardError(w, "Unexpected JSON error", err, http.StatusInternalServerError)
        }
        return
    }

    // Call repository method to add withdrawal
    // ...

    if err != nil {
        if errors.Is(err, ErrInsufficientFunds) {
            a.standardError(w, "Insufficient funds", err, http.StatusPaymentRequired)
            return
        }
        a.standardError(w, "Unexpected error", err, http.StatusInternalServerError)
        return
    }
}

// Orders - get user orders endpoint
// GET /api/user/orders
func (a *API) Orders(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    u, ok := a.userFromContext(w, r)
    if !ok {
        return
    }
    
    // Call repository method to fetch orders
    // ...

    if len(os) == 0 {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    type resDataItem struct {
        OrderID    string    `json:"number"`
        Status     string    `json:"status"`
        Accrual    float64   `json:"accrual,omitempty"`
        UploadedAt time.Time `json:"uploaded_at"`
    }

    var data []resDataItem

    for _, o := range os {
        data = append(data, resDataItem{
            OrderID:    o.ID,
            Status:     string(o.Status),
            Accrual:    o.Accrual,
            UploadedAt: o.UploadedAt,
        })
    }

    b, err := json.Marshal(data)
    if err != nil {
        a.standardError(w, "Failed to JSON marshal", err, http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write(b)
}

// AddOrder - add order endpoint
// POST /api/user/orders
func (a *API) AddOrder(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    u, ok := a.userFromContext(w, r)
    if !ok {
        return
    }

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        a.standardError(w, "Failed to read body", err, http.StatusInternalServerError)
        return
    }

    orderID := string(body)
    if !IsValidOrderID(orderID) {
        a.customError(w, "Invalid order ID", nil, http.StatusUnprocessableEntity)
        return
    }

    // Call repository method to add order
    // ...

    if err != nil {
        if errors.Is(err, ErrOrderAlreadyLoaded) {
            w.WriteHeader(http.StatusOK)
            return
        }
        if errors.Is(err, ErrOrderLoadedByAnotherUser) {
            w.WriteHeader(http.StatusConflict)
            return
        }
        a.standardError(w, "Unexpected error", err, http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusAccepted)
}

// Helper methods
func (a *API) userFromContext(w http.ResponseWriter, r *http.Request) (*User, bool) {
    u, ok := r.Context().Value(userContextKey).(*User)
    if !ok || u == nil {
        a.logger.Error("Unexpectedly empty user")
        w.WriteHeader(http.StatusUnauthorized)
        return nil, false
    }
    return u, true
}
```

### api/auth.go
```go
package api

import (
    "context"
    "net/http"
    "strings"

    "go.uber.org/zap"
)

type contextKey string

var userContextKey contextKey = "user"

// authMiddlewareHandler authenticates requests using Bearer tokens
func (a *API) authMiddlewareHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        auth := r.Header.Get("Authorization")
        token, ok := strings.CutPrefix(auth, "Bearer ")
        if !ok {
            a.logger.Error("Failed to fetch bearer token from header", zap.String("Authorization", auth))
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        
        user, err := a.userRepository.AuthByToken(ctx, token)
        if err != nil {
            a.logger.Error("Failed to authenticate by token", zap.Error(err))
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        
        ctx = context.WithValue(ctx, userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

This template showcases how to implement a properly structured API layer following all the principles outlined in the go-api skill:
1. Independence from other layers through interfaces
2. One method per endpoint
3. Proper error handling
4. Context usage for cancellation and timeouts
5. Authentication middleware pattern
6. Consistent HTTP status codes and response formats