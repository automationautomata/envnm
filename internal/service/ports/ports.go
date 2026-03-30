package ports

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const EnvironmentNameContextKey contextKey = "environment_name"

var (
	ErrEnvironmentWithNameAlreadyExists = errors.New("environment already exists")
)

type AccessPolicyRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error)
	FindByName(ctx context.Context, name string) (*entities.AccessPolicy, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type EnvironmentRepository interface {
	Save(ctx context.Context, env *aggregates.Environment) error
	FindByName(ctx context.Context, name string) (*aggregates.Environment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*aggregates.Environment, error)
	List(ctx context.Context) ([]*aggregates.Environment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type EnvironmentPoliciesRepository interface {
	GetEnvironmentPolicies(ctx context.Context, envID uuid.UUID) ([]*entities.AccessPolicy, error)
	AddToEnvironment(ctx context.Context, envID uuid.UUID, policy *entities.AccessPolicy) error
	DeleteFromEnvironment(ctx context.Context, envID uuid.UUID, policyID uuid.UUID) error
}

type EnvironmentVariablesRepository interface {
	UpdateVariables(ctx context.Context, env *aggregates.Environment) error
	DeleteVariable(ctx context.Context, envID uuid.UUID, key vo.VariableKey) error
}

type ReservedEnvironmentsStorage interface {
	Add(ctx context.Context, env *aggregates.Environment) error
	List(ctx context.Context) (names []string, err error)
	IsReserved(ctx context.Context, envID uuid.UUID) (bool, error)
}

type ClientKeyGenerator interface {
	Generate() string
}

type PasswordVerifier interface {
	Verify(password, hash string) bool
}
