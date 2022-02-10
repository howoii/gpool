package main

import (
	"math"
	"sync"
	"time"

	"github.com/howoii/gpool"
)

const (
	RunTimes = 100
)

func demoFunc() {
	math.Sin(math.Pi / 4)
	time.Sleep(time.Duration(10) * time.Millisecond)
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < RunTimes; i++ {
		wg.Add(1)
		gpool.Run(func() {
			defer wg.Done()
			demoFunc()
		})
	}
	wg.Wait()
}
