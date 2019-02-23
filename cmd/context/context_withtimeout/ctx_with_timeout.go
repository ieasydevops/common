package main

import (
	"math/rand"
	"time"
	"golang.org/x/net/context"
	"log"
)

func test(is_first *bool, prev_map *map[string]string) {
	rand.Seed(int64(time.Now().Nanosecond()))
	ch := make(chan bool)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	go func() {
		if *is_first {
			(*prev_map)["a"] = "1"
		} else {
			(*prev_map)["a"] = "2"
		}
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		ch <- true
	}()

	select {
	case <-ch:
		*is_first = false
		(*prev_map)["b"] = "fast"
		return
	case <-ctx.Done():
		*is_first = true
		(*prev_map)["b"] = "slow"
		return
	}
}

func main() {
	is_first := true
	prev_map := make(map[string]string)
	for {
		test(&is_first, &prev_map)
		time.Sleep(time.Second)
		log.Println(prev_map, is_first)
	}
}
