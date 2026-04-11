package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// encryptor handles AES-256-GCM encryption/decryption for data at rest.
type encryptor struct {
	gcm cipher.AEAD
}

// newEncryptor creates an AES-256-GCM encryptor from a key string.
// The key is hashed via SHA-256 to ensure a consistent 32-byte key.
func newEncryptor(key string) (*encryptor, error) {
	if key == "" {
		return nil, fmt.Errorf("encryption key is empty")
	}

	// Derive 32-byte key from arbitrary-length input via SHA-256
	keyBytes := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &encryptor{gcm: gcm}, nil
}

// encrypt encrypts plaintext and returns nonce + ciphertext.
// Format: [12-byte nonce][ciphertext + 16-byte auth tag]
func (e *encryptor) encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal appends ciphertext + auth tag to the nonce slice
	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data encrypted with encrypt.
// Expects format: [12-byte nonce][ciphertext + 16-byte auth tag]
func (e *encryptor) decrypt(data []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// isEncrypted checks if data looks encrypted (starts with a valid nonce size header).
// We use a simple heuristic: encrypted data is at least nonceSize + tagSize bytes.
func (e *encryptor) isEncrypted(data []byte) bool {
	nonceSize := e.gcm.NonceSize()
	tagSize := e.gcm.Overhead()
	return len(data) >= nonceSize+tagSize
}
