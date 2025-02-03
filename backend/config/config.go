package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Config Struct
type Config struct {
	MongoURI           string
	DBName             string
	JWTSecret          string
	RefreshTokenSecret string
	BCryptCost         int

	// JDOODLE API
	JDoodleClientID     string
	JDoodleClientSecret string
	JDoodleEndpoint     string
}

// Global infra instance
var AppConfig *Config

// Load Config
func LoadConfig() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found, falling back to system environment variables")
	}

	bcryptCost, err := strconv.Atoi(os.Getenv("BCRYPT_COST"))
	if err != nil || bcryptCost < bcrypt.MinCost {
		bcryptCost = 10
	}

	AppConfig = &Config{
		MongoURI:            os.Getenv("MONGO_URI"),
		DBName:              os.Getenv("DB_NAME"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		RefreshTokenSecret:  os.Getenv("REFRESH_TOKEN_SECRET"),
		BCryptCost:          bcryptCost,
		JDoodleClientID:     os.Getenv("JDOODLE_CLIENT_ID"),
		JDoodleClientSecret: os.Getenv("JDOODLE_CLIENT_SECRET"),
		JDoodleEndpoint:     os.Getenv("JDOODLE_ENDPOINT"),
	}
}
