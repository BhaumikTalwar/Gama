package auth

import (
	"fmt"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
)

const (
	ScopeAccess   = "access"
	ScopePreAuth  = "pre_auth_mfa"
	ScopeMFASetup = "pre_auth_setup"
)

type JWTClaims struct {
	UID   string `json:"uid"`
	Roles string `json:"role"` // [role1, role2, ....]
	Ver   int    `json:"ver"`
	Scope string `json:"scope"`

	jwt.RegisteredClaims
}

func GenerateJWT(userID, role string, ver int, scope string) (string, error) {
	appConfig := config.GetAppConfig()
	accessClaims := JWTClaims{
		UID:   userID,
		Roles: role,
		Scope: scope,
		Ver:   ver,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        xid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(appConfig.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    appConfig.AppName,
			Subject:   fmt.Sprintf("user:%s", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessToken.SignedString([]byte(appConfig.JWTKey))
}

func ValidateJWT(tokenString string, allowedScopes ...string) (*JWTClaims, error) {
	appConfig := config.GetAppConfig()
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	if appConfig.JWTKey == "" {
		return nil, fmt.Errorf("missing JWT secret key in configuration")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: expected HS256, got %v", method.Alg())
		}

		return []byte(appConfig.JWTKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	if len(allowedScopes) > 0 {
		scopeValid := false
		for _, s := range allowedScopes {
			if claims.Scope == s {
				scopeValid = true
				break
			}
		}
		if !scopeValid {
			return nil, fmt.Errorf("Invalid Scope of Token")
		}
	}

	return claims, nil
}
