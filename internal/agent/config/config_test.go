package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dario.cat/mergo"
	"github.com/stretchr/testify/assert"
)

func TestAgentConfigPriority(t *testing.T) {
	testCases := []struct {
		name               string
		setupEnv           map[string]string
		setupFile          map[string]interface{}
		setupFlags         map[string]string
		expectedAddress    string
		expectedReportFreq float64
		expectedPollFreq   float64
	}{
		{
			name: "flags override file",
			setupFile: map[string]interface{}{
				"address":        "file:9999",
				"report_interval": 1.0,
				"poll_interval":  2.0,
			},
			setupFlags: map[string]string{
				"a": "flag:8888",
			},
			expectedAddress:    "flag:8888",
			expectedReportFreq: 1.0,
			expectedPollFreq:   2.0,
		},
		{
			name: "file used when no flags/env",
			setupFile: map[string]interface{}{
				"address":        "file:9999",
				"report_interval": 5.0,
				"poll_interval":  3.0,
			},
			expectedAddress:  "file:9999",
			expectedReportFreq: 5.0,
			expectedPollFreq: 3.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()

			var configFile string
			if tc.setupFile != nil {
				configFile = filepath.Join(tempDir, "config.json")
				content, err := json.Marshal(tc.setupFile)
				assert.NoError(t, err)
				err = os.WriteFile(configFile, content, 0644)
				assert.NoError(t, err)
			}

			for k, v := range tc.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			c := &Config{}
			if configFile != "" {
				c.ConfigFile = configFile
			}

			// Load and merge config file
			if c.ConfigFile != "" {
				fileConfig, err := LoadConfigFromFile(c.ConfigFile)
				assert.NoError(t, err)
				mergo.Merge(c, fileConfig, mergo.WithOverride)
			}

			// Apply flags
			if tc.setupFlags != nil {
				if v, ok := tc.setupFlags["a"]; ok {
					c.Address = v
				}
			}

			// Apply env vars
			if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
				c.Address = envAddr
			}

			assert.Equal(t, tc.expectedAddress, c.Address)
			assert.Equal(t, tc.expectedReportFreq, c.ReportInterval)
			assert.Equal(t, tc.expectedPollFreq, c.PollInterval)
		})
	}
}
