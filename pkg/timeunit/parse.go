package timeunit

import (
	"fmt"
)

type Unit int64

const (
	Second Unit = 1
	Minute      = 60 * Second
	Hour        = 60 * Minute
	Day         = 24 * Hour
)

func Parse(unit string) (Unit, error) {
	switch unit {
	case "second":
		return Second, nil
	case "minute":
		return Minute, nil
	case "hour":
		return Hour, nil
	case "day":
		return Day, nil
	default:
		var zero Unit
		return zero, fmt.Errorf("unit %s is not supported", unit)
	}
}
