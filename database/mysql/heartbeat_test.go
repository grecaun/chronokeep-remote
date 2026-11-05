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

package mysql

import (
	"chronokeep/remote/database"
	"chronokeep/remote/types"
	"testing"
	"time"
)

func TestHeartbeats(t *testing.T) {
	db, finalize, err := setupTests(t)
	if err != nil {
		t.Fatalf("Error setting up test. %v", err)
	}
	defer finalize(t)
	setupNotificationTests()
	account1, _ := db.AddAccount(accounts[0])
	account2, _ := db.AddAccount(accounts[1])
	keys := []types.Key{
		{AccountIdentifier: account1.Identifier, Value: "hb-key-1", Type: "write", Name: "reader1"},
		{AccountIdentifier: account1.Identifier, Value: "hb-key-2", Type: "delete", Name: "reader2"},
		{AccountIdentifier: account2.Identifier, Value: "hb-key-3", Type: "write", Name: "reader3"},
	}
	for _, k := range keys {
		if _, err := db.AddKey(k); err != nil {
			t.Fatalf("Error adding key: %v", err)
		}
	}
	first := types.Heartbeat{Device: "Portal", Location: "Portal", Uptime: 60, Hash: "Chronokeep Portal v1.0.0"}
	latest := types.Heartbeat{Device: "Portal", Location: "Portal", Uptime: 120, Hash: "Chronokeep Portal v1.0.0", Reading: true}
	other := types.Heartbeat{Device: "Relay", Location: "Relay", Uptime: 30, Hash: "Chronokeep Relay v0.1.0"}
	before := time.Now().UnixMilli()
	for _, s := range []struct {
		key string
		hb  types.Heartbeat
	}{{"hb-key-1", first}, {"hb-key-1", latest}, {"hb-key-2", other}, {"hb-key-3", other}} {
		if err := db.SaveHeartbeat(s.key, s.hb); err != nil {
			t.Fatalf("Error saving heartbeat: %v", err)
		}
	}
	after := time.Now().UnixMilli()
	// Only the most recent heartbeat per reader, only for this account.
	beats, err := db.GetHeartbeats(account1.Identifier, "")
	if err != nil {
		t.Fatalf("Error getting heartbeats: %v", err)
	}
	if len(beats) != 2 {
		t.Fatalf("Expected 2 heartbeats, found %v", len(beats))
	}
	if beats[0].ReaderName != "reader1" || beats[0].Heartbeat != latest {
		t.Errorf("Expected latest heartbeat for reader1, found %+v", beats[0])
	}
	// Received is the server's time in Unix milliseconds.
	for _, b := range beats {
		if b.Received.UnixMilli() < before || b.Received.UnixMilli() > after {
			t.Errorf("Expected received between %v and %v (Unix ms), found %v", before, after, b.Received.UnixMilli())
		}
	}
	if beats[1].ReaderName != "reader2" || beats[1].Heartbeat != other {
		t.Errorf("Expected heartbeat for reader2, found %+v", beats[1])
	}
	// Filtering by reader name.
	beats, err = db.GetHeartbeats(account1.Identifier, "reader2")
	if err != nil {
		t.Fatalf("Error getting heartbeats: %v", err)
	}
	if len(beats) != 1 || beats[0].ReaderName != "reader2" || beats[0].Heartbeat != other {
		t.Errorf("Expected only reader2's heartbeat, found %+v", beats)
	}
	for _, name := range []string{"reader3", "unknown"} {
		beats, err = db.GetHeartbeats(account1.Identifier, name)
		if err != nil {
			t.Fatalf("Error getting heartbeats: %v", err)
		}
		if len(beats) != 0 {
			t.Errorf("Expected no heartbeats for %v, found %+v", name, beats)
		}
	}
	// A reader whose latest heartbeat is older than HeartbeatMaxAge is left out.
	old := time.Now().Add(-database.HeartbeatMaxAge - time.Minute).UnixMilli()
	if _, err := db.db.Exec("UPDATE heartbeat SET heartbeat_received=? WHERE key_value=?;", old, "hb-key-2"); err != nil {
		t.Fatalf("Error aging heartbeat: %v", err)
	}
	beats, err = db.GetHeartbeats(account1.Identifier, "")
	if err != nil {
		t.Fatalf("Error getting heartbeats: %v", err)
	}
	if len(beats) != 1 || beats[0].ReaderName != "reader1" {
		t.Fatalf("Expected only reader1 after aging reader2, found %+v", beats)
	}
	// Asking for that reader by name still returns its old heartbeat.
	beats, err = db.GetHeartbeats(account1.Identifier, "reader2")
	if err != nil {
		t.Fatalf("Error getting heartbeats: %v", err)
	}
	if len(beats) != 1 || beats[0].ReaderName != "reader2" || beats[0].Received.UnixMilli() != old {
		t.Fatalf("Expected reader2's aged heartbeat when asked by name, found %+v", beats)
	}
	// A deleted key's heartbeat is left out.
	if err := db.DeleteKey(keys[0]); err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}
	beats, err = db.GetHeartbeats(account1.Identifier, "")
	if err != nil {
		t.Fatalf("Error getting heartbeats: %v", err)
	}
	if len(beats) != 0 {
		t.Fatalf("Expected no heartbeats, found %+v", beats)
	}
}
