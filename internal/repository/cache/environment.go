package cache

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	environmentCacheKeyBase         = "env"
	variablesCacheKeyBase           = "env:vars"
	environmentPoliciesCacheKeyBase = "env:policies"
)

var resetScript = redis.NewScript(
	`
	local ttl = redis.call("TTL", KEYS[1])
	redis.call("DEL", KEYS[1])
	redis.call('HSET', KEYS[1], unpack(ARGV, 2))
	redis.call('EXPIRE', KEYS[1], ttl)
	return 1
    `,
)

type EnvironmentCache struct {
	*redisCache
}

func NewEnvironmentCache(settings RedisCacheSettings) *EnvironmentCache {
	return &EnvironmentCache{
		redisCache: newRedisCache(settings),
	}
}

func (c *EnvironmentCache) UpdateCachedVariables(ctx context.Context, envID uuid.UUID, variables entities.Variables) error {
	variablesKey := c.makeVariablesCacheKey(envID.String())

	val, err := c.rdb.Exists(ctx, variablesKey).Result()
	if err != nil {
		return fmt.Errorf("cannot check is variables c exists: %w", err)
	}
	if val == 0 {
		return nil
	}

	variablesData := make(map[string]any)
	for k, v := range variables {
		variablesData[k.String()] = v
	}

	err = resetScript.Run(ctx, c.rdb, []string{variablesKey}, variablesData).Err()
	if err != nil {
		return fmt.Errorf("cannot delete variables: %w", err)
	}
	return nil
}

func (c *EnvironmentCache) GetEnvironment(ctx context.Context, name string) (*aggregates.Environment, error) {
	envKeyAlias := c.makeEnvironmentCacheKey(name)

	envKey, err := c.rdb.Get(ctx, envKeyAlias).Result()
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}

	var dto EnvironmentDTO
	err = c.rdb.Get(ctx, envKey).Scan(&dto)
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}
	if err != nil {
		return nil, err
	}

	policies, err := c.getPolicies(ctx, dto.ID)
	if err == redis.Nil {
		if err != ErrValueNotFound {
			err = fmt.Errorf("cannot get policies from c: %w", err)
		}
		return nil, err
	}

	vars, err := c.GetVariables(ctx, dto.ID)
	if err != nil {
		if err != ErrValueNotFound {
			err = fmt.Errorf("cannot get variables from c: %w", err)
		}
		return nil, fmt.Errorf("cannot get variables from c: %w", err)
	}

	env, err := aggregates.NewEnvironment(dto.Name, "", vars)
	if err != nil {
		return nil, fmt.Errorf("cannot create environment: %w", err)
	}
	if dto.Description != nil {
		env.Description = *dto.Description
	}
	for policyID, canChange := range policies {
		env.AddPolicy(policyID, canChange)
	}

	env.ID = dto.ID
	env.LastVariablesUpdate = dto.LastVariablesUpdate

	keys := []string{
		envKey,
		c.makeEnvironmentCacheKey(dto.Name),
		c.makePoliciesCacheKey(env.ID.String()),
		c.makeVariablesCacheKey(env.ID.String()),
	}
	err = c.updateTTL(ctx, keys...)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (c *EnvironmentCache) updateTTL(ctx context.Context, keys ...string) error {
	_, err := c.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, k := range keys {
			pipe.Expire(ctx, k, c.ttl)
		}
		return nil
	})
	return err
}

func (c *EnvironmentCache) GetVariables(ctx context.Context, envID uuid.UUID) (entities.Variables, error) {
	key := c.makeKey(variablesCacheKeyBase, envID.String())
	res, err := c.rdb.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}

	variables := entities.NewVariables()
	for k, v := range res {
		key, err := vo.NewVariableKey(k)
		if err != nil {
			return nil, err
		}
		variables[key] = vo.NewVariableValue(v)
	}
	return variables, nil
}

func (c *EnvironmentCache) getPolicies(ctx context.Context, envID uuid.UUID) (map[uuid.UUID]bool, error) {
	key := c.makeKey(environmentPoliciesCacheKeyBase, envID.String())
	res, err := c.rdb.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrValueNotFound
	}

	policies := make(map[uuid.UUID]bool)
	for k, v := range res {
		id, err := uuid.Parse(k)
		if err != nil {
			return nil, err
		}

		policies[id], err = strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
	}
	return policies, nil
}

func (c *EnvironmentCache) Remove(ctx context.Context, envID uuid.UUID) error {
	envKey := c.makeEnvironmentCacheKey(envID.String())

	var dto EnvironmentDTO
	err := c.rdb.HGetAll(ctx, envKey).Scan(&dto)
	if err == redis.Nil {
		return ErrValueNotFound
	}
	if err != nil {
		return err
	}

	keys := []string{
		envKey,
		c.makeEnvironmentCacheKey(dto.Name),
		c.makePoliciesCacheKey(envID.String()),
		c.makeVariablesCacheKey(envID.String()),
	}
	return c.rdb.Del(ctx, keys...).Err()
}

type cmdInfo struct {
	redis.Cmder
	key string
}

func (c *EnvironmentCache) SetEnvironment(ctx context.Context, env *aggregates.Environment) error {
	envKeyAlias := c.makeEnvironmentCacheKey(env.Name)
	envKey := c.makeEnvironmentCacheKey(env.ID.String())

	envData := EnvironmentDTO{
		ID:                  env.ID,
		Name:                env.Name,
		Description:         &env.Description,
		LastVariablesUpdate: env.LastVariablesUpdate,
		CreatedAt:           env.CreatedAt,
	}

	variablesKey := c.makeVariablesCacheKey(env.ID.String())
	variablesData := make(map[string]any)
	for k, v := range env.Variables() {
		variablesData[k.String()] = v
	}

	policiesKey := c.makePoliciesCacheKey(env.ID.String())
	policiesData := make(map[string]any)
	for k, v := range env.Policies() {
		policiesData[k.String()] = v
	}

	data := map[string]any{
		envKey:       envData,
		envKeyAlias:  envKey,
		policiesKey:  policiesData,
		variablesKey: variablesData,
	}

	cmds := make([]cmdInfo, 0, len(data)+1)
	_, pipilineErr := c.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		setCmd := pipe.Set(ctx, envKeyAlias, envKey, c.ttl)

		cmds = append(cmds, cmdInfo{setCmd, envKeyAlias})

		for k, v := range data {
			hsetCmd := pipe.HSet(ctx, k, v)
			expireCmd := pipe.Expire(ctx, k, c.ttl)
			cmds = append(cmds,
				cmdInfo{hsetCmd, envKeyAlias},
				cmdInfo{expireCmd, envKeyAlias},
			)
		}
		return nil
	})

	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			cmdErr := fmt.Errorf("cannot cache value with key %q: %w; pipiline error: %w", cmd.key, cmd.Err(), err)
			if pipilineErr != nil {
				return fmt.Errorf("%w; %w", cmdErr, pipilineErr)
			}
			return cmdErr
		}
	}
	return pipilineErr
}

func (c *EnvironmentCache) hsetWithExpire(ctx context.Context, cmd redis.StatefulCmdable, key string, data any) error {
	cmd.HSet(ctx, key, data)
	cmd.Expire(ctx, key, c.ttl)
	return nil
}

func (c *EnvironmentCache) makeEnvironmentCacheKey(suffix string) string {
	return c.makeKey(environmentCacheKeyBase, suffix)
}

func (c *EnvironmentCache) makeVariablesCacheKey(suffix string) string {
	return c.makeKey(variablesCacheKeyBase, suffix)
}

func (c *EnvironmentCache) makePoliciesCacheKey(suffix string) string {
	return c.makeKey(policyCacheKeyBase, suffix)
}
