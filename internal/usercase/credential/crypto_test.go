package credential

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	enc := NewEncryptor("test-password")
	plaintext := "AKIATEST12345678"

	encrypted, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("Encrypted text should not equal plaintext")
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	enc := NewEncryptor("test-password")
	plaintext := "same-text"

	enc1, _ := enc.Encrypt(plaintext)
	enc2, _ := enc.Encrypt(plaintext)

	if enc1 == enc2 {
		t.Error("Same plaintext should produce different ciphertext (due to random nonce)")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	enc := NewEncryptor("test-password")

	_, err := enc.Decrypt("not-valid-base64")
	if err == nil {
		t.Error("Should fail on invalid base64")
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	enc1 := NewEncryptor("password1")
	enc2 := NewEncryptor("password2")

	encrypted, _ := enc1.Encrypt("secret")

	_, err := enc2.Decrypt(encrypted)
	if err == nil {
		t.Error("Should fail with wrong password")
	}
}
