package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// Migrate automatically creates/updates tables for our information.
func Migrate() error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	err = db.AutoMigrate(&types.Account{}).Error
	if err != nil {
		return fmt.Errorf("unable to migrate account: %v", err)
	}
	err = db.AutoMigrate(&types.Read{}).Error
	if err != nil {
		return fmt.Errorf("unable to migrate read: %v", err)
	}
	return nil
}
