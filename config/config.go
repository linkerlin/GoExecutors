package config

import (
	"fmt"
)

func DefaultGoroutinesNum() int32 {
	return 100
}

func LoadConfig() {
	fmt.Println("config loaded.")
}
