package main

import (
	"fmt"
	"gospider/internal"
)

func main() {
	fmt.Println("hello world")

	queue := internal.NewQueue()
	queue.Enqueue("https://aryanranderiya.com")
	queue.Enqueue("https://google.com")
	a, b := queue.Dequeue()

	fmt.Println(a, b)
	a, b = queue.Dequeue()
	fmt.Println(a, b)
}
