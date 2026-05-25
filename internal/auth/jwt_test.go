package auth

import (
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	cfg := &config.AppConfig{
		AppName:             "TestApp",
		JWTKey:              "test_jwt_secret_key_that_is_32_bytes",
		AccessTokenDuration: 24 * time.Hour,
	}
	config.SetTestAppConfig(cfg)
}

func TestGenerateJWT(t *testing.T) {
	token, err := GenerateJWT("user-123", "admin", 1, ScopeAccess)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".")
}

func TestGenerateJWTWithDifferentScopes(t *testing.T) {
	tests := []struct {
		name    string
		scope   string
		wantErr bool
	}{
		{"access scope", ScopeAccess, false},
		{"pre_auth scope", ScopePreAuth, false},
		{"mfa_setup scope", ScopeMFASetup, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT("user-456", "customer", 1, tt.scope)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestValidateJWTValidToken(t *testing.T) {
	token, err := GenerateJWT("user-789", "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, ScopeAccess)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-789", claims.UID)
	assert.Equal(t, "admin", claims.Roles)
	assert.Equal(t, ScopeAccess, claims.Scope)
	assert.Equal(t, 1, claims.Ver)
}

func TestValidateJWTEmptyToken(t *testing.T) {
	claims, err := ValidateJWT("", ScopeAccess)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token string cannot be empty")
}

func TestValidateJWTInvalidSignature(t *testing.T) {
	token, err := GenerateJWT("user-101", "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	cfg := config.GetAppConfig()
	originalKey := cfg.JWTKey
	cfg.JWTKey = "different_secret_key_that_is_32_by"

	claims, err := ValidateJWT(token, ScopeAccess)

	cfg.JWTKey = originalKey

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateJWTDifferentScope(t *testing.T) {
	token, err := GenerateJWT("user-202", "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, ScopePreAuth)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "Invalid Scope")
}

func TestValidateJWTMalformedToken(t *testing.T) {
	claims, err := ValidateJWT("malformed.token.here", ScopeAccess)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateJWTEmptyScope(t *testing.T) {
	token, err := GenerateJWT("user-303", "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, "")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "Invalid Scope")
}

func TestJWTClaimsHaveExpiration(t *testing.T) {
	cfg := config.GetAppConfig()
	originalDuration := cfg.AccessTokenDuration
	cfg.AccessTokenDuration = 1 * time.Hour

	token, err := GenerateJWT("user-404", "customer", 1, ScopeAccess)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, ScopeAccess)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.NotNil(t, claims.ExpiresAt)
	assert.Greater(t, claims.ExpiresAt.Unix(), time.Now().Unix())

	cfg.AccessTokenDuration = originalDuration
}

func TestJWTClaimsHaveIssuer(t *testing.T) {
	token, err := GenerateJWT("user-505", "customer", 1, ScopeAccess)
	assert.NoError(t, err)

	tokenObj, _ := jwt.ParseWithClaims(token, &JWTClaims{}, nil)
	claims := tokenObj.Claims.(*JWTClaims)

	assert.Equal(t, "TestApp", claims.Issuer)
	assert.NotEmpty(t, claims.ID)
	assert.Equal(t, "user:user-505", claims.Subject)
}

func TestGenerateAndValidateRoundTrip(t *testing.T) {
	userID := "test-user-roundtrip"
	role := "editor"
	version := 2
	scope := ScopeAccess

	token, err := GenerateJWT(userID, role, version, scope)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, scope)
	assert.NoError(t, err)

	assert.Equal(t, userID, claims.UID)
	assert.Equal(t, role, claims.Roles)
	assert.Equal(t, version, claims.Ver)
	assert.Equal(t, scope, claims.Scope)
}
