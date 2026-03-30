package repository

import (
	"time"

	"github.com/google/uuid"
)

type EnvironmentDTO struct {
	ID                  uuid.UUID
	Name                string
	Description         *string
	LastVariablesUpdate time.Time
	CreatedAt           time.Time
}
