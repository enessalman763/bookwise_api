package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Gemini   GeminiConfig
	APIs     ExternalAPIsConfig
	Redis    RedisConfig
	Quiz     QuizConfig
}

type ServerConfig struct {
	Port       string
	GinMode    string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type GeminiConfig struct {
	APIKey string
	Model  string
}

type ExternalAPIsConfig struct {
	GoogleBooksAPIKey string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type QuizConfig struct {
	QuestionsCount int
	RetryLimit     int
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
			AllowedOrigins: []string{
				getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
			},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "bookwise_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
			Model:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		},
		APIs: ExternalAPIsConfig{
			GoogleBooksAPIKey: getEnv("GOOGLE_BOOKS_API_KEY", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Quiz: QuizConfig{
			QuestionsCount: getEnvAsInt("QUIZ_QUESTIONS_COUNT", 5),
			RetryLimit:     getEnvAsInt("QUIZ_RETRY_LIMIT", 3),
		},
	}

	// Validate required fields
	if config.Gemini.APIKey == "" {
		log.Println("Warning: GEMINI_API_KEY is not set")
	}

	AppConfig = config
	return config, nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

