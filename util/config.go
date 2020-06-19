package util

import (
	"os"

	"github.com/pkg/errors"
)

// GetConfig returns a config struct filled with values stored in local environment variables
func GetConfig() (*Config, error) {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("DB_NAME not found in environment")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return nil, errors.New("DB_HOST not found in environment")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, errors.New("DB_USER not found in environment")
	}

	dbPassword := os.Getenv("DB_PASSWORD")

	dbDriver := os.Getenv("DB_CONNECTOR")
	if dbDriver == "" {
		dbDriver = "postgres"
	}

	return &Config{
		DBName:     dbName,
		DBHost:     dbHost,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBDriver:   dbDriver,
	}, nil
}

// Config is the struct that holds all of the config values for connecting to a database
type Config struct {
	DBName     string `json:"dbName"`
	DBHost     string `json:"dbHost"`
	DBUser     string `json:"dbUser"`
	DBPassword string `json:"dbPass"`
	DBDriver   string `json:"dbDriver"`
}
