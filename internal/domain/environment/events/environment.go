package events

// import (
// 	environ "envmn/internal/domain/environment"
// 	"time"
// )

// const (
// 	EnvironmentCreateEventName    = "environment.create"
// 	EnvironmentAddPolicyEventName = "environment.policy.add"
// )

// func EnvironmentEventsNames() []string {
// 	return []string{
// 		EnvironmentCreateEventName,
// 		EnvironmentAddPolicyEventName,
// 	}
// }

// type environmentEvent struct {
// 	Environment *environ.Environment
// 	name        string
// 	occurredAt  time.Time
// }

// func newEnvironmentEvent(env *environ.Environment, name string) environmentEvent {
// 	return environmentEvent{
// 		Environment: env,
// 		name:        name,
// 		occurredAt:  time.Now().UTC(),
// 	}
// }

// func (e environmentEvent) Name() string          { return e.name }
// func (e environmentEvent) OccurredAt() time.Time { return e.occurredAt }
// func (e environmentEvent) IsSync() bool          { return false }

// type EnvironmentCreated struct {
// 	environmentEvent
// }

// func NewEnvironmentCreated(env *environ.Environment) EnvironmentCreated {
// 	return EnvironmentCreated{
// 		environmentEvent: newEnvironmentEvent(env, EnvironmentCreateEventName),
// 	}
// }

// type EnvironmentAddPolicy struct {
// 	environmentEvent
// }

// func NewEnvironmentAddedToProfile(env *environ.Environment) EnvironmentAddPolicy {
// 	return EnvironmentAddPolicy{
// 		environmentEvent: newEnvironmentEvent(env, EnvironmentAddPolicyEventName),
// 	}
// }
