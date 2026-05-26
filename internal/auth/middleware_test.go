package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "admin,customer", 1, ScopeAccess)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		uid, exists := c.Get(CtxUserKey)
		assert.True(t, exists)
		assert.IsType(t, uuid.UUID{}, uid)

		roles, exists := c.Get(ctxRolesKey)
		assert.True(t, exists)
		assert.NotNil(t, roles)

		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: "invalid-token"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	originalDuration := config.GetAppConfig().AccessTokenDuration
	config.GetAppConfig().AccessTokenDuration = -1 * time.Hour
	defer func() { config.GetAppConfig().AccessTokenDuration = originalDuration }()

	token, err := GenerateJWT(uuid.New().String(), "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	config.GetAppConfig().AccessTokenDuration = originalDuration

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_WrongScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware(ScopePreAuth))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_ValidScopePreAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "", 1, ScopePreAuth)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware(ScopePreAuth))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_ValidScopeMFASetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "", 1, ScopeMFASetup)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware(ScopeMFASetup))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_MultipleAllowedScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "", 1, ScopePreAuth)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware(ScopePreAuth, ScopeMFASetup))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_EmptyRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT(uuid.New().String(), "", 1, ScopeAccess)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		roles, exists := c.Get(ctxRolesKey)
		assert.True(t, exists)
		roleMap, ok := roles.(map[string]bool)
		assert.True(t, ok)
		assert.Empty(t, roleMap)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_InvalidUserIDInToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	token, err := GenerateJWT("not-a-uuid", "admin", 1, ScopeAccess)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mw.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookie, Value: token})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireStrictAuth_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	router := gin.New()
	router.Use(mw.RequireStrictAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireStrictRoles_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	router := gin.New()
	router.Use(mw.RequireStrictRoles("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRoles_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleWare(nil)

	router := gin.New()
	router.Use(mw.RequireRoles("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
