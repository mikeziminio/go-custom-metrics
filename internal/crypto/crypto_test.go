package crypto

import (
	
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testPubKeyPath = "../../testdata/keys/server-public.pem"
const testPrivKeyPath = "../../testdata/keys/server-private.pem"

func TestLoadPublicKey_Success(t *testing.T) {
	key, err := LoadPublicKey(testPubKeyPath)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &rsa.PublicKey{}, key)
}

func TestLoadPublicKey_InvalidPath(t *testing.T) {
	_, err := LoadPublicKey("/nonexistent/path")
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	tmpFile := t.TempDir() + "/invalid.pem"
	err := os.WriteFile(tmpFile, []byte("not a pem file"), 0600)
	assert.NoError(t, err)
	_, err = LoadPublicKey(tmpFile)
	assert.Equal(t, ErrInvalidPEM, err)
}

func TestLoadPrivateKey_Success(t *testing.T) {
	key, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.IsType(t, &rsa.PrivateKey{}, key)
}

func TestLoadPrivateKey_InvalidPath(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/path")
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	tmpFile := t.TempDir() + "/invalid.pem"
	err := os.WriteFile(tmpFile, []byte("not a pem file"), 0600)
	assert.NoError(t, err)
	_, err = LoadPrivateKey(tmpFile)
	assert.Equal(t, ErrInvalidPEM, err)
}

func TestEncryptWithPublicKey(t *testing.T) {
	pubKey, err := LoadPublicKey(testPubKeyPath)
	assert.NoError(t, err)

	data := []byte("test data")
	encrypted, _ := EncryptWithPublicKey(data, pubKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, data, encrypted)
}

func TestDecryptWithPrivateKey(t *testing.T) {
	privKey, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)

	data := []byte("test data")
	pubKey := privKey.PublicKey
	encrypted, _ := EncryptWithPublicKey(data, &pubKey)
	assert.NoError(t, err)

	decrypted, _ := DecryptWithPrivateKey(encrypted, privKey)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	pubKey, err := LoadPublicKey(testPubKeyPath)
	assert.NoError(t, err)

	privKey, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"small data", []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, _ := EncryptWithPublicKey(tt.data, pubKey)

			decrypted, _ := DecryptWithPrivateKey(encrypted, privKey)

			assert.Equal(t, tt.data, decrypted)
		})
	}
}
