package environment

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"

	"github.com/google/uuid"
)

type service struct {
	envRepo         ports.EnvironmentRepository
	envPolicisRepo  ports.EnvironmentPoliciesRepository
	reservedStorage ports.ReservedEnvironmentsStorage
	publisher       *event.Publisher
}

func New(
	envRepo ports.EnvironmentRepository,
	reservedStorage ports.ReservedEnvironmentsStorage,
	envPolicisRepo ports.EnvironmentPoliciesRepository,
	publisher *event.Publisher,
) *service {
	return &service{
		envRepo:         envRepo,
		reservedStorage: reservedStorage,
		envPolicisRepo:  envPolicisRepo,
		publisher:       publisher,
	}
}

// CreateEnvironment — создаёт новое окружение
func (s *service) CreateEnvironment(ctx context.Context, input dto.CreateEnvironmentInput) (uuid.UUID, error) {
	vars := entities.NewVariables()
	if input.Variables != nil {
		for k, v := range input.Variables {
			key, err := vo.NewVariableKey(k)
			if err != nil {
				return uuid.Nil, errs.ErrInvalidVariableKey
			}
			vars[key] = vo.NewVariableValue(v)
		}
	}

	descr := ""
	if input.Description != nil {
		descr = *input.Description
	}

	env, err := ag.NewEnvironment(input.Name, descr, vars)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.envRepo.Save(ctx, env); err != nil {
		return uuid.Nil, err
	}
	return env.ID, nil
}

// GetAllEnvironments — получить спимок всех окружений, если reserved = true, то только те, которые используются клиентами
func (s *service) GetAllEnvironments(ctx context.Context, input dto.GetAllEnvironmentsInput) ([]*dto.EnvironmentDTO, error) {
	reserved, err := s.reservedStorage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get reserved environmentts: %w", err)
	}

	if input.Reserved {
		res := make([]*dto.EnvironmentDTO, len(reserved))
		for i, name := range reserved {
			env, err := s.getEnvironment(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("failed to get environmentt %q: %w", name, err)
			}
			res[i] = &dto.EnvironmentDTO{
				Name:        env.Name,
				Variables:   s.copyVariables(env.Variables()),
				Reserved:    true,
				Description: env.Description,
			}
		}
		return res, nil
	}

	reservedMap := make(map[string]struct{}, len(reserved))
	for _, v := range reserved {
		reservedMap[v] = struct{}{}
	}

	envs, err := s.envRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get all environmentts: %w", err)
	}

	res := make([]*dto.EnvironmentDTO, len(envs))
	for i, env := range envs {
		_, isReserved := reservedMap[env.Name]
		res[i] = &dto.EnvironmentDTO{
			Name:        env.Name,
			Variables:   s.copyVariables(env.Variables()),
			Reserved:    isReserved,
			Description: env.Description,
		}
	}
	return res, nil
}

func (s *service) GetEnvironmentPolicies(ctx context.Context, input dto.GetEnvironmentPoliciesInput) ([]*dto.PolicyDTO, error) {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return nil, nil
	}

	policies, err := s.envPolicisRepo.GetEnvironmentPolicies(ctx, env.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot get list of environmentt policies: %w", err)
	}

	dtos := make([]*dto.PolicyDTO, len(policies))
	for i, policy := range policies {
		dtos[i] = &dto.PolicyDTO{
			Name:           policy.Name,
			Key:            policy.Key,
			ChangesAllowed: env.CanBeChangedBy(policy.ID),
		}
	}
	return dtos, nil
}

// UpdateEnvironment — меняет имя/описание окружения
func (s *service) UpdateEnvironmentInfo(ctx context.Context, input dto.UpdateEnvironmentInfoInput) error {
	if input.NewName == nil && input.Description == nil {
		return nil
	}

	env, err := s.getEnvironment(ctx, input.OldName)
	if err != nil {
		return err
	}
	if err := s.isEnvironmentReserved(ctx, env.ID); err != nil {
		return err
	}

	upd := ports.EnvironmentInfoUpdate{
		Name: env.Name,
	}
	if input.NewName != nil && *input.NewName != "" {
		upd.Name = *input.NewName
	}
	if input.Description != nil && *input.Description != env.Description {
		env.Description = *input.Description
	}

	if err := s.envRepo.UpdateInfo(ctx, env.ID, upd); err != nil {
		return err
	}
	return nil
}

// DeleteEnvironment — удаляет окружение
func (s *service) DeleteEnvironment(ctx context.Context, input dto.DeleteEnvironmentInput) error {
	env, err := s.getEnvironment(ctx, input.Name)
	if err != nil {
		return err
	}

	if err := s.isEnvironmentReserved(ctx, env.ID); err != nil {
		return err
	}
	if err := s.envRepo.Delete(ctx, env.ID); err != nil {
		return err
	}

	return nil
}

func (s *service) getEnvironment(ctx context.Context, name string) (*ag.Environment, error) {
	env, err := s.envRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find environmentt by name: %w", err)
	}
	if env == nil {
		return nil, errs.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *service) isEnvironmentReserved(ctx context.Context, envID uuid.UUID) error {
	isReserved, err := s.reservedStorage.IsReserved(ctx, envID)
	if err != nil {
		return fmt.Errorf("cannot check reserved envirnomrnt")
	}
	if isReserved {
		return errs.ErrEnvironmentIsReserved
	}
	return nil
}

func (s *service) copyVariables(vars entities.Variables) map[string]string {
	res := make(map[string]string, len(vars))
	for k, v := range vars {
		res[k.String()] = v.String()
	}
	return res
}
