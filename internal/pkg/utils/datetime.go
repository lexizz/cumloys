package utils

import (
	"time"
)

func GetCurrentDatetimeUTC() time.Time {
	now := time.Now()

	return now.UTC()
}
