package service

import (
	"envmn/internal/domain/environment/services"
	"envmn/internal/domain/event"
	"envmn/internal/service/environment"
	"envmn/internal/service/policy"
	"envmn/internal/service/ports"
	"envmn/internal/service/subscription"
	"envmn/internal/service/variables"
)

type ManagmentDependincies struct {
	ports.EnvironmentRepository
	ports.ReservedEnvironmentsStorage
	ports.AccessPolicyRepository
	ports.EnvironmentPoliciesRepository
	ports.EnvironmentVariablesRepository
	*event.EventPublisher
	*services.AccessControlService
}

func ProvideManagement(deps ManagmentDependincies) *Management {
	return NewManagement(
		environment.New(
			deps.EnvironmentRepository,
			deps.ReservedEnvironmentsStorage,
			deps.EnvironmentPoliciesRepository,
			deps.EventPublisher,
		),
		policy.New(
			deps.EnvironmentRepository,
			deps.AccessPolicyRepository,
			deps.EventPublisher,
			deps.AccessControlService,
		),
		variables.New(
			deps.EnvironmentRepository,
			deps.EnvironmentVariablesRepository,
			deps.EventPublisher,
			deps.AccessControlService,
		),
	)
}

type ClientDependincies struct {
	ports.EnvironmentRepository
	ports.EnvironmentVariablesRepository
	*event.EventPublisher
	*services.AccessControlService
	ports.ClientKeyGenerator
	ports.ReservedEnvironmentsStorage
	event.Notifier
}

func ProvideClient(deps ClientDependincies) *Client {
	return NewClient(
		variables.New(
			deps.EnvironmentRepository,
			deps.EnvironmentVariablesRepository,
			deps.EventPublisher,
			deps.AccessControlService,
		),
		subscription.New(
			deps.ClientKeyGenerator,
			deps.EnvironmentRepository,
			deps.ReservedEnvironmentsStorage,
			deps.EventPublisher,
			deps.Notifier,
		),
	)
}
