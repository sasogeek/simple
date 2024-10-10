package main

import (
	"fmt"
	"sync"
)

func executeTask(id interface{}) {
	fmt.Println("Goroutine", id, "is running")
}


func startGoroutines() {
	wg := sync.WaitGroup{}
	numGoroutines := 3
	wg.Add(numGoroutines)
	work := func(id int) {
		defer		wg.Done()
		executeTask(id)
	}


	for i := range numGoroutines {
		go		work(i)
	}
	wg.Wait()
	fmt.Println("All goroutines have finished")
}


func main() {
	startGoroutines()
}
