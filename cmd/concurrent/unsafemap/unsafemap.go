package main

import "fmt"

func main() {

	c := make(map[int]int, 10000)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				fmt.Print("dddddddd")
				c[j] = i
			}
		}()
	}

	select {}

}





