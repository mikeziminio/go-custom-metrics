// Package log provides structured logging using Zap.
//
// It offers a New function to create a configured Zap logger instance
// with JSON encoding and standardized production settings.
package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new Zap logger instance with the specified log level.
//
// Parameters:
//   - level: Log level string (e.g., "debug", "info", "warn", "error")
//
// Returns a *zap.Logger configured with production settings or an error if the
// log level is invalid or logger creation fails.
func New(level string) (*zap.Logger, error) {
	if level == "" {
		return nil, fmt.Errorf("failed to parse %s as log level: level cannot be empty", level)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s as log level: %w", level, err)
	}

	return zap.Must(
		zap.Config{
			Level:             zap.NewAtomicLevelAt(logLevel),
			DisableStacktrace: true,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths:       []string{"stderr"},
			ErrorOutputPaths:  []string{"stderr"},
		}.Build(),
	), nil
}
