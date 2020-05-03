package middlewares

import (
	"github.com/gin-gonic/gin"
)

func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if cip := c.ClientIP(); cip != "" {
			c.Request.RemoteAddr = cip
		}
		c.Next()
	}
}