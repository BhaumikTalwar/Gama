package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashRefToken(t *testing.T) {
	token := "test-token"
	key := "test-secret-key"

	hash1 := HashRefToken(token, key)
	hash2 := HashRefToken(token, key)

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2, "Same input should produce same hash")
}

func TestHashRefToken_DifferentInputs(t *testing.T) {
	key := "test-secret-key"

	hash1 := HashRefToken("token1", key)
	hash2 := HashRefToken("token2", key)

	assert.NotEqual(t, hash1, hash2)
}

func TestHashRefToken_DifferentKeys(t *testing.T) {
	token := "test-token"

	hash1 := HashRefToken(token, "key1")
	hash2 := HashRefToken(token, "key2")

	assert.NotEqual(t, hash1, hash2)
}

func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := GenerateRefreshToken()
		tokens[token] = true
	}
	assert.Greater(t, len(tokens), 1)
}

func TestHashRefToken_Length(t *testing.T) {
	token := "some-token"
	key := "test-key"

	hash := HashRefToken(token, key)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}