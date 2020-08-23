package types

import (
	"github.com/grecaun/apikeys/types"
)

// Account is a structure holding information on accounts that have access
// to this module
type Account struct {
	User           string      `json:"user" gorm:"user"`
	AllowedReaders int         `json:"allowedReaders" gorm:"allowedReaders"`
	Type           string      `json:"type" gorm:"type"`
	Keys           []types.Key `json:"keys" gorm:"keys"`
}
