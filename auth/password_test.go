/*
Chronokeep Remote - Remote Tag Read API
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

package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEqual(t, "correct horse", hash)
	assert.NoError(t, VerifyPassword(hash, "correct horse"))
	assert.Error(t, VerifyPassword(hash, "correct horse "), "trailing space")
	assert.Error(t, VerifyPassword(hash, "Correct horse"), "case")
	assert.Error(t, VerifyPassword(hash, ""), "empty")
	assert.Error(t, VerifyPassword("not-a-hash", "correct horse"), "bad hash")
	// Salted: the same password hashes differently each time.
	again, err := HashPassword("correct horse")
	if assert.NoError(t, err) {
		assert.NotEqual(t, hash, again)
		assert.NoError(t, VerifyPassword(again, "correct horse"))
	}
}

func TestHashPasswordLength(t *testing.T) {
	// bcrypt only takes 72 bytes; longer passwords are refused rather than silently truncated.
	_, err := HashPassword(strings.Repeat("a", 72))
	assert.NoError(t, err, "72 bytes")
	_, err = HashPassword(strings.Repeat("a", 73))
	assert.Error(t, err, "73 bytes")
}
