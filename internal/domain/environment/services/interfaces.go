package services

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
)

type AccessControl interface {
	CreatePolicy(ctx context.Context, name string) (*entities.AccessPolicy, error)
	CanView(ctx context.Context, env *aggregates.Environment, providedKey *string) (bool, error)
	CanChange(ctx context.Context, env *aggregates.Environment, providedKey *string) (bool, error)
}
