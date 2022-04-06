// Package mathutils holds utilities related with mathematics.
package mathutils

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomInt returns random int from provided range.
// "from" should be less or equal than "to".
func RandomInt(from, to int) (int, error) {
	if to < from {
		return 0, fmt.Errorf("could not generate random int because %d is less than %d", from, to)
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Intn(to-from+1) + from, nil
}

// RandomFloat64 returns random float from provided range.
func RandomFloat64(from, to float64) (float64, error) {
	if to < from {
		return 0, fmt.Errorf("could not generate random float because %f is less than %f", from, to)
	}

	return from + rand.Float64()*(to-from), nil
}
