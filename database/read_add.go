package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"

	"github.com/jinzhu/gorm"
)

// AddRead adds a read to the database
func AddRead(read *types.Read) (*types.Read, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	if err = db.Create(read).Error; err != nil {
		return nil, fmt.Errorf("unable to add read: %v", err)
	}
	return read, nil
}

// AddReads adds multiple reads to the database
func AddReads(reads []types.Read) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("unable to establish database connection: %v", err)
	}
	count := 0
	db.Transaction(func(tx *gorm.DB) error {
		var e error
		for _, read := range reads {
			// check if it was successful and increment count if so
			if e = db.Create(&read).Error; e == nil {
				count++
			}
		}
		return nil
	})
	return count, nil
}
