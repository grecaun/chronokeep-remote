package util

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

func GetTokenAndEmail(r *http.Request) (tokenStr, email *string, err error) {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) != 2 {
		return nil, nil, errors.New("unknown authorization header")
	}
	tokenStr = &strArr[1]
	token, err := jwt.Parse(strArr[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("claims not set or token is not valid")
	}
	*email, ok = claims["email"].(string)
	if !ok {
		return nil, nil, errors.New("email not found in token claims")
	}
	return tokenStr, email, nil
}
