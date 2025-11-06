package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Crypto is the interface for encryption and decryption operations
type Crypto interface {
	// Encrypt encrypts plaintext and returns base64 encoded ciphertext
	Encrypt(plaintext string) (string, error)

	// Decrypt decrypts base64 encoded ciphertext and returns plaintext
	Decrypt(ciphertext string) (string, error)

	// EncryptBytes encrypts byte data and returns encrypted bytes
	EncryptBytes(data []byte) ([]byte, error)

	// DecryptBytes decrypts encrypted bytes and returns plaintext bytes
	DecryptBytes(data []byte) ([]byte, error)

	// Hash creates a SHA256 hash of the input
	Hash(data string) string

	// CompareHash compares data with a hash
	CompareHash(data string, hash string) bool
}

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// aesCrypto implements Crypto interface using AES-256-GCM
type aesCrypto struct {
	key []byte
}

// NewCrypto creates a new Crypto instance with the given key
// Key should be 32 bytes for AES-256
func NewCrypto(key string) (Crypto, error) {
	if len(key) == 0 {
		return nil, ErrInvalidKey
	}

	// Use PBKDF2 to derive a 32-byte key from the provided key
	derivedKey := pbkdf2.Key([]byte(key), []byte("hanif-skeleton-salt"), 10000, 32, sha256.New)

	return &aesCrypto{
		key: derivedKey,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64 encoded ciphertext
func (c *aesCrypto) Encrypt(plaintext string) (string, error) {
	encrypted, err := c.EncryptBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts base64 encoded ciphertext using AES-256-GCM
func (c *aesCrypto) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	decrypted, err := c.DecryptBytes(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// EncryptBytes encrypts byte data using AES-256-GCM
func (c *aesCrypto) EncryptBytes(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce (number used once)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptBytes decrypts encrypted bytes using AES-256-GCM
func (c *aesCrypto) DecryptBytes(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// Hash creates a SHA256 hash of the input
func (c *aesCrypto) Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// CompareHash compares data with a hash
func (c *aesCrypto) CompareHash(data string, hash string) bool {
	return c.Hash(data) == hash
}
