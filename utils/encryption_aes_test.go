package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptAES_16ByteKey(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, nonce)
	require.NotEmpty(t, ciphertext)

	decrypted, err := DecryptAES(key, nonce, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptAES_24ByteKey(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key, plaintext)
	require.NoError(t, err)

	decrypted, err := DecryptAES(key, nonce, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptAES_32ByteKey(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key, plaintext)
	require.NoError(t, err)

	decrypted, err := DecryptAES(key, nonce, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptAES_WrongKey(t *testing.T) {
	key1 := []byte("0123456789abcdef0123456789abcdef")
	key2 := []byte("fedcba9876543210fedcba9876543210")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key1, plaintext)
	require.NoError(t, err)

	_, err = DecryptAES(key2, nonce, ciphertext)
	assert.Error(t, err)
}

func TestEncryptDecryptAES_CorruptedCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key, plaintext)
	require.NoError(t, err)

	corrupted := make([]byte, len(ciphertext))
	copy(corrupted, ciphertext)
	corrupted[0] ^= 0xff

	_, err = DecryptAES(key, nonce, corrupted)
	assert.Error(t, err)
}

func TestEncryptDecryptAES_CorruptedNonce(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("hello world")

	nonce, ciphertext, err := EncryptAES(key, plaintext)
	require.NoError(t, err)

	corruptedNonce := make([]byte, len(nonce))
	copy(corruptedNonce, nonce)
	corruptedNonce[0] ^= 0xff

	_, err = DecryptAES(key, corruptedNonce, ciphertext)
	assert.Error(t, err)
}

func TestEncryptDecryptAES_InvalidKeyLength(t *testing.T) {
	key := []byte("short")
	plaintext := []byte("hello world")

	_, _, err := EncryptAES(key, plaintext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AES key length")

	_, err = DecryptAES(key, []byte("nonce"), []byte("ciphertext"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AES key length")
}

func TestEncryptDecryptString_RawKey(t *testing.T) {
	rawKey := []byte("0123456789abcdef0123456789abcdef")
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(rawKey, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := DecryptString(rawKey, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptString_Base64Key(t *testing.T) {
	b64Key := []byte("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(b64Key, plaintext)
	require.NoError(t, err)

	decrypted, err := DecryptString(b64Key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptString_RawURLEncodedKey(t *testing.T) {
	b64Key := []byte("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY")
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(b64Key, plaintext)
	require.NoError(t, err)

	decrypted, err := DecryptString(b64Key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptString_InvalidBase64(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	_, err := EncryptString(key, "plaintext")
	require.NoError(t, err)

	_, err = DecryptString(key, "not-valid-base64!!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode base64")
}

func TestEncryptDecryptString_WrongKey(t *testing.T) {
	key1 := []byte("0123456789abcdef0123456789abcdef")
	key2 := []byte("fedcba9876543210fedcba9876543210")
	plaintext := "totp-secret-value"

	encrypted, err := EncryptString(key1, plaintext)
	require.NoError(t, err)

	_, err = DecryptString(key2, encrypted)
	assert.Error(t, err)
}

func TestEncryptString_EmptyPlaintext(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	encrypted, err := EncryptString(key, "")
	require.NoError(t, err)

	decrypted, err := DecryptString(key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestEncryptString_InvalidKeyLength(t *testing.T) {
	key := []byte("short")
	plaintext := "value"

	_, err := EncryptString(key, plaintext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AES key length")
}

func TestDecryptString_InvalidCiphertextLength(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	_, err := DecryptString(key, "c2hvcnQ=")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ciphertext length")
}
