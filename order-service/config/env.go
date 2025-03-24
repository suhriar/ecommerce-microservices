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
	Server ServerConfig
	MySql  MySqlConfig
	MySql2 MySqlConfig
	MySql3 MySqlConfig
	Redis  RedisConfig
	Jwt    JwtConfig
	Log    LogConfig
	Kafka  KafkaConfig
}

type ServerConfig struct {
	Port string
}

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

type LogConfig struct {
	Level          string
	Type           string
	LogFileEnabled bool
	LogFilePath    string
}

type KafkaConfig struct {
	Host string
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8002"),
		},
		MySql: MySqlConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3307"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASS", ""),
			Name:     getEnv("DB_NAME", "order-db"),
		},
		MySql2: MySqlConfig{
			Host:     getEnv("DB2_HOST", "localhost"),
			Port:     getEnv("DB2_PORT", "3308"),
			User:     getEnv("DB2_USER", "root"),
			Password: getEnv("DB2_PASS", ""),
			Name:     getEnv("DB2_NAME", "order-db"),
		},
		MySql3: MySqlConfig{
			Host:     getEnv("DB3_HOST", "localhost"),
			Port:     getEnv("DB3_PORT", "3309"),
			User:     getEnv("DB3_USER", "root"),
			Password: getEnv("DB3_PASS", ""),
			Name:     getEnv("DB3_NAME", "order-db"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		Jwt: JwtConfig{
			Secret: getEnv("JWT_SECRET_KEY", "secret"),
		},
		Log: LogConfig{
			Level:       getEnv("LOG_LEVEL", "debug"),
			Type:        getEnv("LOG_TYPE", "json"),
			LogFilePath: getEnv("LOG_FILE_PATH", "logs/app.log"),
		},
		Kafka: KafkaConfig{
			Host: getEnv("KAFKA_HOST", "localhost"),
			Port: getEnv("KAFKA_PORT", "9092"),
		},
	}

	AppConfig.Log.LogFileEnabled, _ = strconv.ParseBool(getEnv("LOG_FILE_ENABLED", "true"))

}

// Helper function to get environment variable with a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
