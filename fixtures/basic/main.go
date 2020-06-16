package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	exitcode := flag.Int("exitcode", 0, "exit with this code")
	flag.Parse()

	fmt.Println("out1")
	fmt.Fprintln(os.Stderr, "err1")
	fmt.Println("out2")
	fmt.Fprintln(os.Stderr, "err2")
	fmt.Println("out3")
	fmt.Fprintln(os.Stderr, "err3")

	os.Exit(*exitcode)
}
