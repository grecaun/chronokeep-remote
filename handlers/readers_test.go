package handlers

import (
	"chronokeep/remote/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetReaders(t *testing.T) {
	// GET, /readers
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	// Test no key
	t.Log("Testing no key given.")
	request := httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired key
	t.Log("Testing expired key.")
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["expired"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid key
	t.Log("Testing invalid key.")
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer not-a-valid-key")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid authorization header
	t.Log("Testing invalid authorization header.")
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "not-a-valid-auth-header")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test wrong content type
	t.Log("Testing wrong content type.")
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
	}
	// Test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader("////"))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
	}
	// Test valid
	request = httptest.NewRequest(http.MethodGet, "/readers", strings.NewReader(string("")))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReaders(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadersResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 4, len(resp.Readers))
			keys, err := database.GetAccountKeys(variables.accounts[1].Email)
			if assert.NoError(t, err) {
				assert.Equal(t, len(keys), len(resp.Readers))
				for _, k := range keys {
					found := false
					for _, read := range resp.Readers {
						if k.ReaderName == read.Identifier {
							found = true
							assert.Equal(t, k.ReaderName, read.Identifier)
							assert.Equal(t, k.Name, read.Name)
						}
					}
					assert.True(t, found)
				}
			}
		}
	}
}
