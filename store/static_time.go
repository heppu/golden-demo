//go:build integration

package store

import "time"

func init() {
	now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
}
