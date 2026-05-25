package otp

import "log/slog"

type MockOtpService struct {
	logger *slog.Logger
}

func NewMockOtpService(logger *slog.Logger) OtpService {
	return &MockOtpService{
		logger: logger,
	}
}

func (s *MockOtpService) Send(otp string, to string) (string, error) {
	s.logger.Info("Executing Mock OTP", "otp", otp, "to", to)
	s.logger.Info("Mock SMS sent", "to", to, "otp", otp)
	return "mock-session-id", nil
}
