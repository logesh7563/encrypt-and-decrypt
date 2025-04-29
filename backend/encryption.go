package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// AESKeySize is 32 bytes for AES-256
	AESKeySize = 32
)

// deriveKey derives a 32-byte key from a password for AES-256
func deriveKey(password string) []byte {
	// Use SHA-256 to derive a 32-byte key from the password
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hasher.Sum(nil)
}

// EncryptData encrypts data using AES-256 in GCM mode
func EncryptData(data []byte, password string) ([]byte, error) {
	// Derive 32-byte key for AES-256
	key := deriveKey(password)

	// Create a new AES cipher block using the derived key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCM is a mode of operation for symmetric key cryptographic block ciphers
	// It provides authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce (number used once)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// Encrypt and authenticate data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData decrypts data using AES-256 in GCM mode
func DecryptData(encryptedData []byte, password string) ([]byte, error) {
	// Check if we have data to decrypt
	if len(encryptedData) == 0 {
		return nil, errors.New("no data to decrypt")
	}

	// Derive 32-byte key for AES-256
	key := deriveKey(password)

	// Create a new AES cipher block using the derived key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cipher creation failed: %w", err)
	}

	// GCM is a mode of operation for symmetric key cryptographic block ciphers
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM mode initialization failed: %w", err)
	}

	// Extract the nonce from the encrypted data
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("encrypted data too short (missing nonce)")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt and verify data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		if strings.Contains(err.Error(), "message authentication failed") {
			return nil, errors.New("decryption failed: cipher: message authentication failed - incorrect key or corrupted data")
		}
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptToBase64 encrypts data using AES-GCM and returns it as a base64 string
func EncryptToBase64(data []byte, key string) (string, error) {
	// Create cipher block
	block, err := aes.NewCipher([]byte(key))
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

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptFromBase64 decrypts a base64 encoded string using AES-256-GCM
func DecryptFromBase64(encryptedBase64 string, key string) ([]byte, error) {
	// Decode the base64 string
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("cipher creation failed: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM creation failed: %w", err)
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce from ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
