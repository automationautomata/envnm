package policy

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	svcs "envmn/internal/domain/environment/services"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"

	"github.com/google/uuid"
)

type service struct {
	envRepo         ports.EnvironmentRepository
	policyRepo      ports.AccessPolicyRepository
	envPoliciesRepo ports.EnvironmentPoliciesRepository
	publisher       *event.Publisher
	accessControl   svcs.AccessControl
}

func New(
	envRepo ports.EnvironmentRepository,
	policyRepo ports.AccessPolicyRepository,
	envPoliciesRepo ports.EnvironmentPoliciesRepository,
	publisher *event.Publisher,
	accessControl svcs.AccessControl,
) *service {
	return &service{
		envRepo:         envRepo,
		policyRepo:      policyRepo,
		publisher:       publisher,
		envPoliciesRepo: envPoliciesRepo,
		accessControl:   accessControl,
	}
}

func (s *service) CreateAccessPolicy(ctx context.Context, input dto.CreateAccessPolicyInput) (uuid.UUID, error) {
	policy, err := s.accessControl.CreatePolicy(ctx, input.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot create access policy: %w", err)
	}
	return policy.ID, nil
}

func (s *service) AddPolicyToEnvironment(ctx context.Context, input dto.AddPolicyToEnvironmentInput) error {
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

func (s *service) ListPolicyEnvironments(ctx context.Context, input dto.ListPolicyEnvironments) ([]*dto.PolicyEnvironmentDTO, error) {
	envs, err := s.envPoliciesRepo.ListPolicyEnvironments(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.PolicyEnvironmentDTO, len(envs))
	for i := range envs {
		dtos[i] = &dto.PolicyEnvironmentDTO{
			Name:           envs[i].Name,
			ChangesAllowed: envs[i].ChangesAllowed,
		}
	}
	return dtos, nil
}

func (s *service) RemovePolicyFromEnvironment(ctx context.Context, input dto.RemovePolicyFromEnvironmentInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	policy, err := s.getPolicy(ctx, input.PolicyID)
	if err != nil {
		return err
	}

	if err := s.envPoliciesRepo.DeleteFromEnvironment(ctx, env.ID, policy.ID); err != nil {
		return fmt.Errorf("cannot remove policy from environment: %w", err)
	}
	return nil
}

func (s *service) RemovePolicy(ctx context.Context, input dto.RemovePolicyInput) error {
	policy, err := s.getPolicy(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := s.policyRepo.Delete(ctx, policy.ID); err != nil {
		return fmt.Errorf("cannot remove policy: %w", err)
	}
	return nil
}

func (s *service) getEnvironment(ctx context.Context, name string) (*ag.Environment, error) {
	env, err := s.envRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find environment by name: %w", err)
	}
	if env == nil {
		return nil, errs.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *service) getPolicy(ctx context.Context, id uuid.UUID) (*entities.AccessPolicy, error) {
	policy, err := s.policyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cannot find policy by id: %w", err)
	}
	if policy == nil {
		return nil, errs.ErrAccessPolicyNotFound
	}
	return policy, nil
}
