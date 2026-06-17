package database

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"room-api/internal/config"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	if cfg.MySQLDSN == "" {
		return nil, errors.New("MYSQL_DSN is required")
	}

	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
}
