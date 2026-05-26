//go:build integration

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/caching"
	db "github.com/BhaumikTalwar/Gama/internal/db/gen/sqlc"
	"github.com/BhaumikTalwar/Gama/internal/middleware"
	"github.com/BhaumikTalwar/Gama/internal/repository"
	otp "github.com/BhaumikTalwar/Gama/internal/service/Otp"
	"github.com/BhaumikTalwar/Gama/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlerIntegration(t *testing.T) (*gin.Engine, *repository.Repositories, func()) {
	t.Helper()
	testutil.SetTestConfig()
	config.GetAppConfig().MFASmsEnabled = false

	pool := testutil.SetupTestPool(t)
	rdb := redis.NewClient(&redis.Options{Addr: testutil.GetTestRedisAddr()})

	vm := caching.NewVersionManager(rdb)
	cs := caching.NewNoOpCacheService()
	repos := repository.SetupPostgresRepositories(pool, cs, vm, "test")

	testutil.CleanTestDB(t, pool)

	ctx := context.Background()
	_, err := repos.RBAC.GetRoleByName(ctx, "customer")
	if err != nil {
		_, err = repos.RBAC.CreateRole(ctx, db.CreateRoleParams{Name: "customer"})
		if err != nil {
			t.Fatalf("failed to create customer role: %v", err)
		}
	}

	dummyLogger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	handler := NewAuthHandler(repos, otp.NewOtpService(&config.OTPConfig{}, dummyLogger))
	mw := NewAuthMiddleWare(repos)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware(), middleware.CORSMiddleware("http://localhost:3000"))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "Gama"})
	})
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", CSRFMiddleware(), handler.Refresh)
		authGroup.POST("/refresh/logout", CSRFMiddleware(), handler.Logout)

		mfaGroup := authGroup.Group("/mfa")
		mfaGroup.Use(CSRFMiddleware())
		{
			setupGroup := mfaGroup.Group("/setup")
			setupGroup.Use(mw.JWTAuthMiddleware(ScopeMFASetup))
			{
				setupGroup.POST("/init", handler.InitMFASetup)
				setupGroup.POST("/finalize", handler.FinalizeMFASetup)
			}
			mfaGroup.POST("/verify", mw.JWTAuthMiddleware(ScopePreAuth), handler.VerifyMFA)
			mfaGroup.POST("/resend", mw.JWTAuthMiddleware(ScopePreAuth, ScopeMFASetup), handler.ResendMFAOTP)
		}

		authGroup.GET("/me", mw.JWTAuthMiddleware(), mw.RequireStrictAuth(), handler.Me)
		authGroup.PUT("/profile", CSRFMiddleware(), mw.JWTAuthMiddleware(), mw.RequireStrictAuth(), handler.UpdateProfile)
		authGroup.PUT("/password", CSRFMiddleware(), mw.JWTAuthMiddleware(), mw.RequireStrictAuth(), handler.ChangePassword)
	}

	return router, repos, func() {
		pool.Close()
		rdb.Close()
	}
}

func TestHandlerIntegration_RegisterAndSetupMFA(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	body := map[string]string{
		"email":        "integration@test.com",
		"password":     "password123",
		"first_name":   "Integration",
		"last_name":    "Test",
		"phone_number": "+919999999999",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Register response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "mfa_setup_required", resp["status"])

	cookies := w.Result().Cookies()
	var accessToken string
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessToken = c.Value
		}
	}
	require.NotEmpty(t, accessToken)
}

func TestHandlerIntegration_RegisterDuplicateEmail(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	body := map[string]string{
		"email":        "dupe@test.com",
		"password":     "password123",
		"first_name":   "Dup",
		"last_name":    "Test",
		"phone_number": "+919000000011",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Logf("First register response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandlerIntegration_LoginNoMFA(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	registerBody := map[string]string{
		"email":        "loginnotest@test.com",
		"password":     "password123",
		"first_name":   "Login",
		"last_name":    "NoMFA",
		"phone_number": "+919000000022",
	}
	b, _ := json.Marshal(registerBody)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Logf("Register response: %s", w.Body.String())
	}
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody := map[string]string{"email": "loginnotest@test.com", "password": "password123"}
	b, _ = json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "mfa_setup_required", resp["status"])
}

func TestHandlerIntegration_LoginWrongPassword(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	registerBody := map[string]string{
		"email":        "wrongpwd@test.com",
		"password":     "password123",
		"first_name":   "Wrong",
		"last_name":    "Pwd",
		"phone_number": "+919000000033",
	}
	b, _ := json.Marshal(registerBody)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Logf("Register response body: %s", w.Body.String())
	}
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody := map[string]string{"email": "wrongpwd@test.com", "password": "wrongpassword123"}
	b, _ = json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerIntegration_MeEndpoint(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	cookies := completeMFASetup(t, router, "me@test.com", "password123")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: cookies.accessToken})
	router.ServeHTTP(w, req)

	t.Logf("ME endpoint: status=%d body=%s", w.Code, w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
	var meResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &meResp)
	require.NoError(t, err)
	assert.Equal(t, "me@test.com", meResp["email"])
}

func TestHandlerIntegration_MeUnauthenticated(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandlerIntegration_RefreshTokenFlow(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	cookies := completeMFASetup(t, router, "refresh@test.com", "password123")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: cookies.accessToken})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: cookies.refreshToken, Path: "/api/v1/auth/refresh"})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: cookies.csrfToken})
	req.Header.Set("X-XSRF-Token", cookies.csrfToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerIntegration_RefreshWithoutCSRF(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandlerIntegration_Logout(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	cookies := completeMFASetup(t, router, "logout@test.com", "password123")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: cookies.accessToken})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: cookies.csrfToken})
	req.Header.Set("X-XSRF-Token", cookies.csrfToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerIntegration_UpdateProfile(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	cookies := completeMFASetup(t, router, "profile@test.com", "password123")

	profileBody := map[string]string{"first_name": "New", "last_name": "Profile"}
	pb, _ := json.Marshal(profileBody)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/auth/profile", bytes.NewReader(pb))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "access_token", Value: cookies.accessToken})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: cookies.csrfToken})
	req.Header.Set("X-XSRF-Token", cookies.csrfToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerIntegration_HealthEndpoint(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestHandlerIntegration_HealthLiveEndpoint(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/live", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerIntegration_LoginNonExistentUser(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	loginBody := map[string]string{"email": "nobody@test.com", "password": "password123"}
	b, _ := json.Marshal(loginBody)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerIntegration_CORSHeaders(t *testing.T) {
	router, _, cleanup := setupHandlerIntegration(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/api/v1/auth/register", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

type authCookies struct {
	accessToken  string
	refreshToken string
	csrfToken    string
}

var handlerPhoneCounter int

func nextPhone() string {
	handlerPhoneCounter++
	return fmt.Sprintf("+9190000002%02d", handlerPhoneCounter%100)
}

func completeMFASetup(t *testing.T, router *gin.Engine, email, password string) *authCookies {
	t.Helper()

	registerBody := map[string]string{
		"email":        email,
		"password":     password,
		"first_name":   "MFA",
		"last_name":    "User",
		"phone_number": nextPhone(),
	}
	b, _ := json.Marshal(registerBody)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var setupToken, csrfToken string
	for _, c := range w.Result().Cookies() {
		val, _ := url.QueryUnescape(c.Value)
		switch c.Name {
		case "access_token":
			setupToken = val
		case "csrf_token":
			csrfToken = val
		}
	}
	require.NotEmpty(t, setupToken)
	require.NotEmpty(t, csrfToken)

	initBody := map[string]string{"method": "totp"}
	ib, _ := json.Marshal(initBody)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/mfa/setup/init", bytes.NewReader(ib))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-XSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: setupToken})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var initResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &initResp)
	require.NoError(t, err)
	secret, ok := initResp["secret"].(string)
	require.True(t, ok, "secret should be present in init response")
	require.NotEmpty(t, secret)

	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	finalizeBody := map[string]string{"code": code}
	fb, _ := json.Marshal(finalizeBody)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/mfa/setup/finalize", bytes.NewReader(fb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-XSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: setupToken})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	cookies := &authCookies{}
	for _, c := range w.Result().Cookies() {
		val, _ := url.QueryUnescape(c.Value)
		switch c.Name {
		case "access_token":
			cookies.accessToken = val
		case "refresh_token":
			cookies.refreshToken = val
		case "csrf_token":
			cookies.csrfToken = val
		}
	}
	require.NotEmpty(t, cookies.accessToken)
	require.NotEmpty(t, cookies.refreshToken)
	require.NotEmpty(t, cookies.csrfToken)

	return cookies
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
