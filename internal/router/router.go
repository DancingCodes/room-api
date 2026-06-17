package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"room-api/internal/auth"
	"room-api/internal/config"
	"room-api/internal/handler"
	"room-api/internal/middleware"
	"room-api/internal/realtime"
	"room-api/internal/repository"
	"room-api/internal/response"
	"room-api/internal/service"
)

func New(cfg config.Config, db *gorm.DB) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(corsAll())

	jwtSvc := auth.NewService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	emailCodeRepo := repository.NewEmailCodeRepository(db)
	mailSender, err := service.NewTencentSESMailSender(cfg)
	if err != nil {
		return nil, err
	}
	emailCodeSvc := service.NewEmailCodeService(emailCodeRepo, userRepo, mailSender)
	userSvc := service.NewUserService(userRepo, jwtSvc, emailCodeSvc)
	roomSvc := service.NewRoomService(roomRepo, userRepo)
	messageSvc := service.NewMessageService(messageRepo, roomRepo, userRepo)
	uploadSvc, err := service.NewUploadService(cfg)
	if err != nil {
		return nil, err
	}
	hub := realtime.NewHub()
	userHandler := handler.NewUserHandler(userSvc, emailCodeSvc)
	roomHandler := handler.NewRoomHandler(roomSvc, hub)
	messageHandler := handler.NewMessageHandler(messageSvc, hub)
	uploadHandler := handler.NewUploadHandler(uploadSvc, userSvc)
	wsHandler := handler.NewWSHandler(jwtSvc, roomSvc, hub)

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "up"})
	})

	api := r.Group("/api/v1")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register-code", userHandler.SendRegisterCode)
			authRoutes.POST("/register", userHandler.Register)
			authRoutes.POST("/login", userHandler.Login)
			authRoutes.POST("/password-reset-code", userHandler.SendPasswordResetCode)
			authRoutes.POST("/reset-password", userHandler.ResetPassword)
		}

		api.POST("/uploads/avatar", uploadHandler.UploadAvatar)

		users := api.Group("/users", middleware.Auth(jwtSvc))
		{
			users.GET("/me", userHandler.Me)
			users.PATCH("/me", userHandler.UpdateMe)
			users.POST("/me/avatar", uploadHandler.UpdateMyAvatar)
		}

		rooms := api.Group("/rooms", middleware.Auth(jwtSvc))
		{
			rooms.GET("", roomHandler.List)
			rooms.POST("", roomHandler.Create)
			rooms.GET("/:room_id", roomHandler.Detail)
			rooms.POST("/:room_id/join", roomHandler.Join)
			rooms.POST("/:room_id/leave", roomHandler.Leave)
			rooms.PATCH("/:room_id/mic", roomHandler.UpdateMicStatus)
			rooms.GET("/:room_id/messages", messageHandler.List)
			rooms.POST("/:room_id/messages", messageHandler.Create)
		}

		api.GET("/ws/rooms/:room_id", wsHandler.ConnectRoom)
	}

	return r, nil
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
