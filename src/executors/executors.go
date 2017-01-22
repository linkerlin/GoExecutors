package executors

import (
	"config"
	"fmt"
	"time"
)

type ErrorFuture interface{}

type ErrorTimeout string

func (e ErrorTimeout) Error() string { return string(e) }

type Callable func() interface{}
type CallableQ chan func() (chan interface{}, chan interface{}, Callable)

type Executors struct {
	callableQ CallableQ
}

type Future struct {
	retChan   chan interface{}
	errorChan chan interface{}
}

func NewFuture(retChan, errorChan chan interface{}) *Future {
	return &Future{retChan, errorChan}
}

func (f *Future) GetResult(timeout time.Duration) (ret interface{}, timeoutError error, err ErrorFuture) {
	timer := time.NewTimer(timeout)
	// fmt.Println("future对应的结果chan:", f.retChan)
	select {
	case ret = <-f.retChan:
		fmt.Println("future 获取到了结果：", ret)
		return ret, nil, nil
	case err = <-f.errorChan:
		fmt.Println("future 获取到了错误：", err)
		return nil, nil, err
	case <-timer.C:
		return nil, ErrorTimeout("Callable执行超时错误！"), nil
	}

}

func NewExecutors() *Executors {
	var cq = make(CallableQ, 100)
	var es = &Executors{cq}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			for {
				var err interface{}
				cf := <-cq
				retChan, errorChan, callable := cf()
				defer func() {
					if err = recover(); err != nil {
						fmt.Errorf("捕获了一个错误:%v", err)
						errorChan <- err
					}
				}()
				ret := callable()
				// fmt.Println("result chan:", retChan)
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
	errorChan := make(chan interface{}, 1)
	c := func() (chan interface{}, chan interface{}, Callable) {
		return retChan, errorChan, callable
	}
	es.callableQ <- c
	return NewFuture(retChan, errorChan)
}
