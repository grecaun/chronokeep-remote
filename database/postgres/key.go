package postgres

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (p *Postgres) GetAccountKeys(email string) ([]types.Key, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, valid_until FROM api_key NATURAL JOIN account WHERE key_deleted=FALSE AND account_email=$1;",
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving key: %v", err)
	}
	defer res.Close()
	var outKeys []types.Key
	for res.Next() {
		var key types.Key
		err := res.Scan(
			&key.AccountIdentifier,
			&key.Name,
			&key.Value,
			&key.Type,
			&key.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting key: %v", err)
		}
		outKeys = append(outKeys, key)
	}
	return outKeys, nil
}

func (p *Postgres) GetAccountKeysByKey(key string) ([]types.Key, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, valid_until FROM api_key a WHERE key_deleted=FALSE AND "+
			"EXISTS (SELECT * FROM api_key b WHERE a.account_id=b.account_id AND b.key_value=$1);",
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving key: %v", err)
	}
	defer res.Close()
	var outKeys []types.Key
	for res.Next() {
		var key types.Key
		err := res.Scan(
			&key.AccountIdentifier,
			&key.Name,
			&key.Value,
			&key.Type,
			&key.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting key: %v", err)
		}
		outKeys = append(outKeys, key)
	}
	return outKeys, nil
}

func (p *Postgres) GetKey(key string) (*types.Key, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, valid_until FROM api_key WHERE key_deleted=FALSE AND key_value=$1;",
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving key: %v", err)
	}
	defer res.Close()
	var outKey types.Key
	if res.Next() {
		err := res.Scan(
			&outKey.AccountIdentifier,
			&outKey.Name,
			&outKey.Value,
			&outKey.Type,
			&outKey.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting key: %v", err)
		}
	} else {
		return nil, nil
	}
	return &outKey, nil
}

func (p *Postgres) AddKey(key types.Key) (*types.Key, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"INSERT INTO api_key(account_id, key_name, key_value, key_type, valid_until) VALUES ($1, $2, $3, $4, $5);",
		key.AccountIdentifier,
		key.Name,
		key.Value,
		key.Type,
		key.ValidUntil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to add key: %v", err)
	}
	if res.RowsAffected() < 1 {
		return nil, errors.New("insert appears to be unsuccessful")
	}
	return &types.Key{
		AccountIdentifier: key.AccountIdentifier,
		Name:              key.Name,
		Value:             key.Value,
		Type:              key.Type,
		ValidUntil:        key.ValidUntil,
	}, nil
}

func (p *Postgres) DeleteKey(key types.Key) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE api_key SET key_deleted=TRUE WHERE key_deleted=FALSE AND key_value=$1;",
		key.Value,
	)
	if err != nil {
		return fmt.Errorf("error deleting key: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error deleting key, rows affected: %v", res.RowsAffected())
	}
	return nil
}

func (p *Postgres) UpdateKey(key types.Key) error {
	db, err := p.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"UPDATE api_key SET key_name=$1, key_type=$2, valid_until=$3 WHERE key_deleted=FALSE AND key_value=$4;",
		key.Name,
		key.Type,
		key.ValidUntil,
		key.Value,
	)
	if err != nil {
		return fmt.Errorf("error updating key: %v", err)
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("error updating key, rows affected: %v", res.RowsAffected())
	}
	return nil
}
