package types

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Read is a structure holding the information related to a *read*
// a read is either a chip read from a timing system or a manual entry from
// something like a mobile device
type Read struct {
	Key          string `json:"-"`
	Identifier   string `json:"identifier" validate:"required"`
	Seconds      int64  `json:"seconds" validate:"gte=0"`
	Milliseconds int    `json:"milliseconds" validate:"gte=0"`
	IdentType    string `json:"ident_type"`
	Type         string `json:"type"`
}

// Validate Ensures valid data in the struct
func (r *Read) Validate(validate *validator.Validate) error {
	r.IdentType = strings.ToLower(r.IdentType)
	r.Type = strings.ToLower(r.Type)
	if r.IdentType != "chip" && r.IdentType != "bib" {
		return errors.New("invalid identifier type (bib/chip)")
	}
	if r.Type != "manual" && r.Type != "reader" {
		return errors.New("invalid read type (reader/manual)")
	}
	return validate.Struct(r)
}
