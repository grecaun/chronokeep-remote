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
	"net/http"

	"github.com/labstack/echo/v5"
)

// SaveHeartbeat stores the latest heartbeat sent by a Portal or Relay for the
// key used to send it, replacing the previous one.
func (h Handler) SaveHeartbeat(c *echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	// bind the request to validate it
	var request types.Heartbeat
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	// check if key exists
	mkey, err := database.GetKeyAndAccount(*k)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key/Account", err)
	}
	if mkey == nil || mkey.Key == nil || mkey.Account == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key/Account Not Found", nil)
	}
	// check for expired key
	if mkey.Key.Expired() {
		return getAPIError(c, http.StatusUnauthorized, "Expired Key", nil)
	}
	// Only write keys can send reader data; delete keys are for removing reads, not for readers.
	if msg, err := writeKeyError(mkey.Key); err != nil {
		return getAPIError(c, http.StatusUnauthorized, msg, err)
	}
	if err := request.Validate(h.validate); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Heartbeat", err)
	}
	if err := database.SaveHeartbeat(mkey.Key.Value, request); err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Saving Heartbeat", err)
	}
	return c.NoContent(http.StatusOK)
}

// GetHeartbeats returns the latest heartbeat for every reader on the key's account,
// or just the reader named in the request body.
func (h Handler) GetHeartbeats(c *echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	var request types.GetHeartbeatsRequest
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
	beats, err := database.GetHeartbeats(mkey.Account.Identifier, request.ReaderName)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Heartbeats", err)
	}
	return c.JSON(http.StatusOK, types.GetHeartbeatsResponse{
		Heartbeats: beats,
	})
}
