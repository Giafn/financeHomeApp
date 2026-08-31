package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi yang diambil dari environment variable / file .env.
type Config struct {
	AppEnv  string
	AppPort string
	AppURL  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	JWTExpiryHours int

	CORSAllowOrigins string

	S3Endpoint   string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	S3Region     string

	StorageDriver        string
	StorageLocalDir      string
	StoragePublicBaseURL string
	MaxUploadMB          int

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

// Load membaca file .env (jika ada) lalu memvalidasi env var wajib.
func Load() (*Config, error) {
	_ = godotenv.Load() // abaikan error kalau .env tidak ada, misal di production pakai env var langsung

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppURL:  getEnv("APP_URL", "http://localhost:3000"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "family_finance"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTExpiryHours:   getEnvAsInt("JWT_EXPIRY_HOURS", 72),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),

		S3Endpoint:   getEnv("S3_ENDPOINT", ""),
		S3Bucket:     getEnv("S3_BUCKET", ""),
		S3AccessKey:  getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:  getEnv("S3_SECRET_KEY", ""),
		S3Region:     getEnv("S3_REGION", ""),

		StorageDriver:        getEnv("STORAGE_DRIVER", ""),
		StorageLocalDir:      getEnv("STORAGE_LOCAL_DIR", "./data/uploads"),
		StoragePublicBaseURL: getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080"),
		MaxUploadMB:          getEnvAsInt("MAX_UPLOAD_MB", 10),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi di .env")
	}

	return cfg, nil
}

// DSN mengembalikan connection string Postgres untuk GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v := getEnv(key, "")
	if v == "" {
		return fallback
	}
	var i int
	if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
		return fallback
	}
	return i
}
