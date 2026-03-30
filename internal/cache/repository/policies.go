package repository

import (
	"context"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/environment/services"
	"envmn/internal/service/ports"
	"envmn/logs"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const policyCacheKeyBase = "policy"

type policyCache struct {
	*redisCacheDecorator
	ports.AccessPolicyRepository
	services.AccessPolicyFinderSaver
}

func NewPolicyCache(
	settings RedisCacheSettings,
	policyRepo ports.AccessPolicyRepository,
	finderSaver services.AccessPolicyFinderSaver,
	metrics RepositoryMetrics,
) *policyCache {
	return &policyCache{
		redisCacheDecorator:     newRedisCacheDecorator(settings, metrics),
		AccessPolicyRepository:  policyRepo,
		AccessPolicyFinderSaver: finderSaver,
	}
}

func (cache *policyCache) FindByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	key := cache.makeKey(policyCacheKeyBase, "id", id.String())

	err := cache.rdb.Get(ctx, key).Err()
	if err == nil {
		cache.metrics.Inc("policy", "id", true)
		return nil, nil
	}
	if err != redis.Nil {
		cache.log.Error(
			"cannot get policy by id from cache",
			logs.Args{"id": id, "error": err},
		)
	} else {
		cache.metrics.Inc("policy", "id", false)
	}

	rawPolicy, err, _ := cache.group.Do("repo.policy.find_by_id", func() (any, error) {
		return cache.AccessPolicyRepository.FindByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return rawPolicy.(*entities.AccessPolicy), nil
}

func (cache *policyCache) FindByName(ctx context.Context, name string) (*entities.AccessPolicy, error) {
	key := cache.makeKey(policyCacheKeyBase, "name", name)

	err := cache.rdb.Get(ctx, key).Err()
	if err == nil {
		cache.metrics.Inc("policy", "name", true)
		return nil, nil
	}
	if err != redis.Nil {
		cache.log.Error(
			"cannot get environment by name from cache",
			logs.Args{"id": name, "error": err},
		)
	} else {
		cache.metrics.Inc("policy", "name", false)
	}

	rawPolicy, err, _ := cache.group.Do("repo.policy.find_by_name", func() (any, error) {
		return cache.FindByName(ctx, name)
	})
	if err != nil {
		return nil, err
	}

	go func() {
		err := cache.rdb.Set(ctx, key, nil, cache.ttl).Err()
		if err != nil {
			cache.log.Error(
				"cannot cache policy with an name as key",
				logs.Args{"name": name, "error": err},
			)
		}
	}()

	return rawPolicy.(*entities.AccessPolicy), nil
}

func (cache *policyCache) FindByKey(ctx context.Context, key string) (*entities.AccessPolicy, error) {
	cacheKey := cache.makeKey(policyCacheKeyBase, "key")

	err := cache.rdb.Get(ctx, cacheKey).Err()
	if err == nil {
		return nil, nil
	}
	if err != redis.Nil {
		cache.log.Error(
			"cannot get policy by key from cache",
			logs.Args{"error": err},
		)
	}

	rawPolicy, err, _ := cache.group.Do("repo.policy.find_by_id", func() (any, error) {
		return cache.AccessPolicyFinderSaver.FindByKey(ctx, key)
	})
	if err != nil {
		return nil, err
	}

	return rawPolicy.(*entities.AccessPolicy), nil
}
