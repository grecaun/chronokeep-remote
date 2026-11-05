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

package handlers

import (
	"chronokeep/remote/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// setTokenKeys gives the test config distinct, realistic signing keys
// (setupTests leaves both empty, which hides access/refresh key mix-ups).
func setTokenKeys() {
	config.SecretKey = strings.Repeat("s", 32)
	config.RefreshKey = strings.Repeat("r", 32)
}

func signToken(t *testing.T, method jwt.SigningMethod, key interface{}, claims jwt.MapClaims) string {
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("Error signing token: %v", err)
	}
	return s
}

func TestVerifyTokenTampering(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	setTokenKeys()
	account := variables.accounts[0]
	future := time.Now().Add(time.Hour).Unix()
	verify := func(token string) (*types.Account, error) {
		r := httptest.NewRequest(http.MethodPost, "/r/account", nil)
		r.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		return verifyToken(r)
	}
	// Each token is stored on the account first, so only the token itself can cause the failure.
	use := func(token string) {
		account.Token = token
		if err := database.UpdateTokens(account); err != nil {
			t.Fatalf("Error storing token: %v", err)
		}
	}
	token, refresh, err := createTokens(account.Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	t.Log("Testing valid token.")
	use(*token)
	got, err := verify(*token)
	if assert.NoError(t, err) {
		assert.Equal(t, account.Email, got.Email)
	}
	for name, bad := range map[string]string{
		"refresh token used as access token": *refresh,
		"alg none":                           signToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{"email": account.Email, "exp": future}),
		"signed with wrong key":              signToken(t, jwt.SigningMethodHS256, []byte(strings.Repeat("x", 32)), jwt.MapClaims{"email": account.Email, "exp": future}),
		"no email claim":                     signToken(t, jwt.SigningMethodHS256, []byte(config.SecretKey), jwt.MapClaims{"exp": future}),
		"non-string email claim":             signToken(t, jwt.SigningMethodHS256, []byte(config.SecretKey), jwt.MapClaims{"email": 7, "exp": future}),
		"no expiry, other account's email":   signToken(t, jwt.SigningMethodHS256, []byte(config.SecretKey), jwt.MapClaims{"email": variables.accounts[1].Email}),
	} {
		t.Logf("Testing %v.", name)
		use(bad)
		got, err := verify(bad)
		assert.Error(t, err, name)
		assert.Nil(t, got, name)
	}
	t.Log("Testing token for a deleted account.")
	token, _, err = createTokens(account.Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	use(*token)
	if err := database.DeleteAccount(account.Identifier); err != nil {
		t.Fatalf("Error deleting account: %v", err)
	}
	got, err = verify(*token)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestRefreshTokenTampering(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	setTokenKeys()
	e := echo.New()
	h := Handler{}
	account := variables.accounts[0]
	future := time.Now().Add(time.Hour).Unix()
	refresh := func(token string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(types.RefreshTokenRequest{RefreshToken: token})
		request := httptest.NewRequest(http.MethodPost, "/account/refresh", strings.NewReader(string(body)))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		response := httptest.NewRecorder()
		assert.NoError(t, h.Refresh(e.NewContext(request, response)))
		return response
	}
	use := func(token string) {
		account.RefreshToken = token
		if err := database.UpdateTokens(account); err != nil {
			t.Fatalf("Error storing token: %v", err)
		}
	}
	access, _, err := createTokens(account.Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	for name, bad := range map[string]string{
		"access token used as refresh token": *access,
		"alg none":                           signToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{"email": account.Email, "exp": future}),
		"signed with wrong key":              signToken(t, jwt.SigningMethodHS256, []byte(strings.Repeat("x", 32)), jwt.MapClaims{"email": account.Email, "exp": future}),
		"no email claim":                     signToken(t, jwt.SigningMethodHS256, []byte(config.RefreshKey), jwt.MapClaims{"exp": future}),
		"unknown account":                    signToken(t, jwt.SigningMethodHS256, []byte(config.RefreshKey), jwt.MapClaims{"email": "nobody@test.com", "exp": future}),
		"empty token":                        "",
	} {
		t.Logf("Testing %v.", name)
		use(bad)
		assert.Equal(t, http.StatusUnauthorized, refresh(bad).Code, name)
	}
	t.Log("Testing valid refresh token still works with distinct keys.")
	_, valid, err := createTokens(account.Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	use(*valid)
	response := refresh(*valid)
	if assert.Equal(t, http.StatusOK, response.Code) {
		var resp map[string]string
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			// The new access token works and the new refresh token doesn't work as an access token.
			r := httptest.NewRequest(http.MethodPost, "/r/account", nil)
			r.Header.Set(echo.HeaderAuthorization, "Bearer "+resp["access_token"])
			_, err := verifyToken(r)
			assert.NoError(t, err)
			r.Header.Set(echo.HeaderAuthorization, "Bearer "+resp["refresh_token"])
			_, err = verifyToken(r)
			assert.Error(t, err)
		}
	}
}
