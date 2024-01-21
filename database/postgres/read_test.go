package postgres

import (
	"chronokeep/remote/types"
	"testing"
	"time"
)

var (
	reads []types.Read
	now   int64
)

func setupReadsTests() {
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
				Value:             "030001-1ACSDD-K2389A-00123B",
				Type:              "default",
				Name:              "reader1",
				ValidUntil:        &times[0],
			},
			{
				AccountIdentifier: accounts[0].Identifier,
				Value:             "030001-1ACSDD-K2389A-22123B",
				Type:              "write",
				Name:              "reader2",
				ValidUntil:        &times[1],
			},
			{
				AccountIdentifier: accounts[1].Identifier,
				Value:             "030001-1ACSDD-KH789A-00123B",
				Type:              "delete",
				Name:              "reader3",
				ValidUntil:        &times[2],
			},
			{
				AccountIdentifier: accounts[1].Identifier,
				Value:             "030001-1ACSCT-K2389A-22123B",
				Type:              "write",
				Name:              "reader4",
				ValidUntil:        nil,
			},
			{
				AccountIdentifier: accounts[0].Identifier,
				Value:             "030001-1ACSDD-K2389A-00123B-55223A",
				Type:              "default",
				Name:              "reader1",
				ValidUntil:        &times[0],
			},
		}
	}
	if now < 1 {
		now = 1123341123
	}
	if len(reads) < 1 {
		reads = []types.Read{
			{
				Identifier:   "165123",
				Seconds:      now,
				Milliseconds: 600,
				IdentType:    "chip",
				Type:         "reader",
				Antenna:      2,
				Reader:       "test",
				RSSI:         "-50",
			},
			{
				Identifier:   "1",
				Seconds:      now + 25,
				Milliseconds: 20,
				IdentType:    "chip",
				Type:         "reader",
				Antenna:      2,
				Reader:       "test",
				RSSI:         "-50",
			},
			{
				Identifier:   "15",
				Seconds:      now + 35,
				Milliseconds: 70,
				IdentType:    "chip",
				Type:         "reader",
			},
			{
				Identifier:   "162a",
				Seconds:      now + 55,
				Milliseconds: 123,
				IdentType:    "bib",
				Type:         "manual",
			},
			{
				Identifier:   "82",
				Seconds:      now + 365,
				Milliseconds: 42,
				IdentType:    "chip",
				Type:         "reader",
			},
			{
				Identifier:   "255",
				Seconds:      now + 400,
				Milliseconds: 273,
				IdentType:    "bib",
				Type:         "manual",
			},
			{
				Identifier:   "1365",
				Seconds:      now + 700,
				Milliseconds: 102,
				IdentType:    "chip",
				Type:         "reader",
			},
		}
	}
}

func TestAddReads(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupReadsTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	res, err := db.AddReads(keys[0].Value, reads)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != len(reads) {
		t.Errorf("Expected %v reads to be returned, %v returned.", len(reads), len(res))
	}
	db.AddReads(keys[0].Value, reads)
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != len(reads) {
		t.Errorf("Expected %v reads to be returned, %v returned.", len(reads), len(res))
	}
	for _, outer := range reads {
		found := false
		for _, inner := range res {
			if outer.Equals(&inner) {
				found = true
			}
		}
		if found == false {
			t.Fatalf("Expected to find a read added.")
		}
	}
	res, err = db.AddReads(keys[1].Value, reads[0:2])
	if err != nil {
		t.Fatalf("Error adding reads: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("Expected %v reads to be added, %v added.", 2, len(res))
	}
	res, err = db.AddReads(keys[1].Value, reads[1:3])
	if err != nil {
		t.Fatalf("Error adding reads: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("Expected %v reads to be added, %v added.", 2, len(res))
	}
	res, err = db.GetReads(keys[1].AccountIdentifier, keys[1].Name, now, now+1000)
	if err != nil {
		t.Fatalf("Erorr getting reads: %v", err)
	}
	if len(res) != 3 {
		t.Errorf("Expected %v reads to be returned, %v returned.", 3, len(res))
	}
}

func TestGetReads(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupReadsTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	res, err := db.GetReads(keys[1].AccountIdentifier, keys[1].Name, now, now+1000)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) > 0 {
		t.Fatalf("Found results when none should exist: %v", len(res))
	}
	db.AddReads(keys[0].Value, reads)
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != len(reads) {
		t.Errorf("Expected %v reads to be returned, %v returned.", len(reads), len(res))
	}
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+55)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != 4 {
		t.Errorf("Expected %v reads to be returned, %v returned.", 4, len(res))
	}
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected %v reads to be returned, %v returned.", 1, len(res))
	}
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now+35, now+400)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != 4 {
		t.Errorf("Expected %v reads to be returned, %v returned.", 4, len(res))
	}
	res, err = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now-400)
	if err != nil {
		t.Fatalf("Erorr adding reads: %v", err)
	}
	if len(res) != 4 {
		t.Errorf("Expected %v reads to be returned, %v returned.", 4, len(res))
	}
}

func TestDeleteReads(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupReadsTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	count, err := db.DeleteReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != 0 {
		t.Fatalf("count expected to be %v, deleted %v", 0, count)
	}
	db.AddReads(keys[0].Value, reads)
	count, err = db.DeleteReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != int64(len(reads)) {
		t.Fatalf("count expected to be %v, deleted %v", len(reads), count)
	}
	res, _ := db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if len(res) != 0 {
		t.Fatalf("epected to find %v reads but found %v", 0, len(res))
	}
	db.AddReads(keys[0].Value, reads)
	count, err = db.DeleteReads(keys[0].AccountIdentifier, keys[0].Name, now, now+35)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != 3 {
		t.Fatalf("count expected to be %v, deleted %v", 3, count)
	}
	res, _ = db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if len(res) != 4 {
		t.Fatalf("epected to find %v reads but found %v", 4, len(res))
	}
}

func TestDeleteKeyReads(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupReadsTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	count, err := db.DeleteKeyReads(keys[0].Value)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != 0 {
		t.Fatalf("count expected to be %v, deleted %v", 0, count)
	}
	db.AddReads(keys[0].Value, reads)
	db.AddReads(keys[1].Value, reads)
	count, err = db.DeleteKeyReads(keys[0].Value)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != int64(len(reads)) {
		t.Fatalf("count expected to be %v, deleted %v", len(reads), count)
	}
	res, _ := db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if len(res) != 0 {
		t.Fatalf("epected to find %v reads but found %v", 0, len(res))
	}
	res, _ = db.GetReads(keys[1].AccountIdentifier, keys[1].Name, now, now+1000)
	if len(res) != len(reads) {
		t.Fatalf("epected to find %v reads but found %v", 0, len(res))
	}
}

func TestDeleteReaderReads(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
	defer finalize(t)
	setupReadsTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys[0].AccountIdentifier = account1.Identifier
	keys[1].AccountIdentifier = account1.Identifier
	keys[2].AccountIdentifier = account2.Identifier
	keys[3].AccountIdentifier = account2.Identifier
	keys[4].AccountIdentifier = account1.Identifier
	db.AddKey(keys[0])
	db.AddKey(keys[1])
	count, err := db.DeleteReaderReads(keys[0].AccountIdentifier, keys[0].Name)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != 0 {
		t.Fatalf("count expected to be %v, deleted %v", 0, count)
	}
	db.AddReads(keys[0].Value, reads)
	db.AddReads(keys[1].Value, reads)
	count, err = db.DeleteReaderReads(keys[0].AccountIdentifier, keys[0].Name)
	if err != nil {
		t.Fatalf("error deleting non existant reads: %v", err)
	}
	if count != int64(len(reads)) {
		t.Fatalf("count expected to be %v, deleted %v", 0, count)
	}
	res, _ := db.GetReads(keys[0].AccountIdentifier, keys[0].Name, now, now+1000)
	if len(res) != 0 {
		t.Fatalf("epected to find %v reads but found %v", 0, len(res))
	}
	res, _ = db.GetReads(keys[1].AccountIdentifier, keys[1].Name, now, now+1000)
	if len(res) != len(reads) {
		t.Fatalf("epected to find %v reads but found %v", 0, len(res))
	}
}

func TestBadDatabaseRead(t *testing.T) {
	db := badTestSetup(t)
	_, err := db.GetReads(0, "", 0, 0)
	if err == nil {
		t.Fatal("Expected error on get reads.")
	}
	_, err = db.AddReads("", make([]types.Read, 0))
	if err == nil {
		t.Fatal("Expected error on add reads.")
	}
	_, err = db.DeleteReads(0, "", 0, 0)
	if err == nil {
		t.Fatal("Expected error on delete reads.")
	}
	_, err = db.DeleteKeyReads("")
	if err == nil {
		t.Fatal("Expected error on delete key reads.")
	}
	_, err = db.DeleteReaderReads(0, "")
	if err == nil {
		t.Fatal("Expected error on delete reader reads.")
	}
}

func TestNoDatabaseRead(t *testing.T) {
	db := Postgres{}
	_, err := db.GetReads(0, "", 0, 0)
	if err == nil {
		t.Fatal("Expected error on get reads.")
	}
	_, err = db.AddReads("", make([]types.Read, 0))
	if err == nil {
		t.Fatal("Expected error on add reads.")
	}
	_, err = db.DeleteReads(0, "", 0, 0)
	if err == nil {
		t.Fatal("Expected error on delete reads.")
	}
	_, err = db.DeleteKeyReads("")
	if err == nil {
		t.Fatal("Expected error on delete key reads.")
	}
	_, err = db.DeleteReaderReads(0, "")
	if err == nil {
		t.Fatal("Expected error on delete reader reads.")
	}
}
