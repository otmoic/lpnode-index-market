package main

import (
	"log"
	"time"
)

func main() {
	for {
		time.Sleep(time.Second * 1)
		log.Println("0")
	}
}
