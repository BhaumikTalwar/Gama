package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func normalizeAESKey(key []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(key)
	switch len(trimmed) {
	case 16, 24, 32:
		return trimmed, nil
	}

	keyStr := string(trimmed)
	decoders := []func(string) ([]byte, error){
		base64.StdEncoding.DecodeString,
		base64.RawStdEncoding.DecodeString,
		base64.URLEncoding.DecodeString,
		base64.RawURLEncoding.DecodeString,
	}

	for _, decode := range decoders {
		decoded, err := decode(keyStr)
		if err != nil {
			continue
		}

		switch len(decoded) {
		case 16, 24, 32:
			return decoded, nil
		}
	}

	return nil, fmt.Errorf("invalid AES key length: %d bytes; expected 16, 24, or 32 bytes (raw or base64-decoded)", len(trimmed))
}

func EncryptAES(key, plaintext []byte) ([]byte, []byte, error) {
	normalizedKey, err := normalizeAESKey(key)
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

func DecryptAES(key, nonce, ciphertext []byte) ([]byte, error) {
	normalizedKey, err := normalizeAESKey(key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func EncryptString(key []byte, plaintext string) (string, error) {
	normalizedKey, err := normalizeAESKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize(), aesGCM.NonceSize()+len(plaintext)+aesGCM.Overhead())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func DecryptString(key []byte, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	normalizedKey, err := normalizeAESKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
