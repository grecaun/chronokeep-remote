package types

// Read is a structure holding the information related to a *read*
// a read is either a chip read from a timing system or a manual entry from
// something like a mobile device
type Read struct {
	Key          string `json:"-"`
	Identifier   string `json:"identifier"`
	Seconds      int64  `json:"seconds"`
	Milliseconds int    `json:"milliseconds"`
	IdentType    string `json:"ident_type"`
	Type         string `json:"type"`
}
