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

// portalUpload is the body Chronokeep Portal sends to /reads/add, copied from its
// upload_reads test. Besides the reads it sends "dev", the portal's device name, and each
// read's reader_seconds and reader_milliseconds, none of which this API uses.
const portalUpload = `{"dev":"finish-line-1","reads":[` +
	`{"antenna":2,"ident_type":"chip","identifier":"1001","milliseconds":250,"reader":"Finish","reader_milliseconds":500,"reader_seconds":1700000001,"rssi":"-55","seconds":1700000000,"type":"reader"},` +
	`{"antenna":1,"ident_type":"chip","identifier":"1002","milliseconds":0,"reader":"Finish","reader_milliseconds":0,"reader_seconds":1700000011,"rssi":"-60","seconds":1700000010,"type":"reader"}]}`

func TestAddReadsFromPortal(t *testing.T) {
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()

	upload := func(body string) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(body))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
		response := httptest.NewRecorder()
		if !assert.NoError(t, h.AddReads(e.NewContext(request, response))) {
			t.FailNow()
		}
		return response
	}
	count := func(response *httptest.ResponseRecorder) int64 {
		t.Helper()
		var resp types.UploadReadsResponse
		assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp))
		return resp.Count
	}

	// the portal's upload, dev and all, is accepted and every read is stored
	response := upload(portalUpload)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, int64(2), count(response))

	mkey, err := database.GetKeyAndAccount(variables.knownValues["write"])
	if err != nil || mkey == nil {
		t.Fatalf("Error getting the write key: %v", err)
	}
	stored, err := database.GetReads(mkey.Account.Identifier, mkey.Key.Name, 1699999999, 1700000020)
	if assert.NoError(t, err) && assert.Len(t, stored, 2) {
		byIdentifier := map[string]types.Read{}
		for _, r := range stored {
			byIdentifier[r.Identifier] = r
		}
		first := byIdentifier["1001"]
		assert.Equal(t, int64(1700000000), first.Seconds)
		assert.Equal(t, 250, first.Milliseconds)
		assert.Equal(t, 2, first.Antenna)
		assert.Equal(t, "Finish", first.Reader)
		assert.Equal(t, "-55", first.RSSI)
		assert.Equal(t, "chip", first.IdentType)
		assert.Equal(t, "reader", first.Type)
		assert.Contains(t, byIdentifier, "1002")
	}

	// older portals don't send dev, and other values for it are fine too
	for _, body := range []string{
		strings.Replace(portalUpload, `"dev":"finish-line-1",`, ``, 1),
		strings.Replace(portalUpload, `"dev":"finish-line-1"`, `"dev":""`, 1),
		strings.Replace(portalUpload, `"dev":"finish-line-1"`, `"dev":null`, 1),
	} {
		response := upload(body)
		assert.Equal(t, http.StatusOK, response.Code, body)
		assert.Equal(t, int64(2), count(response), body)
	}
}
