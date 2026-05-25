package utils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptStringWithRawAESKey(t *testing.T) {
	rawKey := []byte("0123456789abcdef0123456789abcdef")
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(rawKey, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := DecryptString(rawKey, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptStringWithBase64EncodedAESKey(t *testing.T) {
	rawKey := []byte("0123456789abcdef0123456789abcdef")
	b64Key := []byte(base64.StdEncoding.EncodeToString(rawKey))
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(b64Key, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := DecryptString(b64Key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptStringRejectsInvalidAESKeyLength(t *testing.T) {
	_, err := EncryptString([]byte("short-key"), "totp-secret-value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AES key length")
}
