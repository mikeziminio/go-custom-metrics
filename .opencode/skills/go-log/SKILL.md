---
name: go-log
description: Skill for implementing zap logger in Go with JSON output to stdout
---

# Go Logging Skill with Zap

This skill provides guidance and templates for implementing zap logger in Go applications with JSON output to stdout, following best practices for structured logging.

## Configuration Requirements

### Always Use JSON Output to stdout

```go
// Example of proper zap logger configuration for JSON output to stdout
package main

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
    "os"
)

func setupLogger() *zap.Logger {
    // Configure encoder for JSON output to stdout
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
    
    // Create logger with JSON encoder output to stdout
    logger := zap.New(zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.AddSync(os.Stdout), // Always output to stdout
        zapcore.InfoLevel,
    ))
    
    return logger
}
```

## Best Practices

### 1. Always Log to stdout (Not Files)
```go
// Correct approach - log to stdout
logger.Info("User login successful", zap.String("user_id", userID))

// Wrong approach - writing to files directly
// This breaks containerized deployment expectations
```

### 2. Structured Logging with Key-Value Pairs
```go
logger.Info("Database connection established",
    zap.String("database_url", dbURL),
    zap.Int("connection_timeout_seconds", timeout),
    zap.Bool("ssl_enabled", sslEnabled))
```

## Key Points to Remember

1. **Always output to stdout**: Never write logs directly to files
2. **Use JSON format**: Ensures compatibility with log aggregation systems
3. **Structured logging**: Use key-value pairs for better querying
4. **Environment-aware configuration**: Different levels for dev vs prod
5. **Proper resource cleanup**: Remember to call `logger.Sync()` before program exit
6. **Context propagation**: Add request IDs and correlation IDs to trace requests
7. **Error logging**: Always include the error object in error log statements
8. **Use zap.With for context**: Leverage `zap.With()` to create contextual loggers with common fields

## Using zap.With for Contextual Logging

The `zap.With()` method is a powerful feature for creating contextual loggers:

```go
// Create a base logger with common fields
baseLogger := setupLogger().With(
    zap.String("service", "my-service"),
    zap.String("version", "v1.0.0"),
)

// Create a contextual logger for a specific request
requestLogger := baseLogger.With(
    zap.String("request_id", "req-123"),
    zap.String("user_id", "user-456"),
)

// Use the contextual logger
requestLogger.Info("Processing user request")
// Output: {"level":"info","ts":"2023-01-01T00:00:00Z","msg":"Processing user request","service":"my-service","version":"v1.0.0","request_id":"req-123","user_id":"user-456"}
```
