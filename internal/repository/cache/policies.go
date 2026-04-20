package cache

import (
	"context"
	"encoding/json"
	"envmn/internal/domain/environment/entities"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const policyCacheKeyBase = "policy"

type PolicyCache struct {
	*redisCache
}

func NewPolicyCache(settings RedisCacheSettings) *PolicyCache {
	return &PolicyCache{
		redisCache: newRedisCache(settings),
	}
}

func (c *PolicyCache) GetByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	key := c.makeKey(policyCacheKeyBase, id.String())
	return c.get(ctx, key)
}

func (c *PolicyCache) GetByKey(ctx context.Context, policyKey string) (*entities.AccessPolicy, error) {
	cacheKeyAlias := c.makeKey(policyCacheKeyBase, policyKey)

	policyKey, err := c.rdb.Get(ctx, cacheKeyAlias).Result()
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}
	if err != nil {
		return nil, err
	}
	return c.get(ctx, policyKey)
}

func (c *PolicyCache) Set(ctx context.Context, policy *entities.AccessPolicy) error {
	keyAlias := c.makeKey(policyCacheKeyBase, policy.ID.String())

	policyKey := c.makeKey(policyCacheKeyBase, policy.ID.String())
	err := c.rdb.Set(ctx, keyAlias, policyKey, c.ttl).Err()
	if err != nil {
		return err
	}
	return c.set(ctx, policy)
}

func (c *PolicyCache) set(ctx context.Context, policy *entities.AccessPolicy) error {
	key := c.makeKey(policyCacheKeyBase, policy.ID.String())
	c.rdb.Set(ctx, key, policy, c.ttl)

	data, err := json.Marshal(&PolicyDTO{
		ID:   policy.ID,
		Name: policy.Name,
		Key:  policy.Key,
	})
	if err != nil {
		return fmt.Errorf("cannot marshal policy: %w", err)
	}
	return c.rdb.Set(ctx, key, data, c.ttl).Err()
}

func (c *PolicyCache) get(ctx context.Context, key string) (*entities.AccessPolicy, error) {
	res, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}
	if err != nil {
		return nil, err
	}

	var dto PolicyDTO
	if err = json.Unmarshal([]byte(res), &dto); err != nil {
		return nil, fmt.Errorf("cannot unmarshal policy: %w", err)
	}

	return &entities.AccessPolicy{
		ID:   dto.ID,
		Name: dto.Name,
		Key:  dto.Key,
	}, nil
}
