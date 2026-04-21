package ports

import (
	"context"
	"envmn/internal/domain/environment/entities"
)

type AccessPolicyFinderSaver interface {
	Save(ctx context.Context, policy *entities.AccessPolicy) error
	FindByKey(ctx context.Context, key string) (*entities.AccessPolicy, error)
}

type KeyGenerator interface {
	Generate() string
}
