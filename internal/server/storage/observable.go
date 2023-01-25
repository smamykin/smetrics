package storage

type Observable interface {
	AddObserver(o *Observer)
}

type Observer interface {
	HandleEvent(e IEvent) error
}

type IEvent interface {
	Payload() interface{}
}

type Event struct {
	payload interface{}
}

func (u Event) Payload() interface{} {
	return u.payload
}

type AfterUpsertEvent struct {
	Event
}

type FuncObserver struct {
	FunctionToInvoke func(e IEvent) error
}

func (fo *FuncObserver) HandleEvent(e IEvent) error {
	return fo.FunctionToInvoke(e)
}
