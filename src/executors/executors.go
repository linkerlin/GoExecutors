package executors

import (
	"config"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

type ErrorTimeout string

func (e ErrorTimeout) Error() string { return string(e) }

type Callable func() (interface{}, error) // result + error
type FutureQ chan *Future

type Executors struct {
	futureQ  FutureQ
	goNum    int32
	stopFlag bool
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
	var es = &Executors{fq, 0, false}
	var goMainFunc = func() {
		atomic.AddInt32(&es.goNum, 1)
		defer atomic.AddInt32(&es.goNum, -1)
		for es.stopFlag == false {
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
	}
	var i int32 = 0
	for ; i < config.DefaultGoroutinesNum(); i++ {
		go goMainFunc()
	}
	es.ControlGoNum(goMainFunc)

	return es

}

func (es *Executors) ControlGoNum(goMainFunc func()) {
	go func() {
		for es.stopFlag == false {
			if es.GetGoNum() < config.DefaultGoroutinesNum() || len(es.futureQ) > 10 {
				runtime.Gosched()
				if es.GetGoNum() < config.DefaultGoroutinesNum() || len(es.futureQ) > 10 {
					fmt.Println("GoNum:", es.GetGoNum(), "len(es.futureQ):", len(es.futureQ))
					go goMainFunc()
				}
			} else {
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()
}

func (es *Executors) GetGoNum() int32 {
	return atomic.LoadInt32(&es.goNum)
}

func (es *Executors) Submit(callable Callable) *Future {
	retChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)
	exceptionChan := make(chan interface{}, 1)
	future := &Future{retChan, errorChan, exceptionChan, callable}
	es.futureQ <- future
	return future
}
