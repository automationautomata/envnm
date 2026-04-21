package environment_test

import (
	"context"
	"errors"
	"testing"

	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	"envmn/internal/service/environment"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/mocks"
	"envmn/internal/service/ports"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateEnvironment(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.CreateEnvironmentInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage, *mocks.MockEnvironmentPoliciesRepository) error
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "successful creation",
			input: dto.CreateEnvironmentInput{
				Name:        "test-env",
				Description: stringPtr("test description"),
				Variables:   nil,
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) error {
				envRepo.On("Save", mock.Anything, mock.MatchedBy(func(env *ag.Environment) bool {
					return env.Name == "test-env" && env.Description == "test description"
				})).Return(nil).Once()
				return nil
			},
			expectedID:  uuid.Nil, // will check not nil
			expectedErr: nil,
		},
		{
			name: "invalid variable key",
			input: dto.CreateEnvironmentInput{
				Name:      "test-env",
				Variables: map[string]string{"invalid key": "value"},
			},
			mockSetup: func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage, *mocks.MockEnvironmentPoliciesRepository) error {
				return nil
			},
			expectedID:  uuid.Nil,
			expectedErr: errs.ErrInvalidVariableKey,
		},
		{
			name: "save error",
			input: dto.CreateEnvironmentInput{
				Name: "test-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) error {
				saveErr := errors.New("save failed")
				envRepo.On("Save", mock.Anything, mock.Anything).Return(saveErr).Once()
				return saveErr
			},
			expectedID:  uuid.Nil,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			reserved := &mocks.MockReservedEnvironmentsStorage{}
			policies := &mocks.MockEnvironmentPoliciesRepository{}
			publisher := &event.Publisher{} // or mock if needed
			expectedErr := tc.mockSetup(envRepo, reserved, policies)

			service := environment.New(envRepo, reserved, policies, publisher)

			id, err := service.CreateEnvironment(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
				assert.Equal(t, tc.expectedID, id)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.expectedID, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}

			envRepo.AssertExpectations(t)
			reserved.AssertExpectations(t)
			policies.AssertExpectations(t)
		})
	}
}

func TestService_GetAllEnvironments(t *testing.T) {
	listFailedErr := errors.New("list failed")

	testCases := []struct {
		name        string
		input       dto.GetAllEnvironmentsInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage, *mocks.MockEnvironmentPoliciesRepository)
		expected    []*dto.EnvironmentDTO
		expectedErr error
	}{
		{
			name:  "get all non-reserved",
			input: dto.GetAllEnvironmentsInput{Reserved: false},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				reserved.On("List", mock.Anything).Return([]string{"reserved-env"}, nil).Once()
				env, _ := ag.NewEnvironment("env1", "desc1", entities.NewVariables())
				envRepo.On("List", mock.Anything).Return([]*ag.Environment{env}, nil).Once()
			},
			expected: []*dto.EnvironmentDTO{
				{Name: "env1", Variables: map[string]string{}, Reserved: false, Description: "desc1"},
			},
			expectedErr: nil,
		},
		{
			name:  "get reserved only",
			input: dto.GetAllEnvironmentsInput{Reserved: true},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				reserved.On("List", mock.Anything).Return([]string{"reserved-env"}, nil).Once()
				env, _ := ag.NewEnvironment("reserved-env", "desc", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "reserved-env").Return(env, nil).Once()
			},
			expected: []*dto.EnvironmentDTO{
				{Name: "reserved-env", Variables: map[string]string{}, Reserved: true, Description: "desc"},
			},
			expectedErr: nil,
		},
		{
			name:  "list reserved error",
			input: dto.GetAllEnvironmentsInput{Reserved: true},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				reserved.On("List", mock.Anything).Return(nil, listFailedErr).Once()
			},
			expected:    nil,
			expectedErr: listFailedErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			reserved := &mocks.MockReservedEnvironmentsStorage{}
			policies := &mocks.MockEnvironmentPoliciesRepository{}
			publisher := &event.Publisher{}
			tc.mockSetup(envRepo, reserved, policies)

			service := environment.New(envRepo, reserved, policies, publisher)

			result, err := service.GetAllEnvironments(context.Background(), tc.input)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			envRepo.AssertExpectations(t)
			reserved.AssertExpectations(t)
			policies.AssertExpectations(t)
		})
	}
}

func TestService_UpdateEnvironmentInfo(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.UpdateEnvironmentInfoInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage, *mocks.MockEnvironmentPoliciesRepository)
		expectedErr error
	}{
		{
			name: "successful update name",
			input: dto.UpdateEnvironmentInfoInput{
				OldName: "old-name",
				NewName: stringPtr("new-name"),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				env := &ag.Environment{}
				env.Name = "old-name"
				env.Description = "desc"
				envRepo.On("FindByName", mock.Anything, "old-name").Return(env, nil).Once()
				reserved.On("IsReserved", mock.Anything, env.ID).Return(false, nil).Once()
				envRepo.On("UpdateInfo", mock.Anything, env.ID, ports.EnvironmentInfoUpdate{Name: "new-name"}).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.UpdateEnvironmentInfoInput{
				OldName: "not-found",
				NewName: stringPtr("new-name"),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
			},
			expectedErr: errs.ErrEnvironmentNotFound,
		},
		{
			name: "environment is reserved",
			input: dto.UpdateEnvironmentInfoInput{
				OldName: "reserved",
				NewName: stringPtr("new-name"),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				env := &ag.Environment{}
				env.Name = "reserved"
				envRepo.On("FindByName", mock.Anything, "reserved").Return(env, nil).Once()
				reserved.On("IsReserved", mock.Anything, env.ID).Return(true, nil).Once()
			},
			expectedErr: errs.ErrEnvironmentIsReserved,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			reserved := &mocks.MockReservedEnvironmentsStorage{}
			policies := &mocks.MockEnvironmentPoliciesRepository{}
			publisher := &event.Publisher{}
			tc.mockSetup(envRepo, reserved, policies)

			service := environment.New(envRepo, reserved, policies, publisher)

			err := service.UpdateEnvironmentInfo(context.Background(), tc.input)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			reserved.AssertExpectations(t)
			policies.AssertExpectations(t)
		})
	}
}

func TestService_DeleteEnvironment(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.DeleteEnvironmentInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockReservedEnvironmentsStorage, *mocks.MockEnvironmentPoliciesRepository)
		expectedErr error
	}{
		{
			name:  "successful delete",
			input: dto.DeleteEnvironmentInput{Name: "env-to-delete"},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				env := &ag.Environment{}
				env.Name = "env-to-delete"
				envRepo.On("FindByName", mock.Anything, "env-to-delete").Return(env, nil).Once()
				reserved.On("IsReserved", mock.Anything, env.ID).Return(false, nil).Once()
				envRepo.On("Delete", mock.Anything, env.ID).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			name:  "environment not found",
			input: dto.DeleteEnvironmentInput{Name: "not-found"},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
			},
			expectedErr: errs.ErrEnvironmentNotFound,
		},
		{
			name:  "environment is reserved",
			input: dto.DeleteEnvironmentInput{Name: "reserved"},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, reserved *mocks.MockReservedEnvironmentsStorage, policies *mocks.MockEnvironmentPoliciesRepository) {
				env := &ag.Environment{}
				env.Name = "reserved"
				envRepo.On("FindByName", mock.Anything, "reserved").Return(env, nil).Once()
				reserved.On("IsReserved", mock.Anything, env.ID).Return(true, nil).Once()
			},
			expectedErr: errs.ErrEnvironmentIsReserved,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			reserved := &mocks.MockReservedEnvironmentsStorage{}
			policies := &mocks.MockEnvironmentPoliciesRepository{}
			publisher := &event.Publisher{}
			tc.mockSetup(envRepo, reserved, policies)

			service := environment.New(envRepo, reserved, policies, publisher)

			err := service.DeleteEnvironment(context.Background(), tc.input)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			reserved.AssertExpectations(t)
			policies.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
