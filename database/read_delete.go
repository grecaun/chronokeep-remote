package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// DeleteUserReads deletes all of the reads from a user
func DeleteUserReads(user string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	return db.Where("user = ?", user).Delete(&types.Read{}).Error
}

// DeleteReaderReads delets all of the reads from a specific reader
func DeleteReaderReads(reader, user string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	return db.Where("readerID = ? AND user = ?", reader, user).Delete(&types.Read{}).Error
}

// DeleteAllReads delets all reads from the database
func DeleteAllReads() error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	return db.Delete(&types.Read{}).Error
}
