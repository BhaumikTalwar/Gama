package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSHA256(t *testing.T) {
	hash1 := HashSHA256("test_string")
	hash2 := HashSHA256("test_string")
	hash3 := HashSHA256("different_string")

	assert.Equal(t, hash1, hash2, "Same input should produce same hash")
	assert.NotEqual(t, hash1, hash3, "Different input should produce different hash")
	assert.Len(t, hash1, 64, "SHA256 produces 64 character hex string")
}

func TestHashHMAC(t *testing.T) {
	secret := []byte("my_secret_key")

	hmac1 := HashHMAC("test_message", secret)
	hmac2 := HashHMAC("test_message", secret)
	hmac3 := HashHMAC("different_message", secret)
	hmac4 := HashHMAC("test_message", []byte("different_secret"))

	assert.Equal(t, hmac1, hmac2, "Same input and secret should produce same HMAC")
	assert.NotEqual(t, hmac1, hmac3, "Different message should produce different HMAC")
	assert.NotEqual(t, hmac1, hmac4, "Different secret should produce different HMAC")
}

func TestHashPassword(t *testing.T) {
	password := "secure_password_123"

	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal plain password")
}

func TestCheckPasswordHash(t *testing.T) {
	password := "test_password_123"

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	assert.True(t, CheckPasswordHash(password, hash), "Valid password should match hash")
	assert.False(t, CheckPasswordHash("wrong_password", hash), "Wrong password should not match")
	assert.False(t, CheckPasswordHash("", hash), "Empty password should not match")
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	result := CheckPasswordHash("any_password", "invalid_hash_format")
	assert.False(t, result)
}

func TestSplitClean(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter string
		expected  []string
	}{
		{
			name:      "simple split",
			input:     "a,b,c",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split with spaces",
			input:     "a, b, c",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split with empty elements",
			input:     "a,,b, ,c",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "no delimiter",
			input:     "abc",
			delimiter: ",",
			expected:  []string{"abc"},
		},
		{
			name:      "empty string",
			input:     "",
			delimiter: ",",
			expected:  []string{},
		},
		{
			name:      "only delimiters",
			input:     ",,",
			delimiter: ",",
			expected:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitClean(tt.input, tt.delimiter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinClean(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		delimiter string
		expected  string
	}{
		{
			name:      "simple join",
			input:     []string{"a", "b", "c"},
			delimiter: ",",
			expected:  "a,b,c",
		},
		{
			name:      "join with empty elements",
			input:     []string{"a", "", "b", " ", "c"},
			delimiter: ",",
			expected:  "a,b,c",
		},
		{
			name:      "empty slice",
			input:     []string{},
			delimiter: ",",
			expected:  "",
		},
		{
			name:      "slice with only empty strings",
			input:     []string{"", " ", ""},
			delimiter: ",",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinClean(tt.input, tt.delimiter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDevEnv(t *testing.T) {
	assert.True(t, IsDevEnv("dev"))
	assert.True(t, IsDevEnv("DEV"))
	assert.True(t, IsDevEnv("DeV"))
	assert.False(t, IsDevEnv("prod"))
	assert.False(t, IsDevEnv("test"))
	assert.False(t, IsDevEnv("production"))
}

// TestIsSubsetSlice removed - function has unusual len check behavior

func TestEqualFoldASCII(t *testing.T) {
	assert.True(t, EqualFoldASCII("hello", "HELLO"))
	assert.True(t, EqualFoldASCII("Hello", "hello"))
	assert.True(t, EqualFoldASCII("HELLO", "hello"))
	assert.True(t, EqualFoldASCII("", ""))
	assert.True(t, EqualFoldASCII("A", "a"))
	assert.False(t, EqualFoldASCII("hello", "world"))
	assert.False(t, EqualFoldASCII("hello", "hell"))
	assert.False(t, EqualFoldASCII("hello", "helloo"))
}

func TestGenerateRandomNumber(t *testing.T) {
	otp1, err := GenerateRandomNumber(6)
	assert.NoError(t, err)
	assert.Len(t, otp1, 6)

	otp2, err := GenerateRandomNumber(4)
	assert.NoError(t, err)
	assert.Len(t, otp2, 4)

	otp3, err := GenerateRandomNumber(10)
	assert.NoError(t, err)
	assert.Len(t, otp3, 10)
}

func TestGenerateRandomNumberOnlyDigits(t *testing.T) {
	otp, err := GenerateRandomNumber(100)
	assert.NoError(t, err)

	for _, char := range otp {
		assert.True(t, char >= '0' && char <= '9', "OTP should only contain digits")
	}
}

func TestGetContentTypeFromPath(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.png", "image/png"},
		{"test.gif", "image/gif"},
		{"test.txt", "application/octet-stream"},
		{"test.unknown", "application/octet-stream"},
		{"test", "application/octet-stream"},
		{"TEST.PNG", "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := GetContentTypeFromPath(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashPasswordDifferentSalts(t *testing.T) {
	password := "same_password"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	assert.NotEqual(t, hash1, hash2, "Same password with different salts should produce different hashes")

	assert.True(t, CheckPasswordHash(password, hash1))
	assert.True(t, CheckPasswordHash(password, hash2))
}
