package subscription_test

import (
	"context"
	"errors"
	"testing"

	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/mocks"
	"envmn/internal/service/subscription"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_SubscribeOnUpdates(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.SubscribeOnUpdatesInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage) error
		expectedKey string
		expectedErr error
	}{
		{
			name: "successful subscription",
			input: dto.SubscribeOnUpdatesInput{
				EnvironmentName: "prod-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, storage *mocks.MockReservedEnvironmentsStorage) error {
				env, _ := ag.NewEnvironment("prod-env", "production", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				storage.On("Add", mock.Anything, env).Return(nil).Once()
				return nil
			},
			expectedKey: "",
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.SubscribeOnUpdatesInput{
				EnvironmentName: "not-found",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, storage *mocks.MockReservedEnvironmentsStorage) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return errs.ErrEnvironmentNotFound
			},
			expectedKey: "",
			expectedErr: nil,
		},
		{
			name: "find environment error",
			input: dto.SubscribeOnUpdatesInput{
				EnvironmentName: "prod-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, storage *mocks.MockReservedEnvironmentsStorage) error {
				findErr := errors.New("db connection failed")
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(nil, findErr).Once()
				return findErr
			},
			expectedKey: "",
			expectedErr: nil,
		},
		{
			name: "reserve environment error",
			input: dto.SubscribeOnUpdatesInput{
				EnvironmentName: "prod-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, storage *mocks.MockReservedEnvironmentsStorage) error {
				env, _ := ag.NewEnvironment("prod-env", "production", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				reserveErr := errors.New("reserve failed")
				storage.On("Add", mock.Anything, env).Return(reserveErr).Once()
				return reserveErr
			},
			expectedKey: "",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			storage := &mocks.MockReservedEnvironmentsStorage{}
			keyGen := &mocks.MockClientKeyGenerator{}
			notifier := &mocks.MockNotifier{}
			publisher := event.NewPublisher()
			expectedErr := tc.mockSetup(envRepo, storage)

			// Setup key generator for successful cases
			if expectedErr == nil {
				keyGen.On("Generate").Return("key-123").Once()
			}

			service := subscription.New(keyGen, envRepo, storage, publisher, notifier)

			key, err := service.SubscribeOnUpdates(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
				assert.Equal(t, tc.expectedKey, key)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.expectedKey, key)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)
				assert.Contains(t, key, "prod-env")
			}

			envRepo.AssertExpectations(t)
			storage.AssertExpectations(t)
			keyGen.AssertExpectations(t)
		})
	}
}
