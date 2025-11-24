package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HexHash(data []byte, key []byte) string {
	hasher := hmac.New(sha256.New, key)
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
