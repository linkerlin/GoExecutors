package main

import (
	"config"
	"fmt"
)

func main() {
	config.LoadConfig()
	fmt.Println("Default goroutines number is ", config.DefaultGoroutinesNum())

}
