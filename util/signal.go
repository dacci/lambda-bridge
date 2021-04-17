package util

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var Running = false

func Notify() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	Running = true
	s := <-c
	log.Println("caught signal:", s)
	Running = false
}
