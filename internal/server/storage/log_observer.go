package storage

import "log"

type LogObserver struct {
	logger *log.Logger
}

func (l LogObserver) HandleEvent(e IEvent) {
	if _, ok := e.(AfterUpsertEvent); ok {
		l.logger.Printf("upsert %#v\n", e.Payload())
	}
}
