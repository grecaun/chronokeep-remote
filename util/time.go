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

package util

import (
	"time"

	"github.com/pkg/errors"
)

func TimeSinceEpochSeconds(t time.Time) (int64, error) {
	epoch := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	seconds := t.Sub(epoch).Milliseconds() / 1000
	if seconds < 0 {
		return 0, errors.New("time value given before epoch")
	}
	return seconds, nil
}

