package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Addr string `flag:"a"`
	Port int    `flag:"p"`
}

func TestFillFlags(t *testing.T) {
	testCases := []struct {
		name string
		dst  *TestConfig
		def  *TestConfig
	}{
		{
			name: "success",
			dst:  &TestConfig{},
			def: &TestConfig{
				Addr: "localhost",
				Port: 8080,
			},
		},
		{
			name: "nil dst",
			dst:  nil,
			def:  &TestConfig{},
		},
		{
			name: "nil def",
			dst:  &TestConfig{},
			def:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FillFlags(tc.dst, tc.def)
			if tc.name == "success" {
				assert.NoError(t, err)
				assert.Equal(t, "localhost", tc.dst.Addr)
				assert.Equal(t, 8080, tc.dst.Port)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMergeOnlyFlags(t *testing.T) {
	testCases := []struct {
		name   string
		dst    *TestConfig
		src    *TestConfig
		flags  map[string]struct{}
		expDst *TestConfig
	}{
		{
			name: "one flag set",
			dst:  &TestConfig{Addr: "from-dst", Port: 9999},
			src:  &TestConfig{Addr: "from-src", Port: 8888},
			flags: map[string]struct{}{
				"a": {},
			},
			expDst: &TestConfig{Addr: "from-src", Port: 9999},
		},
		{
			name:   "no flags set",
			dst:    &TestConfig{Addr: "from-dst", Port: 9999},
			src:    &TestConfig{Addr: "from-src", Port: 8888},
			flags:  map[string]struct{}{},
			expDst: &TestConfig{Addr: "from-dst", Port: 9999},
		},
		{
			name: "all flags set",
			dst:  &TestConfig{Addr: "from-dst", Port: 9999},
			src:  &TestConfig{Addr: "from-src", Port: 8888},
			flags: map[string]struct{}{
				"a": {},
				"p": {},
			},
			expDst: &TestConfig{Addr: "from-src", Port: 8888},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := *tc.dst
			src := *tc.src
			err := MergeOnlyFlags(&dst, &src, tc.flags)
			assert.NoError(t, err)
			assert.Equal(t, tc.expDst.Addr, dst.Addr)
			assert.Equal(t, tc.expDst.Port, dst.Port)
		})
	}
}

func TestMergeOnlyFlags_Nil(t *testing.T) {
	err := MergeOnlyFlags(nil, &TestConfig{}, map[string]struct{}{})
	assert.Error(t, err)

	err = MergeOnlyFlags(&TestConfig{}, nil, map[string]struct{}{})
	assert.Error(t, err)
}

func TestFillConfigFromFile(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		def     *TestConfig
		expDst  *TestConfig
	}{
		{
			name: "full config",
			content: `{
				"addr": "from-file",
				"port": 1234
			}`,
			def:    &TestConfig{Addr: "default-addr", Port: 5678},
			expDst: &TestConfig{Addr: "from-file", Port: 1234},
		},
		{
			name: "partial config",
			content: `{
				"addr": "from-file"
			}`,
			def:    &TestConfig{Addr: "default-addr", Port: 5678},
			expDst: &TestConfig{Addr: "from-file", Port: 5678},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			err := os.WriteFile(configPath, []byte(tc.content), 0644)
			assert.NoError(t, err)

			dst := &TestConfig{}
			err = FillConfigFromFile(dst, tc.def, configPath)
			assert.NoError(t, err)
			assert.Equal(t, tc.expDst.Addr, dst.Addr)
			assert.Equal(t, tc.expDst.Port, dst.Port)
		})
	}
}

func TestFillConfigFromFile_ErrorCases(t *testing.T) {
	err := FillConfigFromFile(nil, &TestConfig{}, "/tmp/test.json")
	assert.Error(t, err)

	err = FillConfigFromFile(&TestConfig{}, nil, "/tmp/test.json")
	assert.Error(t, err)

	err = FillConfigFromFile(&TestConfig{}, &TestConfig{}, "/nonexistent/path")
	assert.Error(t, err)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	err = os.WriteFile(configPath, []byte("not json"), 0644)
	assert.NoError(t, err)

	err = FillConfigFromFile(&TestConfig{}, &TestConfig{}, configPath)
	assert.Error(t, err)
}

func TestIntegration_Full(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"addr": "json-server:9999",
		"port": 3000
	}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	def := &TestConfig{
		Addr: "localhost:8080",
		Port: 8080,
	}

	dst := &TestConfig{}
	err = FillConfigFromFile(dst, def, configPath)
	assert.NoError(t, err)

	assert.Equal(t, "json-server:9999", dst.Addr)
	assert.Equal(t, 3000, dst.Port)

	flags := map[string]struct{}{
		"a": {},
	}

	src := &TestConfig{
		Addr: "flag-server:7777",
		Port: 7000,
	}

	err = MergeOnlyFlags(dst, src, flags)
	assert.NoError(t, err)

	assert.Equal(t, "flag-server:7777", dst.Addr)
	assert.Equal(t, 3000, dst.Port)
}
