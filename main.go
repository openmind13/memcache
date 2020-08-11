package main

import (
	"fmt"
	"log"
	"time"

	"github.com/openmind13/memcache/memcache"
)

func main() {
	fmt.Printf("memcache\n")

	cache := memcache.New(5*time.Minute, 10*time.Minute)
	cache.Add("mykey", "myvalue", 5*time.Minute)
	cache.Add("second", "second value", 5*time.Minute)
	cache.Add("third", "test test", time.Second)

	fmt.Println(cache.Exist("third"))
	fmt.Println(cache.Count())

	time.Sleep(6 * time.Second)
	fmt.Println(cache.Count())
	data, err := cache.Get("third")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(data)
}
