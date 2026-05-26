package otp

import (
	"log/slog"

	"github.com/BhaumikTalwar/Gama/config"
)

type OtpService interface {
	Send(otp string, mno string) (string, error)
}

func NewOtpService(cfg *config.OTPConfig, logger *slog.Logger) OtpService {
	if cfg.APIKey != "" {
		logger.Warn("OTP API key configured but no real SMS provider implemented, using mock service")
	}
	return NewMockOtpService(logger)
}
