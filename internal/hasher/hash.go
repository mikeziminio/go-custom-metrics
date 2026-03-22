// Package hasher provides HMAC-SHA256 hashing functionality.
//
// It offers utilities for computing cryptographic hashes using HMAC-SHA256
// and returning them as hexadecimal strings.
package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HexHash computes HMAC-SHA256 hash of data using the provided key.
//
// Parameters:
//   - data: The data to hash
//   - key: The secret key for HMAC
//
// Returns the hash as a hexadecimal string.
func HexHash(data []byte, key []byte) string {
	hasher := hmac.New(sha256.New, key)
	_, _ = hasher.Write(data) // hash.Hash.Write() never returns errors
	return hex.EncodeToString(hasher.Sum(nil))
}
