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

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestDefaultKeyTypeIsRead(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	account := variables.accounts[1]
	token, refresh, err := createTokens(account.Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account.Token = *token
	account.RefreshToken = *refresh
	if err := database.UpdateTokens(account); err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	send := func(method, target string, body interface{}, handle func(*echo.Context) error) types.Key {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Error encoding request: %v", err)
		}
		request := httptest.NewRequest(method, target, strings.NewReader(string(b)))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
		response := httptest.NewRecorder()
		assert.NoError(t, handle(e.NewContext(request, response)))
		var resp types.ModifyKeyResponse
		if assert.Equal(t, http.StatusOK, response.Code) {
			assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp))
		}
		return resp.Key
	}
	t.Log("Testing add with default type.")
	added := send(http.MethodPost, "/r/key/add", types.AddKeyRequest{Key: types.RequestKey{Type: "default", Name: "default-reader"}}, h.AddKey)
	assert.Equal(t, "read", added.Type)
	stored, err := database.GetKey(added.Value)
	if assert.NoError(t, err) && assert.NotNil(t, stored) {
		assert.Equal(t, "read", stored.Type)
	}
	t.Log("Testing update with default type.")
	value := variables.keys[account.Email][0].Value
	updated := send(http.MethodPut, "/r/key/update", types.UpdateKeyRequest{Key: types.RequestKey{Value: value, Type: "default", Name: "default-update"}}, h.UpdateKey)
	assert.Equal(t, "read", updated.Type)
	stored, err = database.GetKey(value)
	if assert.NoError(t, err) && assert.NotNil(t, stored) {
		assert.Equal(t, "read", stored.Type)
	}
}
