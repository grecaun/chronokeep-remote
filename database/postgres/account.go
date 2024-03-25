package postgres

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	MaxLoginAttempts = 4
)

func (p *Postgres) getAccountInternal(email, key *string, id *int64) (*types.Account, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	var res pgx.Rows
	if email != nil {
		res, err = db.Query(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE AND account_email=$1;",
			email,
		)
	} else if key != nil {
		res, err = db.Query(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, account_wrong_pass, account_token, account_refresh_token FROM account NATURAL JOIN api_key WHERE account_deleted=FALSE AND key_deleted=FALSE AND key_value=$1;",
			key,
		)
	} else if id != nil {
		res, err = db.Query(
			ctx,
			"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE AND account_id=$1;",
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
func (p *Postgres) GetAccount(email string) (*types.Account, error) {
	return p.getAccountInternal(&email, nil, nil)
}

// GetAccountByKey Gets an account based upon an API key provided.
func (p *Postgres) GetAccountByKey(key string) (*types.Account, error) {
	return p.getAccountInternal(nil, &key, nil)
}

// GetAccoutByID Gets an account based upon the Account ID.
func (p *Postgres) GetAccountByID(id int64) (*types.Account, error) {
	return p.getAccountInternal(nil, nil, &id)
}

// GetAccounts Get all accounts that have not been deleted.
func (p *Postgres) GetAccounts() ([]types.Account, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT account_id, account_name, account_email, account_type, account_password, account_locked, account_wrong_pass, account_token, account_refresh_token FROM account WHERE account_deleted=FALSE;",
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
func (p *Postgres) AddAccount(account types.Account) (*types.Account, error) {
	// Check if password has been hashed.
	if !account.PasswordIsHashed() {
		return nil, errors.New("password not hashed")
	}
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	var id int64
	err = db.QueryRow(
		ctx,
		"INSERT INTO account(account_name, account_email, account_type, account_password) VALUES ($1, $2, $3, $4) RETURNING (account_id);",
		account.Name,
		account.Email,
		account.Type,
		account.Password,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("unable to add account: %v", err)
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
func (p *Postgres) DeleteAccount(id int64) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_deleted=TRUE WHERE account_id=$1",
		id,
	)
	if err != nil {
		return fmt.Errorf("error deleting account: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error deleting account, rows affected: %v", res.RowsAffected())
	}
	_, err = db.Exec(
		ctx,
		"UPDATE api_key SET key_deleted=TRUE WHERE account_id=$1",
		id,
	)
	if err != nil {
		return fmt.Errorf("error deleting keys attached to account: %v", err)
	}
	return nil
}

// ResurrectAccount Brings an account out of the deleted state.
func (p *Postgres) ResurrectAccount(email string) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_deleted=FALSE WHERE account_email=$1",
		email,
	)
	if err != nil {
		return fmt.Errorf("error resurrecting account: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error resurrecting account, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// GetDeletedAccount Returns a deleted account.
func (p *Postgres) GetDeletedAccount(email string) (*types.Account, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT account_id, account_name, account_email, account_type FROM account WHERE account_deleted=TRUE AND account_email=$1;",
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
func (p *Postgres) UpdateAccount(account types.Account) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_name=$1, account_type=$2 WHERE account_deleted=FALSE AND account_email=$3",
		account.Name,
		account.Type,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("error updating account: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error updating account, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// ChangePassword Updates a user's password. It can also force a logout of the user. Only checks first value in the logout array if values are specified.
func (p *Postgres) ChangePassword(email, newPassword string, logout ...bool) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	stmt := "UPDATE account SET account_password=$1 WHERE account_email=$2;"
	if len(logout) > 0 && logout[0] {
		stmt = "UPDATE account SET account_password=$1, account_token='', account_refresh_token='' WHERE account_email=$2;"
	}
	res, err := db.Exec(
		ctx,
		stmt,
		newPassword,
		email,
	)
	if err != nil {
		return fmt.Errorf("error changing password: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error changing password, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// UpdateTokens Updates a user's tokens.
func (p *Postgres) UpdateTokens(account types.Account) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_token=$1, account_refresh_token=$2 WHERE account_email=$3;",
		account.Token,
		account.RefreshToken,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("error updating tokens: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error updating tokens, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// ChangeEmail Updates an account email. Also forces a logout of the impacted account.
func (p *Postgres) ChangeEmail(oldEmail, newEmail string) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_email=$1, account_token='', account_refresh_token='' WHERE account_email=$2;",
		newEmail,
		oldEmail,
	)
	if err != nil {
		return fmt.Errorf("erorr updating account email: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error changing email, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// InvalidPassword Increments/locks an account due to an invalid password.
func (p *Postgres) InvalidPassword(account types.Account) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	pAcc, err := p.GetAccount(account.Email)
	if err != nil {
		return fmt.Errorf("error trying to retrieve account: %v", err)
	}
	locked := false
	if pAcc.WrongPassAttempts >= MaxLoginAttempts {
		locked = true
	}
	stmt := "UPDATE account SET account_locked=$1, account_wrong_pass=account_wrong_pass + 1 WHERE account_email=$2;"
	if locked {
		stmt = "UPDATE account SET account_locked=$1, account_wrong_pass=account_wrong_pass + 1, account_token='', account_refresh_token='' WHERE account_email=$2;"
	}
	res, err := db.Exec(
		ctx,
		stmt,
		locked,
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr updating invalid password information: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error updating invalid password information, rows affected: %v", res.RowsAffected())
	}
	return nil
}

// ValidPassword Resets the incorrect password on an account.
func (p *Postgres) ValidPassword(account types.Account) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	acc, err := p.GetAccount(account.Email)
	if err != nil {
		return fmt.Errorf("error retrieving account to check locked status: %v", err)
	}
	if acc.Locked {
		return errors.New("account locked")
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	_, err = db.Exec(
		ctx,
		"UPDATE account SET account_wrong_pass=0 WHERE account_email=$1;",
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr updating valid password information: %v", err)
	}
	return nil
}

// UnlockAccount Unlocks an account that's been locked.
func (p *Postgres) UnlockAccount(account types.Account) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	if !account.Locked {
		return errors.New("account not locked")
	}
	res, err := db.Exec(
		ctx,
		"UPDATE account SET account_wrong_pass=0, account_locked=FALSE WHERE account_email=$1;",
		account.Email,
	)
	if err != nil {
		return fmt.Errorf("erorr unlocking account: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error unlocking account, rows affected: %v", res.RowsAffected())
	}
	return nil
}
