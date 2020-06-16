package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

func main() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("tty")
	} else {
		fmt.Println("pipe")
	}
}
