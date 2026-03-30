package event

import (
	"context"
	"encoding/json"
	"time"
)

type EventName string

type Event interface {
	Name() EventName
	OccurredAt() time.Time
	HasPayload() bool
	Payload() json.Marshaler
	IsSync() bool
}

type Notifier interface {
	Notify(ctx context.Context, event Event)
}
