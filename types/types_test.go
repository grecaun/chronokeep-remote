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

package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAccountValidate(t *testing.T) {
	v := validator.New()
	valid := Account{Name: "Jane", Email: "jane@test.com", Type: "free"}
	for _, typ := range []string{"admin", "free", "paid"} {
		a := valid
		a.Type = typ
		assert.NoError(t, a.Validate(v), typ)
	}
	for _, typ := range []string{"", "Admin", "FREE", "owner"} {
		a := valid
		a.Type = typ
		assert.Error(t, a.Validate(v), "type %q", typ)
	}
	a := valid
	a.Name = ""
	assert.Error(t, a.Validate(v), "empty name")
	for _, email := range []string{"", "jane", "jane@", "@test.com", "jane test@test.com"} {
		a := valid
		a.Email = email
		assert.Error(t, a.Validate(v), "email %q", email)
	}
}

func TestAccountEquals(t *testing.T) {
	a := Account{Identifier: 1, Name: "Jane", Email: "jane@test.com", Type: "free", Locked: true, Password: "x"}
	assert.False(t, a.Equals(nil))
	// Identifier, password and locked state aren't compared.
	assert.True(t, a.Equals(&Account{Identifier: 2, Name: "Jane", Email: "jane@test.com", Type: "free"}))
	assert.False(t, a.Equals(&Account{Name: "Janet", Email: "jane@test.com", Type: "free"}))
	assert.False(t, a.Equals(&Account{Name: "Jane", Email: "janet@test.com", Type: "free"}))
	assert.False(t, a.Equals(&Account{Name: "Jane", Email: "jane@test.com", Type: "paid"}))
}

func TestPasswordIsHashed(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	assert.True(t, (&Account{Password: string(hash)}).PasswordIsHashed())
	for _, p := range []string{"", "password", "$2a$", "$2a$10$short"} {
		assert.False(t, (&Account{Password: p}).PasswordIsHashed(), "password %q", p)
	}
	// A hash of the empty string doesn't count as hashed (empty should never be a real password).
	empty, _ := bcrypt.GenerateFromPassword([]byte(""), bcrypt.MinCost)
	assert.False(t, (&Account{Password: string(empty)}).PasswordIsHashed())
}

func TestKeyValidate(t *testing.T) {
	v := validator.New()
	for _, typ := range []string{"read", "write", "delete"} {
		k := Key{Name: "reader", Type: typ}
		assert.NoError(t, k.Validate(v), typ)
		assert.Equal(t, typ, k.Type)
	}
	// default becomes read.
	k := Key{Name: "reader", Type: "default"}
	assert.NoError(t, k.Validate(v))
	assert.Equal(t, "read", k.Type)
	for _, typ := range []string{"", "READ", "admin"} {
		k := Key{Name: "reader", Type: typ}
		assert.Error(t, k.Validate(v), "type %q", typ)
	}
	assert.Error(t, (&Key{Type: "read"}).Validate(v), "empty name")
}

func TestRequestKeyValidate(t *testing.T) {
	v := validator.New()
	for _, typ := range []string{"read", "write", "delete"} {
		k := RequestKey{Name: "reader", Type: typ}
		assert.NoError(t, k.Validate(v), typ)
		assert.Equal(t, typ, k.Type)
	}
	// default becomes read.
	k := RequestKey{Name: "reader", Type: "default"}
	assert.NoError(t, k.Validate(v))
	assert.Equal(t, "read", k.Type)
	for _, typ := range []string{"", "Write", "admin"} {
		assert.Error(t, (&RequestKey{Name: "reader", Type: typ}).Validate(v), "type %q", typ)
	}
	assert.Error(t, (&RequestKey{Type: "read"}).Validate(v), "empty name")
}

func TestKeyExpired(t *testing.T) {
	past := time.Now().Add(-time.Second)
	future := time.Now().Add(time.Hour)
	assert.False(t, Key{}.Expired(), "no expiry")
	assert.True(t, Key{ValidUntil: &past}.Expired())
	assert.False(t, Key{ValidUntil: &future}.Expired())
}

func TestKeyEqual(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sameInstant := when.In(time.FixedZone("PST", -8*60*60))
	other := when.Add(time.Second)
	base := Key{AccountIdentifier: 1, Name: "reader", Value: "abc", Type: "write"}
	withTime := func(k Key, t *time.Time) *Key { k.ValidUntil = t; return &k }
	assert.True(t, withTime(base, nil).Equal(withTime(base, nil)))
	assert.True(t, withTime(base, &when).Equal(withTime(base, &sameInstant)), "same instant, different zone")
	assert.False(t, withTime(base, &when).Equal(withTime(base, &other)))
	assert.False(t, withTime(base, &when).Equal(withTime(base, nil)))
	assert.False(t, withTime(base, nil).Equal(withTime(base, &when)))
	for _, k := range []Key{
		{AccountIdentifier: 2, Name: "reader", Value: "abc", Type: "write"},
		{AccountIdentifier: 1, Name: "reader2", Value: "abc", Type: "write"},
		{AccountIdentifier: 1, Name: "reader", Value: "abcd", Type: "write"},
		{AccountIdentifier: 1, Name: "reader", Value: "abc", Type: "read"},
	} {
		assert.False(t, base.Equal(&k), "%+v", k)
	}
}

func TestRequestKeyValidUntil(t *testing.T) {
	endOfDay := time.Date(2026, 3, 4, 23, 59, 59, 0, time.Local)
	for _, s := range []string{"2026/03/04", "2026/3/4", "03/04/2026", "3/4/2026", "2026-03-04", "2026-3-4", "03-04-2026", "3-4-2026"} {
		got := RequestKey{ValidUntil: s}.GetValidUntil()
		if assert.NotNil(t, got, s) {
			assert.True(t, endOfDay.Equal(*got), "%v: got %v", s, got)
		}
		k := RequestKey{ValidUntil: s}.ToKey()
		if assert.NotNil(t, k.ValidUntil, s) {
			assert.True(t, endOfDay.Equal(*k.ValidUntil), "%v: got %v", s, k.ValidUntil)
		}
	}
	// RFC3339 is used as given, with no end-of-day adjustment.
	exact := time.Date(2026, 3, 4, 10, 30, 0, 0, time.UTC)
	got := RequestKey{ValidUntil: exact.Format(time.RFC3339)}.GetValidUntil()
	if assert.NotNil(t, got) {
		assert.True(t, exact.Equal(*got))
	}
	// Empty, nonsense and impossible dates mean no expiry.
	for _, s := range []string{"", "tomorrow", "2026-02-30", "2026-13-01", "2026.03.04"} {
		assert.Nil(t, RequestKey{ValidUntil: s}.GetValidUntil(), s)
		assert.Nil(t, RequestKey{ValidUntil: s}.ToKey().ValidUntil, s)
	}
}

func TestRequestKeyToKey(t *testing.T) {
	k := RequestKey{Name: "  reader 1 \t", Value: "abc", Type: "delete"}.ToKey()
	assert.Equal(t, "reader 1", k.Name, "name is trimmed")
	assert.Equal(t, "abc", k.Value)
	assert.Equal(t, "delete", k.Type)
	assert.Equal(t, int64(0), k.AccountIdentifier)
	assert.Equal(t, "read", RequestKey{Name: "reader", Type: "default"}.ToKey().Type, "default becomes read")
}

func TestReadValidate(t *testing.T) {
	v := validator.New()
	valid := Read{Identifier: "1001", Seconds: 10, Milliseconds: 5, IdentType: "chip", Type: "reader"}
	// Types are case-insensitive and normalized to lower case.
	r := valid
	r.IdentType, r.Type = "BIB", "Manual"
	if assert.NoError(t, r.Validate(v)) {
		assert.Equal(t, "bib", r.IdentType)
		assert.Equal(t, "manual", r.Type)
	}
	r = valid
	r.Seconds, r.Milliseconds = 0, 0
	assert.NoError(t, r.Validate(v), "zero time")
	for name, change := range map[string]func(*Read){
		"empty identifier": func(r *Read) { r.Identifier = "" },
		"negative seconds": func(r *Read) { r.Seconds = -1 },
		"negative ms":      func(r *Read) { r.Milliseconds = -1 },
		"bad ident type":   func(r *Read) { r.IdentType = "tag" },
		"empty ident type": func(r *Read) { r.IdentType = "" },
		"bad read type":    func(r *Read) { r.Type = "auto" },
		"empty read type":  func(r *Read) { r.Type = "" },
	} {
		r := valid
		change(&r)
		assert.Error(t, r.Validate(v), name)
	}
}

func TestReadEquals(t *testing.T) {
	a := Read{Key: "k1", Identifier: "1001", Seconds: 10, Milliseconds: 5, IdentType: "chip", Type: "reader", Antenna: 1, Reader: "1", RSSI: "-50"}
	b := a
	b.Key = "k2"
	assert.True(t, a.Equals(&b), "key isn't compared")
	for name, change := range map[string]func(*Read){
		"identifier":   func(r *Read) { r.Identifier = "1002" },
		"seconds":      func(r *Read) { r.Seconds++ },
		"milliseconds": func(r *Read) { r.Milliseconds++ },
		"ident type":   func(r *Read) { r.IdentType = "bib" },
		"type":         func(r *Read) { r.Type = "manual" },
		"antenna":      func(r *Read) { r.Antenna++ },
		"reader":       func(r *Read) { r.Reader = "2" },
		"rssi":         func(r *Read) { r.RSSI = "-51" },
	} {
		c := a
		change(&c)
		assert.False(t, a.Equals(&c), name)
	}
}

func TestRequestNotificationValidate(t *testing.T) {
	v := validator.New()
	when := time.Now().UTC().Format(time.RFC3339)
	for _, typ := range []string{"SHUTTING_DOWN", "RESTARTING", "HIGH_TEMP", "MAX_TEMP", "BATTERY_LOW", "BATTERY_CRITICAL"} {
		assert.NoError(t, (&RequestNotification{Type: typ, When: when}).Validate(v), typ)
	}
	// UPS notifications are no longer used
	for _, typ := range []string{"", "battery_low", "UNKNOWN",
		"UPS_DISCONNECTED", "UPS_CONNECTED", "UPS_ON_BATTERY", "UPS_LOW_BATTERY", "UPS_ONLINE"} {
		assert.Error(t, (&RequestNotification{Type: typ, When: when}).Validate(v), "type %q", typ)
	}
	for _, w := range []string{"", "2026-01-02", "2026-01-02 03:04:05", "1767323045"} {
		assert.Error(t, (&RequestNotification{Type: "MAX_TEMP", When: w}).Validate(v), "when %q", w)
	}
}

func TestRequestNotificationToNotification(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	n, err := RequestNotification{Type: "MAX_TEMP", When: when.Format(time.RFC3339)}.ToNotification()
	if assert.NoError(t, err) {
		assert.Equal(t, "MAX_TEMP", n.Type)
		assert.True(t, when.Equal(n.When))
	}
	n, err = RequestNotification{Type: "MAX_TEMP", When: "not a date"}.ToNotification()
	assert.Error(t, err)
	assert.Nil(t, n)
}

func TestHeartbeatValidate(t *testing.T) {
	v := validator.New()
	assert.NoError(t, (&Heartbeat{}).Validate(v), "zero value")
	assert.NoError(t, (&Heartbeat{Now: -1}).Validate(v), "now isn't range checked")
	strFields := map[string]struct {
		max int
		set func(*Heartbeat, string)
	}{
		"dev":       {100, func(h *Heartbeat, s string) { h.Device = s }},
		"location":  {100, func(h *Heartbeat, s string) { h.Location = s }},
		"battery":   {50, func(h *Heartbeat, s string) { h.Battery = s }},
		"hash":      {100, func(h *Heartbeat, s string) { h.Hash = s }},
		"chip_type": {20, func(h *Heartbeat, s string) { h.ChipType = s }},
	}
	for name, f := range strFields {
		h := Heartbeat{}
		f.set(&h, strings.Repeat("a", f.max))
		assert.NoError(t, h.Validate(v), "%v at max length", name)
		f.set(&h, strings.Repeat("a", f.max+1))
		assert.Error(t, h.Validate(v), "%v over max length", name)
	}
	// Length is counted in characters, not bytes.
	assert.NoError(t, (&Heartbeat{Device: strings.Repeat("é", 100)}).Validate(v))
	for name, set := range map[string]func(*Heartbeat){
		"uptime":            func(h *Heartbeat) { h.Uptime = -1 },
		"upload_errors":     func(h *Heartbeat) { h.UploadErrors = -1 },
		"readers_connected": func(h *Heartbeat) { h.ReadersConnected = -1 },
		"readers_total":     func(h *Heartbeat) { h.ReadersTotal = -1 },
	} {
		h := Heartbeat{}
		set(&h)
		assert.Error(t, h.Validate(v), "negative %v", name)
	}
	// Temperature is optional, and checked against a sensible range in °C when sent.
	for _, temp := range []float64{-100, 0, 54.5, 200} {
		assert.NoError(t, (&Heartbeat{Temperature: &temp}).Validate(v), "temperature %v", temp)
	}
	for _, temp := range []float64{-100.1, 200.1} {
		assert.Error(t, (&Heartbeat{Temperature: &temp}).Validate(v), "temperature %v", temp)
	}
}

func TestHeartbeatJSON(t *testing.T) {
	// Relay sends no battery fields; unknown fields from newer clients are ignored.
	var hb Heartbeat
	err := json.Unmarshal([]byte(`{"dev":"Relay","now":5,"uptime":60,"reading":true,"chip_type":"HEX","future_field":1}`), &hb)
	if assert.NoError(t, err) {
		assert.Equal(t, Heartbeat{Device: "Relay", Now: 5, Uptime: 60, Reading: true, ChipType: "HEX"}, hb)
	}
	// The portal sends its temperature; without one, temperature is left out of the response.
	hb = Heartbeat{}
	if assert.NoError(t, json.Unmarshal([]byte(`{"dev":"Portal","temperature":54.5}`), &hb)) && assert.NotNil(t, hb.Temperature) {
		assert.Equal(t, 54.5, *hb.Temperature)
	}
	if out, err := json.Marshal(Heartbeat{Device: "Relay"}); assert.NoError(t, err) {
		assert.NotContains(t, string(out), "temperature")
	}
	// Received is written as an RFC 3339 date-time string.
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	out, err := json.Marshal(ReaderHeartbeat{ReaderName: "reader1", Received: when})
	if assert.NoError(t, err) {
		assert.Contains(t, string(out), `"received":"2026-01-02T03:04:05Z"`)
		assert.Contains(t, string(out), `"reader":"reader1"`)
	}
}
