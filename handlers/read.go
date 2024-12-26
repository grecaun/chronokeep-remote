package handlers

import (
	"chronokeep/remote/types"
	"errors"
	"fmt"
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
	note, err := database.GetNotification(mkey.Account.Identifier, request.ReaderName)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Notification", err)
	}
	return c.JSON(http.StatusOK, types.GetReadsResponse{
		Count: int64(len(reads)),
		Reads: reads,
		Note:  note,
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
	// Check to ensure a write/delete key
	if mkey.Key.Type != "write" && mkey.Key.Type != "delete" {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("read key attempting to write"))
	}
	// validate read data
	upload := make([]types.Read, 0)
	for _, r := range request.Reads {
		if err := r.Validate(h.validate); err == nil {
			upload = append(upload, r)
		}
	}
	// update reads
	uploaded, err := database.AddReads(mkey.Key.Value, upload)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Adding Keys to Database", err)
	}
	return c.JSON(http.StatusOK, types.UploadReadsResponse{
		Count: int64(len(uploaded)),
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
	if mkey.Key.Type != "delete" {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("attempt to delete with read/write key"))
	}
	// delete reads
	var count int64
	if request.Start != nil && request.End != nil {
		count, err = database.DeleteReaderReads(mkey.Account.Identifier, request.ReaderName, *request.Start, *request.End)
	} else if request.End != nil {
		count, err = database.DeleteReaderReadsBefore(mkey.Account.Identifier, request.ReaderName, *request.End)
	} else {
		count, err = database.DeleteReaderReadsBetween(mkey.Account.Identifier, request.ReaderName)
	}
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Deleting Reads", fmt.Errorf("delete returned error: %v", err))
	}
	return c.JSON(http.StatusOK, types.UploadReadsResponse{
		Count: count,
	})
}
