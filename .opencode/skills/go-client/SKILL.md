---
name: go-client
description: Skill for creating modern HTTP clients in Go for microservice communication
---

# Go HTTP Client Skill

## Overview
This skill provides comprehensive guidance for creating modern, robust HTTP clients in Go for microservice communication. It follows Go 1.26+ best practices and focuses on reliability, error handling, and performance.

## Key Principles

### 1. Client Design Best Practices

#### Use a Dedicated Client Type
Always define a dedicated client struct that encapsulates HTTP configuration and behavior:

```go
type Client struct {
    httpClient *http.Client
    baseURL    string
    // other fields as needed
}
```

#### Configure HTTP Client Appropriately
Customize the `http.Client` with appropriate timeouts and transport settings:

```go
func NewClient(baseURL string, timeout time.Duration) *Client {
    transport := &http.Transport{
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    
    httpClient := &http.Client{
        Transport: transport,
        Timeout:   timeout,
    }
    
    return &Client{
        httpClient: httpClient,
        baseURL:    baseURL,
    }
}
```

### 2. Context Usage

Always use context for request lifecycle management:

```go
func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    // Process response...
}
```

### 3. Error Handling Patterns

Implement proper error handling with contextual wrapping:

```go
// Good error handling
if err != nil {
    return nil, fmt.Errorf("failed to fetch order %s: %w", id, err)
}

// Avoid generic errors
return nil, errors.New("failed to fetch order") // Bad
```

### 4. HTTP Status Code Handling

Explicitly handle different HTTP status codes:

```go
switch res.StatusCode {
case http.StatusOK:
    // Success case
case http.StatusNotFound:
    return nil, ErrOrderNotFound
case http.StatusTooManyRequests:
    return nil, ErrRateLimitExceeded
case http.StatusInternalServerError:
    return nil, ErrInternalServer
default:
    return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
}
```

### 5. Response Body Management

Properly manage response bodies to prevent resource leaks:

```go
defer func() {
    if res.Body != nil {
        res.Body.Close()
    }
}()

body, err := io.ReadAll(res.Body)
if err != nil {
    return nil, fmt.Errorf("failed to read response body: %w", err)
}
```

### 6. Request Construction

Use `url.JoinPath` for safe URL construction:

```go
fullURL, err := url.JoinPath(c.baseURL, path)
if err != nil {
    return nil, fmt.Errorf("failed to construct URL: %w", err)
}
```

### 7. Structured Data Handling

Define internal response structs for JSON marshaling:

```go
type orderInfoResponse struct {
    ID      string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual,omitempty"`
}
```

### 8. Timeout Configuration

Set appropriate timeouts for different operations:

```go
// Connection timeout
Transport: &http.Transport{
    TLSHandshakeTimeout: 10 * time.Second,
    // ...other settings
}

// Request timeout
client := &http.Client{
    Timeout: 30 * time.Second, // Overall request timeout
}
```

### 9. Retry Logic (Optional)

Implement retry mechanisms with exponential backoff for transient failures:

```go
func (c *Client) WithRetry(maxRetries int) *Client {
    // Implementation for retry logic
    return c
}
```

### 10. Testing Considerations

Design clients to be easily testable:

```go
type Client interface {
    GetOrder(ctx context.Context, id string) (*Order, error)
}

// Test with mock implementations
```

## Complete Example Template

```go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"

    "your-module/models"
)

type Client struct {
    httpClient *http.Client
    baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
    transport := &http.Transport{
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    
    httpClient := &http.Client{
        Transport: transport,
        Timeout:   timeout,
    }
    
    return &Client{
        httpClient: httpClient,
        baseURL:    baseURL,
    }
}

func (c *Client) GetOrder(ctx context.Context, id string) (*models.Order, error) {
    path := fmt.Sprintf("/api/orders/%s", id)
    fullURL, err := url.JoinPath(c.baseURL, path)
    if err != nil {
        return nil, fmt.Errorf("failed to construct URL: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, http.NoBody)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer func() {
        if resp.Body != nil {
            resp.Body.Close()
        }
    }()

    if resp.StatusCode != http.StatusOK {
        switch resp.StatusCode {
        case http.StatusNotFound:
            return nil, fmt.Errorf("order not found: %w", ErrOrderNotFound)
        case http.StatusTooManyRequests:
            return nil, fmt.Errorf("rate limit exceeded: %w", ErrRateLimitExceeded)
        default:
            return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, resp.Status)
        }
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    var orderResp orderInfoResponse
    if err := json.Unmarshal(body, &orderResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &models.Order{
        ID:      orderResp.ID,
        Status:  models.OrderStatus(orderResp.Status),
        Accrual: orderResp.Accrual,
    }, nil
}

type orderInfoResponse struct {
    ID      string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual,omitempty"`
}
```

## Best Practices Summary

1. **Always use context** for request cancellation and timeout control
2. **Set reasonable timeouts** for both connection and request durations
3. **Handle HTTP status codes explicitly** rather than assuming success
4. **Properly close response bodies** to prevent resource leaks
5. **Wrap errors appropriately** with contextual information
6. **Use `url.JoinPath`** for safe URL construction
7. **Configure HTTP transport** with appropriate settings for production use
8. **Define clear interfaces** for easier testing and mocking
9. **Validate and sanitize data** before processing
10. **Consider implementing retry logic** for transient failures where appropriate

This skill ensures that HTTP clients in Go applications follow modern best practices, are resilient to network issues, and provide clear error messages for debugging.