package executors

type Executors struct {
	callableQ chan func() interface{}
}

func NewExecutors() *Executors {
	var cq = make(chan func() interface{}, 100)
	go func() {
		var callable func() interface{}
		callable <- cq
	}()
	return &Executors{cq}

}

func (es *Executors) Submit(callable func() interface{}) {
	es.callableQ <- callable
}
