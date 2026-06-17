package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret            string
	MySQLDSN             string
	TencentSecretID      string
	TencentSecretKey     string
	TencentSESRegion     string
	TencentSESFrom       string
	TencentSESTemplateID string
	COSRegion            string
	COSBucket            string
	COSBaseURL           string
}

func Load() Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("load .env: %v", err)
	}

	return Config{
		JWTSecret:            getEnv("JWT_SECRET"),
		MySQLDSN:             getEnv("MYSQL_DSN"),
		TencentSecretID:      getEnv("TENCENT_SECRET_ID"),
		TencentSecretKey:     getEnv("TENCENT_SECRET_KEY"),
		TencentSESRegion:     getEnv("TENCENT_SES_REGION"),
		TencentSESFrom:       getEnv("TENCENT_SES_FROM"),
		TencentSESTemplateID: getEnv("TENCENT_SES_TEMPLATE_ID"),
		COSRegion:            getEnv("COS_REGION"),
		COSBucket:            getEnv("COS_BUCKET"),
		COSBaseURL:           getEnv("COS_BASE_URL"),
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}
