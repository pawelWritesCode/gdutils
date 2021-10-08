package mathutils

import (
	"fmt"
	"math/rand"
	"time"
)

//RandomInt returns random int from provided range
//"from" should be less or equal than "to" otherwise func will panic
func RandomInt(from, to int) int {
	if to < from {
		panic(fmt.Sprintf("could not generate random int because %d is less than %d", from, to))
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Intn(to-from+1) + from
}
