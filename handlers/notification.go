package handlers

import (
	"chronokeep/remote/types"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetNotifications(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	var request types.GetNotificationsRequest
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
	note, err := database.GetNotification(mkey.Account.Identifier, request.ReaderName)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Notification", err)
	}
	if note == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, types.GetNotificationsResponse{
		ReaderName: request.ReaderName,
		Note:       *note,
	})
}

func (h Handler) SaveNotification(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	// bind the request to validate it
	var request types.SaveNotificationRequest
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
	if err := request.Note.Validate(h.validate); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Notification", err)
	}
	return c.NoContent(http.StatusOK)
}
