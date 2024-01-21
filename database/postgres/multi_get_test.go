package postgres

import (
	"chronokeep/remote/types"
	"testing"
	"time"
)

func setupMultiTests() {
	if len(accounts) < 1 {
		accounts = []types.Account{
			{
				Name:     "John Smith",
				Email:    "j@test.com",
				Type:     "admin",
				Password: testHashPassword("password"),
			},
			{
				Name:     "Rose MacDonald",
				Email:    "rose2004@test.com",
				Type:     "paid",
				Password: testHashPassword("password"),
			},
			{
				Name:     "Tia Johnson",
				Email:    "tiatheway@test.com",
				Type:     "free",
				Password: testHashPassword("password"),
			},
		}
	}
}

func TestGetKeyAndAccount(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("Error setting up test. %v", err)
	}
	defer finalize(t)
	setupMultiTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	times := []time.Time{
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		time.Date(2016, 4, 1, 4, 11, 5, 0, time.Local),
	}
	keys := []types.Key{
		{
			AccountIdentifier: account1.Identifier,
			Value:             "030001-1ACSDD-K2389A-00123B",
			Type:              "default",
			Name:              "reader1",
			ValidUntil:        &times[0],
		},
		{
			AccountIdentifier: account2.Identifier,
			Value:             "030001-1ACSDD-KH789A-00123B",
			Type:              "delete",
			Name:              "reader2",
			ValidUntil:        &times[1],
		},
	}
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	mult, err := db.GetKeyAndAccount(keys[0].Value)
	if err != nil {
		t.Fatalf("Error getting key and account: %v", err)
	}
	if mult == nil || mult.Key == nil || mult.Account == nil {
		t.Fatal("Key or Account was nil.")
	}
	if !mult.Account.Equals(account1) || !mult.Key.Equal(&keys[0]) {
		t.Errorf("Account expected: %+v; Found %+v;\nKey expected: %+v; Found %+v;", *account1, *mult.Account, keys[0], *mult.Key)
	}
	mult, err = db.GetKeyAndAccount(keys[1].Value)
	if err != nil {
		t.Fatalf("Error getting key and account: %v", err)
	}
	if mult == nil || mult.Key == nil || mult.Account == nil {
		t.Fatal("Key or Account was nil.")
	}
	if !mult.Account.Equals(account2) || !mult.Key.Equal(&keys[1]) {
		t.Errorf("Account expected: %+v; Found %+v;\nKey expected: %+v; Found %+v;", *account2, *mult.Account, keys[1], *mult.Key)
	}
}

func TestBadDatabaseMultiGet(t *testing.T) {
	db := badTestSetup(t)
	_, err := db.GetKeyAndAccount("")
	if err == nil {
		t.Fatal("Expected error on get account and key.")
	}
}

func TestNoDatabaseMultiGet(t *testing.T) {
	db := Postgres{}
	_, err := db.GetKeyAndAccount("")
	if err == nil {
		t.Fatal("Expected error on get account and key.")
	}
}
