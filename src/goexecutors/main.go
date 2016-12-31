package main

import (
	"config"
	"executors"
	"fmt"
	"time"
)

func main() {
	config.LoadConfig()
	fmt.Println("Default goroutines number is ", config.DefaultGoroutinesNum())
	es := executors.NewExecutors()
	es.Submit(func() interface{} {
		fmt.Println("这是从一个Callable内部发出的声音。")
		return 1
	})
	time.Sleep(10000)
}
