package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	PassPepper   string
	ClientOrigin string
	CertFile     string
	KeyFile      string
	ServerPort   string
	ServerHost   string
	HMACKey      string
)

// LoadEnv loads the environment variables from a .env file and sets up global variables.
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Set global variables
	PassPepper = os.Getenv("PASS_PEPPER")
	ClientOrigin = os.Getenv("CLIENT_ORIGIN")
	CertFile = os.Getenv("CERT_FILE_PATH")
	KeyFile = os.Getenv("KEY_FILE_PATH")
	ServerPort = os.Getenv("SERVER_PORT")
	ServerHost = os.Getenv("SERVER_HOST")
	HMACKey = os.Getenv("HMAC_KEY")
}

// GetDSN constructs the database connection string from environment variables.
func GetDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("SSL_MODE")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}
