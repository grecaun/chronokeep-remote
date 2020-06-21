package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// GetAccount gets a specific account related to the user string
func GetAccount(user string) (*types.Account, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := types.Account{}
	err = db.Where("user = ?", user).Find(&output).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find account: %v", err)
	}
	return &output, nil
}

// GetAllAccounts gets every account in the database
func GetAllAccounts() ([]types.Account, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("unable to establish database connection: %v", err)
	}
	output := make([]types.Account, 0)
	err = db.Find(&output).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get accounts from database: %v", err)
	}
	return output, nil
}
