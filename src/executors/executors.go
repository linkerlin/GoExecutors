package executors

import (
	"config"
)

type Executors struct {
	callableQ chan func() interface{}
}

func NewExecutors() *Executors {
	var cq = make(chan func() interface{}, 100)
	for i := 0; i < config.DefaultGoroutinesNum(); i++ {
		go func() {
			var callable func() interface{}
			for {
				callable = <-cq
				callable()
			}
		}()
	}

	return &Executors{cq}

}

func (es *Executors) Submit(callable func() interface{}) {
	es.callableQ <- callable
}
