/*
Chronokeep Desktop - Race Scoring Software
Copyright (C) 2026 James Sentinella

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package postgres

import (
	"chronokeep/remote/types"
	"context"
	"fmt"
	"time"
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
			"key_value, key_type, key_name, valid_until "+
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
			&outVal.Key.Name,
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

