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
	cache *cache.PolicyCache
	group *singleflight.Group
	log   logs.Logger
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
		cache:                   cache.NewPolicyCache(settings),
		group:                   &singleflight.Group{},
		log:                     log,
		AccessPolicyRepository:  polisyRepo,
		AccessPolicyFinderSaver: finderSaver,
	}
}

func (r *policyRepositoryCache) FindByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	groupKey := fmt.Sprintf("repo.policy.find_by_id.%s", id)

	rawPolicy, err, _ := r.group.Do(groupKey, func() (any, error) {
		policy, err := r.cache.GetByID(ctx, id)
		if err == nil {
			r.cache.Hit("policy", "id")
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
			r.cache.Miss("policy", "id")
			return nil, nil
		}

		go func() {
			err := r.cache.Set(ctx, policy)
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

func (r *policyRepositoryCache) FindByKey(ctx context.Context, key string) (*entities.AccessPolicy, error) {
	groupKey := fmt.Sprintf("repo.policy.find_by_key.%s", key)

	rawPolicy, err, _ := r.group.Do(groupKey, func() (any, error) {
		policy, err := r.cache.GetByKey(ctx, key)
		if err == nil {
			r.cache.Hit("policy", "id")
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
			r.cache.Miss("policy", "id")
			return nil, nil
		}

		go func() {
			err := r.cache.Set(ctx, policy)
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
