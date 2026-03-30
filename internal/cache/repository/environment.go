package repository

import (
	"context"
	"encoding/json"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/service/ports"
	"envmn/logs"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	environmentCacheKeyBase         = "env"
	variablesCacheKeyBase           = "env:vars"
	environmentPoliciesCacheKeyBase = "env:policies"
)

var errValueNotFound = errors.New("value not found in cache")

type environmentCache struct {
	*redisCacheDecorator

	ports.EnvironmentRepository
	ports.EnvironmentVariablesRepository
	ports.EnvironmentPoliciesRepository
}

func NewEnvironmentCache(
	settings RedisCacheSettings,
	envRepo ports.EnvironmentRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	envPoliciesRepo ports.EnvironmentPoliciesRepository,
	metrics RepositoryMetrics,
) *environmentCache {
	return &environmentCache{
		redisCacheDecorator:            newRedisCacheDecorator(settings, metrics),
		EnvironmentRepository:          envRepo,
		EnvironmentPoliciesRepository:  envPoliciesRepo,
		EnvironmentVariablesRepository: envVarsRepo,
	}
}

func (cache *environmentCache) FindByID(ctx context.Context, id uuid.UUID) (*aggregates.Environment, error) {
	rawEnv, err, _ := cache.group.Do("repo.env.find_by_id", func() (any, error) {
		env, err := cache.getEnvironment(ctx, "id", id.String())
		if err != nil && err != errValueNotFound {
			cache.log.Error(
				"cannot get environment by id from cache",
				logs.Args{"id": id, "error": err},
			)
		}
		if err != errValueNotFound {
			return env, nil
		}

		env, err = cache.EnvironmentRepository.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		go func() {
			err := cache.setEnvironment(context.Background(), "id", id.String(), env)
			if err != nil {
				cache.log.Error(
					"cannot cache environment with a id as key",
					logs.Args{"id": id, "error": err},
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

func (cache *environmentCache) FindByName(ctx context.Context, name string) (*aggregates.Environment, error) {
	rawEnv, err, _ := cache.group.Do("repo.env.find_by_id", func() (any, error) {
		env, err := cache.getEnvironment(ctx, "name", name)
		if err != nil && err != errValueNotFound {
			cache.log.Error(
				"cannot get environment by name from cache",
				logs.Args{"name": name, "error": err},
			)
		}
		if err != errValueNotFound {
			return env, nil
		}

		env, err = cache.EnvironmentRepository.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		go func() {
			err := cache.setEnvironment(context.Background(), "name", name, env)
			if err != nil {
				cache.log.Error(
					"cannot cache environment with a name as key",
					logs.Args{"name": name, "error": err},
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

func (cache *environmentCache) UpdateVariables(ctx context.Context, env *aggregates.Environment) error {
	err := cache.UpdateVariables(ctx, env)
	if err != nil {
		return err
	}
	go func() {
		err := cache.updateCachedVariables(context.Background(), env.ID, env.Variables())
		if err != nil {
			cache.log.Error(
				"cannot update environment varables cache",
				logs.Args{"environment_id": env.ID, "error": err},
			)
		}
	}()
	return nil
}

func (cache *environmentCache) updateCachedVariables(ctx context.Context, envID uuid.UUID, variables entities.Variables) error {
	varsCacheKey := cache.makeKey(variablesCacheKeyBase, envID.String())

	val, err := cache.rdb.Exists(ctx, varsCacheKey).Result()
	if err != nil {
		return fmt.Errorf("cannot check is variables cache exists: %w", err)
	}
	if val < 0 {
		return nil
	}

	varsData, err := json.Marshal(variables)
	if err != nil {
		return fmt.Errorf("cannot marshal variables: %w", err)
	}

	err = cache.rdb.Set(ctx, varsCacheKey, varsData, redis.KeepTTL).Err()
	if err != nil {
		return fmt.Errorf("cannot cache variables: %w", err)
	}
	return nil
}

func (cache *environmentCache) getEnvironment(ctx context.Context, field, value string) (*aggregates.Environment, error) {
	envCacheKey := cache.makeKey(environmentCacheKeyBase, field, value)

	res, err := cache.rdb.Get(ctx, envCacheKey).Result()
	if err == redis.Nil {
		return nil, errValueNotFound
	}
	if err != nil {
		return nil, err
	}

	var envDTO EnvironmentDTO
	if err = json.Unmarshal([]byte(res), &envDTO); err != nil {
		return nil, fmt.Errorf("cannot unmarshal environment: %w", err)
	}

	res, err = cache.rdb.Get(ctx, envCacheKey).Result()
	if err == redis.Nil {
		return nil, errValueNotFound
	}

	policiesCacheKey := cache.makeKey(environmentPoliciesCacheKeyBase, envDTO.ID.String())
	res, err = cache.rdb.Get(ctx, policiesCacheKey).Result()
	if err == redis.Nil {
		return nil, err
	}

	vars, err := cache.getVariables(ctx, envDTO.ID)
	if err != nil {
		if err != errValueNotFound {
			err = fmt.Errorf("cannot get variables from cache: %w", err)
		}
		return nil, fmt.Errorf("cannot get variables from cache: %w", err)
	}

	policies, err := cache.getPolicies(ctx, envDTO.ID)
	if err != nil {
		if err != errValueNotFound {
			err = fmt.Errorf("cannot get policies from cache: %w", err)
		}
		return nil, err
	}

	env, err := aggregates.NewEnvironment(envDTO.Name, "", vars)
	if err != nil {
		return nil, fmt.Errorf("cannot create environment: %w", err)
	}
	if envDTO.Description != nil {
		env.Description = *envDTO.Description
	}
	for policyID, canChange := range policies {
		env.AddPolicy(policyID, canChange)
	}

	env.ID = envDTO.ID
	env.LastVariablesUpdate = envDTO.LastVariablesUpdate

	return env, nil
}

func (cache *environmentCache) getVariables(ctx context.Context, envID uuid.UUID) (entities.Variables, error) {
	key := cache.makeKey(variablesCacheKeyBase, envID.String())
	res, err := cache.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errValueNotFound
	}

	var variables entities.Variables
	if err = json.Unmarshal([]byte(res), &variables); err != nil {
		return nil, fmt.Errorf("cannot unmarshal variables: %w", err)
	}
	return variables, nil
}

func (cache *environmentCache) getPolicies(ctx context.Context, envID uuid.UUID) (map[uuid.UUID]bool, error) {
	key := cache.makeKey(environmentPoliciesCacheKeyBase, envID.String())
	res, err := cache.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errValueNotFound
	}

	var policies map[uuid.UUID]bool
	if err = json.Unmarshal([]byte(res), &policies); err != nil {
		return nil, fmt.Errorf("cannot unmarshal variables: %w", err)
	}
	return policies, nil
}

func (cache *environmentCache) setEnvironment(ctx context.Context, field, value string, env *aggregates.Environment) error {
	_, err := cache.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {

		envCacheKey := cache.makeKey(environmentCacheKeyBase, field, value)
		dto := EnvironmentDTO{
			ID:          env.ID,
			Name:        env.Name,
			Description: &env.Description,
		}

		envData, err := json.Marshal(dto)
		if err != nil {
			return fmt.Errorf("cannot marshal environment: %w", err)
		}

		err = pipe.Set(ctx, envCacheKey, envData, cache.ttl).Err()
		if err != nil {
			return fmt.Errorf("cannot cache environment: %w", err)
		}

		varsData, err := json.Marshal(env.Variables())
		if err != nil {
			return fmt.Errorf("cannot marshal variables: %w", err)
		}

		varsCacheKey := cache.makeKey(variablesCacheKeyBase, env.ID.String())
		err = pipe.Set(ctx, varsCacheKey, varsData, cache.ttl).Err()
		if err != nil {
			return fmt.Errorf("cannot cache variables: %w", err)
		}

		policiesData, err := json.Marshal(env.Policies())
		if err != nil {
			return fmt.Errorf("cannot marshal policies: %w", err)
		}

		policiesCacheKey := cache.makeKey(environmentPoliciesCacheKeyBase, env.ID.String())
		err = pipe.Set(ctx, policiesCacheKey, policiesData, cache.ttl).Err()
		if err != nil {
			return fmt.Errorf("cannot cache polcices: %w", err)
		}
		return nil
	})
	return err
}
