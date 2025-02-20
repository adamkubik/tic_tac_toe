package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"tic_tac_toe/internal/tic_tac_toe/models"

	_ "github.com/lib/pq"
)

func LoadConfig() *models.Config {
	return &models.Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
	}
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if defaultValue == "" {
			log.Fatalf("Environment variable %s is not set", key)
		}
		return defaultValue
	}
	return value
}

func GetDBConnectionString(c *models.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func InitDB(cfg *models.Config) *sql.DB {
	connStr := GetDBConnectionString(cfg)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}
