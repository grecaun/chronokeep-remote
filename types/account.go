package types

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Account is a structure holding information on accounts that have access
// to this module
type Account struct {
	Identifier        int64  `json:"-"`
	Password          string `json:"-"`
	Name              string `json:"name" validate:"required"`
	Email             string `json:"email" validate:"email,required"`
	Type              string `json:"type" validate:"required"`
	Locked            bool   `json:"locked"`
	WrongPassAttempts int    `json:"-"`
	Token             string `json:"-"`
	RefreshToken      string `json:"-"`
}

// Equals is used to check if the fields of an Account other than the identifier are identical.
func (a *Account) Equals(other *Account) bool {
	if other == nil {
		return false
	}
	return a.Name == other.Name &&
		a.Email == other.Email &&
		a.Type == other.Type
}

// PasswordIsHashed is used to check if a password has been hashed for insert into the database.
func (a *Account) PasswordIsHashed() bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(""))
	switch err.(type) {
	case bcrypt.InvalidHashPrefixError:
		return false
	case nil:
		// CompareHashAndPassword only returns nil if the password matches, empty string should NEVER be a correct password.
		return false
	default:
		// Check for hash too short (i.e. starts with $ but not long enough to be a hash value)
		if err.Error() == "crypto/bcrypt: hashedSecret too short to be a bcrypted password" {
			return false
		}
	}
	return true
}

// Validate Ensures that the struct is viable for entry into the database.
func (a *Account) Validate(validate *validator.Validate) error {
	valid := false
	switch a.Type {
	case "admin":
		valid = true
	case "free":
		valid = true
	case "paid":
		valid = true
	}
	if !valid {
		return errors.New("invalid account type specified")
	}
	return validate.Struct(a)
}
