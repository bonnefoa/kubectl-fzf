package util

import (
	"strconv"
	"time"
)

func ParseTimestamp(s string) (time.Time, error) {
	var tm time.Time
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return tm, err
	}
	tm = time.Unix(i, 0)
	return tm, nil
}
