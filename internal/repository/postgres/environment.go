package postgres

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/repository/postgres/dbtypes"
	"envmn/internal/repository/postgres/queries"
	"envmn/internal/service/ports"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type environmentsRepository struct {
	conn *connection
}

func NewEnvironmentsRepository(conn *connection) *environmentsRepository {
	return &environmentsRepository{conn: conn}
}

func (r *environmentsRepository) Save(ctx context.Context, env *aggregates.Environment) error {
	return r.conn.transaction(ctx, func(q *queries.Queries) error {
		err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
			ID:                  env.ID,
			Name:                env.Name,
			Description:         nullableString(env.Description),
			LastVariablesUpdate: env.LastVariablesUpdate,
			CreatedAt:           env.CreatedAt,
		})
		if err != nil {
			return fmt.Errorf("create environment: %w", err)
		}

		variables := env.Variables()
		inserted := make([]dbtypes.VariableEntry, 0, len(variables))
		for k, v := range variables {
			inserted = append(inserted, dbtypes.VariableEntry{
				Key:           k.String(),
				Value:         v.String(),
				EnvironmentID: env.ID,
			})
		}

		err = q.UpsertVariables(ctx, inserted)
		if err != nil {
			return fmt.Errorf("upsert variable: %w", err)
		}
		return nil
	})
}

func (r *environmentsRepository) UpdateInfo(ctx context.Context, envID uuid.UUID, upd ports.EnvironmentInfoUpdate) error {
	return r.conn.UpdateEnvironmentInfo(ctx, queries.UpdateEnvironmentInfoParams{
		ID:          envID,
		Name:        upd.Name,
		Description: upd.Description,
	})
}

func (r *environmentsRepository) SetLastVariablesUpdate(ctx context.Context, envID uuid.UUID, upd time.Time) error {
	return r.conn.UpdateEnvironmentLastVariablesUpdate(ctx, queries.UpdateEnvironmentLastVariablesUpdateParams{
		ID:                  envID,
		LastVariablesUpdate: upd,
	})
}

func (r *environmentsRepository) FindByName(ctx context.Context, name string) (*aggregates.Environment, error) {
	env, err := r.conn.GetEnvironmentByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return r.toEnvironment(ctx, env)
}

func (r *environmentsRepository) List(ctx context.Context) ([]*aggregates.Environment, error) {
	rows, err := r.conn.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*aggregates.Environment, 0, len(rows))

	for _, row := range rows {
		env, err := r.toEnvironment(ctx, row)
		if err != nil {
			return nil, err
		}
		result = append(result, env)
	}

	return result, nil
}

func (r *environmentsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.conn.DeleteEnvironment(ctx, id)
}

func (r *environmentsRepository) toEnvironment(ctx context.Context, env queries.Environment) (*aggregates.Environment, error) {
	varsRows, err := r.conn.GetVariablesByEnv(ctx, env.ID)
	if err != nil {
		return nil, err
	}

	policyRows, err := r.conn.GetPoliciesByEnv(ctx, env.ID)
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
