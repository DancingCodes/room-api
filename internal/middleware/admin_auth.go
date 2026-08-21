package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"room-api/internal/auth"
	"room-api/internal/response"
)

func AdminAuth(adminSvc *auth.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, 401, "未登录")
			c.Abort()
			return
		}

		if _, err := adminSvc.Parse(parts[1]); err != nil {
			response.Error(c, 401, "未登录")
			c.Abort()
			return
		}
		c.Next()
	}
}
