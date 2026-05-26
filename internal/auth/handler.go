package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/internal/repository"
	"github.com/BhaumikTalwar/Gama/internal/service/Otp"
	"github.com/BhaumikTalwar/Gama/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pquerna/otp/totp"
)

type AuthHandler struct {
	repos         *repository.Repositories
	SmsOtpService otp.OtpService
}

func NewAuthHandler(repo *repository.Repositories, smsOtpService otp.OtpService) *AuthHandler {
	return &AuthHandler{
		repos:         repo,
		SmsOtpService: smsOtpService,
	}
}

type LoginRequest struct {
	UserEmail string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone_number" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	phoneNumber := strings.TrimSpace(req.Phone)

	if email == "" || firstName == "" || lastName == "" || phoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	_, err := h.repos.User.GetByEmail(c.Request.Context(), email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "User with this email already exists"})
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to validate registration input"})
		return
	}

	_, err = h.repos.User.GetByPhone(c.Request.Context(), phoneNumber)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "User with this phone number already exists"})
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to validate registration input"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
		return
	}

	var createdUser *db.User
	err = h.repos.ExecTx(c.Request.Context(), func(repos *repository.Repositories) error {
		user, err := repos.User.Create(c.Request.Context(), db.CreateUserParams{
			Email:        &email,
			Username:     email,
			PhoneNumber:  phoneNumber,
			PasswordHash: hashedPassword,
			FirstName:    &firstName,
			LastName:     &lastName,
		})
		if err != nil {
			return err
		}

		role, err := repos.RBAC.GetRoleByName(c.Request.Context(), "customer")
		if err != nil {
			return fmt.Errorf("customer role not found: %w", err)
		}

		_, err = repos.RBAC.AssignRole(c.Request.Context(), db.AssignRoleToUserParams{
			UserID: user.ID,
			RoleID: role.ID,
		})
		if err != nil {
			return err
		}

		createdUser = user
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key", "users_username_key":
				c.JSON(http.StatusConflict, gin.H{"message": "User with this email already exists"})
				return
			case "users_phone_number_key":
				c.JSON(http.StatusConflict, gin.H{"message": "User with this phone number already exists"})
				return
			default:
				c.JSON(http.StatusConflict, gin.H{"message": "User already exists"})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user", "error": err.Error()})
		return
	}

	appConfig := config.GetAppConfig()
	setupToken, err := GenerateJWT(createdUser.ID.String(), "", defaultVer, ScopeMFASetup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate setup token"})
		return
	}

	csrfTk := GenerateCSRFToken([]byte(appConfig.AppSecret))
	isProd := utils.EqualFoldASCII(appConfig.Env, "prod")

	options := []string{"totp"}
	if appConfig.MFASmsEnabled {
		options = append(options, "sms")
	}

	SetAccessTokenCookie(c, setupToken, int(PreAuthTokenTTL.Seconds()), isProd)
	SetCsrfCookie(c, csrfTk, int(appConfig.RefreshTokenDuration.Seconds()), isProd)

	c.JSON(http.StatusCreated, gin.H{
		"status":            "mfa_setup_required",
		"available_methods": options,
		"message":           "User created successfully. Please set up MFA to complete registration.",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	req := LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	user, err := h.repos.User.GetByEmail(c.Request.Context(), req.UserEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No Such User Exist"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password is incorrect"})
		return
	}

	appConfig := config.GetAppConfig()
	isProd := utils.EqualFoldASCII(appConfig.Env, "prod")
	csrfTk := GenerateCSRFToken([]byte(appConfig.AppSecret))

	mfaSettings, err := h.repos.MFA.GetSettings(c.Request.Context(), user.ID)
	if err != nil || !mfaSettings.UserMfaSetting.Enabled {
		setupToken, err := GenerateJWT(user.ID.String(), "", defaultVer, ScopeMFASetup)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cant generate a Pre Auth Token"})
			return
		}

		options := []string{"totp"}
		if appConfig.MFASmsEnabled {
			options = append(options, "sms")
		}

		SetAccessTokenCookie(c, setupToken, int(PreAuthTokenTTL.Seconds()), isProd)
		SetCsrfCookie(c, csrfTk, int(appConfig.RefreshTokenDuration.Seconds()), isProd)
		c.JSON(http.StatusOK, gin.H{
			"status":            "mfa_setup_required",
			"available_methods": options,
		})
		return
	}

	loginToken, err := GenerateJWT(user.ID.String(), "", defaultVer, ScopePreAuth)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cant generate a Pre Auth Token"})
		return
	}

	var sessionID string

	if mfaSettings.UserMfaSetting.Method == db.MfaTypeSms {
		phoneNumber := mfaSettings.UserMfaSetting.PhoneNumber
		if phoneNumber == nil || *phoneNumber == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "No phone number found for SMS MFA"})
			return
		}

		code, err := utils.GenerateRandomNumber(6)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cant generate the OTP"})
			return
		}

		sessionID, err = h.SmsOtpService.Send(code, *phoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send OTP", "error": err.Error()})
			return
		}

		hashedCode, err := utils.HashPassword(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cant hash the OTP"})
			return
		}

		externalID := sessionID
		_, err = h.repos.Token.CreateVerificationToken(c.Request.Context(), db.CreateVerificationTokenParams{
			UserID:     user.ID,
			TokenHash:  hashedCode,
			TokenType:  db.TokenTypeSmsOtp,
			ExternalID: &externalID,
			ExpiresAt:  time.Now().Add(5 * time.Minute),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create verification token"})
			return
		}
	}

	SetAccessTokenCookie(c, loginToken, int(PreAuthTokenTTL.Seconds()), isProd)
	SetCsrfCookie(c, csrfTk, int(appConfig.RefreshTokenDuration.Seconds()), isProd)

	response := gin.H{
		"status": "mfa_verification_required",
		"method": mfaSettings.UserMfaSetting.Method,
	}
	if mfaSettings.UserMfaSetting.Method == db.MfaTypeSms {
		response["session_id"] = sessionID
	}
	c.JSON(http.StatusOK, response)
}

type InitMFARequest struct {
	Method      string `json:"method" binding:"required,oneof=totp sms"`
	PhoneNumber string `json:"phone_number"`
}

func (h *AuthHandler) InitMFASetup(c *gin.Context) {
	var req InitMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Request Payload"})
		return
	}

	if req.Method == "sms" && !config.GetAppConfig().MFASmsEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMS MFA is not available"})
		return
	}

	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)
	user, err := h.repos.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No Such User Exist"})
		return
	}

	appConfig := config.GetAppConfig()
	if req.Method == string(db.MfaTypeTotp) {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      appConfig.AppName,
			AccountName: user.Email,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate TOTP key"})
			return
		}

		encryptedSecret, e := utils.EncryptString([]byte(appConfig.AESKey), key.Secret())
		if e != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cant Generate the TOTP QR"})
			return
		}

		h.repos.MFA.UpsertSettings(c.Request.Context(), db.UpsertMFASettingsParams{
			UserID:    userID,
			SecretKey: &encryptedSecret,
			Method:    db.MfaTypeTotp,
			Enabled:   false,
		})

		var buf bytes.Buffer
		img, err := key.Image(200, 200)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate TOTP QR image"})
			return
		}
		png.Encode(&buf, img)

		c.JSON(http.StatusOK, gin.H{
			"qr_code_url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
			"secret":      key.Secret(),
		})

	} else if req.Method == string(db.MfaTypeSms) {
		phoneNumber := user.PhoneNumber
		if req.PhoneNumber != "" && phoneNumber != user.PhoneNumber {
			phoneNumber = req.PhoneNumber
		}

		code, err := utils.GenerateRandomNumber(6)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cant generate the OTP"})
			return
		}

		sessionID, err := h.SmsOtpService.Send(code, phoneNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send OTP", "error": err.Error()})
			return
		}

		hashedCode, err := utils.HashPassword(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Cant hash the OTP"})
			return
		}

		externalID := sessionID
		_, err = h.repos.Token.CreateVerificationToken(c.Request.Context(), db.CreateVerificationTokenParams{
			UserID:     userID,
			TokenHash:  hashedCode,
			TokenType:  db.TokenTypeSmsOtp,
			ExternalID: &externalID,
			ExpiresAt:  time.Now().Add(5 * time.Minute),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create verification token"})
			return
		}

		h.repos.MFA.UpsertSettings(c.Request.Context(), db.UpsertMFASettingsParams{
			UserID:      userID,
			Method:      db.MfaTypeSms,
			PhoneNumber: &phoneNumber,
			Enabled:     false,
		})
		c.JSON(http.StatusOK, gin.H{
			"message":    "OTP sent",
			"session_id": sessionID,
		})
	}
}

type MFAVerifyRequest struct {
	Code      string `json:"code" binding:"required"`
	SessionID string `json:"session_id"`
}

func (h *AuthHandler) FinalizeMFASetup(c *gin.Context) {
	var req MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid payload"})
		return
	}

	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	settings, err := h.repos.MFA.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	appConfig := config.GetAppConfig()
	valid := false
	if settings.UserMfaSetting.Method == db.MfaTypeTotp {
		secret := settings.UserMfaSetting.SecretKey
		if secret == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No MFA TOTP KEY SET"})
			return
		}

		ogKey, err := utils.DecryptString([]byte(appConfig.AESKey), *secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cant Find teh Secret Key Set"})
			return
		}

		cleanCode := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, req.Code)

		cleanKey := strings.Map(func(r rune) rune {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '2' && r <= '7') {
				return r
			}
			return -1
		}, ogKey)

		valid = totp.Validate(cleanCode, strings.ToUpper(cleanKey))
	} else {
		var token *db.VerificationToken
		var err error

		if req.SessionID != "" {
			token, err = h.repos.Token.GetVerificationTokenByExternalID(c.Request.Context(), db.GetVerificationTokenByExternalIDParams{
				UserID:     userID,
				TokenType:  db.TokenTypeSmsOtp,
				ExternalID: &req.SessionID,
			})
		} else {
			token, err = h.repos.Token.GetLatestVerificationTokenForUser(c.Request.Context(), db.GetLatestVerificationTokenForUserParams{
				UserID:    userID,
				TokenType: db.TokenTypeSmsOtp,
			})
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid OTP token"})
			return
		}

		if token.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "OTP has expired"})
			return
		}

		if !utils.CheckPasswordHash(req.Code, token.TokenHash) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid OTP"})
			return
		}
		valid = true
		h.repos.Token.MarkTokenUsed(c.Request.Context(), token.ID)
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid code"})
		return
	}

	var backupCodes []string
	var backupBcryptCodes []string
	for range 3 {
		bk := utils.GetRandomKeyB64(32)
		backupCodes = append(backupCodes, bk)

		bkHash, _ := utils.HashPassword(bk)
		backupBcryptCodes = append(backupBcryptCodes, bkHash)
	}

	h.repos.MFA.SetEnabled(c.Request.Context(), db.EnableMFAParams{
		UserID:      userID,
		BackupCodes: backupBcryptCodes,
	})

	user, err := h.repos.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	rolesRows, err := h.repos.RBAC.GetUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user roles"})
		return
	}

	roles := []string{}
	for _, row := range rolesRows {
		roles = append(roles, row.Name)
	}

	accessToken, err := GenerateJWT(user.ID.String(), utils.JoinClean(roles, ","), defaultVer, ScopeAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token error"})
		return
	}

	userAgent := c.Request.UserAgent()
	tkParams, reftoken := GetRefreshTokenParam(user.ID, &userAgent, nil)
	_, err = h.repos.Token.CreateRefreshToken(c.Request.Context(), *tkParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate refresh token"})
		return
	}

	csrfTk := GenerateCSRFToken([]byte(appConfig.AppSecret))
	isProd := utils.EqualFoldASCII(appConfig.Env, "prod")
	SetAuthCookies(c, accessToken, reftoken, csrfTk, isProd)
	c.JSON(http.StatusOK, gin.H{
		"message":      "MFA Setup Complete",
		"backup_codes": backupCodes,
	})
}

func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing verification code"})
		return
	}
	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	settings, err := h.repos.MFA.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	if !settings.UserMfaSetting.Enabled || settings.UserMfaSetting.SecretKey == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	appConfig := config.GetAppConfig()
	valid := false
	if settings.UserMfaSetting.Method == db.MfaTypeTotp {
		secret := settings.UserMfaSetting.SecretKey
		if secret == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No MFA TOTP KEY SET"})
			return
		}

		ogKey, err := utils.DecryptString([]byte(appConfig.AESKey), *secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cant Find teh Secret Key Set"})
			return
		}

		cleanCode := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, req.Code)

		cleanKey := strings.Map(func(r rune) rune {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '2' && r <= '7') {
				return r
			}
			return -1
		}, ogKey)

		valid = totp.Validate(cleanCode, strings.ToUpper(cleanKey))
	} else {
		var token *db.VerificationToken
		var err error

		if req.SessionID != "" {
			token, err = h.repos.Token.GetVerificationTokenByExternalID(c.Request.Context(), db.GetVerificationTokenByExternalIDParams{
				UserID:     userID,
				TokenType:  db.TokenTypeSmsOtp,
				ExternalID: &req.SessionID,
			})
		} else {
			token, err = h.repos.Token.GetLatestVerificationTokenForUser(c.Request.Context(), db.GetLatestVerificationTokenForUserParams{
				UserID:    userID,
				TokenType: db.TokenTypeSmsOtp,
			})
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid OTP token"})
			return
		}

		if token.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "OTP has expired"})
			return
		}

		if !utils.CheckPasswordHash(req.Code, token.TokenHash) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid OTP"})
			return
		}
		valid = true
		h.repos.Token.MarkTokenUsed(c.Request.Context(), token.ID)
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid code"})
		return
	}

	user, err := h.repos.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	rolesRows, err := h.repos.RBAC.GetUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user roles"})
		return
	}

	roles := []string{}
	for _, row := range rolesRows {
		roles = append(roles, row.Name)
	}

	accessToken, err := GenerateJWT(user.ID.String(), utils.JoinClean(roles, ","), defaultVer, ScopeAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token error"})
		return
	}

	userAgent := c.Request.UserAgent()
	tkParams, reftoken := GetRefreshTokenParam(user.ID, &userAgent, nil)
	_, err = h.repos.Token.CreateRefreshToken(c.Request.Context(), *tkParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate refresh token"})
		return
	}

	csrfTk := GenerateCSRFToken([]byte(appConfig.AppSecret))
	isProd := utils.EqualFoldASCII(appConfig.Env, "prod")
	SetAuthCookies(c, accessToken, reftoken, csrfTk, isProd)
	c.JSON(http.StatusOK, gin.H{
		"message": "Verified",
	})
}

func (h *AuthHandler) ResendMFAOTP(c *gin.Context) {
	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	settings, err := h.repos.MFA.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "MFA not configured"})
		return
	}

	if settings.UserMfaSetting.Method != db.MfaTypeSms {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "SMS MFA not enabled for this user"})
		return
	}

	phoneNumber := settings.UserMfaSetting.PhoneNumber
	if phoneNumber == nil || *phoneNumber == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No phone number found"})
		return
	}

	code, err := utils.GenerateRandomNumber(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate OTP"})
		return
	}

	sessionID, err := h.SmsOtpService.Send(code, *phoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send OTP", "error": err.Error()})
		return
	}

	hashedCode, err := utils.HashPassword(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash OTP"})
		return
	}

	externalID := sessionID
	_, err = h.repos.Token.CreateVerificationToken(c.Request.Context(), db.CreateVerificationTokenParams{
		UserID:     userID,
		TokenHash:  hashedCode,
		TokenType:  db.TokenTypeSmsOtp,
		ExternalID: &externalID,
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create verification token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "OTP resent",
		"session_id": sessionID,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	appConfig := config.GetAppConfig()
	isProd := utils.EqualFoldASCII(appConfig.Env, "prod")

	refToken, err := c.Cookie(RefreshTokenCookie)
	if err != nil || refToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Refresh token not found"})
		return
	}

	tokenHash := HashRefToken(refToken, appConfig.AppSecret)
	tokenModel, err := h.repos.Token.GetRefreshToken(c.Request.Context(), tokenHash)

	if err != nil || tokenModel.Revoked {
		ClearAuthCookies(c, isProd)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or revoked token"})
		return
	}

	if tokenModel.ExpiresAt.Before(time.Now()) {
		reason := "Expired"
		_ = h.repos.Token.RevokeRefreshToken(c.Request.Context(), db.RevokeRefreshTokenParams{
			ID:            tokenModel.ID,
			RevokedReason: &reason,
		})
		ClearAuthCookies(c, isProd)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Refresh token expired"})
		return
	}

	user, err := h.repos.User.GetByID(c.Request.Context(), tokenModel.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	rolesRows, err := h.repos.RBAC.GetUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}
	roles := []string{}
	for _, row := range rolesRows {
		roles = append(roles, row.Name)
	}

	accessToken, err := GenerateJWT(user.ID.String(), utils.JoinClean(roles, ","), defaultVer, ScopeAccess)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token generation error"})
		return
	}

	csrfTk := GenerateCSRFToken([]byte(appConfig.AppSecret))

	remainingTTL := time.Until(tokenModel.ExpiresAt)
	if remainingTTL < appConfig.RefreshRotationThreshold {
		rotatedReason := "Rotated"
		err = h.repos.Token.RevokeRefreshToken(c.Request.Context(), db.RevokeRefreshTokenParams{
			ID:            tokenModel.ID,
			RevokedReason: &rotatedReason,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to rotate token"})
			return
		}

		userAgent := c.Request.UserAgent()
		tkParams, newRefTokenStr := GetRefreshTokenParam(user.ID, &userAgent, nil)

		_, err = h.repos.Token.CreateRefreshToken(c.Request.Context(), *tkParams)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to save new refresh token"})
			return
		}

		SetAuthCookies(c, accessToken, newRefTokenStr, csrfTk, isProd)
		c.JSON(http.StatusOK, gin.H{"message": "Token rotated"})
		return
	}

	SetAccessTokenCookie(c, accessToken, int(appConfig.AccessTokenDuration.Seconds()), isProd)
	SetCsrfCookie(c, csrfTk, int(appConfig.RefreshTokenDuration.Seconds()), isProd)

	c.JSON(http.StatusOK, gin.H{"message": "Set new Access Token"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	defer ClearAuthCookies(c, utils.EqualFoldASCII(config.GetAppConfig().Env, "prod"))

	refToken, err := c.Cookie(RefreshTokenCookie)
	if err != nil || refToken == "" {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	tokenModel, err := h.repos.Token.GetRefreshToken(c.Request.Context(), HashRefToken(refToken, config.GetAppConfig().AppSecret))
	if err != nil || tokenModel.Revoked {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Refresh token not Invalid"})
		return
	}

	reason := "Expired"
	_ = h.repos.Token.RevokeRefreshToken(c.Request.Context(), db.RevokeRefreshTokenParams{
		ID:            tokenModel.ID,
		RevokedReason: &reason,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Invalid User ID in context"})
		return
	}

	user, err := h.repos.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	rolesRows, err := h.repos.RBAC.GetUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user roles"})
		return
	}

	roles := make([]string, len(rolesRows))
	for i, row := range rolesRows {
		roles[i] = row.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID.String(),
		"email":        user.Email,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"phone_number": user.PhoneNumber,
		"roles":        roles,
		"avatar_url":   user.AvatarUrl,
	})
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Invalid User ID in context"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	firstName := req.FirstName
	lastName := req.LastName
	avatarURL := req.AvatarURL

	_, err := h.repos.User.Update(c.Request.Context(), db.UpdateUserParams{
		ID:        userID,
		FirstName: &firstName,
		LastName:  &lastName,
		AvatarUrl: &avatarURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update profile", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
	TOTPCode    string `json:"totp_code" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDVal, exists := c.Get(CtxUserKey)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	user, err := h.repos.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Incorrect old password"})
		return
	}

	settings, err := h.repos.MFA.GetSettings(c.Request.Context(), userID)
	if err != nil || !settings.UserMfaSetting.Enabled || settings.UserMfaSetting.SecretKey == nil || settings.UserMfaSetting.Method != db.MfaTypeTotp {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "TOTP MFA must be enabled to change password"})
		return
	}

	appConfig := config.GetAppConfig()
	ogKey, err := utils.DecryptString([]byte(appConfig.AESKey), *settings.UserMfaSetting.SecretKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to decrypt TOTP key"})
		return
	}

	cleanCode := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, req.TOTPCode)

	cleanKey := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '2' && r <= '7') {
			return r
		}
		return -1
	}, ogKey)

	if !totp.Validate(cleanCode, strings.ToUpper(cleanKey)) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid TOTP code"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash new password"})
		return
	}

	err = h.repos.User.UpdatePassword(c.Request.Context(), db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
