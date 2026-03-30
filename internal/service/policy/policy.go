package policy

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/environment/services"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"

	"github.com/google/uuid"
)

type policy struct {
	envRepo         ports.EnvironmentRepository
	policyRepo      ports.AccessPolicyRepository
	envPoliciesRepo ports.EnvironmentPoliciesRepository
	publisher       *event.EventPublisher
	accessControl   *services.AccessControlService
}

func New(
	envRepo ports.EnvironmentRepository,
	policyRepo ports.AccessPolicyRepository,
	publisher *event.EventPublisher,
	accessControl *services.AccessControlService,
) *policy {
	return &policy{
		envRepo:       envRepo,
		policyRepo:    policyRepo,
		publisher:     publisher,
		accessControl: accessControl,
	}
}

func (s *policy) CreateAccessPolicy(ctx context.Context, input dto.CreateAccessPolicyInput) (uuid.UUID, error) {
	policy, err := s.accessControl.CreatePolicy(ctx, input.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot create access policy: %w", err)
	}
	return policy.ID, nil
}

func (s *policy) AddPolicyToEnvironment(ctx context.Context, input dto.AddPolicyToEnvironmentInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	policy, err := s.getPolicy(ctx, input.PolicyID)
	if err != nil {
		return err
	}

	if err = s.envPoliciesRepo.AddToEnvironment(ctx, env.ID, policy); err != nil {
		return fmt.Errorf("cannot add policy to environment: %w", err)
	}
	return nil
}

func (s *policy) RemovePolicyFromEnvironment(ctx context.Context, input dto.RemovePolicyFromEnvironmentInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	policy, err := s.getPolicy(ctx, input.PolicyID)
	if err != nil {
		return err
	}

	if err := s.envPoliciesRepo.DeleteFromEnvironment(ctx, env.ID, policy.ID); err != nil {
		return fmt.Errorf("cannot add policy to environment: %w", err)
	}
	return nil
}

func (s *policy) RemovePolicy(ctx context.Context, input dto.RemovePolicyInput) error {
	policy, err := s.getPolicy(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := s.policyRepo.Delete(ctx, policy.ID); err != nil {
		return fmt.Errorf("cannot add policy to environment: %w", err)
	}
	return nil
}

func (s *policy) getEnvironment(ctx context.Context, name string) (*ag.Environment, error) {
	env, err := s.envRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find environment by name: %w", err)
	}
	if env == nil {
		return nil, errs.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *policy) getPolicy(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	policy, err := s.policyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cannot find policy by id: %w", err)
	}
	if policy == nil {
		return nil, errs.ErrAccessPolicyNotFound
	}
	return policy, nil
}
