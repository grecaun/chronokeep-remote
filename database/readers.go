package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// GetReaders gets a list of reader names associated with a user
func GetReaders(user string) ([]string, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	reads := make([]types.Read, 0)
	if err = db.Select("DISTINCT(readerID)").Find(&reads).Error; err != nil {
		return nil, fmt.Errorf("unable to get reader names: %v", err)
	}
	output := make([]string, 0)
	for _, read := range reads {
		output = append(output, read.ReaderID)
	}
	return output, nil
}
