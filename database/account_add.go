package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// AddAccount adds an account to the database
func AddAccount(account *types.Account) (*types.Account, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	if err := db.Create(account).Error; err != nil {
		return nil, fmt.Errorf("unable to add account: %v", err)
	}
	return account, nil
}
