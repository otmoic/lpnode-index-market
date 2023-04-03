package main

import (
	"errors"
	"fmt"
	"log"
)

func te() error {
	err := "00000"
	return errors.New(fmt.Errorf("错误:%s", err).Error())
}
func main() {
	err := te()
	if err != nil {
		log.Println(err.Error())
	}
}
