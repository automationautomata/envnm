package events

import (
	"encoding/json"
	"time"

	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"
)

const (
	VariablesCreateEventName event.EventName = "variables.create"
	VariablesChangeEventName event.EventName = "variables.change"
	VariablesDeleteEventName event.EventName = "variables.delete"
)

func VariableEventsNames() []event.EventName {
	return []event.EventName{
		VariablesCreateEventName,
		VariablesChangeEventName,
		VariablesDeleteEventName,
	}
}

type variablesEvent struct {
	name       event.EventName
	occurredAt time.Time
}

func newVariableEvent(name event.EventName) variablesEvent {
	return variablesEvent{
		name:       name,
		occurredAt: time.Now().UTC(),
	}
}

func (e variablesEvent) Name() event.EventName   { return e.name }
func (e variablesEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e variablesEvent) IsSync() bool            { return false }
func (e variablesEvent) HasPayload() bool        { return false }
func (e variablesEvent) Payload() json.Marshaler { return make(payloadMap) }

type variablesEventWithValues struct {
	variablesEvent
	Variables entities.Variables
}

func (e variablesEventWithValues) Payload() json.Marshaler {
	payload := make(payloadMap)
	for k, v := range e.Variables {
		payload[k.String()] = v
	}
	return payload
}

type VariablesCreated variablesEventWithValues

func NewVariablesCreated(vars entities.Variables) VariablesCreated {
	return VariablesCreated{
		variablesEvent: newVariableEvent(VariablesCreateEventName),
		Variables:      vars,
	}
}

type VariablesChanged variablesEventWithValues

func NewVariablesChanged(vars entities.Variables) VariablesChanged {
	return VariablesChanged{
		variablesEvent: newVariableEvent(VariablesChangeEventName),
		Variables:      vars,
	}
}

type VariableDeleted struct {
	variablesEvent
	Keys []vo.VariableKey
}

func NewVariableDeleted(keys ...vo.VariableKey) VariableDeleted {
	return VariableDeleted{
		variablesEvent: newVariableEvent(VariablesDeleteEventName),
		Keys:           keys,
	}
}

func (e VariableDeleted) Payload() json.Marshaler {
	payload := make(payloadArray, len(e.Keys))
	for i, key := range e.Keys {
		payload[i] = key
	}
	return payload
}
