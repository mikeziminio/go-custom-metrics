package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name       string
		inputLevel string
		expectErr  bool
	}{
		{
			name:       "debug level",
			inputLevel: "debug",
			expectErr:  false,
		},
		{
			name:       "info level",
			inputLevel: "info",
			expectErr:  false,
		},
		{
			name:       "warn level",
			inputLevel: "warn",
			expectErr:  false,
		},
		{
			name:       "error level",
			inputLevel: "error",
			expectErr:  false,
		},
		{
			name:       "dpanic level",
			inputLevel: "dpanic",
			expectErr:  false,
		},
		{
			name:       "panic level",
			inputLevel: "panic",
			expectErr:  false,
		},
		{
			name:       "fatal level",
			inputLevel: "fatal",
			expectErr:  false,
		},
		{
			name:       "invalid level",
			inputLevel: "invalid",
			expectErr:  true,
		},
		{
			name:       "empty level",
			inputLevel: "",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, err := New(tc.inputLevel)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestNewLevelConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("valid level returns configured logger", func(t *testing.T) {
		t.Parallel()

		logger, err := New("info")
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		core := logger.Core()
		assert.NotNil(t, core)
	})
}

func TestNewParseLevelError(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"invalid",
		"INVALID",
		"",
		"debugg",
		"infoo",
	}

	for _, level := range testCases {
		t.Run(level, func(t *testing.T) {
			t.Parallel()

			logger, err := New(level)
			assert.Error(t, err)
			assert.Errorf(t, err, "failed to parse %s as log level", level)
			assert.Nil(t, logger)
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		level    string
		expected zapcore.Level
		err      bool
	}{
		{"debug", zapcore.DebugLevel, false},
		{"info", zapcore.InfoLevel, false},
		{"warn", zapcore.WarnLevel, false},
		{"error", zapcore.ErrorLevel, false},
		{"dpanic", zapcore.DPanicLevel, false},
		{"panic", zapcore.PanicLevel, false},
		{"fatal", zapcore.FatalLevel, false},
		{"invalid", zapcore.InvalidLevel, true},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			t.Parallel()

			level, err := zapcore.ParseLevel(tc.level)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, level)
			}
		})
	}
}
