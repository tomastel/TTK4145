// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"sync"
	"time"
)

var i = 0
var wg sync.WaitGroup

func number_server(increment, decrement, get chan int) {
	for {
		select {
		case <-increment:
			i++
			Println(i)
		case <-decrement:
			i--
			Println(i)

		case <-get:
			Println("i iz done")
			return
		}
	}
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(3)
	wg.Add(2)

	increment := make(chan int)
	decrement := make(chan int)
	get := make(chan int)

	// TODO: Spawn both functions as goroutines
	go func() {
		for j := 0; j < 110; j++ {
			increment <- 0
		}
		wg.Done()

	}()

	go func() {
		for j := 0; j < 100; j++ {
			decrement <- 0
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		get <- 0
	}()

	number_server(increment, decrement, get)

	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	time.Sleep(500 * time.Millisecond)
	Println("The magic number is:", i)
}
