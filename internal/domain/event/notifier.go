package event

type NotifierWithConstArg[T any] struct {
	NotifierWith[T]
	Arg T
}

func NewNotifyWithConstArg[T any](arg T, notifier NotifierWith[T]) *NotifierWithConstArg[T] {
	return &NotifierWithConstArg[T]{notifier, arg}
}

func (notifier *NotifierWithConstArg[T]) Notify(event Event) {
	notifier.NotifyWith(notifier.Arg, event)
}
