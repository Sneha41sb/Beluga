package crypto

import (
	"bytes"
	"testing"
)

func TestAESGCMEncryptDecrypt(t *testing.T) {
	originalText := []byte("CONFIDENTIAL ULTRASONIC PAYLOAD: Secret Passcode 123456")
	passphrase := "MySuperSecretPassphrase!2026"

	// 1. Encrypt
	ciphertext, err := Encrypt(originalText, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Ensure ciphertext is different from original
	if bytes.Equal(ciphertext, originalText) {
		t.Fatalf("Ciphertext should not match original plaintext")
	}

	// 2. Decrypt with correct passphrase
	decryptedText, err := Decrypt(ciphertext, passphrase)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(decryptedText, originalText) {
		t.Errorf("Decrypted text mismatch: expected '%s', got '%s'", string(originalText), string(decryptedText))
	}

	// 3. Test wrong passphrase rejection
	_, err = Decrypt(ciphertext, "WrongPassphrase123")
	if err == nil {
		t.Fatalf("Expected error when decrypting with wrong passphrase, got nil")
	}
}
