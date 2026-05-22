package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser           string
	DBPass           string
	DBHost           string
	DBPort           string
	DBName           string
	ElevenLabsAPIKey string
	AppPort          string
}

// LoadConfig memuat semua environment variables dari file .env
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan env default sistem.")
	}

	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		DBUser:           os.Getenv("DB_USER"),
		DBPass:           os.Getenv("DB_PASS"),
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           os.Getenv("DB_PORT"),
		DBName:           os.Getenv("DB_NAME"),
		ElevenLabsAPIKey: os.Getenv("ELEVENLABS_API_KEY"),
		AppPort:          appPort,
	}
}

// BuildDSN menyusun Data Source Name untuk koneksi GORM MySQL
func (c *Config) BuildDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}
