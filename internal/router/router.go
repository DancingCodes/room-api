package router

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"room-api/internal/auth"
	"room-api/internal/config"
	"room-api/internal/handler"
	"room-api/internal/middleware"
	"room-api/internal/realtime"
	"room-api/internal/repository"
	"room-api/internal/service"
)

func New(cfg config.Config, db *gorm.DB) (*gin.Engine, error) {
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET不能为空")
	}
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return nil, errors.New("ADMIN_USERNAME和ADMIN_PASSWORD不能为空")
	}

	r := gin.Default()
	r.Use(corsAll())

	jwtSvc := auth.NewService(cfg.JWTSecret)
	adminSvc := auth.NewAdminService(cfg.JWTSecret, cfg.AdminUsername, cfg.AdminPassword)
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	emailCodeRepo := repository.NewEmailCodeRepository(db)
	appVersionRepo := repository.NewAppVersionRepository(db)
	mailSender, err := service.NewTencentSESMailSender(cfg)
	if err != nil {
		return nil, err
	}
	emailCodeSvc := service.NewEmailCodeService(emailCodeRepo, userRepo, mailSender)
	userSvc := service.NewUserService(userRepo, jwtSvc, emailCodeSvc)
	roomSvc := service.NewRoomService(roomRepo, userRepo)
	messageSvc := service.NewMessageService(messageRepo, roomRepo, userRepo)
	appVersionSvc := service.NewAppVersionService(appVersionRepo)
	voiceTokenSvc := service.NewAgoraTokenService(cfg.AgoraAppID, cfg.AgoraAppCertificate)
	uploadSvc, err := service.NewUploadService(cfg)
	if err != nil {
		return nil, err
	}
	hub := realtime.NewHub()
	userHandler := handler.NewUserHandler(userSvc, emailCodeSvc)
	roomHandler := handler.NewRoomHandler(roomSvc, voiceTokenSvc, hub)
	messageHandler := handler.NewMessageHandler(messageSvc, hub)
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	wsHandler := handler.NewWSHandler(jwtSvc, roomSvc, hub)
	appVersionHandler := handler.NewAppVersionHandler(appVersionSvc)
	adminHandler := handler.NewAdminHandler(adminSvc, roomSvc, appVersionSvc)
	adminAppVersionHandler := handler.NewAdminAppVersionHandler(appVersionSvc)

	app := r.Group("/api/v1/app")
	{

		app.GET("/version/latest", appVersionHandler.Latest)

		authRoutes := app.Group("/auth")
		{
			authRoutes.POST("/email-code", userHandler.SendEmailCode)
			authRoutes.POST("/email-login", userHandler.EmailLogin)
		}

		users := app.Group("/users", middleware.Auth(jwtSvc))
		{
			users.GET("/me", userHandler.Me)
			users.PATCH("/me", userHandler.UpdateMe)
		}

		uploads := app.Group("/uploads", middleware.Auth(jwtSvc))
		{
			uploads.POST("/image", uploadHandler.UploadImage)
		}

		rooms := app.Group("/rooms", middleware.Auth(jwtSvc))
		{
			rooms.GET("", roomHandler.List)
			rooms.POST("", roomHandler.Create)
			rooms.GET("/:room_id", roomHandler.Detail)
			rooms.GET("/:room_id/rtc-token", roomHandler.RTCToken)
			rooms.POST("/:room_id/join", roomHandler.Join)
			rooms.POST("/:room_id/leave", roomHandler.Leave)
			rooms.PATCH("/:room_id/mic", roomHandler.UpdateMicStatus)
			rooms.GET("/:room_id/messages", messageHandler.List)
			rooms.POST("/:room_id/messages", messageHandler.Create)
		}

		app.GET("/ws/rooms/:room_id", wsHandler.ConnectRoom)
	}

	admin := r.Group("/api/v1/admin")
	{
		admin.POST("/auth/login", adminHandler.Login)

		protected := admin.Group("", middleware.AdminAuth(adminSvc))
		{
			protected.GET("/dashboard", adminHandler.Dashboard)
			protected.GET("/rooms", adminHandler.Rooms)
			protected.GET("/rooms/:room_id", adminHandler.RoomDetail)
			protected.GET("/app-versions", adminAppVersionHandler.List)
			protected.POST("/app-versions", adminAppVersionHandler.Create)
			protected.PUT("/app-versions/:id", adminAppVersionHandler.Update)
			protected.POST("/app-versions/:id/publish", adminAppVersionHandler.Publish)
			protected.POST("/app-versions/:id/unpublish", adminAppVersionHandler.Unpublish)
		}
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
