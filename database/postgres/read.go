package postgres

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (p *Postgres) GetReads(account int64, reader_name string, from, to int64) ([]types.Read, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	toVal := to
	if to < from {
		toVal = from + 360
	}
	res, err := db.Query(
		ctx,
		"SELECT key_value, identifier, seconds, milliseconds, ident_type, type, antenna,"+
			" reader, rssi FROM read NATURAL JOIN api_key WHERE account_id=$1 AND "+
			"reader_name=$2 AND seconds>=$3 AND seconds<=$4;",
		account,
		reader_name,
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
			&read.Antenna,
			&read.Reader,
			&read.RSSI,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting read: %v", err)
		}
		outReads = append(outReads, read)
	}
	return outReads, nil
}

func (p *Postgres) AddReads(key string, reads []types.Read) ([]types.Read, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction to add reads: %v", err)
	}
	for _, read := range reads {
		_, err := tx.Exec(
			ctx,
			"INSERT INTO read("+
				"key_value, "+
				"identifier, "+
				"seconds, "+
				"milliseconds, "+
				"ident_type, "+
				"type, "+
				"antenna, "+
				"reader, "+
				"rssi"+
				") VALUES ("+
				"$1, "+
				"$2, "+
				"$3, "+
				"$4, "+
				"$5, "+
				"$6, "+
				"$7, "+
				"$8, "+
				"$9 "+
				") "+
				"ON CONFLICT(key_value, identifier, seconds, milliseconds, ident_type) DO NOTHING;",
			key,
			read.Identifier,
			read.Seconds,
			read.Milliseconds,
			read.IdentType,
			read.Type,
			read.Antenna,
			read.Reader,
			read.RSSI,
		)
		if err != nil {
			return nil, err
		}
	}
	return reads, tx.Commit(ctx)
}

func (p *Postgres) DeleteReads(account int64, reader_name string, from, to int64) (int64, error) {
	if to < from {
		return 0, errors.New("second input variable must be greater than first")
	}
	db, err := p.GetDB()
	if err != nil {
		return 0, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"DELETE FROM read r WHERE seconds>=$1 AND seconds<=$2 AND EXISTS (SELECT * "+
			"FROM api_key a WHERE a.key_value=r.key_value AND a.account_id=$3 AND "+
			"a.reader_name=$4);",
		from,
		to,
		account,
		reader_name,
	)
	if err != nil {
		return 0, fmt.Errorf("unable to delete reads: %v", err)
	}
	return res.RowsAffected(), nil
}

func (p *Postgres) DeleteKeyReads(key string) (int64, error) {
	db, err := p.GetDB()
	if err != nil {
		return 0, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"DELETE FROM read WHERE key_value=$1;",
		key,
	)
	if err != nil {
		return 0, fmt.Errorf("unable to delete reads: %v", err)
	}
	return res.RowsAffected(), nil
}

func (p *Postgres) DeleteReaderReads(account int64, reader_name string) (int64, error) {
	db, err := p.GetDB()
	if err != nil {
		return 0, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Exec(
		ctx,
		"DELETE FROM read r WHERE EXISTS (SELECT * FROM api_key a WHERE "+
			"a.key_value=r.key_value AND a.account_id=$1 AND a.reader_name=$2);",
		account,
		reader_name,
	)
	if err != nil {
		return 0, fmt.Errorf("unable to delete reads: %v", err)
	}
	return res.RowsAffected(), nil
}
