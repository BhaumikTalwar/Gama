package auth

import (
	"net/netip"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CSRFHeader = "X-XSRF-Token"
	CSRFCookie = "csrf_token"
	ctxCSRFKey = "csrf_token"
	defaultVer = 1

	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"

	refTokenCookiePath = "/api/v1/auth/refresh"

	CtxUserKey  = "user_id"
	ctxRolesKey = "user_roles"

	PreAuthTokenTTL = 5 * time.Minute
)

func GetRefreshTokenParam(userID uuid.UUID, userAgent *string, IPAddrs *netip.Addr) (*db.CreateRefreshTokenParams, string) {
	token := GenerateRefreshToken()
	appConfig := config.GetAppConfig()
	return &db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: HashRefToken(token, appConfig.AppSecret),
		ExpiresAt: time.Now().Add(appConfig.RefreshTokenDuration),
		UserAgent: userAgent,
		IpAddress: IPAddrs,
	}, token
}

func GenerateRefreshToken() string {
	return utils.GetRandomKeyB64(32)
}

func HashRefToken(token, key string) string {
	return utils.HashHMAC(token, []byte(key))
}

func SetAccessTokenCookie(c *gin.Context, accessToken string, duration int, secure bool) {
	c.SetCookie(
		AccessTokenCookie,
		accessToken,
		int(duration),
		"/",
		"",
		secure,
		true,
	)
}

func SetCsrfCookie(c *gin.Context, csrfToken string, duration int, secure bool) {
	c.SetCookie(
		CSRFCookie,
		csrfToken,
		int(duration),
		"/",
		"",
		secure,
		false,
	)
}

func SetRefreshTokenCookie(c *gin.Context, reftoken string, duration int, secure bool) {
	c.SetCookie(
		RefreshTokenCookie,
		reftoken,
		int(duration),
		refTokenCookiePath,
		"",
		secure,
		true,
	)
}

func SetAuthCookies(c *gin.Context, accessToken, refreshToken, csrfToken string, secure bool) {
	appConfig := config.GetAppConfig()
	SetAccessTokenCookie(c, accessToken, int(appConfig.AccessTokenDuration.Seconds()), secure)
	SetRefreshTokenCookie(c, refreshToken, int(appConfig.RefreshTokenDuration.Seconds()), secure)
	SetCsrfCookie(c, csrfToken, int(appConfig.RefreshTokenDuration.Seconds()), secure)
}

func ClearAuthCookies(c *gin.Context, secure bool) {
	c.SetCookie(AccessTokenCookie, "", -1, "/", "", secure, true)
	c.SetCookie(RefreshTokenCookie, "", -1, refTokenCookiePath, "", secure, true)
	c.SetCookie(CSRFCookie, "", -1, "/", "", secure, false)
}
