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
	CurrentVersion        = 2
	MaxLoginAttempts      = 4
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
	GetReads(account int64, reader_name string, from, to int64) ([]types.Read, error)
	AddReads(key string, reads []types.Read) ([]types.Read, error)
	DeleteReaderReads(account int64, reader_name string, from, to int64) (int64, error)
	DeleteKeyReads(key string) (int64, error)
	DeleteReaderReadsBefore(account int64, reader_name string, to int64) (int64, error)
	DeleteReaderReadsBetween(account int64, reader_name string) (int64, error)
	// Key Functions
	GetAccountKeys(email string) ([]types.Key, error)
	GetAccountKeysByKey(key string) ([]types.Key, error)
	GetKey(key string) (*types.Key, error)
	AddKey(key types.Key) (*types.Key, error)
	DeleteKey(key types.Key) error
	UpdateKey(key types.Key) error
	// Multi-get Functions
	GetKeyAndAccount(key string) (*types.MultiKey, error)
	// Notification settings
	GetNotification(account int64, reader_name string) (*types.Notification, error)
	SaveNotification(notificaiton *types.RequestNotification, key string) error
	// Close the database
	Close()
}

