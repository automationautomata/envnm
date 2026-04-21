package subscription

import (
	"context"
	envevents "envmn/internal/domain/environment/events"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/ports"
	"fmt"
)

type subscription struct {
	keyGen    ports.ClientKeyGenerator
	envRepo   ports.EnvironmentRepository
	storage   ports.ReservedEnvironmentsStorage
	pub       *event.Publisher
	notifiler event.Notifier
}

func New(
	keyGen ports.ClientKeyGenerator,
	envRepo ports.EnvironmentRepository,
	storage ports.ReservedEnvironmentsStorage,
	pub *event.Publisher,
	notifiler event.Notifier,
) *subscription {
	return &subscription{
		keyGen:    keyGen,
		envRepo:   envRepo,
		storage:   storage,
		pub:       pub,
		notifiler: notifiler,
	}
}

func (s *subscription) SubscribeOnUpdates(ctx context.Context, input dto.SubscribeOnUpdatesInput) (key string, err error) {
	env, err := s.envRepo.FindByName(ctx, input.EnvironmentName)
	if err != nil {
		return "", fmt.Errorf("cannot find environment: %w", err)
	}
	if env == nil {
		return "", errs.ErrEnvironmentNotFound
	}

	if err = s.storage.Add(ctx, env); err != nil {
		return "", fmt.Errorf("cannot reserve environment: %w", err)
	}

	key = fmt.Sprintf("%s.%s", env.Name, s.keyGen.Generate())

	s.pub.Subscribe(s.notifiler, envevents.VariableEventsNames()...)

	return key, nil
}
