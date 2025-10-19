package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

func Encrypt(plaintext string, password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Derive a 32-byte key from the password using PBKDF2
	key, err := pbkdf2.Key(sha256.New, password, salt, 10_000, 32)
	if err != nil {
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Prepend salt to the result
	result := append(salt, ciphertext...)

	// Use Base64 instead of hex
	return base64.StdEncoding.EncodeToString(result), nil
}

func Decrypt(encryptedB64 string, password string) (string, error) {
	// Decode from Base64
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", err
	}

	// Extract salt (first 32 bytes)
	salt := data[:32]
	ciphertext := data[32:]

	// Derive the same key using the salt
	key, err := pbkdf2.Key(sha256.New, password, salt, 10_000, 32)
	if err != nil {
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce and encrypted data
	nonceSize := gcm.NonceSize()
	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
