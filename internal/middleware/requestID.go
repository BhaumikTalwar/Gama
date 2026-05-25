package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

const (
	Header       = "X-Request-ID"
	RequestIdKey = "requestID"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(Header)
		if requestID == "" {
			requestID = xid.New().String()
		}

		c.Header(Header, requestID)
		c.Set(RequestIdKey, requestID)
		c.Next()
	}
}
