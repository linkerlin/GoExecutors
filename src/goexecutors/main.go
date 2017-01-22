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
	var ret, t, e = future.GetResult(time.Millisecond * 1500)
	switch {
	case t == nil && e == nil:
		fmt.Println("正常", ret)
	case t != nil:
		fmt.Println("超时！")
	case e != nil:
		fmt.Println("出错", e)
	default:
		fmt.Println("不会到这里", ret)
	}
	fTimeout := func() interface{} {
		time.Sleep(time.Second * 1)
		fmt.Println("这是第二次从Callable内部发出的声音。")
		return 2
	}
	fmt.Println("=================")
	time.Sleep(100)
	future = es.Submit(fTimeout)
	ret2, t, err := future.GetResult(time.Millisecond * 500)
	switch {
	case t == nil && err == nil:
		fmt.Println("执行成功", ret2)
	case err != nil:
		fmt.Println("执行出错", err)
	case t != nil:
		fmt.Println("超时！", t)
	default:
		fmt.Println("不会到这里", ret2)
	}
	fPanic := func() interface{} {
		fmt.Println("这是第三次从Callable内部发出的声音。")
		panic(100)
	}
	future = es.Submit(fPanic)
	ret3, t, err := future.GetResult(time.Millisecond * 500)
	switch {
	case err == nil && t == nil:
		fmt.Println("执行失败,没有捕获到错误", ret3)
	case t != nil:
		fmt.Println("执行失败,超时", t)
	case err != nil:
		fmt.Println("执行成功,捕获到", err)
	default:
		fmt.Println("不会到这里", ret3)
	}
	time.Sleep(time.Second * 6)
}
