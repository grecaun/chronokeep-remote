/*
Chronokeep Remote - Remote Tag Read API
Copyright (C) 2026 James Sentinella

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package sqlite

import (
	"chronokeep/remote/database"
	"chronokeep/remote/types"
	"testing"
)

func TestSetupNilConfig(t *testing.T) {
	o := SQLite{}
	if err := o.Setup(nil); err == nil {
		t.Fatal("Expected error setting up with no config.")
	}
}

func TestSetupCreatesAdminOnce(t *testing.T) {
	o := &SQLite{}
	config := getTestConfig()
	if err := o.Setup(config); err != nil {
		t.Fatalf("Error setting up database: %v", err)
	}
	defer o.dropTables()
	accounts, err := o.GetAccounts()
	if err != nil {
		t.Fatalf("Error getting accounts: %v", err)
	}
	if len(accounts) != 1 || accounts[0].Email != config.AdminEmail || accounts[0].Type != "admin" {
		t.Fatalf("Expected one admin account %v, found %+v", config.AdminEmail, accounts)
	}
	admin, err := o.GetAccount(config.AdminEmail)
	if err != nil || admin == nil {
		t.Fatalf("Error getting admin account: %v", err)
	}
	if !admin.PasswordIsHashed() || admin.Password == config.AdminPass {
		t.Error("Expected the admin password to be stored hashed.")
	}
	// Running setup again on a current database doesn't add another admin.
	if err := o.Setup(config); err != nil {
		t.Fatalf("Error running setup a second time: %v", err)
	}
	accounts, err = o.GetAccounts()
	if err != nil {
		t.Fatalf("Error getting accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected one account after second setup, found %v", len(accounts))
	}
}

func TestSetupWithoutAdminCredentials(t *testing.T) {
	for _, field := range []string{"name", "email", "pass"} {
		o := &SQLite{}
		config := getTestConfig()
		switch field {
		case "name":
			config.AdminName = ""
		case "email":
			config.AdminEmail = ""
		case "pass":
			config.AdminPass = ""
		}
		err := o.Setup(config)
		if err == nil {
			t.Errorf("Expected error setting up a new database without admin %v.", field)
		}
		o.dropTables()
		o.Close()
	}
}

func TestSetupUpgradesOldDatabase(t *testing.T) {
	old, err := setupOld()
	if err != nil {
		t.Fatalf("Error setting up version 1 database: %v", err)
	}
	old.Close()
	o := &SQLite{}
	if err := o.Setup(getTestConfig()); err != nil {
		t.Fatalf("Error upgrading database through Setup: %v", err)
	}
	defer o.dropTables()
	if v := o.checkVersion(); v != database.CurrentVersion {
		t.Fatalf("Expected version %v after Setup, found %v", database.CurrentVersion, v)
	}
	// The tables added by upgrades are usable.
	acc, err := o.AddAccount(types.Account{Name: "Up Grade", Email: "upgrade@test.com", Type: "free", Password: testHashPassword("password")})
	if err != nil || acc == nil {
		t.Fatalf("Error adding account: %v", err)
	}
	if _, err := o.AddKey(types.Key{AccountIdentifier: acc.Identifier, Value: "upgrade-key", Type: "write", Name: "reader"}); err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	if err := o.SaveHeartbeat("upgrade-key", types.Heartbeat{Device: "Portal"}); err != nil {
		t.Fatalf("Error saving heartbeat after upgrade: %v", err)
	}
	beats, err := o.GetHeartbeats(acc.Identifier, "")
	if err != nil || len(beats) != 1 {
		t.Fatalf("Expected one heartbeat after upgrade, found %+v (%v)", beats, err)
	}
}

func TestDeleteReaderReadsBackwardsRange(t *testing.T) {
	o := SQLite{}
	count, err := o.DeleteReaderReads(1, "reader", 10, 5)
	if err == nil {
		t.Fatal("Expected error deleting reads with an end before the start.")
	}
	if count != 0 {
		t.Fatalf("Expected no reads deleted, found %v", count)
	}
}
