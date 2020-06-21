package database

import (
	"fmt"

	"github.com/grecaun/chronokeepremote/types"
)

// DeleteAccount deletes an account from the database
func DeleteAccount(account *types.Account) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	return db.Where("user = ?", account.User).Delete(&types.Account{}).Error
}

// DeleteAllAccounts deletes all accounts in the database
func DeleteAllAccounts() error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("unable to establish database connection: %v", err)
	}
	return db.Delete(&types.Account{}).Error
}
