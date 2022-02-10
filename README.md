# gpool
a toy goroutine pool inspired by net/http and [ants](https://github.com/panjf2000/ants)

## Test
测试函数如下：
```go
func demoFunc() {
	math.Sin(math.Pi / 4)
	time.Sleep(time.Duration(10) * time.Millisecond)
}
```
测试1000000个goroutine并发：
```shell
$ go test -run=. -v
=== RUN   TestNoPool
    gpool_test.go:96: Goroutines to execute: 1000000
    gpool_test.go:114: memory usage:192 MB
--- PASS: TestNoPool (0.86s)
=== RUN   TestGPool
    gpool_test.go:118: Goroutines to execute: 1000000
    gpool_test.go:134: pool status: &gpool.PoolStatus{Running:1000, Idle:1000, Waiting:0, ReuseRate:0.86053}
    gpool_test.go:138: memory usage:60 MB
--- PASS: TestGPool (0.94s)
```
测试100000个goroutine并发：
```shell
$ go test -run=. -v
=== RUN   TestNoPool
    gpool_test.go:96: Goroutines to execute: 100000
    gpool_test.go:114: memory usage:46 MB
--- PASS: TestNoPool (0.25s)
=== RUN   TestGPool
    gpool_test.go:118: Goroutines to execute: 100000
    gpool_test.go:134: pool status: &gpool.PoolStatus{Running:1000, Idle:1000, Waiting:0, ReuseRate:0.30742}
    gpool_test.go:138: memory usage:24 MB
--- PASS: TestGPool (0.26s)
```
测试10000个goroutine并发：
```shell
$ go test -run=. -v
=== RUN   TestNoPool
    gpool_test.go:96: Goroutines to execute: 10000
    gpool_test.go:114: memory usage:5 MB
--- PASS: TestNoPool (0.12s)
=== RUN   TestGPool
    gpool_test.go:118: Goroutines to execute: 10000
    gpool_test.go:134: pool status: &gpool.PoolStatus{Running:1000, Idle:1000, Waiting:0, ReuseRate:0}
    gpool_test.go:138: memory usage:4 MB
--- PASS: TestGPool (0.12s)
```
内存分配优化效果明显，然后运行效率反而降低了，是我的实现方式有问题吗？

## Bench
```shell
goos: darwin
goarch: amd64
pkg: github.com/howoii/gpool
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkGPool-12                      5         239903121 ns/op        17704430 B/op     323375 allocs/op
BenchmarkGoroutine-12                  8         139307052 ns/op         9624650 B/op     200001 allocs/op
PASS
```
内存占用和运行效率都变差了。。我不理解