package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HexHash(data []byte, key []byte) string {
	hasher := hmac.New(sha256.New, key)
	_, _ = hasher.Write(data) // hash.Hash.Write() never returns errors
	return hex.EncodeToString(hasher.Sum(nil))
}
