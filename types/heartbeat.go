/*
Chronokeep Remote - Remote Tag Read API
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
	"time"

	"github.com/go-playground/validator/v10"
)

// Heartbeat is the periodic status message sent by Chronokeep Portal and
// Chronokeep Relay. Battery and Charging are only sent by Portal.
type Heartbeat struct {
	Device           string `json:"dev" validate:"max=100"`
	Location         string `json:"location" validate:"max=100"`
	Now              int64  `json:"now"`                     // device time, Unix milliseconds
	Uptime           int64  `json:"uptime" validate:"min=0"` // seconds
	Battery          string `json:"battery" validate:"max=50"`
	Charging         bool   `json:"charging"`
	Hash             string `json:"hash" validate:"max=100"`
	UploadErrors     int64  `json:"upload_errors" validate:"min=0"`
	ReadersConnected int    `json:"readers_connected" validate:"min=0"`
	ReadersTotal     int    `json:"readers_total" validate:"min=0"`
	Reading          bool   `json:"reading"`
	ChipType         string `json:"chip_type" validate:"max=20"`
	// Temperature is the device's CPU temperature in degrees Celsius, or nil when it couldn't be read.
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,min=-100,max=200"`
}

// Validate Ensures valid data in the structure.
func (h *Heartbeat) Validate(validate *validator.Validate) error {
	return validate.Struct(h)
}

// ReaderHeartbeat is the most recent heartbeat received for a key (reader).
type ReaderHeartbeat struct {
	ReaderName string    `json:"reader"`
	Heartbeat  Heartbeat `json:"heartbeat"`
	Received   time.Time `json:"received"` // server time the heartbeat arrived (stored in Unix milliseconds)
}
