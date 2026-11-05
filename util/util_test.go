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

package util

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestTimeSinceEpochSeconds(t *testing.T) {
	epoch := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	s, err := TimeSinceEpochSeconds(epoch)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(0), s, "epoch itself")
	}
	s, err = TimeSinceEpochSeconds(epoch.Add(90*time.Second + 999*time.Millisecond))
	if assert.NoError(t, err) {
		assert.Equal(t, int64(90), s, "partial seconds are dropped")
	}
	// The same instant in another time zone gives the same answer.
	s, err = TimeSinceEpochSeconds(epoch.Add(time.Hour).In(time.FixedZone("PST", -8*60*60)))
	if assert.NoError(t, err) {
		assert.Equal(t, int64(3600), s)
	}
	_, err = TimeSinceEpochSeconds(epoch.Add(-time.Second))
	assert.Error(t, err, "before epoch")
}

// setConfigEnv sets a complete, valid environment for GetConfig.
func setConfigEnv(t *testing.T) {
	for k, v := range map[string]string{
		"DB_NAME":         "remote",
		"DB_HOST":         "localhost",
		"DB_PORT":         "5432",
		"DB_USER":         "user",
		"DB_PASSWORD":     "",
		"DB_CONNECTOR":    "",
		"RECORD_INTERVAL": "",
		"PORT":            "",
		"VERSION":         "",
		"AUTOTLS":         "",
		"SECRET_KEY":      strings.Repeat("s", 20),
		"REFRESH_KEY":     strings.Repeat("r", 20),
		"ADMIN_EMAIL":     "",
		"ADMIN_NAME":      "",
		"ADMIN_PASS":      "",
		"DOMAIN":          "",
	} {
		t.Setenv(k, v)
	}
}

func TestGetConfig(t *testing.T) {
	t.Log("Testing defaults.")
	setConfigEnv(t)
	c, err := GetConfig()
	if assert.NoError(t, err) {
		assert.Equal(t, "remote", c.DBName)
		assert.Equal(t, 5432, c.DBPort)
		assert.Equal(t, "mysql", c.DBDriver, "default driver")
		assert.Equal(t, 300, c.RecordInterval, "default record interval")
		assert.Equal(t, 8181, c.Port, "default port")
		assert.True(t, c.Development, "anything but production is development")
		assert.False(t, c.AutoTLS)
	}
	t.Log("Testing set values.")
	t.Setenv("DB_CONNECTOR", "postgres")
	t.Setenv("RECORD_INTERVAL", "120")
	t.Setenv("PORT", "8080")
	t.Setenv("VERSION", "production")
	t.Setenv("AUTOTLS", "enabled")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("ADMIN_EMAIL", "admin@test.com")
	t.Setenv("ADMIN_NAME", "Admin")
	t.Setenv("ADMIN_PASS", "password")
	t.Setenv("DOMAIN", "remote.test")
	c, err = GetConfig()
	if assert.NoError(t, err) {
		assert.Equal(t, "postgres", c.DBDriver)
		assert.Equal(t, 120, c.RecordInterval)
		assert.Equal(t, 8080, c.Port)
		assert.False(t, c.Development)
		assert.True(t, c.AutoTLS)
		assert.Equal(t, "pass", c.DBPassword)
		assert.Equal(t, "admin@test.com", c.AdminEmail)
		assert.Equal(t, "Admin", c.AdminName)
		assert.Equal(t, "password", c.AdminPass)
		assert.Equal(t, "remote.test", c.Domain)
	}
	t.Log("Testing values that are only on when exactly matched.")
	t.Setenv("VERSION", "Production")
	t.Setenv("AUTOTLS", "true")
	c, err = GetConfig()
	if assert.NoError(t, err) {
		assert.True(t, c.Development)
		assert.False(t, c.AutoTLS)
	}
}

func TestGetConfigMinimums(t *testing.T) {
	for _, c := range []struct {
		interval, port         string
		wantInterval, wantPort int
	}{
		{"abc", "abc", 300, 8181},  // not numbers
		{"10", "80", 300, 8181},    // below the minimum
		{"59", "199", 300, 8181},   // just below
		{"-5", "-5", 300, 8181},    // negative
		{"60", "200", 60, 200},     // exactly the minimum
		{"61", "201", 61, 201},     // just above
		{"900", "8080", 900, 8080}, // well above
	} {
		setConfigEnv(t)
		t.Setenv("RECORD_INTERVAL", c.interval)
		t.Setenv("PORT", c.port)
		cfg, err := GetConfig()
		if assert.NoError(t, err) {
			assert.Equal(t, c.wantInterval, cfg.RecordInterval, "RECORD_INTERVAL=%v", c.interval)
			assert.Equal(t, c.wantPort, cfg.Port, "PORT=%v", c.port)
		}
	}
}

func TestGetConfigErrors(t *testing.T) {
	for name, env := range map[string][2]string{
		"missing DB_NAME":     {"DB_NAME", ""},
		"missing DB_HOST":     {"DB_HOST", ""},
		"missing DB_PORT":     {"DB_PORT", ""},
		"non-numeric DB_PORT": {"DB_PORT", "abc"},
		"negative DB_PORT":    {"DB_PORT", "-1"},
		"missing DB_USER":     {"DB_USER", ""},
		"missing SECRET_KEY":  {"SECRET_KEY", ""},
		"short SECRET_KEY":    {"SECRET_KEY", strings.Repeat("s", 19)},
		"missing REFRESH_KEY": {"REFRESH_KEY", ""},
		"short REFRESH_KEY":   {"REFRESH_KEY", strings.Repeat("r", 19)},
	} {
		t.Run(name, func(t *testing.T) {
			setConfigEnv(t)
			t.Setenv(env[0], env[1])
			c, err := GetConfig()
			assert.Error(t, err)
			assert.Nil(t, c)
		})
	}
}

func TestGetTokenAndEmail(t *testing.T) {
	secret := strings.Repeat("s", 20)
	t.Setenv("SECRET_KEY", secret)
	sign := func(method jwt.SigningMethod, key interface{}, claims jwt.MapClaims) string {
		s, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatalf("Error signing token: %v", err)
		}
		return s
	}
	call := func(auth string) (*string, *string, error) {
		r := httptest.NewRequest("GET", "/", nil)
		if auth != "" {
			r.Header.Set("Authorization", auth)
		}
		return GetTokenAndEmail(r)
	}
	future := time.Now().Add(time.Hour).Unix()
	valid := sign(jwt.SigningMethodHS256, []byte(secret), jwt.MapClaims{"email": "j@test.com", "exp": future})
	tok, email, err := call("Bearer " + valid)
	if assert.NoError(t, err) && assert.NotNil(t, email) && assert.NotNil(t, tok) {
		assert.Equal(t, "j@test.com", *email)
		assert.Equal(t, valid, *tok)
	}
	for name, auth := range map[string]string{
		"no header":        "",
		"no bearer":        valid,
		"extra part":       "Bearer " + valid + " extra",
		"not a token":      "Bearer not-a-token",
		"wrong secret":     "Bearer " + sign(jwt.SigningMethodHS256, []byte(strings.Repeat("x", 20)), jwt.MapClaims{"email": "j@test.com", "exp": future}),
		"expired":          "Bearer " + sign(jwt.SigningMethodHS256, []byte(secret), jwt.MapClaims{"email": "j@test.com", "exp": time.Now().Add(-time.Hour).Unix()}),
		"alg none":         "Bearer " + sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{"email": "j@test.com", "exp": future}),
		"no email":         "Bearer " + sign(jwt.SigningMethodHS256, []byte(secret), jwt.MapClaims{"exp": future}),
		"non-string email": "Bearer " + sign(jwt.SigningMethodHS256, []byte(secret), jwt.MapClaims{"email": 5, "exp": future}),
	} {
		tok, email, err := call(auth)
		assert.Error(t, err, name)
		assert.Nil(t, tok, name)
		assert.Nil(t, email, name)
	}
}
