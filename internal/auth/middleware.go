package auth

import (
	"net/http"
	"strings"

	"github.com/BhaumikTalwar/Gama/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	repos *repository.Repositories
}

func NewAuthMiddleWare(repo *repository.Repositories) *AuthMiddleware {
	return &AuthMiddleware{
		repos: repo,
	}
}

func (am *AuthMiddleware) JWTAuthMiddleware(scope ...string) gin.HandlerFunc {
	var scopeStr string = ScopeAccess
	if len(scope) != 0 {
		scopeStr = scope[0]
	}

	return func(ctx *gin.Context) {
		accessToken, err := ctx.Cookie(AccessTokenCookie)
		if accessToken == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Access Token"})
			return
		}

		claims, err := ValidateJWT(accessToken, scopeStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			return
		}

		uid, err := uuid.Parse(claims.UID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID in Token"})
			return
		}

		ctx.Set(CtxUserKey, uid)

		roleMap := make(map[string]bool)
		if claims.Roles != "" {
			for r := range strings.SplitSeq(claims.Roles, ",") {
				roleMap[strings.TrimSpace(r)] = true
			}
		}

		ctx.Set(ctxRolesKey, roleMap)
		ctx.Next()
	}
}

func (am *AuthMiddleware) RequireStrictAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIDVal, exists := ctx.Get(CtxUserKey)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
			return
		}

		user, err := am.repos.User.GetByID(ctx.Request.Context(), userIDVal.(uuid.UUID))
		if err != nil || user.Disabled {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User disabled or not found"})
			return
		}

		ctx.Next()
	}
}

func (am *AuthMiddleware) RequireStrictRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(CtxUserKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userID := userIDVal.(uuid.UUID)
		rolesRows, err := am.repos.RBAC.GetUserRoles(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify permissions"})
			return
		}

		roleMap := make(map[string]bool)
		for _, row := range rolesRows {
			roleMap[row.Name] = true
		}

		hasRole := false
		for _, allowed := range allowedRoles {
			if roleMap[allowed] {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions (Strict Check)"})
			return
		}

		c.Set(ctxRolesKey, roleMap)
		c.Next()
	}
}

func (am *AuthMiddleware) RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(CtxUserKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userID := userIDVal.(uuid.UUID)
		roles, err := am.repos.RBAC.GetUserRoles(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user roles"})
			c.Abort()
			return
		}

		roleMap := make(map[string]bool)
		for _, role := range roles {
			roleMap[role.Name] = true
		}

		hasRole := false
		for _, allowedRole := range allowedRoles {
			if roleMap[allowedRole] {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Set(ctxRolesKey, roles)
		c.Next()
	}
}
