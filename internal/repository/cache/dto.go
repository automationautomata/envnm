package cache

import (
	"time"

	"github.com/google/uuid"
)

type EnvironmentDTO struct {
	ID                  uuid.UUID `redis:"id"`
	Name                string    `redis:"name"`
	Description         *string   `redis:"description"`
	LastVariablesUpdate time.Time `redis:"last_variables_update"`
	CreatedAt           time.Time `redis:"description"`
}

type PolicyDTO struct {
	ID   uuid.UUID
	Name string
	Key  string
}
