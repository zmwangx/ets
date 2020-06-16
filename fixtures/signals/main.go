package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exit := make(chan bool, 1)
	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT:
				fmt.Println("ignored SIGINT")
			case syscall.SIGTERM:
				fmt.Println("shutting down after receiving SIGTERM")
				exit <- true
				return
			}
		}
	}()
	done := false
	for !done {
		select {
		case <-exit:
			done = true
		case <-time.After(200 * time.Millisecond):
			fmt.Println("busy waiting")
		}
	}
}
