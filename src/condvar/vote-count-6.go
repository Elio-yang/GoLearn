package main

import (
	"awesomeProject/tools"
	"time"
)
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())

	count := 0
	ch := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ch <- tools.RequestVote()
		}()
	}
	for i := 0; i < 10; i++ {
		v := <-ch
		if v {
			count += 1
		}
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}
