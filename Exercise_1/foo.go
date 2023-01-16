// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var i = 0

func incrementing() {
	//TODO: increment i 1000000 times
	for j := 0; j < 1000001; j++ {
		//lock
		i++
		//release
	}
}

func decrementing() {
	//TODO: decrement i 1000000 times
	for k := 0; k < 1000000; k++ {
		i--
	}
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(2)

	// TODO: Spawn both functions as goroutines
	go decrementing()
	go incrementing()


	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	time.Sleep(500 * time.Millisecond)
	Println("The magic number is:", i)
}
