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
	var future = es.Submit(func() interface{} {
		fmt.Println("这是从一个Callable内部发出的声音。")
		//		time.Sleep(time.Second * 1)
		return 1
	})
	var ret = future.GetResult(time.Millisecond * 1500)
	switch ret {
	case nil:
		fmt.Println("超时！")
	default:
		fmt.Println("执行成功", ret)
	}
	time.Sleep(time.Second * 3)
}
