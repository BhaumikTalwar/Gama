package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	cfg := &config.AppConfig{
		AppName:              "TestApp",
		JWTKey:               "test_jwt_secret_key_that_is_32_bytes",
		AccessTokenDuration:  24 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		AppSecret:            "test_app_secret_key",
		Env:                  "test",
	}
	config.SetTestAppConfig(cfg)
}

func TestGenerateCSRFToken(t *testing.T) {
	key := []byte("test_secret_key")

	token1 := GenerateCSRFToken(key)
	assert.NotEmpty(t, token1)

	token2 := GenerateCSRFToken(key)
	assert.NotEqual(t, token1, token2, "Tokens should be unique")
}

func TestValidateCSRFToken(t *testing.T) {
	key := []byte("test_secret_key")

	token := GenerateCSRFToken(key)

	assert.True(t, ValidateCSRFToken(token, token))
	assert.False(t, ValidateCSRFToken(token, "different_token"))
	assert.False(t, ValidateCSRFToken(token, ""))
	assert.False(t, ValidateCSRFToken("", token))
}

func TestCSRFMiddlewareAllowsGetRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddlewareAllowsHeadRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddlewareAllowsOptionsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddlewareBlocksPostWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewareBlocksPutWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewareBlocksDeleteWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewareAllowsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	csrfToken := GenerateCSRFToken([]byte("test_secret_key"))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-XSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddlewareBlocksInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-XSRF-Token", "invalid_token")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "different_token"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewareBlocksMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "some_token"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewareBlocksMissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-XSRF-Token", "some_token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	SetAccessTokenCookie(c, "test_token", 3600, false)

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "access_token", cookies[0].Name)
	assert.Equal(t, "test_token", cookies[0].Value)
	assert.Equal(t, "/", cookies[0].Path)
	assert.False(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)
}

func TestSetRefreshTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	SetRefreshTokenCookie(c, "refresh_token", 86400, true)

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "refresh_token", cookies[0].Name)
	assert.Equal(t, "refresh_token", cookies[0].Value)
	assert.Equal(t, "/api/v1/auth/refresh", cookies[0].Path)
	assert.True(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)
}

func TestSetCsrfCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	SetCsrfCookie(c, "csrf_token", 3600, false)

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "csrf_token", cookies[0].Name)
	assert.Equal(t, "csrf_token", cookies[0].Value)
	assert.Equal(t, "/", cookies[0].Path)
	assert.False(t, cookies[0].Secure)
	assert.False(t, cookies[0].HttpOnly)
}

func TestClearAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	ClearAuthCookies(c, false)

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 3)

	cookieMap := make(map[string]string)
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}

	assert.Equal(t, "", cookieMap["access_token"])
	assert.Equal(t, "", cookieMap["refresh_token"])
	assert.Equal(t, "", cookieMap["csrf_token"])
}

func TestAuthConstants(t *testing.T) {
	assert.Equal(t, "X-XSRF-Token", CSRFHeader)
	assert.Equal(t, "csrf_token", CSRFCookie)
	assert.Equal(t, "access_token", AccessTokenCookie)
	assert.Equal(t, "refresh_token", RefreshTokenCookie)
	assert.Equal(t, "/api/v1/auth/refresh", refTokenCookiePath)
	assert.Equal(t, "user_id", CtxUserKey)
	assert.Equal(t, "user_roles", ctxRolesKey)
}

func TestGenerateRefreshToken(t *testing.T) {
	token1 := GenerateRefreshToken()
	token2 := GenerateRefreshToken()

	assert.NotEmpty(t, token1)
	assert.NotEqual(t, token1, token2, "Tokens should be unique")
	assert.GreaterOrEqual(t, len(token1), 32, "Token should be at least 32 characters")
}
