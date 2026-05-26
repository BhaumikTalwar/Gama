package routes

import (
	"github.com/BhaumikTalwar/Gama/internal/auth"
	"github.com/BhaumikTalwar/Gama/internal/repository"
	"github.com/BhaumikTalwar/Gama/internal/service/Otp"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.RouterGroup, repos *repository.Repositories, otpService otp.OtpService) {
	authGroup := r.Group("/auth")
	authHandler := auth.NewAuthHandler(repos, otpService)
	authMiddleware := auth.NewAuthMiddleWare(repos)

	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	authGroup.POST("/refresh", auth.CSRFMiddleware(), authHandler.Refresh)
	authGroup.POST("/refresh/logout", auth.CSRFMiddleware(), authHandler.Logout)

	mfaGroup := authGroup.Group("/mfa")
	mfaGroup.Use(auth.CSRFMiddleware())
	{
		setupGroup := mfaGroup.Group("/setup")
		setupGroup.Use(authMiddleware.JWTAuthMiddleware(auth.ScopeMFASetup))
		{
			setupGroup.POST("/init", authHandler.InitMFASetup)
			setupGroup.POST("/finalize", authHandler.FinalizeMFASetup)
		}
		mfaGroup.POST("/verify", authMiddleware.JWTAuthMiddleware(auth.ScopePreAuth), authHandler.VerifyMFA)
		mfaGroup.POST("/resend", authMiddleware.JWTAuthMiddleware(auth.ScopePreAuth, auth.ScopeMFASetup), authHandler.ResendMFAOTP)
	}

	authGroup.GET("/me", authMiddleware.JWTAuthMiddleware(), authMiddleware.RequireStrictAuth(), authHandler.Me)
	authGroup.PUT("/profile", authMiddleware.JWTAuthMiddleware(), authMiddleware.RequireStrictAuth(), authHandler.UpdateProfile)
	authGroup.PUT("/password", authMiddleware.JWTAuthMiddleware(), authMiddleware.RequireStrictAuth(), authHandler.ChangePassword)
}
