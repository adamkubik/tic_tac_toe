package db

import (
	"database/sql"
	"fmt"
	"log"
	"tic_tac_toe/internal/tic_tac_toe/models"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func LoadConfig() *models.Config {
	viper.SetConfigName("cred_db")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	return &models.Config{
		DBHost:     viper.GetString("db.host"),
		DBPort:     viper.GetString("db.port"),
		DBUser:     viper.GetString("db.user"),
		DBPassword: viper.GetString("db.password"),
		DBName:     viper.GetString("db.name"),
	}
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
