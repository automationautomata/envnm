package event

import (
	"encoding/json"
	"time"
)

type Event interface {
	Name() string
	OccurredAt() time.Time
	HasPayload() bool
	Payload() json.Marshaler
	IsSync() bool
}

type EventNotifier interface {
	Notify(event Event)
}

type NotifierWith[T any] interface {
	NotifyWith(arg T, event Event)
}
