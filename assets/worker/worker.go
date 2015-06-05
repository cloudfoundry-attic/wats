package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Running Worker")
		time.Sleep(time.Second)
	}
}
