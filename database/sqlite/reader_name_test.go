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
	"chronokeep/remote/types"
	"testing"
	"time"
)

// Reader names are the API key names; Reader1 and reader1 are the same reader.
func TestReaderNameIgnoresCase(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("Error setting up test. %v", err)
	}
	defer finalize(t)
	setupNotificationTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	if _, err := db.AddKey(types.Key{AccountIdentifier: account1.Identifier, Value: "case-key-1", Type: "write", Name: "Reader1"}); err != nil {
		t.Fatalf("Error adding key: %v", err)
	}
	// Same name, different case, same account: rejected.
	if _, err := db.AddKey(types.Key{AccountIdentifier: account1.Identifier, Value: "case-key-2", Type: "write", Name: "reader1"}); err == nil {
		t.Error("Expected error adding a key whose name only differs by case.")
	}
	// Another account can use the name in any case.
	if _, err := db.AddKey(types.Key{AccountIdentifier: account2.Identifier, Value: "case-key-3", Type: "write", Name: "READER1"}); err != nil {
		t.Fatalf("Error adding key on another account: %v", err)
	}
	reads := []types.Read{
		{Identifier: "100", Seconds: 10, IdentType: "chip", Type: "reader"},
		{Identifier: "101", Seconds: 20, IdentType: "chip", Type: "reader"},
	}
	if _, err := db.AddReads("case-key-1", reads); err != nil {
		t.Fatalf("Error adding reads: %v", err)
	}
	if _, err := db.AddReads("case-key-3", reads[:1]); err != nil {
		t.Fatalf("Error adding reads: %v", err)
	}
	for _, name := range []string{"Reader1", "reader1", "READER1", "rEaDeR1"} {
		got, err := db.GetReads(account1.Identifier, name, 0, 100)
		if err != nil || len(got) != 2 {
			t.Errorf("GetReads(%q): expected 2 reads, found %v (%v)", name, len(got), err)
		}
	}
	if got, _ := db.GetReads(account1.Identifier, "Reader 1", 0, 100); len(got) != 0 {
		t.Errorf("Expected no reads for a different name, found %v", len(got))
	}
	when := time.Now().UTC().Format(time.RFC3339)
	if err := db.SaveNotification(&types.RequestNotification{Type: "MAX_TEMP", When: when}, "case-key-1"); err != nil {
		t.Fatalf("Error saving notification: %v", err)
	}
	if note, err := db.GetNotification(account1.Identifier, "READER1"); err != nil || note == nil {
		t.Errorf("Expected notification for READER1, found %v (%v)", note, err)
	}
	if err := db.SaveHeartbeat("case-key-1", types.Heartbeat{Device: "Portal"}); err != nil {
		t.Fatalf("Error saving heartbeat: %v", err)
	}
	if beats, err := db.GetHeartbeats(account1.Identifier, "reader1"); err != nil || len(beats) != 1 || beats[0].ReaderName != "Reader1" {
		t.Errorf("Expected heartbeat for reader1 named as stored (Reader1), found %+v (%v)", beats, err)
	}
	count, err := db.DeleteReaderReads(account1.Identifier, "READER1", 0, 15)
	if err != nil || count != 1 {
		t.Errorf("DeleteReaderReads: expected 1 deleted, found %v (%v)", count, err)
	}
	count, err = db.DeleteReaderReadsBefore(account1.Identifier, "rEaDeR1", 100)
	if err != nil || count != 1 {
		t.Errorf("DeleteReaderReadsBefore: expected 1 deleted, found %v (%v)", count, err)
	}
	// The other account's reads are untouched.
	if got, _ := db.GetReads(account2.Identifier, "reader1", 0, 100); len(got) != 1 {
		t.Errorf("Expected other account's read to remain, found %v", len(got))
	}
	count, err = db.DeleteReaderReadsBetween(account2.Identifier, "Reader1")
	if err != nil || count != 1 {
		t.Errorf("DeleteReaderReadsBetween: expected 1 deleted, found %v (%v)", count, err)
	}
}
