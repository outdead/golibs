package random

import (
	crand "crypto/rand"
	"math/big"
	mrand "math/rand"
)

func IntWeak(from, to int) int {
	return from + mrand.Intn(to-from+1) //nolint:gosec // I'm ok with unsecure here.
}

func IntStrong(from, to int) int {
	num, _ := crand.Int(crand.Reader, big.NewInt(int64(to-from+1)))

	return from + int(num.Int64())
}
