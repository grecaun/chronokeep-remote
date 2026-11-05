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

func heartbeatRequest(method, target, body, auth string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if auth != "" {
		request.Header.Set(echo.HeaderAuthorization, auth)
	}
	return request
}

func TestSaveHeartbeat(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	hb := types.Heartbeat{
		Device:           "Portal 1",
		Location:         "Portal 1",
		Now:              1762300000000,
		Uptime:           3600,
		Battery:          "4.10",
		Charging:         true,
		Hash:             "Chronokeep Portal v1.0.0",
		UploadErrors:     2,
		ReadersConnected: 1,
		ReadersTotal:     2,
		Reading:          true,
		ChipType:         "DEC",
	}
	body, err := json.Marshal(hb)
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	send := func(b, auth string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		c := e.NewContext(heartbeatRequest(http.MethodPost, "/heartbeat", b, auth), response)
		assert.NoError(t, h.SaveHeartbeat(c))
		return response
	}
	t.Log("Testing no key given.")
	assert.Equal(t, http.StatusUnauthorized, send(string(body), "").Code)
	t.Log("Testing expired key.")
	assert.Equal(t, http.StatusUnauthorized, send(string(body), "Bearer "+variables.knownValues["expired"]).Code)
	t.Log("Testing invalid key.")
	assert.Equal(t, http.StatusUnauthorized, send(string(body), "Bearer not-a-valid-key").Code)
	t.Log("Testing read key.")
	assert.Equal(t, http.StatusUnauthorized, send(string(body), "Bearer "+variables.knownValues["read"]).Code)
	t.Log("Testing bad request.")
	assert.Equal(t, http.StatusBadRequest, send("////", "Bearer "+variables.knownValues["write"]).Code)
	t.Log("Testing invalid heartbeat.")
	bad := hb
	bad.Uptime = -1
	badBody, _ := json.Marshal(bad)
	assert.Equal(t, http.StatusBadRequest, send(string(badBody), "Bearer "+variables.knownValues["write"]).Code)
	t.Log("Testing valid heartbeat with write key.")
	assert.Equal(t, http.StatusOK, send(string(body), "Bearer "+variables.knownValues["write"]).Code)
	t.Log("Testing relay-style heartbeat (no battery fields) replaces the previous one.")
	relay := `{"dev":"Relay","location":"Relay","now":1762300060000,"uptime":60,"hash":"Chronokeep Relay v0.1.0","upload_errors":0,"readers_connected":0,"readers_total":1,"reading":false,"chip_type":""}`
	assert.Equal(t, http.StatusOK, send(relay, "Bearer "+variables.knownValues["write"]).Code)
	t.Log("Testing delete key can't send heartbeats.")
	response := send(string(body), "Bearer "+variables.knownValues["delete3"])
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Body.String(), "Delete Keys Cannot Write")
	beats, err := database.GetHeartbeats(variables.accounts[0].Identifier, "")
	if assert.NoError(t, err) && assert.Len(t, beats, 1) {
		assert.Equal(t, "reader2", beats[0].ReaderName)
		assert.Equal(t, "Relay", beats[0].Heartbeat.Device)
		assert.Equal(t, int64(60), beats[0].Heartbeat.Uptime)
		assert.Equal(t, "", beats[0].Heartbeat.Battery)
		assert.Equal(t, 1, beats[0].Heartbeat.ReadersTotal)
	}
}

func TestGetHeartbeats(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	hb := types.Heartbeat{Device: "Portal 6", Location: "Portal 6", Uptime: 10, ReadersTotal: 1, Hash: "Chronokeep Portal v1.0.0"}
	if err := database.SaveHeartbeat(variables.knownValues["write2"], hb); err != nil {
		t.Fatalf("Unexpected error saving heartbeat: %v", err)
	}
	hb7 := types.Heartbeat{Device: "Relay 7", Location: "Relay 7", Uptime: 20, Hash: "Chronokeep Relay v0.1.0"}
	if err := database.SaveHeartbeat(variables.knownValues["delete2"], hb7); err != nil {
		t.Fatalf("Unexpected error saving heartbeat: %v", err)
	}
	get := func(body, auth string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		c := e.NewContext(heartbeatRequest(http.MethodGet, "/heartbeats", body, auth), response)
		assert.NoError(t, h.GetHeartbeats(c))
		return response
	}
	decode := func(response *httptest.ResponseRecorder) []types.ReaderHeartbeat {
		var resp types.GetHeartbeatsResponse
		if !assert.Equal(t, http.StatusOK, response.Code) || !assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			return nil
		}
		assert.NotNil(t, resp.Heartbeats)
		return resp.Heartbeats
	}
	readKey := "Bearer " + variables.knownValues["read"]
	t.Log("Testing no key given.")
	assert.Equal(t, http.StatusUnauthorized, get("", "").Code)
	t.Log("Testing expired key.")
	assert.Equal(t, http.StatusUnauthorized, get("", "Bearer "+variables.knownValues["expired"]).Code)
	t.Log("Testing invalid key.")
	assert.Equal(t, http.StatusUnauthorized, get("", "Bearer not-a-valid-key").Code)
	t.Log("Testing bad request.")
	assert.Equal(t, http.StatusBadRequest, get("////", readKey).Code)
	t.Log("Testing no body returns every reader.")
	if beats := decode(get("", readKey)); assert.Len(t, beats, 2) {
		assert.Equal(t, "reader6", beats[0].ReaderName)
		assert.Equal(t, hb, beats[0].Heartbeat)
		assert.False(t, beats[0].Received.IsZero())
		assert.Equal(t, "reader7", beats[1].ReaderName)
		assert.Equal(t, hb7, beats[1].Heartbeat)
	}
	t.Log("Testing empty reader returns every reader.")
	assert.Len(t, decode(get(`{"reader":""}`, readKey)), 2)
	t.Log("Testing single reader.")
	if beats := decode(get(`{"reader":"reader7"}`, readKey)); assert.Len(t, beats, 1) {
		assert.Equal(t, "reader7", beats[0].ReaderName)
		assert.Equal(t, hb7, beats[0].Heartbeat)
	}
	t.Log("Testing reader without a heartbeat.")
	assert.Len(t, decode(get(`{"reader":"reader4"}`, readKey)), 0)
	t.Log("Testing reader on another account.")
	assert.Len(t, decode(get(`{"reader":"reader6"}`, "Bearer "+variables.knownValues["write"])), 0)
	t.Log("Testing other account sees none.")
	assert.Len(t, decode(get("", "Bearer "+variables.knownValues["write"])), 0)
}

func TestHeartbeatRequestEdgeCases(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	post := func(body, key string) int {
		response := httptest.NewRecorder()
		c := e.NewContext(heartbeatRequest(http.MethodPost, "/heartbeat", body, "Bearer "+key), response)
		assert.NoError(t, h.SaveHeartbeat(c))
		return response.Code
	}
	get := func(body, key string) (int, []types.ReaderHeartbeat) {
		response := httptest.NewRecorder()
		c := e.NewContext(heartbeatRequest(http.MethodGet, "/heartbeats", body, "Bearer "+key), response)
		assert.NoError(t, h.GetHeartbeats(c))
		var resp types.GetHeartbeatsResponse
		if response.Code == http.StatusOK {
			assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp))
		}
		return response.Code, resp.Heartbeats
	}
	write := variables.knownValues["write2"] // reader6, account 2
	t.Log("Testing wrong JSON types.")
	for _, body := range []string{`{"uptime":"60"}`, `{"reading":"yes"}`, `{"dev":5}`, `[]`, `"heartbeat"`} {
		assert.Equal(t, http.StatusBadRequest, post(body, write), body)
	}
	t.Log("Testing unknown fields are ignored.")
	assert.Equal(t, http.StatusOK, post(`{"dev":"Portal","uptime":5,"new_field":{"a":1}}`, write))
	t.Log("Testing empty object is a valid heartbeat.")
	assert.Equal(t, http.StatusOK, post(`{}`, write))
	t.Log("Testing over-long device name.")
	assert.Equal(t, http.StatusBadRequest, post(`{"dev":"`+strings.Repeat("a", 101)+`"}`, write))
	_, beats := get(`{"reader":"reader6"}`, write)
	if assert.Len(t, beats, 1) {
		assert.Equal(t, types.Heartbeat{}, beats[0].Heartbeat, "latest heartbeat is the empty one")
	}
	t.Log("Testing reader name match ignores case but not spacing.")
	code, beats := get(`{"reader":"READER6"}`, write)
	assert.Equal(t, http.StatusOK, code)
	if assert.Len(t, beats, 1) {
		assert.Equal(t, "reader6", beats[0].ReaderName, "name is returned as stored")
	}
	code, beats = get(`{"reader":" reader6"}`, write)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, beats, 0)
	t.Log("Testing wrong JSON type for reader.")
	code, _ = get(`{"reader":6}`, write)
	assert.Equal(t, http.StatusBadRequest, code)
	t.Log("Testing deleted key can't send or be listed.")
	keys, err := database.GetAccountKeysByKey(write)
	if err != nil {
		t.Fatalf("Error getting keys: %v", err)
	}
	for _, k := range keys {
		if k.Value == write {
			if err := database.DeleteKey(k); err != nil {
				t.Fatalf("Error deleting key: %v", err)
			}
		}
	}
	assert.Equal(t, http.StatusUnauthorized, post(`{}`, write))
	code, beats = get(`{"reader":"reader6"}`, variables.knownValues["read"])
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, beats, 0)
	code, beats = get("", variables.knownValues["read"])
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, beats, 0)
}
