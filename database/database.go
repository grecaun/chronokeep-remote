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
	SetSetting(name, value string) error
	// Account Functions
	GetAccount(email string) (*types.Account, error)
	GetAccountByKey(key string) (*types.Account, error)
	GetAccountByID(id int64) (*types.Account, error)
	GetAccounts() ([]types.Account, error)
	AddAccount(account types.Account) (*types.Account, error)
	DeleteAccount(id int64) error
	ResurrectAccount(email string) error
	GetDeletedAccount(email string) (*types.Account, error)
	UpdateAccount(account types.Account) error
	ChangePassword(email, newPassword string, logout ...bool) error
	ChangeEmail(oldEmail, newEmail string) error
	InvalidPassword(account types.Account) error
	ValidPassword(account types.Account) error
	UnlockAccount(account types.Account) error
	UpdateTokens(account types.Account) error
	// Read Functions
	GetReads(account int64, name string, from, to int64) ([]types.Read, error)
	AddReads(key string, reads []types.Read) ([]types.Read, error)
	DeleteReads(account int64, name string, from, to int64) (int64, error)
	DeleteKeyReads(key string) (int64, error)
	// Key Functions
	GetAccountKeys(identifier string) ([]types.Key, error)
	GetKey(key string) (*types.Key, error)
	AddKey(key types.Key) (*types.Key, error)
	DeleteKey(key types.Key) error
	UpdateKey(key types.Key) error
	// Multi-get Functions
	GetKeyAndAccount(key string) (*types.MultiKey, error)
	// Close the database
	Close()
}
