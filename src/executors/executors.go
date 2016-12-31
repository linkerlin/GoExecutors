package executors

import (
	"config"
	"fmt"
	"time"
)

type ErrorTimeout string

func (e ErrorTimeout) Error() string { return string(e) }

type Callable func() interface{}
type CallableQ chan func() (chan interface{}, Callable)

type Executors struct {
	callableQ CallableQ
}

type Future struct {
	retChan chan interface{}
}

func NewFuture(retChan chan interface{}) *Future {
	return &Future{retChan}
}

func (f *Future) GetResult(timeout time.Duration) (interface{}, error) {
	timer := time.NewTimer(timeout)
	var ret interface{}
	fmt.Println("future对应的结果chan:", f.retChan)
	select {
	case ret = <-f.retChan:
		fmt.Println("future 获取到了结果：", ret)
		return ret, nil
	case <-timer.C:
		return nil, ErrorTimeout("Callable执行超时错误！")
	}

}

func NewExecutors() *Executors {
	var cq = make(CallableQ, 100)
	var es = &Executors{cq}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			for {
				pairFunc := <-cq
				retChan, callable := pairFunc()
				ret := callable()
				fmt.Println("result chan:", retChan)
				if retChan != nil {
					retChan <- ret
				}
				fmt.Println(".")
			}
		}()
	}

	return es

}

func (es *Executors) Submit(callable Callable) *Future {
	retChan := make(chan interface{}, 1)
	c := func() (chan interface{}, Callable) {
		return retChan, callable
	}
	es.callableQ <- c
	return NewFuture(retChan)
}
