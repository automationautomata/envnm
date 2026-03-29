package event

type EventPublisher struct {
	notifiers map[string][]EventNotifier
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		notifiers: make(map[string][]EventNotifier),
	}
}

func (pub *EventPublisher) Subscribe(notifier EventNotifier, eventsNames ...string) {
	for _, name := range eventsNames {
		pub.notifiers[name] = append(pub.notifiers[name], notifier)
	}
}

// Notify уведомляет EventNotifier о том, что произошло определенное событие.
// для асинхронных событий обработчики запускаются в горутинах.
func (pub *EventPublisher) Notify(event Event) {
	if event.IsSync() {
		for _, notifier := range pub.notifiers[event.Name()] {
			notifier.Notify(event)
		}
		return
	}
	for _, notifier := range pub.notifiers[event.Name()] {
		go notifier.Notify(event)
	}
}
