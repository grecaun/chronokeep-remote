package types

// Read is a structure holding the information related to a *read*
// a read is either a chip read from a timing system or a manual entry from
// something like a mobile device
type Read struct {
	ChipNumber   string `json:"chipNumber" gorm:"chipNumber"`
	Seconds      int    `json:"seconds" gorm:"seconds"`
	Milliseconds int    `json:"milliseconds" gorm:"milliseconds"`
	Antenna      int    `json:"antenna" gorm:"antenna"`
	Reader       string `json:"reader" gorm:"reader"`
	LogIndex     int    `json:"logIndex" gorm:"logIndex"`
	RSSI         int    `json:"rssi" gorm:"rssi"`
	IsRewind     int    `json:"rewind" gorm:"rewind"`
	Bib          string `json:"bib" gorm:"bib"`
	Type         string `json:"type" gorm:"type"`
	ReaderID     string `json:"readerID" gorm:"readerID"`
	User         string `json:"user" gorm:"user"`
}
