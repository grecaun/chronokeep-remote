package handlers

import (
	"chronokeep/remote/auth"
	"chronokeep/remote/database/sqlite"
	"chronokeep/remote/types"
	"chronokeep/remote/util"
	"os"
	"strconv"
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
	t.Log("Adding Keys and reads.")
	reads := make([]types.Read, 0)
	for i := 0; i < 300; i++ {
		reads = append(reads, types.Read{
			Identifier:   strconv.Itoa(1000 + i),
			Seconds:      int64(25 * i),
			Milliseconds: 0,
			IdentType:    "chip",
			Type:         "reader",
		})
	}
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
	output.knownValues["writeName"] = "reader2"
	output.knownValues["write2"] = "030001-1ACSCT-K2389A-22423BAA"
	for _, key := range []types.Key{
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Value:             "030001-1ACSDD-K2389A-00123B",
			Type:              "write",
			Name:              "reader1",
			ValidUntil:        &times[0],
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Value:             "030001-1ACSDD-K2389A-22123B",
			Type:              "write",
			Name:              "reader2",
			ValidUntil:        &times[1],
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Value:             "0030001-1ACSCT-K2389A-22023BAA",
			Type:              "delete",
			Name:              "reader3",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22023B",
			Type:              "delete",
			Name:              "reader4",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22423B",
			Type:              "read",
			Name:              "user",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22423BAA",
			Type:              "write",
			Name:              "reader6",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[1].Identifier,
			Value:             "030001-1ACSCT-K2389A-22023BAA",
			Type:              "delete",
			Name:              "reader7",
			ValidUntil:        nil,
		},
		{
			AccountIdentifier: output.accounts[0].Identifier,
			Value:             "030001-1ACSDD-K2389A-001230B",
			Type:              "delete",
			Name:              "reader8",
			ValidUntil:        &times[0],
		},
	} {
		k, err := database.AddKey(key)
		if err != nil {
			t.Errorf("Error adding key: %v", err)
		}
		if k == nil {
			t.Errorf("Error adding key: %v -- %v : %v", key, key.AccountIdentifier, key.Name)
		} else {
			t.Logf("Adding reads for key: %v", k.Value)
			r, err := database.AddReads(k.Value, reads)
			if err != nil {
				t.Errorf("Error adding reads: %v", err)
			}
			if len(r) < 1 {
				t.Errorf("Keys not returned on add read call.")
			}
		}
	}
	when := time.Now()
	notes := []types.RequestNotification{
		{
			Type: "UPS_DISCONNECTED",
			When: when.UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_LOW_BATTERY",
			When: when.Add(time.Minute * -10).UTC().Format(time.RFC3339),
		},
	}
	err = database.SaveNotification(&notes[0], "030001-1ACSCT-K2389A-22423BAA")
	if err != nil {
		t.Fatalf("Unexpected error saving notification: %v", err)
	}
	err = database.SaveNotification(&notes[1], "030001-1ACSCT-K2389A-22023BAA")
	if err != nil {
		t.Fatalf("Unexpected error saving notification: %v", err)
	}
	output.keys[output.accounts[0].Email], err = database.GetAccountKeys(output.accounts[0].Email)
	if err != nil {
		t.Fatalf("Unexptected error getting keys: %v", err)
	}
	output.keys[output.accounts[1].Email], err = database.GetAccountKeys(output.accounts[1].Email)
	if err != nil {
		t.Fatalf("Unexptected error getting keys: %v", err)
	}
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
