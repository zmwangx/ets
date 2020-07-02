package main

import (
	"fmt"
	"log"
	"os"

	"github.com/creack/pty"
)

func main() {
	rows, cols, err := pty.Getsize(os.Stdin)
	if err != nil {
		log.Println("not a tty")
	} else {
		fmt.Printf("%dx%d\n", cols, rows)
	}
}
