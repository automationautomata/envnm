package aggregate

import (
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/environment/events"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"

	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidName      = errors.New("invalid environment name")
	ErrNoAccessPolicyID = errors.New("access policy not found")
	ErrNoVaraible       = errors.New("variable not found")
)

type Environment struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	variables           entities.Variables
	accessPolicies      map[uuid.UUID]bool
	lastVariablesUpdate time.Time
	CreatedAt           time.Time

	events []event.Event
}

func NewEnvironment(name string, description string, vars entities.Variables) (*Environment, error) {
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}

	now := time.Now().UTC()
	return &Environment{
		ID:                  uuid.New(),
		Name:                name,
		Description:         description,
		variables:           vars.Copy(),
		accessPolicies:      make(map[uuid.UUID]bool),
		lastVariablesUpdate: now,
		CreatedAt:           now,
	}, nil
}

func (e *Environment) Variables() entities.Variables {
	return e.variables.Copy()
}

func (e *Environment) LastVariablesUpdate() time.Time {
	return e.lastVariablesUpdate
}

func (e *Environment) UpdateVariables(vars entities.Variables) (new []vo.VariableKey, changed entities.Variables) {
	newVars := entities.NewVariables()
	changed = entities.NewVariables()

	for k, v := range vars {
		if old, ok := e.variables[k]; ok && old != v {
			changed[k] = v
		} else if !ok {
			newVars[k] = v
		}

		e.variables[k] = v
	}

	if len(changed) != 0 && len(new) != 0 {
		e.lastVariablesUpdate = time.Now().UTC()
	}

	if len(changed) != 0 {
		e.raise(events.NewVariablesChanged(vars))
	}
	if len(new) != 0 {
		e.raise(events.NewVariablesCreated(newVars))
	}

	return newVars.Keys(), changed
}

func (e *Environment) RemoveVariable(key vo.VariableKey) error {
	if _, ok := e.variables[key]; !ok {
		return ErrNoVaraible
	}
	delete(e.variables, key)

	e.lastVariablesUpdate = time.Now().UTC()
	e.raise(events.NewVariableDeleted(key))
	return nil
}

func (e *Environment) CanBeChangedBy(accessPolicyID uuid.UUID) bool {
	if len(e.accessPolicies) == 0 {
		return true
	}

	canChange, ok := e.accessPolicies[accessPolicyID]
	return ok && canChange
}

func (e *Environment) HasAccess(accessPolicyID uuid.UUID) bool {
	if len(e.accessPolicies) == 0 {
		return true
	}
	_, ok := e.accessPolicies[accessPolicyID]
	return ok
}

func (e *Environment) AccessPoliciesCount() int {
	return len(e.accessPolicies)
}

func (e *Environment) AddAccessPolicy(policy *entities.AccessPolicy) {
	e.accessPolicies[policy.ID] = policy.ChangesAllowed
}

func (e *Environment) RemoveAccessKeyID(accessPolicyID uuid.UUID) error {
	if _, ok := e.accessPolicies[accessPolicyID]; !ok {
		return ErrNoAccessPolicyID
	}
	delete(e.accessPolicies, accessPolicyID)
	return nil
}

func (e *Environment) raise(event event.Event) {
	e.events = append(e.events, event)
}

func (e *Environment) PullEvents() []event.Event {
	events := e.events
	e.events = make([]event.Event, 0)
	return events
}
