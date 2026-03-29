package entities

import (
	"github.com/google/uuid"
)

type AccessPolicy struct {
	ID             uuid.UUID
	Name           string
	Key            string
	ChangesAllowed bool
}

func NewAccessPolicy(name, key string, changesAllowed bool) *AccessPolicy {
	return &AccessPolicy{
		ID:             uuid.New(),
		Name:           name,
		Key:            key,
		ChangesAllowed: changesAllowed,
	}
}
