package sqlite

import (
	"chronokeep/remote/database"
	"chronokeep/remote/types"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *SQLite) getAccountInternal(email, key *string, id *int64) (*types.Account, error) {
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	var res *sql.Rows
	if email != nil {
		res, err = db.QueryContext(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, "+
				"account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE "+
				"AND account_email=?;",
			email,
		)
	} else if key != nil {
		res, err = db.QueryContext(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, "+
				"account_wrong_pass, account_token, account_refresh_token FROM account NATURAL JOIN api_key WHERE "+
				"account_deleted=FALSE AND key_deleted=FALSE AND key_value=?;",
			key,
		)
	} else if id != nil {
		res, err = db.QueryContext(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, "+
				"account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE "+
				"AND account_id=?;",
			id,
		)
	} else {
		return nil, errors.New("no valid identifying value provided to internal method")
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving account: %v", err)
	}
	defer res.Close()
	var outAccount types.Account
	if res.Next() {
		err := res.Scan(
			&outAccount.Identifier,
			&outAccount.Name,
			&outAccount.Email,
			&outAccount.Type,
			&outAccount.Password,
			&outAccount.Locked,
			&outAccount.WrongPassAttempts,
			&outAccount.Token,
			&outAccount.RefreshToken,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting account information: %v", err)
		}
	} else {
		return nil, nil
	}
	return &outAccount, nil
}

// GetAccount Gets an account based on the email address provided.
func (s *SQLite) GetAccount(email string) (*types.Account, error) {
	return s.getAccountInternal(&email, nil, nil)
}

// GetAccountByKey Gets an account based upon an API key provided.
func (s *SQLite) GetAccountByKey(key string) (*types.Account, error) {
	return s.getAccountInternal(nil, &key, nil)
}

// GetAccoutByID Gets an account based upon the Account ID.
func (s *SQLite) GetAccountByID(id int64) (*types.Account, error) {
	return s.getAccountInternal(nil, nil, &id)
}

// GetAccounts Get all accounts that have not been deleted.
func (s *SQLite) GetAccounts() ([]types.Account, error) {
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, "+
			"account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE;",
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving accounts: %v", err)
	}
	defer res.Close()
	var outAccounts []types.Account
	for res.Next() {
		var account types.Account
		err := res.Scan(
			&account.Identifier,
			&account.Name,
			&account.Email,
			&account.Type,
			&account.Password,
			&account.Locked,
			&account.WrongPassAttempts,
			&account.Token,
			&account.RefreshToken,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting account information: %v", err)
		}
		outAccounts = append(outAccounts, account)
	}
	return outAccounts, nil
}

// AddAccount Adds an account to the database.
func (s *SQLite) AddAccount(account types.Account) (*types.Account, error) {
	// Check if password has been hashed.
	if !account.PasswordIsHashed() {
		return nil, errors.New("password not hashed")
	}
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"INSERT INTO account(account_name, account_email, account_type, account_password) VALUES (?, ?, ?, ?)",
		account.Name,
		account.Email,
		account.Type,
		account.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to add account: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("unable to determine ID for account: %v", err)
	}
	return &types.Account{
		Identifier: id,
		Name:       account.Name,
		Email:      account.Email,
		Type:       account.Type,
	}, nil
}

// DeleteAccount Deletes an account from view, does not permanently delete from database.
// This does not delete events associated with this account, but does set keys to deleted.
func (s *SQLite) DeleteAccount(id int64) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %v", err)
	}
	res, err := tx.ExecContext(
		ctx,
		"UPDATE account SET account_deleted=TRUE WHERE account_id=?",
		id,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting account: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error checking rows affected on delete account: %v", err)
	}
	if rows != 1 {
		tx.Rollback()
		return fmt.Errorf("error deleting account, rows affected: %v", rows)
	}
	_, err = tx.ExecContext(
		ctx,
		"UPDATE api_key SET key_deleted=TRUE WHERE account_id=?",
		id,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting keys attached to account: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

// ResurrectAccount Brings an account out of the deleted state.
func (s *SQLite) ResurrectAccount(email string) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_deleted=FALSE WHERE account_email=?",
		email,
	)
	if err != nil {
		return fmt.Errorf("error resurrecting account: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on resurrect account: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error resurrecting account, rows affected: %v", rows)
	}
	return nil
}

// GetDeletedAccount Returns a deleted account.
func (s *SQLite) GetDeletedAccount(email string) (*types.Account, error) {
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT account_id, account_name, account_email, account_type FROM account WHERE account_deleted=TRUE AND account_email=?;",
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving account: %v", err)
	}
	defer res.Close()
	var outAccount types.Account
	if res.Next() {
		err := res.Scan(
			&outAccount.Identifier,
			&outAccount.Name,
			&outAccount.Email,
			&outAccount.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting account information: %v", err)
		}
	} else {
		return nil, nil
	}
	return &outAccount, nil
}

// UpdateAccount Updates account information in the database.
func (s *SQLite) UpdateAccount(account types.Account) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_name=?, account_type=? WHERE account_deleted=FALSE AND account_email=?",
		account.Name,
		account.Type,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("error updating account: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on update account: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error updating account, rows affected: %v", rows)
	}
	return nil
}

// ChangePassword Updates a user's password. It can also force a logout of the user. Only checks first value in the logout array if values are specified.
func (s *SQLite) ChangePassword(email, newPassword string, logout ...bool) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	stmt := "UPDATE account SET account_password=? WHERE account_email=?;"
	if len(logout) > 0 && logout[0] {
		stmt = "UPDATE account SET account_password=?, account_token='', account_refresh_token='' WHERE account_email=?;"
	}
	res, err := db.ExecContext(
		ctx,
		stmt,
		newPassword,
		email,
	)
	if err != nil {
		return fmt.Errorf("error changing password: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on password change: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error changing password, rows affected: %v", rows)
	}
	return nil
}

// UpdateTokens Updates a user's tokens.
func (s *SQLite) UpdateTokens(account types.Account) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_token=?, account_refresh_token=? WHERE account_email=?;",
		account.Token,
		account.RefreshToken,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("error updating tokens: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on token update: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error updating tokens, rows affected: %v", rows)
	}
	return nil
}

// ChangeEmail Updates an account email. Also forces a logout of the impacted account.
func (s *SQLite) ChangeEmail(oldEmail, newEmail string) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_email=?, account_token='', account_refresh_token='' WHERE account_email=?;",
		newEmail,
		oldEmail,
	)
	if err != nil {
		return fmt.Errorf("erorr updating account email: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on email change: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error changing email, rows affected: %v", rows)
	}
	return nil
}

// InvalidPassword Increments/locks an account due to an invalid password.
func (s *SQLite) InvalidPassword(account types.Account) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	pAcc, err := s.GetAccount(account.Email)
	if err != nil {
		return fmt.Errorf("error trying to retrieve account: %v", err)
	}
	locked := false
	if pAcc.WrongPassAttempts >= database.MaxLoginAttempts {
		locked = true
	}
	stmt := "UPDATE account SET account_locked=?, account_wrong_pass=account_wrong_pass + 1 WHERE account_email=?;"
	if locked {
		stmt = "UPDATE account SET account_locked=?, account_wrong_pass=account_wrong_pass + 1, account_token='', " +
			"account_refresh_token='' WHERE account_email=?;"
	}
	res, err := db.ExecContext(
		ctx,
		stmt,
		locked,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr updating invalid password information: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on invalid password information update: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error updating invalid password information, rows affected: %v", rows)
	}
	return nil
}

// ValidPassword Resets the incorrect password on an account.
func (s *SQLite) ValidPassword(account types.Account) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	acc, err := s.GetAccount(account.Email)
	if err != nil {
		return fmt.Errorf("error retrieving account to check locked status: %v", err)
	}
	if acc.Locked {
		return errors.New("account locked")
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_wrong_pass=0 WHERE account_email=?;",
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr updating valid password information: %v", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on valid password information update: %v", err)
	}
	return nil
}

// UnlockAccount Unlocks an account that's been locked.
func (s *SQLite) UnlockAccount(account types.Account) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	if !account.Locked {
		return errors.New("account not locked")
	}
	res, err := db.ExecContext(
		ctx,
		"UPDATE account SET account_wrong_pass=0, account_locked=FALSE WHERE account_email=?;",
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr unlocking account: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected on account unlock: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error unlocking account, rows affected: %v", rows)
	}
	return nil
}
