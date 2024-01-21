package handlers

import (
	"chronokeep/remote/types"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h Handler) GetKeys(c echo.Context) error {
	var request types.GetKeysRequest
	err := c.Bind(&request)
	if err != nil {
		return getAPIError(c, http.StatusBadRequest, "Bad Request", err)
	}
	account, err := verifyToken(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized Token", err)
	}
	if account.Locked {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("account locked"))
	}
	if account.Type != "admin" && request.Email != nil && account.Email != *request.Email {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("not admin / ownership error"))
	}
	email := account.Email
	if request.Email != nil {
		keyAccount, err := database.GetAccount(*request.Email)
		if err != nil {
			return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key Account", err)
		}
		if keyAccount == nil {
			return getAPIError(c, http.StatusNotFound, "Account Not Found", nil)
		}
		email = keyAccount.Email
	}
	keys, err := database.GetAccountKeys(email)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Keys", err)
	}
	return c.JSON(http.StatusOK, types.GetKeysResponse{
		Keys: keys,
	})
}

func (h Handler) AddKey(c echo.Context) error {
	// Adding new Key; Binding request body to key struct.
	var request types.AddKeyRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	// Verifying token.
	account, err := verifyToken(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized Token", err)
	}
	if account.Locked {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("account locked"))
	}
	// Validating key.
	if err := request.Key.Validate(h.validate); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Field(s)", err)
	}
	// Checking for admin or ownership.
	if account.Type != "admin" && request.Email != nil && account.Email != *request.Email {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("not admin / ownership error"))
	}
	// If email is set we add a key to that account, otherwise add it to the calling person's account.
	accountid := account.Identifier
	if request.Email != nil {
		// Getting key account holder for id value.
		keyAccount, err := database.GetAccount(*request.Email)
		if err != nil {
			return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key Account", err)
		}
		if keyAccount == nil {
			return getAPIError(c, http.StatusNotFound, "Account Not Found", nil)
		}
		accountid = keyAccount.Identifier
	}
	// Adding key to database.
	// Create new API Key for our key to add.
	newKey, err := uuid.NewRandom()
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Key Generation Error", err)
	}
	key, err := database.AddKey(types.Key{
		AccountIdentifier: accountid,
		Name:              strings.TrimSpace(request.Key.Name),
		Value:             newKey.String(),
		Type:              request.Key.Type,
		ValidUntil:        request.Key.GetValidUntil(),
	})
	if err != nil || key == nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Adding Key", err)
	}
	return c.JSON(http.StatusOK, types.ModifyKeyResponse{
		Key: *key,
	})
}

func (h Handler) DeleteKey(c echo.Context) error {
	var request types.DeleteKeyRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	account, err := verifyToken(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized Token", err)
	}
	if account.Locked {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("account locked"))
	}
	if len(request.Key) < 1 {
		return getAPIError(c, http.StatusBadRequest, "Bad Request", errors.New("no key specified"))
	}
	// Get Key to be deleted.
	multiKey, err := database.GetKeyAndAccount(request.Key)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key/Account", err)
	}
	if multiKey == nil || multiKey.Key == nil || multiKey.Account == nil {
		return getAPIError(c, http.StatusNotFound, "Key Not Found", nil)
	}
	// Deny access to non admins who do not own the key
	if account.Type != "admin" && account.Email != multiKey.Account.Email {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("not an admin / ownership error"))
	}
	err = database.DeleteKey(*multiKey.Key)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Deleting Key", err)
	}
	return c.NoContent(http.StatusOK)
}

func (h Handler) UpdateKey(c echo.Context) error {
	var request types.UpdateKeyRequest
	if err := c.Bind(&request); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Request Body", err)
	}
	account, err := verifyToken(c.Request())
	if err != nil {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized Token", err)
	}
	if account.Locked {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("account locked"))
	}
	if err := request.Key.Validate(h.validate); err != nil {
		return getAPIError(c, http.StatusBadRequest, "Invalid Field(s)", err)
	}
	// Get Account associated with this key
	keyAccount, err := database.GetAccountByKey(request.Key.Value)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Key Account", err)
	}
	if keyAccount == nil {
		return getAPIError(c, http.StatusNotFound, "Key Not Found", nil)
	}
	// Deny access to non admins who do not own the key
	if account.Type != "admin" && account.Email != keyAccount.Email {
		return getAPIError(c, http.StatusUnauthorized, "Unauthorized", errors.New("not an admin / ownership error"))
	}
	err = database.UpdateKey(request.Key.ToKey())
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Unable To Update Key", err)
	}
	key, err := database.GetKey(request.Key.Value)
	if err != nil {
		return getAPIError(c, http.StatusInternalServerError, "Error Retrieving Updated Key", err)
	}
	if key == nil {
		return getAPIError(c, http.StatusNotFound, "Key Not Found After Update", nil)
	}
	return c.JSON(http.StatusOK, types.ModifyKeyResponse{
		Key: *key,
	})
}
