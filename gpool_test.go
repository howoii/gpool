package gpool

import (
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

var curMem uint64

const (
	RunTimes = 100000
	MiB      = 1 << 20
)

func demoFunc() {
	math.Sin(math.Pi / 4)
	time.Sleep(time.Duration(100) * time.Millisecond)
}

func BenchmarkGPool(b *testing.B) {
	var wg sync.WaitGroup

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for n := 0; n < RunTimes; n++ {
			Run(func() {
				defer wg.Done()
				demoFunc()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkGoroutine(b *testing.B) {
	var wg sync.WaitGroup

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for n := 0; n < RunTimes; n++ {
			go func() {
				defer wg.Done()
				demoFunc()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
}

func TestNoPool(t *testing.T) {
	t.Logf("Goroutines to execute: %d", RunTimes)

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc / MiB

	var wg sync.WaitGroup
	for i := 0; i < RunTimes; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestGPool(t *testing.T) {
	t.Logf("Goroutines to execute: %d", RunTimes)

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc / MiB

	var wg sync.WaitGroup
	for i := 0; i < RunTimes; i++ {
		wg.Add(1)
		Run(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("pool status: %#v\n", Status())

	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}
