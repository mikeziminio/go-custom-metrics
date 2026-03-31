// Package crypto provides RSA-OAEP encryption and decryption utilities.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var (
	ErrKeyNotFound   = errors.New("key file not found")
	ErrInvalidPEM    = errors.New("invalid PEM format")
	ErrDecryptFailed = errors.New("decryption failed")
	ErrEncryptFailed = errors.New("encryption failed")
)

// LoadPublicKey loads a public key from a PEM-encoded file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrKeyNotFound
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEM
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidPEM
	}

	return rsaKey, nil
}

// LoadPrivateKey loads a private key from a PEM-encoded file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrKeyNotFound
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEM
	}

	var key *rsa.PrivateKey
	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrInvalidPEM
		}
	}

	return key, nil
}

// EncryptWithPublicKey encrypts data using RSA-OAEP with SHA-256.
func EncryptWithPublicKey(data []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, nil)
}

// DecryptWithPrivateKey decrypts data using RSA-OAEP with SHA-256.
func DecryptWithPrivateKey(data []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, data, nil)
}
