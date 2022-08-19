package handlers

import (
	"chronokeep/remote/auth"
	"chronokeep/remote/database/sqlite"
	"chronokeep/remote/types"
	"chronokeep/remote/util"
	"os"
	"testing"
	"time"
)

func setupTests(t *testing.T) (SetupVariables, func(t *testing.T)) {
	t.Log("Setting up sqlite database.")
	database = &sqlite.SQLite{}
	config = &util.Config{
		DBName:     "./remote_test.sqlite",
		DBHost:     "",
		DBUser:     "",
		DBPassword: "",
		DBPort:     0,
		DBDriver:   "sqlite3",
	}
	database.Setup(config)
	t.Log("Setting up config variables to export.")
	output := SetupVariables{
		testPassword1: "amazingpassword",
		testPassword2: "othergoodpassword",
		knownValues:   make(map[string]string),
		keys:          make(map[string][]types.Key),
		reads:         make(map[string]map[string][]types.Read),
	}
	// add accounts
	t.Log("Adding accounts.")
	output.knownValues["admin"] = "j@test.com"
	for _, account := range []types.Account{
		{
			Name:     "John Smith",
			Email:    "j@test.com",
			Type:     "admin",
			Password: testHashPassword(output.testPassword1),
		},
		{
			Name:     "Jerry Garcia",
			Email:    "jgarcia@test.com",
			Type:     "free",
			Password: testHashPassword(output.testPassword1),
		},
		{
			Name:     "Rose MacDonald",
			Email:    "rose2004@test.com",
			Type:     "paid",
			Password: testHashPassword(output.testPassword2),
		},
	} {
		database.AddAccount(account)
	}
	var err error
	output.accounts, err = database.GetAccounts()
	if err != nil {
		t.Fatalf("Unexpected error adding accounts: %v", err)
	}
	t.Log("Adding Keys.")
	// add keys, one expired, one with a timer, one write, one read, one delete, two different accounts
	times := []time.Time{
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		time.Now().Add(time.Hour * 20).Truncate(time.Second),
	}
	output.knownValues["expired"] = "030001-1ACSDD-K2389A-00123B"
	output.knownValues["expired2"] = "030001-1ACSDD-K2389A-001230B"
	output.knownValues["delete"] = "030001-1ACSCT-K2389A-22023B"
	output.knownValues["delete2"] = "030001-1ACSCT-K2389A-22023BAA"
	output.knownValues["delete3"] = "0030001-1ACSCT-K2389A-22023BAA"
	output.knownValues["read"] = "030001-1ACSCT-K2389A-22423B"
	output.knownValues["write"] = "030001-1ACSDD-K2389A-22123B"
	output.knownValues["write2"] = "030001-1ACSCT-K2389A-22423BAA"
	for _, key := range []types.Key{
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Name:              "expired1",
			Value:             "030001-1ACSDD-K2389A-00123B",
			Type:              "read",
			ReaderName:        "reader1",
			ValidUntil:        &times[0],
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Name:              "write",
			Value:             "030001-1ACSDD-K2389A-22123B",
			Type:              "write",
			ReaderName:        "reader2",
			ValidUntil:        &times[1],
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Value:             "0030001-1ACSCT-K2389A-22023BAA",
			Name:              "delete3",
			Type:              "delete",
			ReaderName:        "reader3",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22023B",
			Name:              "delete",
			Type:              "delete",
			ReaderName:        "reader4",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22423B",
			Name:              "read",
			Type:              "read",
			ReaderName:        "reader5",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22423BAA",
			Name:              "write2",
			Type:              "write",
			ReaderName:        "reader6",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22023BAA",
			Name:              "delete2",
			Type:              "delete",
			ReaderName:        "reader7",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Name:              "expired2",
			Value:             "030001-1ACSDD-K2389A-001230B",
			Type:              "delete",
			ReaderName:        "reader8",
			ValidUntil:        &times[0],
		},
	} {
		k, err := database.AddKey(key)
		if err != nil {
			t.Errorf("Error adding key: %v", err)
		}
		if k == nil {
			t.Errorf("Error adding key: %v -- %v : %v", key, key.AccountIdentifier, key.ReaderName)
		}
	}
	output.keys[output.accounts[0].Email], err = database.GetAccountKeys(output.accounts[0].Email)
	if err != nil {
		t.Fatalf("Unexptected error getting keys: %v", err)
	}
	output.keys[output.accounts[1].Email], err = database.GetAccountKeys(output.accounts[1].Email)
	if err != nil {
		t.Fatalf("Unexptected error getting keys: %v", err)
	}
	// add reads //->//
	t.Log("Adding reads.")
	return output, func(t *testing.T) {
		t.Log("Deleting old database.")
		database.Close()
		err := os.Remove(config.DBName)
		if err != nil {
			t.Fatalf("Error deleting database: %v", err)
		}
		t.Log("Cleanup successful.")
	}
}

func testHashPassword(pass string) string {
	hash, _ := auth.HashPassword(pass)
	return hash
}

type SetupVariables struct {
	accounts      []types.Account
	testPassword1 string
	testPassword2 string
	keys          map[string][]types.Key
	reads         map[string]map[string][]types.Read
	knownValues   map[string]string
}
