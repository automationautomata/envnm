package cache

import (
	"envmn/internal/domain/environment/services"
	"envmn/internal/service/ports"
)

type EnvironmentRepositoryCache interface {
	ports.EnvironmentRepository
	ports.EnvironmentVariablesRepository
	ports.EnvironmentPoliciesRepository
}

type PolicyRepositoryCache interface {
	ports.AccessPolicyRepository
	services.AccessPolicyFinderSaver
}
