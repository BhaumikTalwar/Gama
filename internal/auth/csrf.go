package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateCSRFToken(csrfKey []byte) string {
	h := hmac.New(sha256.New, csrfKey)
	fmt.Fprintf(h, "%d", time.Now().UnixNano())
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func ValidateCSRFToken(tokenFromCookie, tokenFromHeader string) bool {
	return hmac.Equal([]byte(tokenFromCookie), []byte(tokenFromHeader))
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			csrfFromHeader := c.GetHeader(CSRFHeader)
			if csrfFromHeader == "" {
				c.JSON(http.StatusForbidden, gin.H{"error": "missing CSRF header"})
				c.Abort()
				return
			}

			csrfFromCookie, err := c.Cookie(CSRFCookie)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "missing CSRF cookie"})
				c.Abort()
				return
			}

			if !ValidateCSRFToken(csrfFromCookie, csrfFromHeader) {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid CSRF token"})
				c.Abort()
				return
			}
		}

		c.Set(ctxCSRFKey, CSRFCookie)
		c.Next()
	}
}
