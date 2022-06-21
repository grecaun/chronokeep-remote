package postgres

import (
	"chronokeep/remote/types"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

// GetKeyAndAccount Gets an account and key based upon the key value.
func (p *Postgres) GetKeyAndAccount(key string) (*types.MultiKey, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT "+
			"account_id, account_name, account_email, account_type, account_locked, "+
			"key_value, key_type, allowed_hosts, valid_until "+
			"FROM account NATURAL JOIN api_key WHERE account_deleted=FALSE AND key_deleted=FALSE AND key_value=$1",
		key,
	)
	if err != nil {
		res.Close()
		return nil, fmt.Errorf("error getting account and event from database: %v", err)
	}
	defer res.Close()
	if res.Next() {
		outVal := types.MultiKey{
			Key:     &types.Key{},
			Account: &types.Account{},
		}
		err := res.Scan(
			&outVal.Account.Identifier,
			&outVal.Account.Name,
			&outVal.Account.Email,
			&outVal.Account.Type,
			&outVal.Account.Locked,
			&outVal.Key.Value,
			&outVal.Key.Type,
			&outVal.Key.AllowedHosts,
			&outVal.Key.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting values for account and event: %v", err)
		}
		outVal.Key.AccountIdentifier = outVal.Account.Identifier
		return &outVal, nil
	}
	return nil, nil
}
