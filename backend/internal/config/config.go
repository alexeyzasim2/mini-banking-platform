package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string

	InitialBalanceUSDCents int64
	InitialBalanceEURCents int64

	DefaultPage  int
	DefaultLimit int
	MaxLimit     int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnvRequired("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "banking_platform"),
		JWTSecret:  getEnvRequired("JWT_SECRET"),
		Port:       getEnv("SERVER_PORT", "8080"),

		InitialBalanceUSDCents: getEnvInt64("INITIAL_BALANCE_USD_CENTS", 100000),
		InitialBalanceEURCents: getEnvInt64("INITIAL_BALANCE_EUR_CENTS", 50000),

		DefaultPage:  getEnvInt("DEFAULT_PAGE", 1),
		DefaultLimit: getEnvInt("DEFAULT_LIMIT", 10),
		MaxLimit:     getEnvInt("MAX_LIMIT", 100),
	}

	if len(config.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return config, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("ERROR: required environment variable %s is not set", key))
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvInt64(key string, defaultValue int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultValue
	}
	return v
}
