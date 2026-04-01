package credential

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Encryptor struct {
	key []byte
}

func NewEncryptor(password string) *Encryptor {
	hash := sha256.Sum256([]byte(password))
	return &Encryptor{key: hash[:]}
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func EncryptCredentials(cred *string, password string) error {
	if *cred == "" {
		return nil
	}
	enc := NewEncryptor(password)
	encrypted, err := enc.Encrypt(*cred)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}
	*cred = encrypted
	return nil
}

func DecryptCredentials(cred *string, password string) error {
	if *cred == "" {
		return nil
	}
	enc := NewEncryptor(password)
	decrypted, err := enc.Decrypt(*cred)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}
	*cred = decrypted
	return nil
}
