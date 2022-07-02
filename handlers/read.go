package handlers

import (
	"chronokeep/remote/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetReads(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	var request types.GetReadsRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	mkey, err := database.GetKeyAndAccount(*k)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key/Account", err)
	}
	if mkey == nil || mkey.Key == nil || mkey.Account == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key/Account Not Found", nil)
	}
	// Check for expired key
	if mkey.Key.Expired() {
		return getAPIError(c, http.StatusUnauthorized, "Expired Key", nil)
	}
	reads, err := database.GetReads(mkey.Account.Identifier, request.ReaderName, request.Start, request.End)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Reads", err)
	}
	return c.JSON(http.StatusOK, types.GetReadsResponse{
		Count: len(reads),
		Reads: reads,
	})
}

func (h Handler) AddReads(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	// bind the request to validate it
	var request types.UploadReadsRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	// check if key exists
	mkey, err := database.GetKeyAndAccount(*k)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key/Account", err)
	}
	// check if we have return values for everything
	if mkey == nil || mkey.Key == nil || mkey.Account == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key/Account Not Found", nil)
	}
	// check for expired key
	if mkey.Key.Expired() {
		return getAPIError(c, http.StatusUnauthorized, "Expired Key", nil)
	}
	// update reads
	uploaded, err := database.AddReads(mkey.Key.Value, request.Reads)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Adding Keys to Database", err)
	}
	return c.JSON(http.StatusOK, types.UploadReadsResponse{
		Count: len(uploaded),
	})
}

func (h Handler) DeleteReads(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	// bind the request to validate it
	var request types.DeleteReadsRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	// check if key exists
	mkey, err := database.GetKeyAndAccount(*k)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key/Account", err)
	}
	//delKey, err := database.GetKey(request.ReaderName)
	// check if we have return values for everything
	if mkey == nil || mkey.Key == nil || mkey.Account == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key/Account Not Found", nil)
	}
	// check for expired key
	if mkey.Key.Expired() {
		return getAPIError(c, http.StatusUnauthorized, "Expired Key", nil)
	}
	// delete reads
	if request.Start != nil && request.End != nil {
		database.DeleteReads(mkey.Account.Identifier, request.ReaderName, *request.Start, *request.End)
	} else {
		database.DeleteReaderReads(mkey.Account.Identifier, request.ReaderName)
	}
	return c.NoContent(http.StatusNotImplemented)
}
