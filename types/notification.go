/*
Chronokeep Desktop - Race Scoring Software
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

package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Notification struct {
	Identifier int64     `json:"id"`
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

