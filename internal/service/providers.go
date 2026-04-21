package service

import (
	envsvc "envmn/internal/domain/environment/services"
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
	*event.Publisher
	envsvc.AccessControl
}

func ProvideManagementServices(deps ManagmentDependincies) *ManagementServices {
	return NewManagementServices(
		environment.New(
			deps.EnvironmentRepository,
			deps.ReservedEnvironmentsStorage,
			deps.EnvironmentPoliciesRepository,
			deps.Publisher,
		),
		policy.New(
			deps.EnvironmentRepository,
			deps.AccessPolicyRepository,
			deps.EnvironmentPoliciesRepository,
			deps.Publisher,
			deps.AccessControl,
		),
		variables.New(
			deps.EnvironmentRepository,
			deps.EnvironmentVariablesRepository,
			deps.Publisher,
			deps.AccessControl,
		),
	)
}

type DistributionDependincies struct {
	ports.EnvironmentRepository
	ports.EnvironmentVariablesRepository
	*event.Publisher
	envsvc.AccessControl
	ports.ClientKeyGenerator
	ports.ReservedEnvironmentsStorage
	event.Notifier
}

func ProvideDistributionServices(deps DistributionDependincies) *DistributionServices {
	return NewDistributionServices(
		variables.New(
			deps.EnvironmentRepository,
			deps.EnvironmentVariablesRepository,
			deps.Publisher,
			deps.AccessControl,
		),
		subscription.New(
			deps.ClientKeyGenerator,
			deps.EnvironmentRepository,
			deps.ReservedEnvironmentsStorage,
			deps.Publisher,
			deps.Notifier,
		),
	)
}
