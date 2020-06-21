package database

import (
	"fmt"
	"time"

	"github.com/grecaun/chronokeepremote/types"
)

var (
	// 1980-01-01T00:00:00Z (00:00:00 on January 1 1980)
	epoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
)

// GetReaderReads gets the reads from a specific user's reader
func GetReaderReads(reader, user string) ([]types.Read, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := make([]types.Read, 0)
	if err = db.Where("readerID = ? AND user = ?", reader, user).Find(&output).Error; err != nil {
		return nil, fmt.Errorf("unable to find reads: %v", err)
	}
	return output, nil
}

// GetReaderReadsTime gets all reads from a specific user's reader from a specified time until the other specified time
func GetReaderReadsTime(reader, user string, from, until time.Time) ([]types.Read, error) {
	// Convert time values into time.Duration since epoch, then convert to a seconds value
	fromInt := int(from.Sub(epoch) / time.Second)
	untilInt := int(from.Sub(epoch) / time.Second)
	return GetReaderReadsInt(reader, user, fromInt, untilInt)
}

// GetReaderReadsInt gets all reads from a specific user's reader from a specified time until the other specified time (time defined as seconds since Jan 1 1980)
func GetReaderReadsInt(reader, user string, from, until int) ([]types.Read, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := make([]types.Read, 0)
	if err = db.Where("readerID = ? AND user = ? AND seconds >= ? AND seconds <= ?", reader, user, from, until).Find(&output).Error; err != nil {
		return nil, fmt.Errorf("unable to find reads: %v", err)
	}
	return output, nil
}

// GetReaderReadsAfterTime gets all reads from a specified user's reader that occurred after a specified time
func GetReaderReadsAfterTime(reader, user string, from time.Time) ([]types.Read, error) {
	// Convert time values into time.Duration since epoch, then convert to a seconds value
	fromInt := int(from.Sub(epoch) / time.Second)
	return GetReaderReadsAfterInt(reader, user, fromInt)
}

// GetReaderReadsAfterInt gets all reads from a specified user's reader that occurred after a specified time (time defined as seconds since Jan 1 1980)
func GetReaderReadsAfterInt(reader, user string, from int) ([]types.Read, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := make([]types.Read, 0)
	if err = db.Where("readerID = ? AND user = ? AND seconds >= ?", reader, user, from).Find(&output).Error; err != nil {
		return nil, fmt.Errorf("unable to find reads: %v", err)
	}
	return output, nil
}

// GetUserReads gets all reads from a specified user
func GetUserReads(user string) ([]types.Read, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := make([]types.Read, 0)
	if err = db.Where("user = ?", user).Find(&output).Error; err != nil {
		return nil, fmt.Errorf("unabel to find reads: %v", err)
	}
	return output, nil
}
