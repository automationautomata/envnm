package entities

import (
	"github.com/google/uuid"
)

type AccessPolicy struct {
	ID   uuid.UUID
	Name string
	Key  string
}

func NewAccessPolicy(name, key string) *AccessPolicy {
	return &AccessPolicy{
		ID:   uuid.New(),
		Name: name,
		Key:  key,
	}
}
