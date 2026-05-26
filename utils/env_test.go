package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvInt_Exists(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	var val int
	GetEnvInt("TEST_INT", &val)
	assert.Equal(t, 42, val)
}

func TestGetEnvInt_NotExists(t *testing.T) {
	var val int
	GetEnvInt("NONEXISTENT_INT", &val)
	assert.Equal(t, 0, val)
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_INT_INVALID", "notanumber")
	defer os.Unsetenv("TEST_INT_INVALID")

	var val int
	GetEnvInt("TEST_INT_INVALID", &val)
	assert.Equal(t, 0, val)
}

func TestGetEnvBool_ExistsTrue(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	var val bool
	GetEnvBool("TEST_BOOL", &val)
	assert.True(t, val)
}

func TestGetEnvBool_ExistsFalse(t *testing.T) {
	os.Setenv("TEST_BOOL_FALSE", "false")
	defer os.Unsetenv("TEST_BOOL_FALSE")

	var val bool
	GetEnvBool("TEST_BOOL_FALSE", &val)
	assert.False(t, val)
}

func TestGetEnvBool_NotExists(t *testing.T) {
	var val bool
	GetEnvBool("NONEXISTENT_BOOL", &val)
	assert.False(t, val)
}

func TestGetEnvBool_Invalid(t *testing.T) {
	os.Setenv("TEST_BOOL_INVALID", "maybe")
	defer os.Unsetenv("TEST_BOOL_INVALID")

	var val bool
	GetEnvBool("TEST_BOOL_INVALID", &val)
	assert.False(t, val)
}
