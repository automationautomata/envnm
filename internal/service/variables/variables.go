package variables

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	envsvc "envmn/internal/domain/environment/services"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"
)

type service struct {
	envRepo       ports.EnvironmentRepository
	varsRepo      ports.EnvironmentVariablesRepository
	pub           *event.Publisher
	accessControl envsvc.AccessControl
}

func New(
	envRepo ports.EnvironmentRepository,
	varsRepo ports.EnvironmentVariablesRepository,
	pub *event.Publisher,
	accessControl envsvc.AccessControl,
) *service {
	return &service{
		envRepo:       envRepo,
		varsRepo:      varsRepo,
		pub:           pub,
		accessControl: accessControl,
	}
}

func (s *service) GetClientVariables(ctx context.Context, input dto.GetClientVariablesInput) (map[string]string, error) {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string)
	for k, v := range env.Variables() {
		vars[k.String()] = v.String()
	}
	return vars, nil
}

// UpdateEnvironmentVariables — обновляет (добавляет/изменяет) переменные
func (s *service) UpdateEnvironmentVariables(ctx context.Context, input dto.UpdateEnvironmentVariablesInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	canView, err := s.accessControl.CanView(ctx, env, input.AccessKey)
	if err != nil {
		return fmt.Errorf("cannot check client access to view: %w", err)
	}
	if !canView {
		return errs.ErrAccessDenied
	}

	return s.updateVariables(ctx, env, input.Variables)
}

// UpdateVariablesByClient — обновляет (добавляет/изменяет) переменные
func (s *service) UpdateVariablesByClient(ctx context.Context, input dto.UpdateVariablesByClientInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	canChange, err := s.accessControl.CanChange(ctx, env, &input.AccessKey)
	if err != nil {
		return fmt.Errorf("cannot check client access to change: %w", err)
	}
	if !canChange {
		return errs.ErrAccessDenied
	}
	return s.updateVariables(ctx, env, input.Variables)
}

// RemoveVariableFromEnvironment — удаляет переменную из окружения
func (s *service) RemoveVariableFromEnvironment(ctx context.Context, input dto.RemoveVariableFromEnvironmentInput) error {
	env, err := s.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	key, err := vo.NewVariableKey(input.VariableKey)
	if err != nil {
		return errs.ErrInvalidVariableKey
	}

	if err := s.varsRepo.DeleteVariable(ctx, env.ID, key); err != nil {
		return fmt.Errorf("cannot delete variable: %w", err)
	}

	for _, e := range env.PullEvents() {
		s.pub.Publish(ctx, e)
	}
	return nil
}

func (s *service) getEnvironment(ctx context.Context, name string) (*ag.Environment, error) {
	env, err := s.envRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find environment: %w", err)
	}
	if env == nil {
		return nil, errs.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *service) updateVariables(ctx context.Context, env *ag.Environment, variables map[string]string) error {
	envVars := entities.NewVariables()
	for k, v := range variables {
		key, err := vo.NewVariableKey(k)
		if err != nil {
			return errs.ErrInvalidVariableKey
		}
		envVars[key] = vo.NewVariableValue(v)
	}

	_, _ = env.UpdateVariables(envVars)
	if err := s.varsRepo.UpdateVariables(ctx, env); err != nil {
		return fmt.Errorf("cannot update variables: %w", err)
	}

	ctx = context.WithValue(ctx, ports.EnvironmentNameContextKey, env.Name)
	for _, e := range env.PullEvents() {
		s.pub.Publish(ctx, e)
	}
	return nil
}
