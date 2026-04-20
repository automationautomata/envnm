package postgres

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/repository/postgres/dbtypes"
	"envmn/internal/repository/postgres/queries"

	"github.com/google/uuid"
)

type variablesRepository struct {
	conn *connection
}

func NewVariablesRepository(conn *connection) *variablesRepository {
	return &variablesRepository{conn: conn}
}

func (r *variablesRepository) UpdateVariables(ctx context.Context, env *aggregates.Environment) error {
	variables := env.Variables()
	inserted := make([]dbtypes.VariableEntry, 0, len(variables))
	for k, v := range variables {
		inserted = append(inserted, dbtypes.VariableEntry{
			Key:           k.String(),
			Value:         v.String(),
			EnvironmentID: env.ID,
		})
	}

	err := r.conn.UpsertVariables(ctx, inserted)
	if err != nil {
		return err
	}
	return nil
}

func (r *variablesRepository) DeleteVariable(ctx context.Context, envID uuid.UUID, key vo.VariableKey) error {
	return r.conn.DeleteVariable(ctx, queries.DeleteVariableParams{
		EnvironmentID: envID,
		Key:           string(key),
	})
}
