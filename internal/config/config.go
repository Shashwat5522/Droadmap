package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string

	// PostgreSQL (Master DB)
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string

	// MongoDB (Tenant DBs)
	MongoHost string
	MongoPort string
	MongoUser string
	MongoPass string

	// MinIO / S3
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIOBucket    string

	// AI Services
	GeminiAPIKey string // Google Gemini API Key (Free Tier)
	OpenAIAPIKey string // OpenAI API Key (Deprecated)
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnv("POSTGRES_DB", "master_db"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres123"),
		MongoHost:        getEnv("MONGO_HOST", "localhost"),
		MongoPort:        getEnv("MONGO_PORT", "27017"),
		MongoUser:        getEnv("MONGO_USER", ""),
		MongoPass:        getEnv("MONGO_PASS", ""),
		MinIOEndpoint:    getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:   getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOUseSSL:      getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOBucket:      getEnv("MINIO_BUCKET", "pdf-uploads"),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
	}
}

// PostgresConnString returns the PostgreSQL connection string
func (c *Config) PostgresConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}

// MongoConnString returns the MongoDB connection string
func (c *Config) MongoConnString() string {
	if c.MongoUser != "" && c.MongoPass != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", c.MongoUser, c.MongoPass, c.MongoHost, c.MongoPort)
	}
	return fmt.Sprintf("mongodb://%s:%s", c.MongoHost, c.MongoPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
