package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/openmind13/memcache"
)

func main() {
	fmt.Printf("memcache\n")

	go printNumGorutines()

	cache := memcache.New(5*time.Minute, 10*time.Minute)

	cache.Add("test", "hello", 5*time.Second)
	cache.Add("first", "azazaza", 10*time.Minute)

	cache.Destroy()

	os.Exit(0)
}

func printNumGorutines() {
	for {
		time.Sleep(1 * time.Second)
		fmt.Println(runtime.NumGoroutine())
	}
}
