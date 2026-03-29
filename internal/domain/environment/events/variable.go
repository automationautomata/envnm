package events

import (
	"encoding/json"
	"time"

	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
)

const (
	VariablesCreateEventName = "variables.create"
	VariablesChangeEventName = "variables.change"
	VariablesDeleteEventName = "variables.delete"
)

func VariableEventsNames() []string {
	return []string{
		VariablesCreateEventName,
		VariablesChangeEventName,
		VariablesDeleteEventName,
	}
}

type payloadMap map[string]any

func (p payloadMap) MarshalJSON() ([]byte, error) { return json.Marshal(map[string]any(p)) }

type variablesEvent struct {
	name       string
	occurredAt time.Time
}

func newVariableEvent(name string) variablesEvent {
	return variablesEvent{
		name:       name,
		occurredAt: time.Now().UTC(),
	}
}

func (e variablesEvent) Name() string            { return e.name }
func (e variablesEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e variablesEvent) IsSync() bool            { return false }
func (e variablesEvent) HasPayload() bool        { return false }
func (e variablesEvent) Payload() json.Marshaler { return make(payloadMap) }

type VariablesCreated struct {
	variablesEvent
	Variables entities.Variables
}

func NewVariablesCreated(vars entities.Variables) VariablesCreated {
	return VariablesCreated{
		variablesEvent: newVariableEvent(VariablesCreateEventName),
		Variables:      vars,
	}
}

type VariablesChanged struct {
	variablesEvent
	Variables entities.Variables
}

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
