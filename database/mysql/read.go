package mysql

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (m *MySQL) GetReads(account, name string, from, to int64) ([]types.Read, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	toVal := to
	if to < from {
		toVal = from + 360
	}
	res, err := db.QueryContext(
		ctx,
		"SELECT key_value, identifier, seconds, milliseconds, ident_type, type FROM a_read NATURAL JOIN api_key WHERE account_id=? AND reader_name=? AND seconds>=? AND seconds<=?;",
		account,
		name,
		from,
		toVal,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving reads: %v", err)
	}
	defer res.Close()
	var outReads []types.Read
	for res.Next() {
		var read types.Read
		err := res.Scan(
			&read.Key,
			&read.Identifier,
			&read.Seconds,
			&read.Milliseconds,
			&read.IdentType,
			&read.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting read: %v", err)
		}
		outReads = append(outReads, read)
	}
	return outReads, nil
}

func (m *MySQL) AddReads(key string, reads []types.Read) ([]types.Read, error) {
	db, err := m.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	stmt, err := db.PrepareContext(
		ctx,
		"INSERT IGNORE INTO a_read("+
			"key_value, "+
			"identifier, "+
			"seconds, "+
			"milliseconds, "+
			"ident_type, "+
			"type) VALUES (?, ?, ?, ?, ?, ?);",
	)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare statement for read add: %v", err)
	}
	defer stmt.Close()
	var outReads []types.Read
	for _, read := range reads {
		_, err := stmt.ExecContext(
			ctx,
			key,
			read.Identifier,
			read.Seconds,
			read.Milliseconds,
			read.IdentType,
			read.Type,
		)
		if err != nil {
			return outReads, fmt.Errorf("error adding reads to database: %v", err)
		}
		outReads = append(outReads, read)
	}
	return reads, nil
}

func (m *MySQL) DeleteReads(account, name string, from, to int64) (int64, error) {
	if to < from {
		return 0, errors.New("second input variable must be greater than first")
	}
	db, err := m.GetDB()
	if err != nil {
		return 0, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"DELETE r FROM a_read r WHERE r.seconds>=? AND r.seconds<=? AND EXISTS (SELECT * FROM api_key a WHERE r.key_value=a.key_value AND a.account_id=? AND a.reader_name=?);",
		from,
		to,
		account,
		name,
	)
	if err != nil {
		return 0, fmt.Errorf("unable to delete reads: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("unable to determine rows affected by delete: %v", err)
	}
	return rows, nil
}

func (m *MySQL) DeleteKeyReads(key string) (int64, error) {
	db, err := m.GetDB()
	if err != nil {
		return 0, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"DELETE FROM a_read WHERE key_value=?;",
		key,
	)
	if err != nil {
		return 0, fmt.Errorf("unable to delete reads: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("unable to determine rows affected by delete: %v", err)
	}
	return rows, nil
}