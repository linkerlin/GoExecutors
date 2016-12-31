package executors

import (
	"config"
	"fmt"
	"sync"
	"time"
)

type Callable *func() interface{}
type CallableQ chan Callable

type Executors struct {
	callableQ CallableQ
	lock      *sync.Mutex
	rets      map[Callable]chan interface{}
}

type Future struct {
	es       *Executors
	callable Callable
}

func NewFuture(es *Executors, callable Callable) *Future {
	return &Future{es, callable}
}

func (f *Future) GetResult(timeout time.Duration) interface{} {
	fmt.Println("future get result.callable:", f.callable)
	timer := time.NewTimer(timeout)
	var ret interface{}

	defer func() {
		f.es.lock.Lock()
		delete(f.es.rets, f.callable)
		f.es.lock.Unlock()
	}()
	fmt.Println("future对应的结果chan:", f.es.rets[f.callable])
	select {
	case ret = <-f.es.rets[f.callable]:
		fmt.Println("future 获取到了结果：", ret)
		return ret
	case <-timer.C:
		return nil
	}

}

func NewExecutors() *Executors {
	var cq = make(CallableQ, 100)
	var es = &Executors{cq, &sync.Mutex{}, make(map[Callable]chan interface{})}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			var callable Callable
			for {
				callable = <-cq
				var ret = (*callable)()
				es.lock.Lock()
				var retChan = es.rets[callable]
				es.lock.Unlock()
				fmt.Println("callable:", callable, " ret:", ret, " to ", retChan)
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
	es.callableQ <- callable
	es.lock.Lock()
	es.rets[callable] = make(chan interface{}, 1) // 保证map里面有东西
	defer es.lock.Unlock()
	return NewFuture(es, callable)
}
