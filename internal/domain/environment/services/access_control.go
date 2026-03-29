package services

import (
	"context"
	ag "envmn/internal/domain/environment/aggregate"
	"envmn/internal/domain/environment/entities"
	"errors"
	"fmt"
)

var ErrInvalidAccessPolicyKey = errors.New("invalid access policy key")

type AccessControlService struct {
	policyStor AccessPolicyFinderSaver
	keyGen     KeyGenerator
}

func NewAccessControlService(policyStor AccessPolicyFinderSaver, keyGen KeyGenerator) *AccessControlService {
	return &AccessControlService{
		policyStor: policyStor,
		keyGen:     keyGen,
	}
}

func (s *AccessControlService) CreatePolicy(ctx context.Context, name string, changesAllowed bool) (*entities.AccessPolicy, error) {
	policy := entities.NewAccessPolicy(name, s.keyGen.Generate(), changesAllowed)
	if err := s.policyStor.Save(ctx, policy); err != nil {
		return nil, fmt.Errorf("cannot save policy: %w", err)
	}
	return policy, nil
}

// CanView — может ли клиент просматривать окружение
// Если к окружению нет привязанных политик — любой может смотреть
// Если политики есть — нужен валидный ключ
func (s *AccessControlService) CanView(ctx context.Context, env *ag.Environment, providedKeyValue string) (bool, error) {
	if providedKeyValue == "" {
		return env.AccessPoliciesCount() == 0, nil
	}

	policy, err := s.policyStor.FindByKey(ctx, providedKeyValue)
	if errors.Is(err, ErrAccessPolicyNotFound) {
		return false, ErrInvalidAccessPolicyKey
	} else if err != nil {
		return false, fmt.Errorf("cannot find access policy: %w", err)
	}

	return env.HasAccess(policy.ID), nil
}

// CanChange — может ли клиент изменять окружение - всегда требуется ключ
func (s *AccessControlService) CanChange(ctx context.Context, env *ag.Environment, providedKey string) (bool, *entities.AccessPolicy, error) {
	if providedKey == "" {
		return false, nil, nil
	}

	policy, err := s.policyStor.FindByKey(ctx, providedKey)
	if errors.Is(err, ErrAccessPolicyNotFound) {
		return false, nil, ErrInvalidAccessPolicyKey
	} else if err != nil {
		return false, nil, fmt.Errorf("cannot find access policy: %w", err)
	}

	if !policy.ChangesAllowed && env.CanBeChangedBy(policy.ID) {
		return false, policy, nil
	}

	return true, policy, nil
}
