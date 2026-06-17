package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"room-api/internal/auth"
	"room-api/internal/response"
)

const userIDKey = "user_id"

func Auth(jwtSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, 401, "未登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, 401, "未登录")
			c.Abort()
			return
		}

		claims, err := jwtSvc.Parse(parts[1])
		if err != nil {
			response.Error(c, 401, "未登录")
			c.Abort()
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(userIDKey)
	if !ok {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}
