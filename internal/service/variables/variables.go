package variables

import (
	"context"
	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/environment/services"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"
)

type useCase struct {
	envRepo       ports.EnvironmentRepository
	varsRepo      ports.EnvironmentVariablesRepository
	pub           *event.EventPublisher
	accessControl *services.AccessControlService
}

func New(
	envRepo ports.EnvironmentRepository,
	varsRepo ports.EnvironmentVariablesRepository,
	pub *event.EventPublisher,
	accessControl *services.AccessControlService,
) *useCase {
	return &useCase{
		envRepo:       envRepo,
		varsRepo:      varsRepo,
		pub:           pub,
		accessControl: accessControl,
	}
}

func (uc *useCase) GetClientVariables(ctx context.Context, input dto.GetClientVariablesInput) (map[string]string, error) {
	env, err := uc.getEnvironment(ctx, input.EnvironmentName)
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
func (uc *useCase) UpdateEnvironmentVariables(ctx context.Context, input dto.UpdateEnvironmentVariablesInput) error {
	env, err := uc.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	canView, err := uc.accessControl.CanView(ctx, env, input.AccessKey)
	if err != nil {
		return fmt.Errorf("cannot check dtoient access to view: %w", err)
	}
	if !canView {
		return errs.ErrAccessDenied
	}

	return uc.updateVariables(ctx, env, input.Variables)
}

// UpdateVariablesByClient — обновляет (добавляет/изменяет) переменные
func (uc *useCase) UpdateVariablesByClient(ctx context.Context, input dto.UpdateVariablesByClientInput) error {
	env, err := uc.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	canChange, _, err := uc.accessControl.CanChange(ctx, env, input.AccessKey)
	if err != nil {
		return fmt.Errorf("cannot check dtoient access to change: %w", err)
	}
	if !canChange {
		return errs.ErrAccessDenied
	}
	return uc.updateVariables(ctx, env, input.Variables)
}

// RemoveVariableFromEnvironment — удаляет переменную из окружения
func (uc *useCase) RemoveVariableFromEnvironment(ctx context.Context, input dto.RemoveVariableFromEnvironmentInput) error {
	env, err := uc.getEnvironment(ctx, input.EnvironmentName)
	if err != nil {
		return err
	}

	key, err := vo.NewVariableKey(input.VariableKey)
	if err != nil {
		return errs.ErrInvalidVariableKey
	}

	if err := uc.varsRepo.DeleteVariable(ctx, env.ID, key); err != nil {
		return fmt.Errorf("cannot delete variable: %w", err)
	}

	for _, e := range env.PullEvents() {
		uc.pub.Publish(ctx, e)
	}
	return nil
}

func (uc *useCase) getEnvironment(ctx context.Context, name string) (*ag.Environment, error) {
	env, err := uc.envRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find environment: %w", err)
	}
	if env == nil {
		return nil, errs.ErrEnvironmentNotFound
	}
	return env, nil
}

func (uc *useCase) updateVariables(ctx context.Context, env *ag.Environment, variables map[string]string) error {
	envVars := entities.NewVariables()
	for k, v := range variables {
		key, err := vo.NewVariableKey(k)
		if err != nil {
			return errs.ErrInvalidVariableKey
		}
		envVars[key] = vo.NewVariableValue(v)
	}

	_, _ = env.UpdateVariables(envVars)
	if err := uc.varsRepo.UpdateVariables(ctx, env); err != nil {
		return fmt.Errorf("cannot update variables: %w", err)
	}

	ctx = context.WithValue(ctx, ports.EnvironmentNameContextKey, env.Name)
	for _, e := range env.PullEvents() {
		uc.pub.Publish(ctx, e)
	}
	return nil
}
