package database

import (
	"time"

	"chronokeep/remote/types"
	"chronokeep/remote/util"
)

const (
	MaxOpenConnections    = 20
	MaxIdleConnections    = 20
	MaxConnectionLifetime = time.Minute * 5
	CurrentVersion        = 1
)

type Database interface {
	// Database Base Functions
	Setup(config *util.Config) error
	SetSettings(name, value string) error
	// Read Functions
	GetReads(account, name string, from, to int64) ([]types.Read, error)
	AddReads(key string, reads []types.Read) error
	DeleteReads(account, name string, from, to int64) (int64, error)
	DeleteKeyReads(key string) (int64, error)
	// Key Functions
	GetAccountKeys(identifier string) ([]types.Key, error)
	GetKey(key string) (*types.Key, error)
	AddKey(key types.Key) (*types.Key, error)
	DeleteKey(key types.Key) error
	UpdateKey(key types.Key) error
	// Close the database
	Close()
}
