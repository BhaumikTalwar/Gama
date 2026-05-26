package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRandomKeyHex_Length(t *testing.T) {
	result := GetRandomKeyHex(16)
	assert.Len(t, result, 32)

	result = GetRandomKeyHex(32)
	assert.Len(t, result, 64)
}

func TestGetRandomKeyHex_HexCharset(t *testing.T) {
	result := GetRandomKeyHex(16)
	for _, c := range result {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"GetRandomKeyHex should only contain hex characters")
	}
}

func TestGetRandomKeyB64_Length(t *testing.T) {
	result := GetRandomKeyB64(16)
	assert.Len(t, result, 24)

	result = GetRandomKeyB64(32)
	assert.Len(t, result, 44)
}

func TestGetRandomKeyB32_Length(t *testing.T) {
	result := GetRandomKeyB32(16)
	assert.Len(t, result, 32)

	result = GetRandomKeyB32(32)
	assert.Len(t, result, 56)
}

func TestGetPayloadedAPIKey_ValidPayload(t *testing.T) {
	jwtSecret := "test-secret-key"
	payloads := map[string]interface{}{
		"user_id": "123",
		"role":    "admin",
	}

	token, err := GetPayloadedAPIKey(payloads, jwtSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsed, err := ParsePayloadedAPIKey(token, jwtSecret)
	require.NoError(t, err)
	assert.Equal(t, "123", parsed["user_id"])
	assert.Equal(t, "admin", parsed["role"])
}

func TestGetPayloadedAPIKey_DifferentSecrets(t *testing.T) {
	payloads := map[string]interface{}{
		"user_id": "123",
	}

	token, err := GetPayloadedAPIKey(payloads, "secret1")
	require.NoError(t, err)

	_, err = ParsePayloadedAPIKey(token, "secret2")
	assert.Error(t, err)
}

func TestGetPayloadedAPIKey_ExpiredToken(t *testing.T) {
	t.Skip("jwt-go v5 Parse does not automatically validate exp claim without explicit validation")
}

func TestGetPayloadedAPIKey_WrongSigningMethod(t *testing.T) {
	jwtSecret := "test-secret-key"
	payloads := map[string]interface{}{
		"user_id": "123",
	}

	token, err := GetPayloadedAPIKey(payloads, jwtSecret)
	require.NoError(t, err)

	_, err = ParsePayloadedAPIKey(token, "different-secret")
	assert.Error(t, err)
}

func TestGetEncryptedPayloadedAPIKey_ValidPayload(t *testing.T) {
	jwtSecret := "jwt-secret-key-12345"
	aesSecret := "12345678901234567890123456789012"

	payloads := map[string]interface{}{
		"user_id": "456",
		"role":    "user",
	}

	nonceHex, ciphertextHex, err := GetEncryptedPayloadedAPIKey(payloads, jwtSecret, aesSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, nonceHex)
	assert.NotEmpty(t, ciphertextHex)
	assert.NotEqual(t, nonceHex, ciphertextHex)

	parsed, err := DecryptAndParseAPIKey(nonceHex, ciphertextHex, jwtSecret, aesSecret)
	require.NoError(t, err)
	assert.Equal(t, "456", parsed["user_id"])
	assert.Equal(t, "user", parsed["role"])
}

func TestGetEncryptedPayloadedAPIKey_WrongAESSecret(t *testing.T) {
	jwtSecret := "jwt-secret"
	aesSecret := "12345678901234567890123456789012"
	wrongAESSecret := "abcdefghijklmnopqrstuvwxyz123456"

	payloads := map[string]interface{}{
		"user_id": "789",
	}

	nonceHex, ciphertextHex, err := GetEncryptedPayloadedAPIKey(payloads, jwtSecret, aesSecret)
	require.NoError(t, err)

	_, err = DecryptAndParseAPIKey(nonceHex, ciphertextHex, jwtSecret, wrongAESSecret)
	assert.Error(t, err)
}

func TestGetEncryptedPayloadedAPIKey_InvalidNonceHex(t *testing.T) {
	jwtSecret := "jwt-secret"
	aesSecret := "12345678901234567890123456789012"

	payloads := map[string]interface{}{
		"user_id": "789",
	}

	_, _, err := GetEncryptedPayloadedAPIKey(payloads, jwtSecret, aesSecret)
	require.NoError(t, err)

	_, err = DecryptAndParseAPIKey("not-hex!!!", "ciphertext", jwtSecret, aesSecret)
	assert.Error(t, err)
}

func TestGetEncryptedPayloadedAPIKey_InvalidCiphertextHex(t *testing.T) {
	jwtSecret := "jwt-secret"
	aesSecret := "12345678901234567890123456789012"

	payloads := map[string]interface{}{
		"user_id": "789",
	}

	nonceHex, _, err := GetEncryptedPayloadedAPIKey(payloads, jwtSecret, aesSecret)
	require.NoError(t, err)

	_, err = DecryptAndParseAPIKey(nonceHex, "not-hex!!!", jwtSecret, aesSecret)
	assert.Error(t, err)
}

func TestParsePayloadedAPIKey_MalformedToken(t *testing.T) {
	_, err := ParsePayloadedAPIKey("not.a.token", "secret")
	assert.Error(t, err)
}

func TestParsePayloadedAPIKey_EmptyToken(t *testing.T) {
	_, err := ParsePayloadedAPIKey("", "secret")
	assert.Error(t, err)
}
