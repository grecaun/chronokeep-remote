package mysql

import (
	"chronokeep/remote/types"
	"testing"
	"time"
)

func setupNotificationTests() {
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

func TestSaveNotification(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("Error setting up test. %v", err)
	}
	defer finalize(t)
	setupNotificationTests()
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
	when := time.Now()
	notifications := []types.RequestNotification{
		{
			Type: "invalid_type",
			When: when.UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_DISCONNECTED",
			When: "invalid date",
		},
		{
			Type: "UPS_DISCONNECTED",
			When: when.Add(time.Second * -9).UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_CONNECTED",
			When: when.Add(time.Second * -8).UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_ON_BATTERY",
			When: when.Add(time.Second * -7).UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_LOW_BATTERY",
			When: when.Add(time.Second * -6).UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_ONLINE",
			When: when.Add(time.Second * -5).UTC().Format(time.RFC3339),
		},
		{
			Type: "SHUTTING_DOWN",
			When: when.Add(time.Second * -4).UTC().Format(time.RFC3339),
		},
		{
			Type: "RESTARTING",
			When: when.Add(time.Second * -3).UTC().Format(time.RFC3339),
		},
		{
			Type: "HIGH_TEMP",
			When: when.Add(time.Second * -2).UTC().Format(time.RFC3339),
		},
		{
			Type: "MAX_TEMP",
			When: when.Add(time.Second * -1).UTC().Format(time.RFC3339),
		},
		{
			Type: "MAX_TEMP",
			When: when.Add(time.Second * -9).UTC().Format(time.RFC3339),
		},
	}
	err = db.SaveNotification(&notifications[0], keys[0].Value)
	if err == nil {
		t.Fatalf("expected error saving notification with invalid type but no error found")
	}
	err = db.SaveNotification(&notifications[1], keys[0].Value)
	if err == nil {
		t.Fatalf("expected error saving notification with invalid date but no error found")
	}
	err = db.SaveNotification(&notifications[2], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[3], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[4], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[5], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[6], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[7], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[8], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[9], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[10], keys[0].Value)
	if err != nil {
		t.Fatalf("error found saving notification: %v", err)
	}
	err = db.SaveNotification(&notifications[11], keys[0].Value)
	if err == nil {
		t.Fatalf("expected error when adding notification with duplicate when value but no error was found")
	}
	err = db.SaveNotification(&notifications[2], "invalid key")
	if err == nil {
		t.Fatalf("expected error when adding notification with invalid key but no error was found")
	}
}

func TestGetNotification(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("Error setting up test. %v", err)
	}
	defer finalize(t)
	setupNotificationTests()
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
	when := time.Now()
	notifications := []types.RequestNotification{
		{
			Type: "UPS_DISCONNECTED",
			When: when.UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_DISCONNECTED",
			When: when.Add(time.Minute * -6).UTC().Format(time.RFC3339),
		},
		{
			Type: "UPS_ON_BATTERY",
			When: when.Add(time.Second * -10).UTC().Format(time.RFC3339),
		},
	}
	// No notifications saved.
	note, err := db.GetNotification(account1.Identifier, keys[0].Name)
	if err != nil {
		t.Fatalf("error when trying to get notification: %v", err)
	}
	if note != nil {
		t.Fatalf("found notification when none was expected: %v", note)
	}
	_ = db.SaveNotification(&notifications[0], keys[0].Value)
	_ = db.SaveNotification(&notifications[1], keys[1].Value)
	_ = db.SaveNotification(&notifications[2], keys[0].Value)
	// Saved notification, within time period
	note, err = db.GetNotification(account1.Identifier, keys[0].Name)
	if err != nil {
		t.Fatalf("error when trying to get notification: %v", err)
	}
	if note == nil {
		t.Fatalf("expected a notification but didn't find anything")
	}
	if note.Type != notifications[0].Type {
		t.Fatalf("expected to find %v for the notification type, found %v", notifications[0].Type, note.Type)
	}
	// Notification too long ago
	note, err = db.GetNotification(account2.Identifier, keys[1].Name)
	if err != nil {
		t.Fatalf("error when trying to get notification: %v", err)
	}
	if note != nil {
		t.Fatalf("found notification when none was expected: %v", note)
	}
	// Invalid key
	note, err = db.GetNotification(account1.Identifier, "invalid key")
	if err != nil {
		t.Fatalf("error when trying to get notification: %v", err)
	}
	if note != nil {
		t.Fatalf("found notification when none was expected: %v", note)
	}
}
