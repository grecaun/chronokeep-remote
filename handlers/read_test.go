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

func TestGetReads(t *testing.T) {
	// GET, /reads
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	// Test no key
	t.Log("Testing no key given.")
	body, err := json.Marshal(types.GetReadsRequest{
		ReaderName: "reader5",
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired key
	t.Log("Testing expired key.")
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["expired"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid key
	t.Log("Testing invalid key.")
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer not-a-valid-key")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid authorization header
	t.Log("Testing invalid authorization header.")
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "not-a-valid-auth-header")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test wrong content type
	t.Log("Testing wrong content type.")
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader("////"))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test account mis-match.
	t.Log("Testing account mis-match.")
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader1",
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
			assert.Equal(t, 0, len(resp.Reads))
		}
	}
	// Test invalid reader
	t.Log("Testing invalid reader (reader not found).")
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "unknownreader",
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
			assert.Equal(t, 0, len(resp.Reads))
		}
	}
	// Test valid
	t.Log("Testing valid request.")
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader5",
		Start:      0,
		End:        10000,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 300, len(resp.Reads))
			assert.Equal(t, 300, resp.Count)
			assert.Equal(t, "chip", resp.Reads[0].IdentType)
			assert.Equal(t, "reader", resp.Reads[0].Type)
		}
	}
	// Test start/end values
	t.Log("Testing start/end values.")
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader5",
		Start:      35,
		End:        0,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, resp.Count, len(resp.Reads))
			for _, read := range resp.Reads {
				assert.True(t, read.Seconds <= 35+360 && read.Seconds >= 35)
				assert.Equal(t, "chip", read.IdentType)
				assert.Equal(t, "reader", read.Type)
			}
		}
	}
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader5",
		Start:      135,
		End:        550,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, resp.Count, len(resp.Reads))
			for _, read := range resp.Reads {
				assert.True(t, read.Seconds <= 550 && read.Seconds >= 135)
				assert.Equal(t, "chip", read.IdentType)
				assert.Equal(t, "reader", read.Type)
			}
		}
	}
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader5",
		Start:      10000,
		End:        10050,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/reads", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.GetReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.GetReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, len(resp.Reads))
		}
	}
}

func TestAddReads(t *testing.T) {
	// POST, /reads/add
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	h.Setup()
	// Test no key
	t.Log("Testing no key given.")
	body, err := json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired key
	t.Log("Testing expired key.")
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["expired"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid key
	t.Log("Testing invalid key.")
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer not-a-valid-key")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid authorization header
	t.Log("Testing invalid authorization header.")
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "not-a-valid-auth-header")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test wrong content type
	t.Log("Testing wrong content type.")
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader("////"))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test read key
	t.Log("Testing read key.")
	body, err = json.Marshal(types.GetReadsRequest{
		ReaderName: "reader1",
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test valid
	t.Log("Testing valid request.")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "manual",
				Identifier:   "102",
				IdentType:    "bib",
				Milliseconds: 0,
				Seconds:      405,
			},
			{
				Type:         "manual",
				Identifier:   "104",
				IdentType:    "bib",
				Milliseconds: 0,
				Seconds:      465,
			},
			{
				Type:         "manual",
				Identifier:   "108",
				IdentType:    "bib",
				Milliseconds: 0,
				Seconds:      415,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 3, resp.Count)
		}
	}
	// Test validation -- IdentType
	t.Log("Testing validation -- IdentType")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "manual",
				Identifier:   "102",
				IdentType:    "wrong",
				Milliseconds: 0,
				Seconds:      405,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
		}
	}
	// Test validation -- Type
	t.Log("Testing validation -- Type")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "wrong",
				Identifier:   "102",
				IdentType:    "chip",
				Milliseconds: 0,
				Seconds:      405,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
		}
	}
	// Test validation -- Seconds
	t.Log("Testing validation -- Seconds")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "manual",
				Identifier:   "102",
				IdentType:    "bib",
				Milliseconds: 0,
				Seconds:      -405,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
		}
	}
	// Test validation -- Milliseconds
	t.Log("Testing validation -- Milliseconds")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "manual",
				Identifier:   "102",
				IdentType:    "bib",
				Milliseconds: -100,
				Seconds:      405,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
		}
	}
	// Test validation -- Identifier
	t.Log("Testing validation -- Identifier")
	body, err = json.Marshal(types.UploadReadsRequest{
		Reads: []types.Read{
			{
				Type:         "manual",
				Identifier:   "",
				IdentType:    "wrong",
				Milliseconds: 0,
				Seconds:      405,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodPost, "/reads/add", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.AddReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, 0, resp.Count)
		}
	}
}

func TestDeleteReads(t *testing.T) {
	// DELETE, /reads/delete
	variables, finalize := setupTests(t)
	defer finalize(t)
	e := echo.New()
	h := Handler{}
	// Test no key
	t.Log("Testing no key given.")
	start := int64(0)
	end := int64(70)
	body, err := json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader1",
		Start:      &start,
		End:        &end,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	c := e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test expired key
	t.Log("Testing expired key.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["expired"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid key
	t.Log("Testing invalid key.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer not-a-valid-key")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test invalid authorization header
	t.Log("Testing invalid authorization header.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "not-a-valid-auth-header")
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test wrong content type
	t.Log("Testing wrong content type.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test bad request
	t.Log("Testing bad request.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader("////"))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	// Test read key
	t.Log("Testing read key.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["read"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test write key
	t.Log("Testing write key.")
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["write"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	}
	// Test account mis-match.
	t.Log("Test account mis-match.")
	start = 0
	end = 70
	body, err = json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader5",
		Start:      &start,
		End:        &end,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, int64(0), resp.Count)
		}
		nr, err := database.GetReads(variables.accounts[1].Identifier, "reader5", 0, 10000)
		if assert.NoError(t, err) {
			assert.Equal(t, 300, len(nr))
		}
	}
	// Test valid
	t.Log("Test valid request.")
	body, err = json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader1",
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		var resp types.UploadReadsResponse
		if assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &resp)) {
			assert.Equal(t, int64(300), resp.Count)
		}
		nr, err := database.GetReads(variables.accounts[0].Identifier, "reader1", 0, 10000)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, len(nr))
		}
	}
	// Test start/end values
	t.Log("Test start/end values.")
	start = 0
	end = 75
	body, err = json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader2",
		Start:      &start,
		End:        &end,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		nr, err := database.GetReads(variables.accounts[0].Identifier, "reader2", 0, 10000)
		if assert.NoError(t, err) {
			for _, r := range nr {
				assert.True(t, r.Seconds < start || r.Seconds > end)
			}
		}
	}
	start = 1000
	end = 1175
	body, err = json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader2",
		Start:      &start,
		End:        &end,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusOK, response.Code)
		nr, err := database.GetReads(variables.accounts[0].Identifier, "reader2", 0, 10000)
		if assert.NoError(t, err) {
			for _, r := range nr {
				assert.True(t, r.Seconds < start || r.Seconds > end)
			}
		}
	}
	t.Log("Testing invalid end time.")
	start = 1000
	end = 175
	body, err = json.Marshal(types.DeleteReadsRequest{
		ReaderName: "reader2",
		Start:      &start,
		End:        &end,
	})
	if err != nil {
		t.Fatalf("Error encoding request body into json object: %v", err)
	}
	request = httptest.NewRequest(http.MethodDelete, "/reads/delete", strings.NewReader(string(body)))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+variables.knownValues["delete3"])
	response = httptest.NewRecorder()
	c = e.NewContext(request, response)
	if assert.NoError(t, h.DeleteReads(c)) {
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	}
}
