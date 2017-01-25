package executors

import (
	"config"
	"fmt"
	"sync/atomic"
	"time"
)

type ErrorTimeout string

func (e ErrorTimeout) Error() string { return string(e) }

type Callable func() (interface{}, error) // result + error
type FutureQ chan *Future

type Executors struct {
	futureQ FutureQ
	goNum   int32
}

type Future struct {
	retChan       chan interface{}
	errorChan     chan error
	exceptionChan chan interface{}
	callable      Callable
}

func (f *Future) GetResult(timeout time.Duration) (ret interface{}, timeoutError error, err error, exception interface{}) {
	timer := time.NewTimer(timeout)
	// fmt.Println("future对应的结果chan:", f.retChan)
	select {
	case exception = <-f.exceptionChan:
		fmt.Println("future 获取到了异常：", err)
		return nil, nil, nil, exception
	case err = <-f.errorChan:
		fmt.Println("future 获取到了错误：", err)
		return nil, nil, err, nil
	case ret = <-f.retChan:
		fmt.Println("future 获取到了结果：", ret)
		return ret, nil, nil, nil
	case <-timer.C:
		return nil, ErrorTimeout("Callable执行超时错误！"), nil, nil
	}
}

func NewExecutors() *Executors {
	var fq = make(FutureQ, 100)
	var es = &Executors{fq, 0}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			atomic.AddInt32(&es.goNum, 1)
			defer atomic.AddInt32(&es.goNum, -1)
			for {
				var err interface{}
				future := <-fq
				defer func() {
					if err = recover(); err != nil {
						fmt.Errorf("捕获了一个错误:%v", err)
						future.exceptionChan <- err
					}
				}()
				ret, callableError := future.callable()
				// fmt.Println("result chan:", retChan)
				if callableError != nil {
					future.errorChan <- callableError
				} else if ret != nil {
					future.retChan <- ret
				}
				fmt.Println(".")
			}
		}()
	}

	return es

}

func (es *Executors) GetGoNum() int32 {
	return es.goNum
}

func (es *Executors) Submit(callable Callable) *Future {
	retChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	exceptionChan := make(chan interface{}, 1)
	future := &Future{retChan, errorChan, exceptionChan, callable}
	es.futureQ <- future
	return future
}
