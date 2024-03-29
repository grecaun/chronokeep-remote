package mysql

import (
	"chronokeep/remote/auth"
	"chronokeep/remote/database"
	"chronokeep/remote/types"
	"chronokeep/remote/util"
	"context"
	"strconv"

	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type MySQL struct {
	db       *sql.DB
	config   *util.Config
	validate *validator.Validate
}

// GetDatabase Used to get a database with given configuration information.
func (m *MySQL) GetDatabase(inCfg *util.Config) (*sql.DB, error) {
	if m.db != nil {
		return m.db, nil
	}
	if inCfg == nil {
		return nil, fmt.Errorf("no valid config supplied")
	}

	m.config = inCfg
	conString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.config.DBUser,
		m.config.DBPassword,
		m.config.DBHost,
		m.config.DBPort,
		m.config.DBName,
	)

	dbCon, err := sql.Open(m.config.DBDriver, conString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}
	dbCon.SetMaxIdleConns(database.MaxIdleConnections)
	dbCon.SetMaxOpenConns(database.MaxOpenConnections)
	dbCon.SetConnMaxLifetime(database.MaxConnectionLifetime)

	m.db = dbCon
	return m.db, nil
}

// GetDB Used as a general way to get a database.
func (m *MySQL) GetDB() (*sql.DB, error) {
	if m.db != nil {
		return m.db, nil
	}
	if m.config != nil {
		return m.GetDatabase(m.config)
	}
	return nil, errors.New("config file not established")
}

// Setup Automatically creates and updates tables for all of our information.
func (m *MySQL) Setup(config *util.Config) error {
	if config == nil {
		return fmt.Errorf("no valid config supplied")
	}
	// Set up Validator
	m.validate = validator.New()
	log.Info("Setting up database.")
	// Connect to DB with database name.
	_, err := m.GetDatabase(config)
	if err != nil {
		return fmt.Errorf("error connecting to database; %v", err)
	}

	dbVersion := m.checkVersion()

	if dbVersion < 1 {
		return m.createTables()
	} else if dbVersion < database.CurrentVersion {
		log.Info(fmt.Sprintf("Updating database from verison %v to %v", dbVersion, database.CurrentVersion))
		return m.updateTables(dbVersion, database.CurrentVersion)
	}

	// Check if there's an account created.
	accounts, err := m.GetAccounts()
	if err != nil {
		return fmt.Errorf("error checking for account: %v", err)
	}
	if len(accounts) < 1 {
		log.Info("Creating admin user.")
		if config.AdminName == "" || config.AdminEmail == "" || config.AdminPass == "" {
			return errors.New("admin account doesn't exist and proper credentions have not been supplied")
		}
		acc := types.Account{
			Name:     config.AdminName,
			Email:    config.AdminEmail,
			Password: config.AdminPass,
			Type:     "admin",
		}
		err = m.validate.Struct(acc)
		if err != nil {
			return fmt.Errorf("error validating base admin account on setup: %v", err)
		}
		acc.Password, err = auth.HashPassword(config.AdminPass)
		if err != nil {
			return fmt.Errorf("error hashing admin account password on setup: %v", err)
		}
		_, err = m.AddAccount(acc)
		if err != nil {
			return fmt.Errorf("error adding admin account on setup: %v", err)
		}
	}
	return nil
}

func (m *MySQL) dropTables() error {
	db, err := m.GetDB()
	if err != nil {
		return fmt.Errorf("error connecting to database to drop tables: %v", err)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	_, err = db.ExecContext(
		ctx,
		"DROP TABLE notification, a_read, api_key, settings, account;",
	)
	if err != nil {
		return fmt.Errorf("error dropping tables: %v", err)
	}
	return nil
}

func (m *MySQL) SetSetting(name, value string) error {
	db, err := m.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO settings(name, value) VALUES (?, ?) ON DUPLICATE KEY UPDATE value=VALUES(value);",
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

func (m *MySQL) createTables() error {
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
		// ACCOUNT TABLE
		{
			name: "AccountTable",
			query: "CREATE TABLE IF NOT EXISTS account(" +
				"account_id BIGINT NOT NULL AUTO_INCREMENT, " +
				"account_name VARCHAR(100) NOT NULL, " +
				"account_email VARCHAR(100) NOT NULL, " +
				"account_password VARCHAR(300) NOT NULL, " +
				"account_type VARCHAR(20) NOT NULL, " +
				"account_wrong_pass INT NOT NULL DEFAULT 0, " +
				"account_locked BOOL DEFAULT FALSE, " +
				"account_token VARCHAR(1000) NOT NULL DEFAULT '', " +
				"account_refresh_token VARCHAR(1000) NOT NULL DEFAULT '', " +
				"account_created_at DATETIME DEFAULT CURRENT_TIMESTAMP, " +
				"account_updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
				"account_deleted BOOL DEFAULT FALSE, " +
				"UNIQUE(account_email), " +
				"PRIMARY KEY (account_id)" +
				");",
		},
		// KEY TABLE
		{
			name: "KeyTable",
			query: "CREATE TABLE IF NOT EXISTS api_key(" +
				"account_id BIGINT NOT NULL, " +
				"key_name VARCHAR(100) NOT NULL," +
				"key_value VARCHAR(100) NOT NULL, " +
				"key_type VARCHAR(20) NOT NULL, " +
				"valid_until DATETIME DEFAULT NULL, " +
				"key_deleted BOOL DEFAULT FALSE, " +
				"key_created_at DATETIME DEFAULT CURRENT_TIMESTAMP, " +
				"key_updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value), " +
				"UNIQUE(account_id, key_name)," +
				"FOREIGN KEY (account_id) REFERENCES account(account_id)" +
				");",
		},
		// READ TABLE
		{
			name: "ReadTable",
			query: "CREATE TABLE IF NOT EXISTS a_read(" +
				"key_value VARCHAR(100) NOT NULL, " +
				"identifier VARCHAR(100) NOT NULL, " +
				"seconds BIGINT NOT NULL DEFAULT 0, " +
				"milliseconds INT NOT NULL DEFAULT 0, " +
				"ident_type VARCHAR(25) NOT NULL DEFAULT 'chip', " +
				"type VARCHAR(25) NOT NULL DEFAULT '', " +
				"antenna INT NOT NULL DEFAULT 0, " +
				"reader VARCHAR(50) NOT NULL DEFAULT '', " +
				"rssi VARCHAR(10) NOT NULL DEFAULT '', " +
				"read_created_at DATETIME DEFAULT CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value, identifier, seconds, milliseconds, ident_type), " +
				"FOREIGN KEY (key_value) REFERENCES api_key(key_value)" +
				");",
		},
		// NOTIFICATIONS TABLE
		{
			name: "NotificationsTable",
			query: "CREATE TABLE IF NOT EXISTS notification(" +
				"notification_id BIGINT NOT NULL AUTO_INCREMENT, " +
				"key_value VARCHAR(100) NOT NULL, " +
				"notification_type VARCHAR(100) NOT NULL, " +
				"notification_when BIGINT NOT NULL, " +
				"notification_created_at DATETIME DEFAULT CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value, notification_when), " +
				"FOREIGN KEY (key_value) REFERENCES api_key(key_value), " +
				"PRIMARY KEY (notification_id)" +
				");",
		},
	}

	if m.db == nil {
		return fmt.Errorf("database not setup")
	}

	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %v", err)
	}
	for _, single := range queries {
		log.Info(fmt.Sprintf("Executing query for: %s", single.name))
		_, err := tx.ExecContext(ctx, single.query)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing %s query: %v", single.name, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to commit transaction: %v", err)
	}

	m.SetSetting("version", strconv.Itoa(database.CurrentVersion))

	return nil
}

func (m *MySQL) checkVersion() int {
	log.Info("Checking database version.")
	if m.db == nil {
		return -1
	}
	res, err := m.db.Query("SELECT * FROM settings WHERE name='version';")
	if err != nil {
		return -1
	}
	if res.Next() {
		var name string
		var version int
		err = res.Scan(&name, &version)
		if err != nil {
			return -1
		}
		return version
	}
	return -1
}

func (m *MySQL) updateTables(oldVersion, newVersion int) error {
	if m.db == nil {
		return fmt.Errorf("database not set up")
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %v", err)
	}
	// Update from version 1 to 2
	if oldVersion < 2 && newVersion >= 2 {
		log.Debug("Updating to database version 2.")
		_, err := tx.ExecContext(
			ctx,
			"CREATE TABLE IF NOT EXISTS notification("+
				"notification_id BIGINT NOT NULL AUTO_INCREMENT, "+
				"key_value VARCHAR(100) NOT NULL, "+
				"notification_type VARCHAR(100) NOT NULL, "+
				"notification_when BIGINT NOT NULL, "+
				"notification_created_at DATETIME DEFAULT CURRENT_TIMESTAMP, "+
				"UNIQUE(key_value, notification_when), "+
				"FOREIGN KEY (key_value) REFERENCES api_key(key_value), "+
				"PRIMARY KEY (notification_id)"+
				");",
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating from verison %d to %d: %v", oldVersion, newVersion, err)
		}
	}
	_, err = tx.ExecContext(
		ctx,
		"UPDATE settings SET value=? WHERE name='version';",
		strconv.Itoa(newVersion),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating from version %d to %d: %v", oldVersion, newVersion, err)
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func (m *MySQL) updateDB(newdb *sql.DB) {
	m.db = newdb
}

func (m *MySQL) Close() {
	m.db.Close()
}
