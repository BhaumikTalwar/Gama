package utils

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func mustReadRand(key []byte) {
	_, err := rand.Read(key)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
}

func GetRandomKeyHex(length uint8) string {
	key := make([]byte, length)
	mustReadRand(key)

	return hex.EncodeToString(key)
}

func GetRandomKeyB64(length uint8) string {
	key := make([]byte, length)
	mustReadRand(key)

	return base64.URLEncoding.EncodeToString(key)
}

func GetRandomKeyB32(length uint8) string {
	key := make([]byte, length)
	mustReadRand(key)

	return base32.StdEncoding.EncodeToString(key)
}

func GetPayloadedAPIKey(payloads map[string]interface{}, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloads))
	return token.SignedString([]byte(jwtSecret))
}

func GetEncryptedPayloadedAPIKey(payloads map[string]interface{}, jwtSecret, aesSecret string) (string, string, error) {
	token, err := GetPayloadedAPIKey(payloads, jwtSecret)
	if err != nil {
		return "", "", err
	}

	noonce, enckey, err := EncryptAES([]byte(aesSecret), []byte(token))
	if err != nil {
		return "", "", err
	}

	return hex.EncodeToString(noonce), hex.EncodeToString(enckey), nil
}

func ParsePayloadedAPIKey(tokenstring, jwtSecret string) (map[string]interface{}, error) {
	payload, err := jwt.Parse(tokenstring, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := payload.Claims.(jwt.MapClaims); ok && payload.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid or unverifiable token claims")
}

func DecryptAndParseAPIKey(nonceHex, ciphertextHex string, jwtSecret, aesSecret string) (map[string]interface{}, error) {
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil, err
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}

	plaintextToken, err := DecryptAES([]byte(aesSecret), nonce, ciphertext)
	if err != nil {
		return nil, err
	}

	return ParsePayloadedAPIKey(string(plaintextToken), jwtSecret)
}
