package tools

import (
	"math/rand"
	"time"
)

func RequestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return rand.Int()%2 == 0
}
