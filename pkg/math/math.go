package math

import (
	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}

	currMin := values[0]
	for _, v := range values {
		if currMin > v {
			currMin = v
		}
	}

	return currMin
}
