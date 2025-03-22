package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var AppConfig *Config

// Config holds all configuration for the application
type Config struct {
	Server  ServerConfig
	MySql   MySqlConfig
	Redis   RedisConfig
	Jwt     JwtConfig
	Logging Logging
}

// ServerConfig holds server related configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database related configuration
type MySqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host string
	Port string
}

type JwtConfig struct {
	Secret string
}

type Logging struct {
	Level          string
	Type           string
	LogFileEnabled bool
	LogFilePath    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		MySql: MySqlConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASS", ""),
			Name:     getEnv("DB_NAME", "userdb"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		Jwt: JwtConfig{
			Secret: getEnv("JWT_SECRET_KEY", "secret"),
		},
		Logging: Logging{
			Level:       getEnv("LOG_LEVEL", "debug"),
			Type:        getEnv("LOG_TYPE", "json"),
			LogFilePath: getEnv("LOG_FILE_PATH", "logs/app.log"),
		},
	}

	AppConfig.Logging.LogFileEnabled, _ = strconv.ParseBool(getEnv("LOG_FILE_ENABLED", "true"))
}

// Helper function to get environment variable with a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
