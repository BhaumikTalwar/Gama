package otp

import (
	"log/slog"

	"github.com/BhaumikTalwar/Gama/config"
)

type OtpService interface {
	Send(otp string, mno string) (string, error)
}

func NewOtpService(cfg *config.OTPConfig, logger *slog.Logger) OtpService {
	return NewMockOtpService(logger)
}
