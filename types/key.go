package types

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	timeFormats = [...]string{
		"2006/01/02",
		"01/02/2006",
		"2006-01-02",
		"01-02-2006",
	}
)

// Key outline for data stored about an PI key
// Account should be a unique value for the account that owns the Key.
// Example types are: read (readonly), delete (read, write, delete), write (read, write)
// Allowed hosts are the hosts the calls are allowed to come from. Default of empty string is all hosts.
type Key struct {
	AccountIdentifier int64      `json:"account_id"`
	Name              string     `json:"name" validate:"required"`
	Value             string     `json:"value"`
	Type              string     `json:"type" validate:"required"`
	ValidUntil        *time.Time `json:"valid_until"`
}

type RequestKey struct {
	Name       string `json:"name" validate:"required"`
	Value      string `json:"value"`
	Type       string `json:"type" validate:"required"`
	ValidUntil string `json:"valid_until"`
}

func (k *Key) Equal(other *Key) bool {
	return k.AccountIdentifier == other.AccountIdentifier &&
		k.Name == other.Name &&
		k.Value == other.Value &&
		k.Type == other.Type &&
		// This next expression is TRUE if both are nil or both are not nil and they are equal.
		((k.ValidUntil != nil && other.ValidUntil != nil && k.ValidUntil.Equal(*other.ValidUntil)) || (k.ValidUntil == nil && other.ValidUntil == nil))
}

// Validate Ensures valid data in the structure.
func (k *Key) Validate(validate *validator.Validate) error {
	valid := false
	switch k.Type {
	case "default":
		k.Type = "write"
		valid = true
	case "read":
		valid = true
	case "write":
		valid = true
	case "delete":
		valid = true
	}
	if !valid {
		return errors.New("invalid key type specified")
	}
	// TODO: validation on the allowed hosts
	return validate.Struct(k)
}

// Validate Ensures valid data in the structure.
func (k RequestKey) Validate(validate *validator.Validate) error {
	valid := false
	switch k.Type {
	case "default":
		k.Type = "read"
		valid = true
	case "read":
		valid = true
	case "write":
		valid = true
	case "delete":
		valid = true
	}
	if !valid {
		return errors.New("invalid key type specified")
	}
	// TODO: validation on the allowed hosts
	return validate.Struct(k)
}

// Expired Reports whether the key has expired.
func (k Key) Expired() bool {
	if k.ValidUntil == nil {
		return false
	}
	return k.ValidUntil.Before(time.Now())
}

// ToKey Returns a Key struct with proper information.
func (k RequestKey) ToKey() Key {
	out := Key{
		Name:  strings.TrimSpace(k.Name),
		Value: k.Value,
		Type:  k.Type,
	}
	valid, err := time.Parse(time.RFC3339, k.ValidUntil)
	if err == nil {
		out.ValidUntil = &valid
		return out
	}
	for _, val := range timeFormats {
		valid, err = time.ParseInLocation(val, k.ValidUntil, time.Local)
		if err == nil {
			valid = valid.Add(time.Hour * 23).Add(time.Minute * 59).Add(time.Second * 59)
			out.ValidUntil = &valid
			return out
		}
	}
	return out
}

// GetValidUntil Returns a *time.Time value if the Request key has a valid time, or it returns nil.
func (k RequestKey) GetValidUntil() *time.Time {
	valid, err := time.Parse(time.RFC3339, k.ValidUntil)
	if err == nil {
		return &valid
	}
	for _, val := range timeFormats {
		valid, err = time.ParseInLocation(val, k.ValidUntil, time.Local)
		if err == nil {
			valid = valid.Add(time.Hour * 23).Add(time.Minute * 59).Add(time.Second * 59)
			return &valid
		}
	}
	return nil
}
