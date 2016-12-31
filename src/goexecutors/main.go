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
	f := func() interface{} {
		fmt.Println("这是从一个Callable内部发出的声音。")
		//		time.Sleep(time.Second * 1)
		return 1
	}

	var future = es.Submit(f)
	var ret, _ = future.GetResult(time.Millisecond * 1500)
	switch ret {
	case nil:
		fmt.Println("超时！")
	default:
		fmt.Println("执行成功", ret)
	}
	fTimeout := func() interface{} {
		time.Sleep(time.Second * 1)
		fmt.Println("这是第二次从Callable内部发出的声音。")
		return 2
	}
	fmt.Println("=================")
	time.Sleep(100)
	future = es.Submit(fTimeout)
	ret2, err := future.GetResult(time.Millisecond * 500)
	switch err {
	case nil:
		fmt.Println("执行成功", ret2)
	default:
		fmt.Println("超时！", err)
	}
	time.Sleep(time.Second * 6)
}
