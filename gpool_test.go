package gpool

import (
	"log"
	"math"
	"testing"
	"time"
)

func testFunc() float64 {
	time.Sleep(10 * time.Millisecond)
	log.Println("hello")
	return math.Sin(3.14)
}

func BenchmarkPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < 10; n++ {
			Run(func() error {
				testFunc()
				return nil
			})
		}
	}
}

func TestPool(t *testing.T) {
	for n := 0; n < 10; n++ {
		Run(func() error {
			testFunc()
			return nil
		})
	}
}
