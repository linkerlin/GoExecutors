package config

import (
	"fmt"
)

func DefaultGoroutinesNum() int {
	return 100
}

func LoadConfig() {
	fmt.Println("config loaded.")
}
