package event

import "context"

type EventPublisher struct {
	notifiers map[EventName][]Notifier
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		notifiers: make(map[EventName][]Notifier),
	}
}

func (pub *EventPublisher) Subscribe(notifier Notifier, eventNames ...EventName) {
	for _, name := range eventNames {
		pub.notifiers[name] = append(pub.notifiers[name], notifier)
	}
}

// Publish уведомляет EventNotifier-ы о том, что произошло определенное событие.
// для асинхронных событий обработчики запускаются в горутинах.
func (pub *EventPublisher) Publish(ctx context.Context, event Event) {
	if event.IsSync() {
		for _, notifier := range pub.notifiers[event.Name()] {
			notifier.Notify(ctx, event)
		}
		return
	}
	for _, notifier := range pub.notifiers[event.Name()] {
		go notifier.Notify(ctx, event)
	}
}
