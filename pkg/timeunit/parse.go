package timeunit

import (
	"fmt"
	"time"
)

func Parse(unit string) (time.Duration, error) {
	switch unit {
	case "second":
		return time.Second, nil
	case "minute":
		return time.Minute, nil
	case "hour":
		return time.Hour, nil
	case "day":
		return 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unit %s is not supported", unit)
	}
}
