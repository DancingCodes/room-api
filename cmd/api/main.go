package main

import (
	"log"

	"room-api/internal/config"
	"room-api/internal/database"
	"room-api/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	r, err := router.New(cfg, db)
	if err != nil {
		log.Fatalf("创建路由失败: %v", err)
	}
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
