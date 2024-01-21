package handlers

import (
	"chronokeep/remote/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetReaders(c echo.Context) error {
	// Get Key from Authorization Header
	k, err := retrieveKey(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Error Getting Key From Authorization Header", err)
	}
	if k == nil {
		return getAPIError(c, http.StatusUnauthorized, "Key Not Provided in Authorization Header", nil)
	}
	// Get account key is attached to
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
	keys, err := database.GetAccountKeys(mkey.Account.Email)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Reader Names", err)
	}
	readers := make([]types.Reader, 0)
	for _, k := range keys {
		if k.Type == "write" {
			readers = append(readers, types.Reader{
				Name: k.Name,
			})
		}
	}
	return c.JSON(http.StatusOK, types.GetReadersResponse{
		Readers: readers,
	})
}
