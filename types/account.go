package types

// Account is a structure holding information on accounts that have access
// to this module
type Account struct {
	User           string `json:"user" gorm:"user"`
	AllowedReaders int    `json:"allowedReaders" gorm:"allowedReaders"`
	Type           string `json:"type" gorm:"type"`
}
