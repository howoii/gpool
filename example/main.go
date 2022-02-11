package main

import (
	"fmt"
	"math"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/howoii/gpool"
)

func demoFunc() {
	math.Sin(math.Pi / 4)
}

func main() {
	go func() {
		for {
			gpool.Run(demoFunc)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("%#v\n", gpool.Status())
			}
		}
	}()

	http.ListenAndServe("0.0.0.0:6060", nil)
}
