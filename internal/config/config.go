package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr            string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiry           time.Duration
	ESIAMode            string // mock | live
	ESIABaseURL         string
	ESIAClientID        string
	ESIAClientSecret    string
	ESIARedirectURI     string
	ESIAScopes          []string
	ESIACertificatePath string
	ESIAPrivateKeyPath  string
	PublicBaseURL       string
	UploadDir           string
	VisionMode          string // mock | openai
	OpenAIAPIKey        string
	OpenAIBaseURL       string
	OpenAIVisionModel   string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://driverhub:driverhub@localhost:5432/driverhub?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-me-in-production"),
		JWTExpiry:           getDurationEnv("JWT_EXPIRY", 24*time.Hour),
		ESIAMode:            getEnv("ESIA_MODE", "mock"),
		ESIABaseURL:         getEnv("ESIA_BASE_URL", "https://esia-portal1.test.gosuslugi.ru"),
		ESIAClientID:        getEnv("ESIA_CLIENT_ID", ""),
		ESIAClientSecret:    getEnv("ESIA_CLIENT_SECRET", ""),
		ESIARedirectURI:     getEnv("ESIA_REDIRECT_URI", "http://localhost:8080/api/v1/auth/esia/callback"),
		ESIAScopes:          strings.Split(getEnv("ESIA_SCOPES", "openid,fullname,birthdate,gender,snils,inn,mobile,email,id_doc,drivers_licence_doc"), ","),
		ESIACertificatePath: getEnv("ESIA_CERT_PATH", ""),
		ESIAPrivateKeyPath:  getEnv("ESIA_KEY_PATH", ""),
		PublicBaseURL:       getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		VisionMode:          getEnv("VISION_MODE", "mock"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:       getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIVisionModel:   getEnv("OPENAI_VISION_MODEL", "gpt-4o-mini"),
	}
	if cfg.VisionMode == "openai" && cfg.OpenAIAPIKey == "" {
		cfg.VisionMode = "mock"
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if hours, err := strconv.Atoi(v); err == nil {
		return time.Duration(hours) * time.Hour
	}
	return fallback
}
