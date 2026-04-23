package decorators

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/repository/cache"
	"envmn/internal/service/ports"
	"envmn/logs"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

type environmentRepositoryCache struct {
	ports.EnvironmentRepository
	ports.EnvironmentPoliciesRepository
	ports.EnvironmentVariablesRepository
	log   logs.Logger
	cache *cache.EnvironmentCache
	group *singleflight.Group
}

func NewEnvironmentRepositoryCache(
	log logs.Logger,
	settings cache.RedisCacheSettings,
	envRepo ports.EnvironmentRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	envPoliciesRepo ports.EnvironmentPoliciesRepository,
) *environmentRepositoryCache {
	return &environmentRepositoryCache{
		cache:                          cache.NewEnvironmentCache(settings),
		group:                          &singleflight.Group{},
		log:                            log,
		EnvironmentRepository:          envRepo,
		EnvironmentPoliciesRepository:  envPoliciesRepo,
		EnvironmentVariablesRepository: envVarsRepo,
	}
}

func (r *environmentRepositoryCache) FindByName(ctx context.Context, name string) (*aggregates.Environment, error) {
	groupKey := fmt.Sprintf("repo.env.find_by_name.%s", name)

	rawEnv, err, _ := r.group.Do(groupKey, func() (any, error) {
		env, err := r.cache.GetEnvironment(ctx, name)

		if err == nil {
			r.cache.Hit("environment", "name")
			return env, nil
		}

		miss := errors.Is(err, cache.ErrValueNotFound)
		if !miss {
			r.log.Error(
				"cannot get environment by name from cache",
				logs.Args{"name": name, "error": err},
			)
		}

		env, err = r.EnvironmentRepository.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if env != nil && miss {
			r.cache.Miss("environment", "name")
			return nil, nil
		}

		go func() {
			err := r.cache.SetEnvironment(context.Background(), env)
			if err != nil {
				r.log.Error(
					"cannot cache environment with a name as key",
					logs.Args{"environment_id": env.ID, "error": err},
				)
			}
		}()
		return env, nil
	})
	if err != nil {
		return nil, err
	}
	return rawEnv.(*aggregates.Environment), nil
}

func (r *environmentRepositoryCache) UpdateInfo(ctx context.Context, envID uuid.UUID, upd ports.EnvironmentInfoUpdate) error {
	err := r.EnvironmentRepository.UpdateInfo(ctx, envID, upd)
	if err != nil {
		return err
	}
	go r.removeEnvironment(envID)
	return nil
}

func (r *environmentRepositoryCache) UpdateVariables(ctx context.Context, env *aggregates.Environment) error {
	err := r.EnvironmentVariablesRepository.UpdateVariables(ctx, env)
	if err != nil {
		return err
	}
	go func() {
		err := r.cache.UpdateCachedVariables(context.Background(), env.ID, env.Variables())
		if err != nil {
			r.log.Error(
				"cannot update environment varables cache",
				logs.Args{"environment_id": env.ID, "error": err},
			)
			go r.removeEnvironment(env.ID)
		}
	}()
	return nil
}

func (r *environmentRepositoryCache) AddToEnvironment(ctx context.Context, envID uuid.UUID, policyID uuid.UUID, canChange bool) error {
	err := r.EnvironmentPoliciesRepository.AddToEnvironment(ctx, envID, policyID, canChange)
	if err != nil {
		return err
	}
	go r.removeEnvironment(envID)
	return nil
}

func (r *environmentRepositoryCache) DeleteFromEnvironment(ctx context.Context, envID uuid.UUID, policyID uuid.UUID) error {
	err := r.EnvironmentPoliciesRepository.DeleteFromEnvironment(ctx, envID, policyID)
	if err != nil {
		return err
	}
	go r.removeEnvironment(envID)
	return nil
}

func (r environmentRepositoryCache) removeEnvironment(envID uuid.UUID) {
	err := r.cache.Remove(context.Background(), envID)
	if err != nil {
		r.log.Error(
			"cannot remove environment",
			logs.Args{"environment_id": envID, "error": err},
		)
	}
}
