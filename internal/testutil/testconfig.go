package testutil

import (
	"time"

	"github.com/BhaumikTalwar/Gama/config"
)

func SetTestConfig() {
	cfg := &config.AppConfig{
		AppName:                  "TestApp",
		Env:                      "test",
		CorsAddresses:            []string{"http://localhost:3000"},
		AppSecret:                "test-app-secret-32-bytes-long!!!",
		JWTKey:                   "test-jwt-key-32-bytes-long-for-test!",
		AESKey:                   "test-aes-key-32-bytes-for-tests!",
		AccessTokenDuration:      15 * time.Minute,
		RefreshTokenDuration:     24 * time.Hour,
		RefreshRotationThreshold: 1 * time.Hour,
		MFASmsEnabled:            false,
	}
	config.SetTestAppConfig(cfg)
}
