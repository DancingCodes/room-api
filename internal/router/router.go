package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"room-api/internal/auth"
	"room-api/internal/config"
	"room-api/internal/handler"
	"room-api/internal/middleware"
	"room-api/internal/repository"
	"room-api/internal/response"
	"room-api/internal/service"
)

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(corsAll())

	jwtSvc := auth.NewService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, jwtSvc)
	userHandler := handler.NewUserHandler(userSvc)

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "up"})
	})

	api := r.Group("/api/v1")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", userHandler.Register)
			authRoutes.POST("/login", userHandler.Login)
		}

		users := api.Group("/users", middleware.Auth(jwtSvc))
		{
			users.GET("/me", userHandler.Me)
			users.PATCH("/me", userHandler.UpdateMe)
		}
	}

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
