package handlers

import (
	db "chronokeep/remote/database"
	"chronokeep/remote/database/mysql"
	"chronokeep/remote/database/postgres"
	"chronokeep/remote/util"
	"errors"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

var (
	database db.Database
	config   *util.Config
)

func Setup(inCfg *util.Config) error {
	config = inCfg
	switch config.DBDriver {
	case "mysql":
		log.Info("Database set to MySQL")
		database = &mysql.MySQL{}
		return database.Setup(config)
	case "postgres":
		log.Info("Database set to Postgresql")
		database = &postgres.Postgres{}
		return database.Setup(config)
	default:
		return errors.New("unknown database driver specified")
	}
}

func Finalize() {
	database.Close()
}

func (h *Handler) Setup() {
	// Set up Validator.
	h.validate = validator.New()
}
