package postgres

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	vo "envmn/internal/domain/environment/valueobjects"
	queries "envmn/internal/repository/queries/postgres"

	"github.com/google/uuid"
)

type variablesRepository struct {
	q *queries.Queries
}

func NewVariablesRepository(q *queries.Queries) *variablesRepository {
	return &variablesRepository{q: q}
}

func (r *variablesRepository) UpdateVariables(ctx context.Context, env *aggregates.Environment) error {
	for k, v := range env.Variables() {
		err := r.q.UpsertVariable(ctx, queries.UpsertVariableParams{
			Key:           string(k),
			Value:         string(v),
			EnvironmentID: env.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *variablesRepository) DeleteVariable(ctx context.Context, envID uuid.UUID, key vo.VariableKey) error {
	return r.q.DeleteVariable(ctx, queries.DeleteVariableParams{
		EnvironmentID: envID,
		Key:           string(key),
	})
}
