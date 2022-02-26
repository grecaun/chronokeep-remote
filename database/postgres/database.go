package postgres

import (
	"chronokeep/remote/database"
	"chronokeep/remote/util"
	"context"
	"strconv"

	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Postgres struct {
	db       *pgxpool.Pool
	config   *util.Config
	validate *validator.Validate
}

// GetDatabase Used to get a database with given configuration information.
func (p *Postgres) GetDatabase(inCfg *util.Config) (*pgxpool.Pool, error) {
	if p.db != nil {
		return p.db, nil
	}

	p.config = inCfg
	conString := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s",
		p.config.DBDriver,
		p.config.DBUser,
		p.config.DBPassword,
		p.config.DBHost,
		p.config.DBPort,
		p.config.DBName,
	)

	if !inCfg.Development {
		conString = conString + "?sslmode=require"
	}

	dbCon, err := pgxpool.Connect(context.Background(), conString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	p.db = dbCon
	return p.db, nil
}

// GetDB Used as a general way to get a database.
func (p *Postgres) GetDB() (*pgxpool.Pool, error) {
	if p.db != nil {
		return p.db, nil
	}
	if p.config != nil {
		return p.GetDatabase(p.config)
	}
	return nil, errors.New("config file not established")
}

// Setup Automatically creates and updates tables for all of our information.
func (p *Postgres) Setup(config *util.Config) error {
	// Set up Validator
	p.validate = validator.New()
	log.Info("Setting up database.")
	// Connect to DB with database name.
	_, err := p.GetDatabase(config)
	if err != nil {
		return fmt.Errorf("error connecting to database; %v", err)
	}

	dbVersion := p.checkVersion()

	if dbVersion < 1 {
		return p.createTables()
	} else if dbVersion < database.CurrentVersion {
		log.Info(fmt.Sprintf("Updating database from verison %v to %v", dbVersion, database.CurrentVersion))
		return p.updateTables(dbVersion, database.CurrentVersion)
	}
	return nil
}

func (p *Postgres) dropTables() error {
	db, err := p.GetDB()
	if err != nil {
		return fmt.Errorf("error connecting to database to drop tables: %v", err)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	_, err = db.Exec(
		ctx,
		"DROP TABLE read, api_key, settings;",
	)
	if err != nil {
		return fmt.Errorf("error dropping tables: %v", err)
	}
	return nil
}

func (p *Postgres) SetSetting(name, value string) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	_, err = db.Exec(
		ctx,
		"INSERT INTO settings(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=$2;",
		name,
		value,
	)
	if err != nil {
		return fmt.Errorf("error setting settings value: %v", err)
	}
	return nil
}

type myQuery struct {
	name  string
	query string
}

func (p *Postgres) createTables() error {
	log.Info("Creating database tables.")
	queries := []myQuery{
		// SETTINGS TABLE
		{
			name: "SettingsTable",
			query: "CREATE TABLE IF NOT EXISTS settings(" +
				"name VARCHAR(200) NOT NULL, " +
				"value VARCHAR(200) NOT NULL, " +
				"UNIQUE (name)" +
				");",
		},
		// KEY TABLE
		{
			name: "KeyTable",
			query: "CREATE TABLE IF NOT EXISTS api_key(" +
				"account_id VARCHAR(200) NOT NULL, " +
				"key_name VARCHAR(100) NOT NULL DEFAULT ''," +
				"key_value VARCHAR(100) NOT NULL, " +
				"key_type VARCHAR(20) NOT NULL, " +
				"reader_name VARCHAR(100) NOT NULL, " +
				"valid_until TIMESTAMPTZ DEFAULT NULL, " +
				"key_deleted BOOL DEFAULT FALSE, " +
				"key_created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"key_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value), " +
				"UNIQUE(account_id, reader_name)" +
				");",
		},
		// READ TABLE
		{
			name: "ReadTable",
			query: "CREATE TABLE IF NOT EXISTS read(" +
				"key_value VARCHAR(100) NOT NULL, " +
				"identifier VARCHAR(100) NOT NULL, " +
				"seconds BIGINT NOT NULL DEFAULT 0, " +
				"milliseconds INT NOT NULL DEFAULT 0, " +
				"ident_type VARCHAR(25) NOT NULL DEFAULT 'chip', " +
				"type VARCHAR(25) NOT NULL DEFAULT '', " +
				"read_created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value, identifier, seconds, milliseconds, ident_type), " +
				"FOREIGN KEY (key_value) REFERENCES api_key(key_value)" +
				");",
		},
		// UPDATE KEY FUNC
		{
			name: "UpdateKeyFunc",
			query: "CREATE OR REPLACE FUNCTION key_timestamp_column() " +
				"RETURNS TRIGGER AS $$ " +
				"BEGIN " +
				"NEW.key_updated_at = now();" +
				"RETURN NEW;" +
				"END;" +
				"$$ language 'plpgsql';",
		},
		// TRIGGERS FOR UPDATING UPDATED_AT timestamps
		{
			name:  "KeyTableTrigger",
			query: "CREATE TRIGGER update_key_timestamp BEFORE UPDATE ON api_key FOR EACH ROW EXECUTE PROCEDURE key_timestamp_column();",
		},
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()

	for _, single := range queries {
		log.Info(fmt.Sprintf("Executing query for: %s", single.name))
		_, err := p.db.Exec(ctx, single.query)
		if err != nil {
			return fmt.Errorf("error executing %s query: %v", single.name, err)
		}
	}

	p.SetSetting("version", strconv.Itoa(database.CurrentVersion))

	return nil
}

func (p *Postgres) checkVersion() int {
	log.Info("Checking database version.")
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	var name string
	var version string
	err := p.db.QueryRow(
		ctx,
		"SELECT name, value FROM settings WHERE name='version';",
	).Scan(&name, &version)
	if err != nil {
		return -1
	}
	v, err := strconv.Atoi(version)
	if err != nil {
		return -1
	}
	return v
}

func (p *Postgres) updateTables(oldVersion, newVersion int) error {
	return nil
}

func (p *Postgres) updateDB(newdb *pgxpool.Pool) {
	p.db = newdb
}

func (p *Postgres) Close() {
	p.db.Close()
}
