package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
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
