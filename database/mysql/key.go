package mysql

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (m *MySQL) GetAccountKeys(email string) ([]types.Key, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, reader_name, valid_until FROM api_key NATURAL JOIN account WHERE key_deleted=FALSE AND account_email=?;",
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
			&key.ReaderName,
			&key.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting key: %v", err)
		}
		outKeys = append(outKeys, key)
	}
	return outKeys, nil
}

func (m *MySQL) GetAccountKeysByKey(key string) ([]types.Key, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, reader_name, valid_until FROM api_key a WHERE key_deleted=FALSE AND "+
			"EXISTS (SELECT * FROM api_key b WHERE a.account_id=b.account_id AND b.key_value=?);",
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
			&key.ReaderName,
			&key.ValidUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting key: %v", err)
		}
		outKeys = append(outKeys, key)
	}
	return outKeys, nil
}

func (m *MySQL) GetKey(key string) (*types.Key, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT account_id, key_name, key_value, key_type, reader_name, valid_until FROM api_key WHERE key_deleted=FALSE AND key_value=?;",
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
			&outKey.ReaderName,
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

func (m *MySQL) AddKey(key types.Key) (*types.Key, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"INSERT INTO api_key(account_id, key_name, key_value, key_type, reader_name, valid_until) VALUES (?, ?, ?, ?, ?, ?);",
		key.AccountIdentifier,
		key.Name,
		key.Value,
		key.Type,
		key.ReaderName,
		key.ValidUntil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to add key: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error checking rows affected: %v", err)
	}
	if rows < 1 {
		return nil, errors.New("insert appears to be unsuccessful")
	}
	return &types.Key{
		AccountIdentifier: key.AccountIdentifier,
		Name:              key.Name,
		Value:             key.Value,
		Type:              key.Type,
		ReaderName:        key.ReaderName,
		ValidUntil:        key.ValidUntil,
	}, nil
}

func (m *MySQL) DeleteKey(key types.Key) error {
	db, err := m.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE api_key SET key_deleted=TRUE WHERE key_value=?;",
		key.Value,
	)
	if err != nil {
		return fmt.Errorf("error deleting key: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error deleting key, rows affected: %v", rows)
	}
	return nil
}

func (m *MySQL) UpdateKey(key types.Key) error {
	db, err := m.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"UPDATE api_key SET key_name=?, key_type=?, reader_name=?, valid_until=? WHERE key_deleted=FALSE AND key_value=?;",
		key.Name,
		key.Type,
		key.ReaderName,
		key.ValidUntil,
		key.Value,
	)
	if err != nil {
		return fmt.Errorf("error updating key: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("error updating key, rows affected: %v", rows)
	}
	return nil
}
