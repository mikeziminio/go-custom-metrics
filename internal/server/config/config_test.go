package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dario.cat/mergo"
	"github.com/stretchr/testify/assert"
)

func TestServerConfigPriority(t *testing.T) {
	testCases := []struct {
		name                    string
		setupEnv                map[string]string
		setupFile               map[string]interface{}
		setupFlags              map[string]string
		expectedAddress         string
		expectedRestore         bool
		expectedStoreInterval   float64
		expectedFileStoragePath string
	}{
		{
			name: "flags override file",
			setupFile: map[string]interface{}{
				"address":         "file:9999",
				"restore":         false,
				"store_interval":  1.0,
				"store_file":      "/file/path.json",
			},
			setupFlags: map[string]string{
				"a": "flag:8888",
				"r": "true",
			},
			expectedAddress:         "flag:8888",
			expectedRestore:         true,
			expectedStoreInterval:   1.0,
			expectedFileStoragePath: "/file/path.json",
		},
		{
			name: "file used when no flags/env",
			setupFile: map[string]interface{}{
				"address":         "file:9999",
				"restore":         true,
				"store_interval":  5.0,
				"store_file":      "/file/path.json",
			},
			expectedAddress:         "file:9999",
			expectedRestore:         true,
			expectedStoreInterval:   5.0,
			expectedFileStoragePath: "/file/path.json",
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
				if v, ok := tc.setupFlags["r"]; ok && v == "true" {
					c.Restore = true
				}
			}

			// Apply env vars
			if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
				c.Address = envAddr
			}

			assert.Equal(t, tc.expectedAddress, c.Address)
			assert.Equal(t, tc.expectedRestore, c.Restore)
			assert.Equal(t, tc.expectedStoreInterval, c.StoreInterval)
			assert.Equal(t, tc.expectedFileStoragePath, c.FileStoragePath)
		})
	}
}
