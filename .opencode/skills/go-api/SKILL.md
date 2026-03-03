---
name: go-api
description: Skill for creating API layers in Go applications using chi router with independence from other layers
---

# Go API Skill

## Overview
This skill provides comprehensive guidance for creating API layers in Go applications using the chi router, following principles of independence from other layers, using interfaces to decrease coupling, and implementing one method per endpoint.

## Key Principles

### 1. Independence from Other Layers
- API layer should be completely independent from business logic and data access layers
- No direct imports from database, repository, or business logic packages
- All dependencies passed through interfaces

### 2. Interface-Based Dependency Injection
- Define clear interfaces for all dependencies
- Use dependency injection to decouple API from concrete implementations
- Interfaces should define exactly what the API needs, nothing more

### 3. One Method Per Endpoint
- Each API endpoint corresponds to one method in the API struct
- Methods should be kept small and focused
- Each method handles exactly one HTTP endpoint

## Structure Overview

### Basic API Structure
```go
package api

import (
    "context"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"
)

// Repository Interfaces - defined in the API layer
type UserRepository interface {
    Register(ctx context.Context, login string, password string) (*model.User, error)
    AuthByLogin(ctx context.Context, login string, password string) (*model.User, error)
    AuthByToken(ctx context.Context, token string) (*model.User, error)
    // ... other methods needed by API
}

type OrderFetcher interface {
    Order(ctx context.Context, orderID string) (*model.Order, error)
}

// API struct holding dependencies
type API struct {
    logger         *zap.Logger
    httpServer     *http.Server
    router         *chi.Mux
    userRepository UserRepository
    orderFetcher   OrderFetcher
}

// Constructor
func NewAPI(
    address string, 
    userRepository UserRepository,
    orderFetcher OrderFetcher, 
    logger *zap.Logger,
) *API {
    // Initialize router and HTTP server
    // ...
}

// Router Registration
func (a *API) RegisterRouters() {
    // Register all routes here
    a.router.Post("/api/user/register", a.Register)
    a.router.Post("/api/user/login", a.Login)
    // ... other routes
}

// Individual endpoint handlers
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
    // Handle register logic
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
    // Handle login logic
}
```

### Best Practices

#### 1. HTTP Status Codes
- Use appropriate HTTP status codes consistently:
```
200 - Success
201 - Created
202 - Accepted
204 - No Content
400 - Bad Request
401 - Unauthorized
402 - Payment Required
403 - Forbidden
404 - Not Found
409 - Conflict
422 - Unprocessable Entity
500 - Internal Server Error
```

#### 2. Error Handling
- Use standard error handling patterns:
```go
func (a *API) standardError(w http.ResponseWriter, logText string, err error, code int) {
    a.logger.Error(logText, zap.Error(err))
    http.Error(w, http.StatusText(code), code)
}

func (a *API) customError(w http.ResponseWriter, errText string, err error, code int) {
    a.logger.Error(errText, zap.Error(err))
    http.Error(w, errText, code)
}
```

#### 3. Context Usage
- All methods should accept context:
```go
func (a *API) SomeEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Use context for cancellation and timeouts
}
```

#### 4. Authentication Middleware
- Implement middleware for authentication:
```go
type contextKey string
var userContextKey contextKey = "user"

func (a *API) authMiddlewareHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from authorization header
        // Authenticate user
        // Store user in context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 5. Response Formatting
- Keep responses consistent:
```go
// For successful JSON responses
w.Header().Set("Content-Type", "application/json")
_, _ = w.Write(data)
```

## Implementation Steps

### Step 1: Define Repository Interfaces
In your API package, define all interfaces your API layer requires:

```go
type UserRepository interface {
    Register(ctx context.Context, login string, password string) (*model.User, error)
    AuthByLogin(ctx context.Context, login string, password string) (*model.User, error)
    AuthByToken(ctx context.Context, token string) (*model.User, error)
    // Add other methods here...
}
```

### Step 2: Create API Struct
```go
type API struct {
    logger         *zap.Logger
    httpServer     *http.Server
    router         *chi.Mux
    userRepository UserRepository
    // Add other dependencies...
}
```

### Step 3: Implement Constructor
```go
func NewAPI(
    address string, 
    userRepository UserRepository,
    // other dependencies...
    logger *zap.Logger,
) *API {
    // Set up chi router
    r := chi.NewRouter()
    
    // Configure HTTP server
    
    return &API{
        logger:         logger,
        httpServer:     httpServer,
        router:         r,
        userRepository: userRepository,
        // Initialize dependencies...
    }
}
```

### Step 4: Register Routes
```go
func (a *API) RegisterRouters() {
    a.router.Use(middleware.StripSlashes)
    
    // Public routes
    a.router.Post("/api/user/register", a.Register)
    a.router.Post("/api/user/login", a.Login)
    
    // Protected routes
    a.router.With(a.authMiddlewareHandler).
        Get("/api/user/balance", a.Balance)
    // ... other protected routes
}
```

### Step 5: Implement Individual Endpoints
Each endpoint should be a method on the API struct:

```go
// Register - user registration endpoint
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    // Validate input
    // Call repository method
    // Handle result/error
    // Send response
}
```

## Example Implementation Pattern

For each endpoint handler, follow this pattern:
1. Extract context from request
2. Read and parse request body if needed
3. Validate input parameters
4. Call repository/business logic methods
5. Handle errors appropriately
6. Format and send response

## Testing Strategy

### Unit Testing API Layer
- Mock all repository interfaces 
- Test each endpoint handler in isolation
- Test error scenarios
- Verify HTTP status codes and headers

### Example Test Structure
```go
func TestAPI_Register(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    api := NewAPI("localhost:8080", mockRepo, nil, zap.NewNop())
    
    // Act
    // Call the handler
    
    // Assert
    // Verify expected behavior
}
```

## Performance Considerations

### Memory Management
- Prefer reusing objects where possible
- Avoid unnecessary allocations in hot paths
- Use sync.Pool for frequently allocated objects

### Concurrency
- All API methods are safe for concurrent execution
- Use mutexes or channels when sharing state
- Leverage context for cancellation and timeouts

## Security Considerations

### Input Validation
- Always validate incoming request data
- Sanitize all inputs
- Use request size limits

### Authentication
- Implement proper authentication middleware
- Handle tokens securely
- Log authentication attempts appropriately

This skill ensures consistent, maintainable API development patterns in Go applications while following established best practices.