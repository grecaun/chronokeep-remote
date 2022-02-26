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
