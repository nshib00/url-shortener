package random

import (
	"math/rand"
	"time"
)

var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewRandomString(aliasLength int) string {
	randSymbols := make([]byte, aliasLength)
	symbols := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

	for i := range aliasLength {
		randSymbols[i] = symbols[rng.Intn(len(symbols))]
	}
	return string(randSymbols)
}
