package event

type EventPublisher struct {
	notifiers map[string][]EventNotifier
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		notifiers: make(map[string][]EventNotifier),
	}
}

func (pub *EventPublisher) Subscribe(handler EventNotifier, events ...Event) {
	for _, event := range events {
		pub.notifiers[event.Name()] = append(pub.notifiers[event.Name()], handler)
	}
}

// Notify уведомляет EventNotifier о том, что произошло определенное событие.
// для асинхронных событий обработчики запускаются в горутинах.
func (pub *EventPublisher) Notify(event Event) {
	if event.IsSync() {
		for _, handler := range pub.notifiers[event.Name()] {
			handler.Notify(event)
		}
		return
	}
	for _, handler := range pub.notifiers[event.Name()] {
		go handler.Notify(event)
	}
}
