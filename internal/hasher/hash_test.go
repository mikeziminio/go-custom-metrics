package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexHash(t *testing.T) {
	t.Run("should return consistent hash for same data and key", func(t *testing.T) {
		data := []byte("hello world")
		key := []byte("secret-key")

		hash1 := HexHash(data, key)
		hash2 := HexHash(data, key)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("should return different hashes for different data", func(t *testing.T) {
		data1 := []byte("hello world")
		data2 := []byte("goodbye world")
		key := []byte("secret-key")

		hash1 := HexHash(data1, key)
		hash2 := HexHash(data2, key)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("should return different hashes for different keys", func(t *testing.T) {
		data := []byte("hello world")
		key1 := []byte("key1")
		key2 := []byte("key2")

		hash1 := HexHash(data, key1)
		hash2 := HexHash(data, key2)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("should handle empty data", func(t *testing.T) {
		data := []byte("")
		key := []byte("secret-key")

		hash := HexHash(data, key)
		assert.NotEmpty(t, hash)
	})

	t.Run("should handle empty key", func(t *testing.T) {
		data := []byte("hello world")
		key := []byte("")

		hash := HexHash(data, key)
		assert.NotEmpty(t, hash)
	})

	t.Run("should return valid hex string", func(t *testing.T) {
		data := []byte("test data")
		key := []byte("test-key")

		hash := HexHash(data, key)
		assert.Len(t, hash, 64) // SHA256 produces 32 bytes = 64 hex chars
	})
}
