package handlers

import (
	"chronokeep/remote/types"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
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

func verifyToken(r *http.Request) (*types.Account, error) {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) != 2 {
		return nil, errors.New("unknown authorization header")
	}
	token, err := jwt.Parse(strArr[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("claims not set or token is not valid")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("email not found in token claims")
	}
	account, err := database.GetAccount(email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	if account.Token != strArr[1] || account.Token == "" {
		return nil, errors.New("token no longer valid")
	}
	return account, nil
}

func createTokens(email string) (*string, *string, error) {
	// Create token
	claims := jwt.MapClaims{}
	claims["email"] = email
	claims["authorized"] = true
	claims["exp"] = time.Now().Add(expirationWindow).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString([]byte(config.SecretKey))
	if err != nil {
		return nil, nil, err
	}
	// Create refresh token
	claims = jwt.MapClaims{}
	claims["email"] = email
	claims["exp"] = time.Now().Add(refreshWindow).Unix()
	r := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refresh, err := r.SignedString([]byte(config.RefreshKey))
	if err != nil {
		return nil, nil, err
	}
	return &token, &refresh, nil
}
