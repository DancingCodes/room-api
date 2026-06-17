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
		log.Fatalf("connect database: %v", err)
	}

	r := router.New(db)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
