package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"room-api/internal/response"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(corsAll())

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "up"})
	})

	_ = db
	return r
}

func corsAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
