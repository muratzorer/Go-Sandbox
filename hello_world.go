package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello")

	pipe := make(chan string)

	go func() {
		fmt.Println(<-pipe)
	}()

	pipe <- "deneme"
	close(pipe)

	fmt.Scanln()
}
