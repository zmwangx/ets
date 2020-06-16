package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(time.Second)
	fmt.Println("out1")
	time.Sleep(time.Second)
	fmt.Println("out2")
	time.Sleep(time.Second)
	fmt.Println("out3")
}
