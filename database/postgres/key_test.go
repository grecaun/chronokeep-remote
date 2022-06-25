package postgres

import (
	"chronokeep/remote/types"
	"testing"
	"time"
)

var (
	keys  []types.Key
	times []time.Time
)

func setupKeyTests() {
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
		}
	}
	if len(times) < 1 {
		times = []time.Time{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
			time.Now().Add(time.Hour * 20).Truncate(time.Second),
			time.Date(2016, 4, 1, 4, 11, 5, 0, time.Local),
		}
	}
	if len(keys) < 1 {
		keys = []types.Key{
			{
				AccountIdentifier: accounts[0].Identifier,
				Name:              "test1",
				Value:             "030001-1ACSDD-K2389A-00123B",
				Type:              "default",
				ReaderName:        "reader1",
				ValidUntil:        &times[0],
			},
			{
				AccountIdentifier: accounts[0].Identifier,
				Value:             "030001-1ACSDD-K2389A-22123B",
				Type:              "write",
				ReaderName:        "reader2",
				ValidUntil:        &times[1],
			},
			{
				AccountIdentifier: accounts[1].Identifier,
				Name:              "test2",
				Value:             "030001-1ACSDD-KH789A-00123B",
				Type:              "delete",
				ReaderName:        "reader3",
				ValidUntil:        &times[2],
			},
			{
				AccountIdentifier: accounts[1].Identifier,
				Value:             "030001-1ACSCT-K2389A-22123B",
				Type:              "write",
				ReaderName:        "reader4",
				ValidUntil:        nil,
			},
			{
				AccountIdentifier: accounts[1].Identifier,
				Name:              "test1",
				Value:             "030001-1ACSDD-K2389A-00123B-55223A",
				Type:              "default",
				ReaderName:        "reader1",
				ValidUntil:        &times[0],
			},
		}
	}
}

func TestAddKey(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupKeyTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	key, err := db.AddKey(keys[0])
	if err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	if !key.Equal(&keys[0]) {
		t.Errorf("Expected key %+v, found %+v", keys[0], *key)
	}
	if key.Name != "test1" {
		t.Errorf("Expected key to be named %s, found %s.", "test1", key.Name)
	}
	key, err = db.AddKey(keys[1])
	if err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	if !key.Equal(&keys[1]) {
		t.Errorf("Expected key %+v, found %+v", keys[1], *key)
	}
	key, err = db.AddKey(keys[2])
	if err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	if !key.Equal(&keys[2]) {
		t.Errorf("Expected key %+v, found %+v", keys[2], *key)
	}
	if key.Name != "test2" {
		t.Errorf("Expected key to be named %s, found %s.", "test2", key.Name)
	}
	key, err = db.AddKey(keys[3])
	if err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	if !key.Equal(&keys[3]) {
		t.Errorf("Expected key %+v, found %+v", keys[3], *key)
	}
	key, err = db.AddKey(keys[3])
	if err == nil {
		t.Errorf("Expected error adding key that exists, found key %+v", key)
	}
	key, err = db.AddKey(keys[4])
	if err == nil {
		t.Errorf("Expected error adding key with duplicate account and reader name, found key %+v", key)
	}
}

func TestGetAccountKeys(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupKeyTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	k, err := db.GetAccountKeys(keys[0].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 0 {
		t.Errorf("Expected no keys found for account but found %v keys.", len(k))
	}
	db.AddKey(keys[0])
	db.AddKey(keys[2])
	k, err = db.GetAccountKeys(keys[0].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 1 {
		t.Errorf("Expected %v keys found for account but found %v keys.", 1, len(k))
	}
	k, err = db.GetAccountKeys(keys[1].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 0 {
		t.Errorf("Expected no keys found for account but found %v keys.", len(k))
	}
	k, err = db.GetAccountKeys(keys[2].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 1 {
		t.Errorf("Expected %v keys found for account but found %v keys.", 1, len(k))
	}
	db.AddKey(keys[1])
	db.AddKey(keys[3])
	k, err = db.GetAccountKeys(keys[1].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 2 {
		t.Errorf("Expected %v keys found for account but found %v keys.", 2, len(k))
	}
	k, err = db.GetAccountKeys(keys[3].Value)
	if err != nil {
		t.Fatalf("Error getting account keys: %v", err)
	}
	if len(k) != 2 {
		t.Errorf("Expected %v keys found for account but found %v keys.", 2, len(k))
	}
}

func TestGetKey(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupKeyTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	db.AddKey(keys[2])
	db.AddKey(keys[3])
	key, err := db.GetKey(keys[0].Value)
	if err != nil {
		t.Fatalf("Error getting key: %v", err)
	}
	if !key.Equal(&keys[0]) {
		t.Errorf("Expected key %+v, found %+v.", keys[0], *key)
	}
	if key.Name != "test1" {
		t.Errorf("Expected key to be named %s, found %s.", "test1", key.Name)
	}
	key, err = db.GetKey(keys[1].Value)
	if err != nil {
		t.Fatalf("Error getting key: %v", err)
	}
	if !key.Equal(&keys[1]) {
		t.Errorf("Expected key %+v, found %+v.", keys[1], *key)
	}
	key, err = db.GetKey(keys[2].Value)
	if err != nil {
		t.Fatalf("Error getting key: %v", err)
	}
	if !key.Equal(&keys[2]) {
		t.Errorf("Expected key %+v, found %+v.", keys[2], *key)
	}
	if key.Name != "test2" {
		t.Errorf("Expected key to be named %s, found %s.", "test2", key.Name)
	}
	key, err = db.GetKey(keys[3].Value)
	if err != nil {
		t.Fatalf("Error getting key: %v", err)
	}
	if !key.Equal(&keys[3]) {
		t.Errorf("Expected key %+v, found %+v.", keys[3], *key)
	}
	key, err = db.GetKey("test-value")
	if err != nil {
		t.Fatalf("Error getting non-existant key: %v", err)
	}
	if key != nil {
		t.Errorf("Expected no key but found %+v.", *key)
	}
}

func TestDeleteKey(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupKeyTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	db.AddKey(keys[2])
	db.AddKey(keys[3])
	err = db.DeleteKey(keys[0])
	if err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}
	k, _ := db.GetKey(keys[0].Value)
	if k != nil {
		t.Errorf("Found deleted key: %+v", k)
	}
	err = db.DeleteKey(keys[1])
	if err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}
	k, _ = db.GetKey(keys[1].Value)
	if k != nil {
		t.Errorf("Found deleted key: %+v", k)
	}
	err = db.DeleteKey(keys[2])
	if err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}
	k, _ = db.GetKey(keys[2].Value)
	if k != nil {
		t.Errorf("Found deleted key: %+v", k)
	}
	err = db.DeleteKey(keys[3])
	if err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}
	k, _ = db.GetKey(keys[3].Value)
	if k != nil {
		t.Errorf("Found deleted key: %+v", k)
	}
	err = db.DeleteKey(keys[3])
	if err == nil {
		t.Error("Expected error from deletion of already deleted key.")
	}
}

func TestUpdateKey(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupKeyTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	keys[0].Type = "write"
	keys[1].Name = "newtest1"
	keys[0].ReaderName = "reader8"
	validTime := time.Now().Add(time.Minute * 30).Truncate(time.Second)
	keys[0].ValidUntil = &validTime
	err = db.UpdateKey(keys[0])
	if err != nil {
		t.Fatalf("Error updating key: %v", err)
	}
	key, _ := db.GetKey(keys[0].Value)
	if !key.Equal(&keys[0]) {
		t.Errorf("Expected key %+v, found %+v.", keys[0], *key)
	}
	if key.Name != keys[0].Name {
		t.Errorf("Expected key name to be %s, found %s.", keys[0].Name, key.Name)
	}
	if key.ReaderName != keys[0].ReaderName {
		t.Errorf("Expected key reader name to be %s, found %s.", keys[0].ReaderName, key.ReaderName)
	}
	keys[1].AccountIdentifier = accounts[0].Identifier + 200
	keys[1].Value = "update-value-test"
	err = db.UpdateKey(keys[1])
	if err == nil {
		t.Error("Expected error from update with no changed values.")
	}
	key, _ = db.GetKey(keys[1].Value)
	if key != nil {
		t.Errorf("Found key with modified key value: %+v", key)
	}
}
