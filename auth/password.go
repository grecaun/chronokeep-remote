package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword is used to encrypt the password before it is stored in the Database
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", fmt.Errorf("unable to hash password: %v", err)
	}
	return string(bytes), nil
}

// VerifyPassword checks the input password while verifying it with the database password.
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
