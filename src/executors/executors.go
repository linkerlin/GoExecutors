package executors

import (
	"config"
	"sync"
)

type Callable func() interface{}
type CallableQ chan Callable

type Executors struct {
	callableQ CallableQ
	lock      *sync.Mutex
	rets      map[*Callable]interface{}
}

func NewExecutors() *Executors {
	var cq = make(CallableQ, 100)
	var es = &Executors{cq, &sync.Mutex{}, make(map[*Callable]interface{})}
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			var callable Callable
			for {
				callable = <-cq
				var ret = callable()
				es.lock.Lock()
				defer es.lock.Unlock()
				es.rets[&callable] = ret
			}
		}()
	}

	return es

}

func (es *Executors) Submit(callable func() interface{}) {
	es.callableQ <- callable
}
