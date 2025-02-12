package ops

import (
	"time"

	"math/rand"
)

const (
	prefix  = "goroot-"
	charset = "abcdefghijklmnopqrstuvwxyz"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func CreateName() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	name := prefix + string(b)

	return name
}
