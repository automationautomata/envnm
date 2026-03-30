package postgres

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	queries "envmn/internal/repository/queries/postgres"
	"fmt"

	"github.com/google/uuid"
)

type environmentsRepository struct {
	q *queries.Queries
}

func NewEnvironmentsRepository(q *queries.Queries) *environmentsRepository {
	return &environmentsRepository{q: q}
}

func (r *environmentsRepository) Save(ctx context.Context, env *aggregates.Environment) error {
	err := r.q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
		ID:                  env.ID,
		Name:                env.Name,
		Description:         nullableString(env.Description),
		LastVariablesUpdate: env.LastVariablesUpdate,
		CreatedAt:           env.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("create environment: %w", err)
	}

	for k, v := range env.Variables() {
		err = r.q.UpsertVariable(ctx, queries.UpsertVariableParams{
			Key:           string(k),
			Value:         string(v),
			EnvironmentID: env.ID,
		})
		if err != nil {
			return fmt.Errorf("upsert variable: %w", err)
		}
	}

	return nil
}

func (r *environmentsRepository) FindByID(ctx context.Context, id uuid.UUID) (*aggregates.Environment, error) {
	env, err := r.q.GetEnvironmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.buildEnvironment(ctx, env)
}

func (r *environmentsRepository) FindByName(ctx context.Context, name string) (*aggregates.Environment, error) {
	env, err := r.q.GetEnvironmentByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return r.buildEnvironment(ctx, env)
}

func (r *environmentsRepository) List(ctx context.Context) ([]*aggregates.Environment, error) {
	rows, err := r.q.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*aggregates.Environment, 0, len(rows))

	for _, row := range rows {
		env, err := r.buildEnvironment(ctx, row)
		if err != nil {
			return nil, err
		}
		result = append(result, env)
	}

	return result, nil
}

func (r *environmentsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteEnvironment(ctx, id)
}

func (r *environmentsRepository) buildEnvironment(
	ctx context.Context,
	env queries.Environment,
) (*aggregates.Environment, error) {

	varsRows, err := r.q.GetVariablesByEnv(ctx, env.ID)
	if err != nil {
		return nil, err
	}

	policyRows, err := r.q.GetPoliciesByEnv(ctx, env.ID)
	if err != nil {
		return nil, err
	}

	vars := entities.NewVariables()

	for _, v := range varsRows {
		key, err := vo.NewVariableKey(v.Key)
		if err != nil {
			return nil, err
		}
		vars[key] = vo.NewVariableValue(v.Value)
	}

	e, err := aggregates.NewEnvironment(env.Name, "", vars)
	if err != nil {
		return nil, err
	}

	if env.Description != nil {
		e.Description = *env.Description
	}

	e.ID = env.ID
	e.CreatedAt = env.CreatedAt
	e.LastVariablesUpdate = env.LastVariablesUpdate

	for _, p := range policyRows {
		e.AddPolicy(p.ID, p.ChangesAllowed)
	}

	return e, nil
}
