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
	COSBaseURL           string
	COSPathPrefix        string
	COSCDNURL            string
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
		COSBaseURL:           getEnv("COS_BASE_URL"),
		COSPathPrefix:        getEnv("COS_PATH_PREFIX"),
		COSCDNURL:            getEnv("COS_CDN_URL"),
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}
