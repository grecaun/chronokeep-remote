package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// UpdateAccount updates an account
func UpdateAccount(account *types.Account) (*types.Account, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	if err = db.Where("user = ?", account.User).Update(account).Error; err != nil {
		return nil, fmt.Errorf("unable to update account: %v", err)
	}
	return account, nil
}
