package util

import (
	"os"
	"strconv"

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

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil || dbPort < 0 {
		return nil, errors.New("DB_PORT not found in environment")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, errors.New("DB_USER not found in environment")
	}

	dbPassword := os.Getenv("DB_PASSWORD")

	dbDriver := os.Getenv("DB_CONNECTOR")
	if dbDriver == "" {
		dbDriver = "mysql"
	}

	recordInterval, err := strconv.Atoi(os.Getenv("RECORD_INTERVAL"))
	if err != nil && recordInterval < 60 {
		recordInterval = 300
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil && port < 200 {
		port = 8181
	}

	development := os.Getenv("VERSION") != "production"

	autotls := os.Getenv("AUTOTLS") == "enabled"

	domain := os.Getenv("DOMAIN")

	return &Config{
		DBName:         dbName,
		DBHost:         dbHost,
		DBPort:         dbPort,
		DBUser:         dbUser,
		DBPassword:     dbPassword,
		DBDriver:       dbDriver,
		RecordInterval: recordInterval,
		Port:           port,
		Development:    development,
		AutoTLS:        autotls,
		Domain:         domain,
	}, nil
}

// Config is the struct that holds all of the config values for connecting to a database
type Config struct {
	DBName         string
	DBHost         string
	DBPort         int
	DBUser         string
	DBPassword     string
	DBDriver       string
	RecordInterval int
	Port           int
	Development    bool
	AutoTLS        bool
	Domain         string
}
