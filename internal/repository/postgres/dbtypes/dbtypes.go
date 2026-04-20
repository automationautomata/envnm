package dbtypes

import "github.com/google/uuid"

type VariableEntry struct {
	Key           string    `db:"key"`
	Value         string    `db:"value"`
	EnvironmentID uuid.UUID `db:"environment_id"`
}
