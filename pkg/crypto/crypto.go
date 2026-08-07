package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// DeriveKey derives a 32-byte (256-bit) AES key from a passphrase using SHA-256
func DeriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-256-GCM with a user passphrase.
// Output Format: Nonce (12 Bytes) | Ciphertext (N Bytes + 16 Bytes Auth Tag)
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	key := DeriveKey(passphrase)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts AES-256-GCM ciphertext using the user passphrase.
func Decrypt(ciphertext []byte, passphrase string) ([]byte, error) {
	key := DeriveKey(passphrase)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid passphrase or corrupted payload")
	}

	return plaintext, nil
}
