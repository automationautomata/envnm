package storages

import (
	"context"
	"envmn/internal/domain/environment/aggregates"
	"sync"

	"github.com/google/uuid"
)

type inmemoryEnvironmentsStorage struct {
	environments sync.Map // т.к. мало операций записи и много операций чтения
}

func NewInmemoryEnvironmentsStorage() *inmemoryEnvironmentsStorage {
	return &inmemoryEnvironmentsStorage{}
}

func (s *inmemoryEnvironmentsStorage) Add(ctx context.Context, env *aggregates.Environment) error {
	s.environments.Store(env.ID, env.Name)
	return nil
}

func (s *inmemoryEnvironmentsStorage) List(ctx context.Context) (names []string, err error) {
	names = make([]string, 0)
	s.environments.Range(func(key any, value any) bool {
		names = append(names, value.(string))
		return true
	})
	return names, nil
}

func (s *inmemoryEnvironmentsStorage) IsReserved(ctx context.Context, envID uuid.UUID) (bool, error) {
	_, ok := s.environments.Load(envID)
	return ok, nil
}
