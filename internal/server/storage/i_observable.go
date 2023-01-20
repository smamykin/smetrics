package storage

type Observable interface {
	AddObserver(o *Observer)
}

type Observer interface {
	HandleEvent(e IEvent)
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

type BeforeUpsertEvent struct {
	Event
}
