package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Notification struct {
	Identifier int64     `json:"-"`
	Type       string    `json:"type"`
	When       time.Time `json:"when"`
}

type RequestNotification struct {
	Type string `json:"type" validate:"required"`
	When string `json:"when" validate:"required"`
}

func (n *RequestNotification) Validate(validate *validator.Validate) error {
	valid := false
	switch n.Type {
	case "UPS_DISCONNECTED", "UPS_CONNECTED", "UPS_ON_BATTERY", "UPS_LOW_BATTERY", "UPS_ONLINE", "SHUTTING_DOWN", "RESTARTING", "HIGH_TEMP", "MAX_TEMP":
		valid = true
	}
	if !valid {
		return fmt.Errorf("%v is not a valid type", n.Type)
	}
	_, err := time.Parse(time.RFC3339, n.When)
	if err != nil {
		return fmt.Errorf("%v is not a valid date", n.When)
	}
	return validate.Struct(n)
}

func (n RequestNotification) ToNotification() (*Notification, error) {
	out := Notification{
		Type: n.Type,
	}
	valid, err := time.Parse(time.RFC3339, n.When)
	if err == nil {
		out.When = valid
		return &out, nil
	}
	return nil, errors.New("invalid time value")
}
