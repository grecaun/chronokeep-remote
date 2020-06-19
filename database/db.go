package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/grecaun/chronokeepremote/util"

	"github.com/jinzhu/gorm"

	// required for database driver string
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	// required for database driver string
	_ "github.com/jinzhu/gorm/dialects/mysql"
	// required for database driver string
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// required for database driver string
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

var (
	db     *gorm.DB
	config *util.Config
)

// GetDatabase returns a gorm DB object based upon the configuration provided
func GetDatabase(inCfg *util.Config) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	config = inCfg

	conString := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s",
		config.DBHost,
		config.DBUser,
		config.DBName,
		config.DBPassword)

	// make sure our database is supported
	switch inCfg.DBDriver {
	case "postgres":
	case "cloudsqlpostgres":
	case "mysql":
	case "mssql":
	default:
		return nil, errors.New("invalid database type given")
	}
	dbCon, err := gorm.Open(inCfg.DBDriver, conString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	dbCon.DB().SetMaxIdleConns(0)
	dbCon.DB().SetMaxOpenConns(10)
	dbCon.DB().SetConnMaxLifetime(time.Minute * 5)

	db = dbCon
	return db, nil
}

// GetDB gets the gorm database object
func GetDB() (*gorm.DB, error) {
	return GetDatabase(config)
}
