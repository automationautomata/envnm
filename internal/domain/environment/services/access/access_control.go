package access

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	serverrors "envmn/internal/domain/environment/services/errors"
	"envmn/internal/domain/environment/services/ports"

	"fmt"
)

type accessControlService struct {
	policyStor ports.AccessPolicyFinderSaver
	keyGen     ports.KeyGenerator
}

func New(policyStor ports.AccessPolicyFinderSaver, keyGen ports.KeyGenerator) *accessControlService {
	return &accessControlService{
		policyStor: policyStor,
		keyGen:     keyGen,
	}
}

func (s *accessControlService) CreatePolicy(ctx context.Context, name string) (*entities.AccessPolicy, error) {
	policy := entities.NewAccessPolicy(name, s.keyGen.Generate())
	if err := s.policyStor.Save(ctx, policy); err != nil {
		return nil, fmt.Errorf("cannot save policy: %w", err)
	}
	return policy, nil
}

// CanView - может ли клиент просматривать окружение
// Если к окружению нет привязанных политик - любой может смотреть
// Если политики есть - нужен валидный ключ
func (s *accessControlService) CanView(ctx context.Context, env *ag.Environment, providedKey *string) (bool, error) {
	if env.AccessPoliciesCount() == 0 {
		return true, nil
	}

	if providedKey == nil {
		return false, nil
	}

	policy, err := s.policyStor.FindByKey(ctx, *providedKey)
	if err != nil {
		return false, fmt.Errorf("cannot find access policy: %w", err)
	}

	if policy == nil {
		return false, serverrors.ErrInvalidAccessKey
	}

	return env.HasAccess(policy.ID), nil
}

// CanChange - может ли клиент изменять окружение - всегда требуется ключ
func (s *accessControlService) CanChange(ctx context.Context, env *ag.Environment, providedKey *string) (bool, error) {
	if providedKey == nil {
		return false, nil
	}

	policy, err := s.policyStor.FindByKey(ctx, *providedKey)
	if err != nil {
		return false, fmt.Errorf("cannot find access policy: %w", err)
	}

	if policy == nil {
		return false, serverrors.ErrInvalidAccessKey
	}

	canChange := env.CanBeChangedBy(policy.ID)
	return canChange, nil
}
