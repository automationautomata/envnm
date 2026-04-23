package decorators

import (
	"context"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/repository/cache"
	"envmn/internal/service/ports"
	"envmn/logs"
	"errors"
	"fmt"

	domainports "envmn/internal/domain/environment/services/ports"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

const policyCacheKeyBase = "policy"

type policyRepositoryCache struct {
	policyCache *cache.PolicyCache
	envCache    *cache.EnvironmentCache
	group       *singleflight.Group
	log         logs.Logger
	ports.AccessPolicyRepository
	domainports.AccessPolicyFinderSaver
}

func NewPolicyRepositoryCache(
	log logs.Logger,
	settings cache.RedisCacheSettings,
	polisyRepo ports.AccessPolicyRepository,
	finderSaver domainports.AccessPolicyFinderSaver,
) *policyRepositoryCache {
	return &policyRepositoryCache{
		policyCache:             cache.NewPolicyCache(settings),
		envCache:                cache.NewEnvironmentCache(settings),
		group:                   &singleflight.Group{},
		log:                     log,
		AccessPolicyRepository:  polisyRepo,
		AccessPolicyFinderSaver: finderSaver,
	}
}

func (r *policyRepositoryCache) FindByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	groupKey := fmt.Sprintf("repo.policy.find_by_id.%s", id)

	rawPolicy, err, _ := r.group.Do(groupKey, func() (any, error) {
		policy, err := r.policyCache.GetByID(ctx, id)
		if err == nil {
			r.policyCache.Hit("policy", "id")
			return policy, nil
		}
		miss := errors.Is(err, cache.ErrValueNotFound)
		if !miss {
			r.log.Error(
				"cannot get policy by id from cache",
				logs.Args{"id": id, "error": err},
			)
		}

		policy, err = r.AccessPolicyRepository.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if policy != nil && miss {
			r.policyCache.Miss("policy", "id")
			return nil, nil
		}

		go func() {
			err := r.policyCache.Set(ctx, policy)
			if err != nil {
				r.log.Error(
					"FindByID: cannot set policy to cache",
					logs.Args{"id": id, "error": err},
				)
			}
		}()
		return policy, nil
	})
	if err != nil {
		return nil, err
	}

	return rawPolicy.(*entities.AccessPolicy), nil
}

func (r *policyRepositoryCache) Remove(ctx context.Context, id uuid.UUID) error {
	envs, err := r.AccessPolicyRepository.ListPolicyEnvironments(ctx, id)
	if err != nil {
		return fmt.Errorf("cache failed: %w", err)
	}

	for _, env := range envs {
		err := r.envCache.RemoveByName(ctx, env.Name)
		if err != nil {
			r.log.Error(
				"Remove: cannot remove environments from cache (by name) while removing policy id",
				logs.Args{"name": env.Name, "policy_id": id, "error": err},
			)
			return errors.New("cache failed on removing old data")
		}
	}

	return r.AccessPolicyRepository.Remove(ctx, id)
}

func (r *policyRepositoryCache) FindByKey(ctx context.Context, key string) (*entities.AccessPolicy, error) {
	groupKey := fmt.Sprintf("repo.policy.find_by_key.%s", key)

	rawPolicy, err, _ := r.group.Do(groupKey, func() (any, error) {
		policy, err := r.policyCache.GetByKey(ctx, key)
		if err == nil {
			r.policyCache.Hit("policy", "id")
			return policy, nil
		}

		miss := errors.Is(err, cache.ErrValueNotFound)
		if !miss {
			r.log.Error(
				"cannot get policy by id from cache",
				logs.Args{"id": key, "error": err},
			)
		}

		policy, err = r.AccessPolicyFinderSaver.FindByKey(ctx, key)
		if err != nil {
			return nil, err
		}

		if policy != nil && miss {
			r.policyCache.Miss("policy", "id")
			return nil, nil
		}

		go func() {
			err := r.policyCache.Set(ctx, policy)
			if err != nil {
				r.log.Error(
					"FindByKey: cannot set policy to cache",
					logs.Args{"id": policy.ID, "error": err},
				)
			}
		}()
		return policy, nil
	})
	if err != nil {
		return nil, err
	}
	return rawPolicy.(*entities.AccessPolicy), nil
}
