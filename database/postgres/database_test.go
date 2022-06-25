package postgres

import (
	"chronokeep/remote/auth"
	"chronokeep/remote/database"
	"chronokeep/remote/util"
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"
)

const (
	dbName     = "remote_test"
	dbHost     = "database.lan"
	dbUser     = "remote_test"
	dbPassword = "remote_test"
	dbPort     = 5432
	dbDriver   = "postgres"
)

func testHashPassword(pass string) string {
	hash, _ := auth.HashPassword(pass)
	return hash
}

func setupTests(t *testing.T) (*Postgres, func(t *testing.T), error) {
	t.Log("Setting up testing database variables.")
	o := Postgres{}
	config := getTestConfig()
	t.Log("Initializing database.")
	// Connect to DB with database name.
	test, err := o.GetDatabase(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to database: %v", err)
	}
	if test == nil {
		return nil, nil, errors.New("database returned was nil")
	}

	// Check our database version.
	dbVersion := o.checkVersion()

	// Error checking version, most likely means tables are not created.
	if dbVersion < 1 {
		err = o.createTables()
		if err != nil {
			return nil, nil, err
		}
		// Otherwise check if our database is out of date and update if necessary.
	} else if dbVersion < database.CurrentVersion {
		err = o.updateTables(dbVersion, database.CurrentVersion)
		if err != nil {
			return nil, nil, err
		}
	}
	t.Log("Database initialized.")
	return &o, func(t *testing.T) {
		t.Log("Deleting old database.")
		err = o.dropTables()
		if err != nil {
			t.Fatalf("Error deleting database. %v", err)
			return
		}
		t.Log("Database successfully deleted.")
	}, nil
}

func setupOld() (*Postgres, error) {
	o := Postgres{}
	config := getTestConfig()
	// Connect to DB with database name.
	db, err := o.GetDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
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
				"account_id BIGSERIAL NOT NULL, " +
				"account_name VARCHAR(100) NOT NULL, " +
				"account_email VARCHAR(100) NOT NULL, " +
				"account_password VARCHAR(300) NOT NULL, " +
				"account_type VARCHAR(20) NOT NULL, " +
				"account_wrong_pass INT NOT NULL DEFAULT 0, " +
				"account_locked BOOL DEFAULT FALSE, " +
				"account_token VARCHAR(1000) NOT NULL DEFAULT '', " +
				"account_refresh_token VARCHAR(1000) NOT NULL DEFAULT '', " +
				"account_created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"account_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP," +
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
				"key_name VARCHAR(100) NOT NULL DEFAULT ''," +
				"key_value VARCHAR(100) NOT NULL, " +
				"key_type VARCHAR(20) NOT NULL, " +
				"reader_name VARCHAR(100) NOT NULL, " +
				"valid_until TIMESTAMPTZ DEFAULT NULL, " +
				"key_deleted BOOL DEFAULT FALSE, " +
				"key_created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"key_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, " +
				"UNIQUE(key_value), " +
				"UNIQUE(account_id, reader_name)," +
				"FOREIGN KEY (account_id) REFERENCES account(account_id)" +
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

	// Get a context and cancel function to create our tables, defer the cancel until we're done.
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()

	for _, single := range queries {
		_, err := db.Exec(ctx, single.query)
		if err != nil {
			return nil, fmt.Errorf("error executing %s query: %v", single.name, err)
		}
	}

	o.SetSetting("version", strconv.Itoa(1))

	return &o, nil
}

func TestSetupAndGet(t *testing.T) {
	t.Log("Setting up testing database variables.")
	o := &Postgres{}
	config := getTestConfig()
	t.Log("Initializing database.")
	err := o.Setup(config)
	defer o.dropTables()
	if err != nil {
		t.Fatalf("Error initializing database. %v", err)
	}
	t.Log("Database initialized.")
	if o.db == nil {
		t.Fatalf("db variable not set")
	}
	o.db.Close()
	o.updateDB(nil)
	_, err = o.GetDatabase(config)
	if err != nil {
		t.Fatalf("error getting database with config values: %v", err)
	}
	o.db.Close()
	o.updateDB(nil)
	_, err = o.GetDB()
	if err != nil {
		t.Fatalf("error getting database without config values: %v", err)
	}
	_, err = o.GetDB()
	if err != nil {
		t.Fatalf("error getting database without config values: %v", err)
	}
	err = o.dropTables()
	if err != nil {
		t.Fatalf("error deleting database: %v", err)
	}
}

func TestCheckVersion(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	version := db.checkVersion()
	if version != database.CurrentVersion {
		t.Fatalf("version found '%v' expected '%v'", version, database.CurrentVersion)
	}
}

func TestUpgrade(t *testing.T) {
	t.Log("Setting up testing database variables.")
	t.Log("Initializing database version 1.")
	db, err := setupOld()
	defer db.dropTables()
	if err != nil {
		t.Fatalf("Error initializing database. %v", err)
	}
	t.Log("Database initialized.")
	if db == nil || db.db == nil {
		t.Fatalf("db variable not set")
	}
	// Verify version 1
	version := db.checkVersion()
	if version != 1 {
		t.Fatalf("Version set to '%v' expected '1'.", version)
	}
	// Check for error on drop tables as well. Because we can.
	err = db.dropTables()
	if err != nil {
		t.Fatalf("error deleting database: %v", err)
	}
}

func getTestConfig() *util.Config {
	return &util.Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBName:     dbName,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBDriver:   dbDriver,
		AdminEmail: "admin@test.com",
		AdminName:  "tester number 1",
		AdminPass:  "password",
	}
}
