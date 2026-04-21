package event

import "context"

type Publisher struct {
	notifiers map[EventName][]Notifier
}

func NewPublisher() *Publisher {
	return &Publisher{
		notifiers: make(map[EventName][]Notifier),
	}
}

func (pub *Publisher) Subscribe(notifier Notifier, eventNames ...EventName) {
	for _, name := range eventNames {
		pub.notifiers[name] = append(pub.notifiers[name], notifier)
	}
}

// Publish уведомляет EventNotifier-ы о том, что произошло определенное событие.
// для асинхронных событий обработчики запускаются в горутинах.
func (pub *Publisher) Publish(ctx context.Context, event Event) {
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
