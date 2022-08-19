package handlers

import (
	"chronokeep/remote/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetKeys(t *testing.T) {
	// POST, /r/key
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	// Test empty auth header
	t.Log("Testing empty auth header.")
	request := httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid auth header
	t.Log("Testing invalid auth header.")
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid token
	t.Log("Testing invalid token.")
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test unknown email
	t.Log("Testing unknown email in token.")
	token, _, err := createTokens("unknown@test.com")
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired token
	t.Log("Testing expired token.")
	token, refresh, err := createExpiredTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account := variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test not logged in
	t.Log("Testing not logged in.")
	token, _, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test locked account
	lockAccount(t, variables.accounts[0].Email, e, h)
	t.Log("Testing locked account.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	account.Locked = true
	err = database.UnlockAccount(account)
	if err != nil {
		t.Fatalf("Error unlocking account during test: %v", err)
	}
	// test bad request
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("////")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test invalid content type
	t.Log("Testing invalid content type.")
	body, err := json.Marshal(types.GetKeysRequest{
		Email: &variables.accounts[0].Email,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test empty request -- valid
	t.Log("Testing empty request -- valid.")
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetKeysResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetAccountKeys(variables.accounts[0].Email)
			if assert.NoError(t, err) {
				assert.Equal(t, len(keys), len(resp.Keys))
				for _, key := range keys {
					found := false
					for _, inner := range resp.Keys {
						if key.Value == inner.Value {
							found = true
							assert.Equal(t, key.Name, inner.Name)
							assert.Equal(t, key.ReaderName, inner.ReaderName)
							assert.Equal(t, key.Type, inner.Type)
							if key.ValidUntil != nil {
								assert.Equal(t, key.ValidUntil.Local(), inner.ValidUntil.Local())
							} else {
								assert.Equal(t, key.ValidUntil, inner.ValidUntil)
							}
						}
					}
					assert.True(t, found)
				}
			}
		}
	}
	// test other email -- admin -- valid
	t.Log("Testing other email -- admin -- valid.")
	body, err = json.Marshal(types.GetKeysRequest{
		Email: &variables.accounts[1].Email,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetKeysResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetAccountKeys(variables.accounts[1].Email)
			if assert.NoError(t, err) {
				assert.Equal(t, len(keys), len(resp.Keys))
				for _, key := range keys {
					found := false
					for _, inner := range resp.Keys {
						if key.Value == inner.Value {
							found = true
							assert.Equal(t, key.Name, inner.Name)
							assert.Equal(t, key.ReaderName, inner.ReaderName)
							assert.Equal(t, key.Type, inner.Type)
							if key.ValidUntil != nil {
								assert.Equal(t, key.ValidUntil.Local(), inner.ValidUntil.Local())
							} else {
								assert.Equal(t, key.ValidUntil, inner.ValidUntil)
							}
						}
					}
					assert.True(t, found)
				}
			}
		}
	}
	// test other email -- non-admin -- invalid
	t.Log("Testing other email -- non-admin -- invalid.")
	token, refresh, err = createTokens(variables.accounts[2].Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	account = variables.accounts[2]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating tokens: %v", err)
	}
	body, err = json.Marshal(types.GetKeysRequest{
		Email: &variables.accounts[1].Email,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// test unknown account
	t.Log("Testing unknown email.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating tokens: %v", err)
	}
	email := "unknown@test.com"
	body, err = json.Marshal(types.GetKeysRequest{
		Email: &email,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
	// test token for wrong account //->//
	t.Log("Test token with embeded email not belonging to account it is attached to.")
	account = variables.accounts[0]
	account.Token = "totally-not-valid-token"
	account.RefreshToken = "not-a-valid-refresh-token"
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetKeys(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
}

func TestAddKey(t *testing.T) {
	// POST, /r/key/add
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	// Test empty auth header
	t.Log("Testing empty auth header.")
	request := httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid auth header
	t.Log("Testing invalid auth header.")
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid token
	t.Log("Testing invalid token.")
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test unknown email
	t.Log("Testing unknown email in token.")
	token, _, err := createTokens("unknown@test.com")
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired token
	t.Log("Testing expired token.")
	token, refresh, err := createExpiredTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account := variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test not logged in
	t.Log("Testing not logged in.")
	token, _, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test locked account
	lockAccount(t, variables.accounts[0].Email, e, h)
	t.Log("Testing locked account.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	account.Locked = true
	err = database.UnlockAccount(account)
	if err != nil {
		t.Fatalf("Error unlocking account during test: %v", err)
	}
	// test empty request
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	t.Log("Testing empty request.")
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("////")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test invalid content type
	t.Log("Testing invalid content type.")
	body, err := json.Marshal(types.AddKeyRequest{
		Email: &variables.accounts[0].Email,
		Key: types.RequestKey{
			Type:       "read",
			ReaderName: "reader12",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test valid request -- self
	t.Log("Testing valid request -- self add.")
	token, refresh, err = createTokens(variables.accounts[1].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.AddKeyRequest{
		Key: types.RequestKey{
			Type:       "read",
			ReaderName: "reader13",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.ModifyKeyResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetKey(resp.Key.Value)
			if assert.NoError(t, err) {
				assert.Equal(t, keys.Value, resp.Key.Value)
				assert.Equal(t, keys.Name, resp.Key.Name)
				assert.Equal(t, keys.ReaderName, resp.Key.ReaderName)
				assert.Equal(t, keys.Type, resp.Key.Type)
				if keys.ValidUntil != nil {
					assert.Equal(t, keys.ValidUntil.Local(), resp.Key.ValidUntil.Local())
				} else {
					assert.Equal(t, keys.ValidUntil, resp.Key.ValidUntil)
				}
			}
		}
	}
	// test valid request -- admin for other
	t.Log("Testing valid request -- admin for other.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.AddKeyRequest{
		Email: &variables.accounts[2].Email,
		Key: types.RequestKey{
			Type:       "read",
			ReaderName: "reader14",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.ModifyKeyResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetKey(resp.Key.Value)
			if assert.NoError(t, err) {
				assert.Equal(t, keys.Value, resp.Key.Value)
				assert.Equal(t, keys.Name, resp.Key.Name)
				assert.Equal(t, keys.ReaderName, resp.Key.ReaderName)
				assert.Equal(t, keys.Type, resp.Key.Type)
				if keys.ValidUntil != nil {
					assert.Equal(t, keys.ValidUntil.Local(), resp.Key.ValidUntil.Local())
				} else {
					assert.Equal(t, keys.ValidUntil, resp.Key.ValidUntil)
				}
			}
		}
	}
	// test invalid request -- non-admin for other
	t.Log("Testing valid request -- non-admin for other.")
	token, refresh, err = createTokens(variables.accounts[1].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.AddKeyRequest{
		Email: &variables.accounts[2].Email,
		Key: types.RequestKey{
			Type:       "read",
			ReaderName: "reader15",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// test unknown email
	t.Log("Testing unknown email.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	email := "unknown@test.com"
	body, err = json.Marshal(types.AddKeyRequest{
		Email: &email,
		Key: types.RequestKey{
			Type:       "read",
			ReaderName: "reader16",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
	// test token for wrong account //->//
	t.Log("Test token with embeded email not belonging to account it is attached to.")
	account = variables.accounts[0]
	account.Token = "totally-not-valid-token"
	account.RefreshToken = "not-a-valid-refresh-token"
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// validation -- type
	t.Log("Testing type validation.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.AddKeyRequest{
		Key: types.RequestKey{
			Type:       "",
			ReaderName: "reader17",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	body, err = json.Marshal(types.AddKeyRequest{
		Key: types.RequestKey{
			Type:       "unknown",
			ReaderName: "reader18",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/r/key/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
}

func TestUpdateKey(t *testing.T) {
	// PUT, /r/key/update
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	// Test empty auth header
	t.Log("Testing empty auth header.")
	request := httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid auth header
	t.Log("Testing invalid auth header.")
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid token
	t.Log("Testing invalid token.")
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test unknown email
	t.Log("Testing unknown email in token.")
	token, _, err := createTokens("unknown@test.com")
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired token
	t.Log("Testing expired token.")
	token, refresh, err := createExpiredTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account := variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test not logged in
	t.Log("Testing not logged in.")
	token, _, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test locked account
	lockAccount(t, variables.accounts[0].Email, e, h)
	t.Log("Testing locked account.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	account.Locked = true
	err = database.UnlockAccount(account)
	if err != nil {
		t.Fatalf("Error unlocking account during test: %v", err)
	}
	// test empty request
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	t.Log("Testing empty request.")
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("////")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test invalid content type
	t.Log("Testing invalid content type.")
	body, err := json.Marshal(types.UpdateKeyRequest{
		Key: types.RequestKey{
			Value:      variables.keys[variables.accounts[0].Email][0].Value,
			Name:       "updated-name",
			Type:       "read",
			ReaderName: "reader11",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test valid request -- self update
	t.Log("Testing valid request -- self add.")
	token, refresh, err = createTokens(variables.accounts[1].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	validUntil := time.Now().Add(50 * time.Hour).Format("2006/01/02 15:04:05")
	key := types.RequestKey{
		Value:      variables.keys[variables.accounts[1].Email][0].Value,
		Name:       "updated-name",
		Type:       "read",
		ReaderName: "reader12",
		ValidUntil: validUntil,
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: key,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.ModifyKeyResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetKey(resp.Key.Value)
			if assert.NoError(t, err) {
				assert.Equal(t, keys.Value, resp.Key.Value)
				assert.Equal(t, keys.Name, resp.Key.Name)
				assert.Equal(t, keys.ReaderName, resp.Key.ReaderName)
				assert.Equal(t, keys.Type, resp.Key.Type)
				if keys.ValidUntil != nil {
					assert.Equal(t, keys.ValidUntil.Local(), resp.Key.ValidUntil.Local())
				} else {
					assert.Equal(t, keys.ValidUntil, resp.Key.ValidUntil)
				}
				assert.Equal(t, key.Name, keys.Name)
				assert.Equal(t, key.Value, keys.Value)
				assert.Equal(t, key.ReaderName, keys.ReaderName)
				assert.Equal(t, key.Type, keys.Type)
			}
		}
	}
	// test valid request -- admin update other
	t.Log("Testing valid request -- admin update other.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	validUntil = time.Now().Add(50 * time.Hour).Format("2006/01/02 15:04:05")
	key = types.RequestKey{
		Value:      variables.keys[variables.accounts[1].Email][2].Value,
		Name:       "updated-name-2",
		Type:       "write",
		ReaderName: "new-read26.com",
		ValidUntil: validUntil,
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: key,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.ModifyKeyResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			keys, err := database.GetKey(resp.Key.Value)
			if assert.NoError(t, err) {
				assert.Equal(t, keys.Value, resp.Key.Value)
				assert.Equal(t, keys.Name, resp.Key.Name)
				assert.Equal(t, keys.ReaderName, resp.Key.ReaderName)
				assert.Equal(t, keys.Type, resp.Key.Type)
				if keys.ValidUntil != nil {
					assert.Equal(t, keys.ValidUntil.Local(), resp.Key.ValidUntil.Local())
				} else {
					assert.Equal(t, keys.ValidUntil, resp.Key.ValidUntil)
				}
				assert.Equal(t, key.Name, keys.Name)
				assert.Equal(t, key.Value, keys.Value)
				assert.Equal(t, key.ReaderName, keys.ReaderName)
				assert.Equal(t, key.Type, keys.Type)
			}
		}
	}
	// test invalid request -- non-admin update other
	t.Log("Testing invalid request -- non-admin update other.")
	token, refresh, err = createTokens(variables.accounts[2].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[2]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	validUntil = time.Now().Add(50 * time.Hour).Format("2006/01/02 15:04:05")
	key = types.RequestKey{
		Value:      variables.keys[variables.accounts[1].Email][2].Value,
		Name:       "updated-name-2",
		Type:       "write",
		ReaderName: "new-read2.com",
		ValidUntil: validUntil,
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: key,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// test unknown key
	t.Log("Test unknown key.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	validUntil = time.Now().Add(50 * time.Hour).Format("2006/01/02 15:04:05")
	key = types.RequestKey{
		Value:      "unknown-key-value",
		Name:       "updated-name-2",
		Type:       "write",
		ReaderName: "new-read21.com",
		ValidUntil: validUntil,
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: key,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
	// test token for wrong account //->//
	t.Log("Test token with embeded email not belonging to account it is attached to.")
	account = variables.accounts[0]
	account.Token = "totally-not-valid-token"
	account.RefreshToken = "not-a-valid-refresh-token"
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// validation -- type
	t.Log("Testing type validation.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: types.RequestKey{
			Value:      variables.keys[variables.accounts[1].Email][2].Value,
			Type:       "",
			ReaderName: "reader15",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	body, err = json.Marshal(types.UpdateKeyRequest{
		Key: types.RequestKey{
			Value:      variables.keys[variables.accounts[1].Email][2].Value,
			Type:       "unknown",
			ReaderName: "reader16",
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/r/key/update", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.UpdateKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
}

func TestDeleteKey(t *testing.T) {
	// DELETE, /r/key/delete
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	// Test empty auth header
	t.Log("Testing empty auth header.")
	request := httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid auth header
	t.Log("Testing invalid auth header.")
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid token
	t.Log("Testing invalid token.")
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer invalid-bearer-token")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test unknown email
	t.Log("Testing unknown email in token.")
	token, _, err := createTokens("unknown@test.com")
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired token
	t.Log("Testing expired token.")
	token, refresh, err := createExpiredTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account := variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test not logged in
	t.Log("Testing not logged in.")
	token, _, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test locked account
	lockAccount(t, variables.accounts[0].Email, e, h)
	t.Log("Testing locked account.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	account.Locked = true
	err = database.UnlockAccount(account)
	if err != nil {
		t.Fatalf("Error unlocking account during test: %v", err)
	}
	// test empty request
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	t.Log("Testing empty request.")
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("////")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test invalid content type
	t.Log("Testing invalid content type.")
	body, err := json.Marshal(types.DeleteKeyRequest{
		Key: variables.keys[variables.accounts[0].Email][0].Value,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// test valid request -- self delete
	t.Log("Testing valid request -- self delete.")
	token, refresh, err = createTokens(variables.accounts[1].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	keyVal := variables.keys[variables.accounts[1].Email][0].Value
	body, err = json.Marshal(types.DeleteKeyRequest{
		Key: keyVal,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	key, err := database.GetKey(keyVal)
	if assert.NoError(t, err) {
		assert.NotNil(t, key)
	}
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		key, err = database.GetKey(keyVal)
		if assert.NoError(t, err) {
			assert.Nil(t, key)
		}
	}
	// test valid request -- admin delete
	t.Log("Testing valid request -- admin delete other.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	keyVal = variables.keys[variables.accounts[1].Email][1].Value
	body, err = json.Marshal(types.DeleteKeyRequest{
		Key: keyVal,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	key, err = database.GetKey(keyVal)
	if assert.NoError(t, err) {
		assert.NotNil(t, key)
	}
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		key, err = database.GetKey(keyVal)
		if assert.NoError(t, err) {
			assert.Nil(t, key)
		}
	}
	// test invalid request -- non-admin delete other
	t.Log("Testing invalid request -- non-admin delete other.")
	token, refresh, err = createTokens(variables.accounts[2].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[2]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	keyVal = variables.keys[variables.accounts[1].Email][2].Value
	body, err = json.Marshal(types.DeleteKeyRequest{
		Key: keyVal,
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	key, err = database.GetKey(keyVal)
	if assert.NoError(t, err) {
		assert.NotNil(t, key)
	}
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
		key, err = database.GetKey(keyVal)
		if assert.NoError(t, err) {
			assert.NotNil(t, key)
		}
	}
	// test unknown key
	t.Log("Testing unknown key.")
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[0]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	body, err = json.Marshal(types.DeleteKeyRequest{
		Key: "invalid-key-value",
	})
	if err != nil {
		t.Fatalf("Error encoding request into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
	// test token for wrong account //->//
	t.Log("Test token with embeded email not belonging to account it is attached to.")
	account = variables.accounts[0]
	account.Token = "totally-not-valid-token"
	account.RefreshToken = "not-a-valid-refresh-token"
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	token, refresh, err = createTokens(variables.accounts[0].Email)
	if err != nil {
		t.Fatalf("Error creating test tokens: %v", err)
	}
	account = variables.accounts[1]
	account.Token = *token
	account.RefreshToken = *refresh
	err = database.UpdateTokens(account)
	if err != nil {
		t.Fatalf("Error updating test tokens: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/r/key/delete", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+*token)
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteKey(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
}
