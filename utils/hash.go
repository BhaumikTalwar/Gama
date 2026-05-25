package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashSHA256(Key string) string {
	hash := sha256.Sum256([]byte(Key))
	return hex.EncodeToString(hash[:])
}

func HashHMAC(Key string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(Key))
	return hex.EncodeToString(mac.Sum(nil))
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
