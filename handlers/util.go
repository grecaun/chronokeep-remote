package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// APIError holds information on an error from the API
type APIError struct {
	Message string `json:"message,omitempty"`
}

func getAPIError(c echo.Context, code int, message string, err error) error {
	log.WithFields(log.Fields{
		"message": message,
		"error":   err,
		"code":    code,
	}).Error("API Error.")
	return c.JSON(code, APIError{Message: message})
}

func retrieveKey(r *http.Request) (*string, error) {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) != 2 {
		return nil, errors.New("unknown authorization header")
	}
	return &strArr[1], nil
}
