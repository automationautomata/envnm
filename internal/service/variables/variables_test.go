package variables_test

import (
	"context"
	"errors"
	"testing"

	ag "envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"envmn/internal/domain/event"
	"envmn/internal/service/dto"
	errs "envmn/internal/service/errors"
	"envmn/internal/service/mocks"
	"envmn/internal/service/variables"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func stringPtr(s string) *string {
	return &s
}

func TestService_GetClientVariables(t *testing.T) {
	testCases := []struct {
		name         string
		input        dto.GetClientVariablesInput
		mockSetup    func(*mocks.MockEnvironmentRepository) error
		expectedVars map[string]string
		expectedErr  error
	}{
		{
			name: "successful get variables",
			input: dto.GetClientVariablesInput{
				EnvironmentName: "prod-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository) error {
				vars := entities.NewVariables()
				key, _ := vo.NewVariableKey("API_KEY")
				vars[key] = vo.NewVariableValue("secret123")
				env, _ := ag.NewEnvironment("prod-env", "prod", vars)
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				return nil
			},
			expectedVars: map[string]string{"API_KEY": "secret123"},
			expectedErr:  nil,
		},
		{
			name: "environment not found",
			input: dto.GetClientVariablesInput{
				EnvironmentName: "not-found",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return errs.ErrEnvironmentNotFound
			},
			expectedVars: nil,
			expectedErr:  nil,
		},
		{
			name: "find error",
			input: dto.GetClientVariablesInput{
				EnvironmentName: "prod-env",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository) error {
				findErr := errors.New("db error")
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(nil, findErr).Once()
				return findErr
			},
			expectedVars: nil,
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			varsRepo := &mocks.MockEnvironmentVariablesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			expectedErr := tc.mockSetup(envRepo)

			service := variables.New(envRepo, varsRepo, publisher, accessControl)

			vars, err := service.GetClientVariables(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
				assert.Nil(t, vars)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, vars)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVars, vars)
			}

			envRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateEnvironmentVariables(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.UpdateEnvironmentVariablesInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockEnvironmentVariablesRepository, *mocks.MockAccessControl) error
		expectedErr error
	}{
		{
			name: "successful update with view access",
			input: dto.UpdateEnvironmentVariablesInput{
				EnvironmentName: "prod-env",
				Variables:       map[string]string{"API_KEY": "new-key"},
				AccessKey:       stringPtr("valid-key"),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				accessControl.On("CanView", mock.Anything, env, stringPtr("valid-key")).Return(true, nil).Once()
				varsRepo.On("UpdateVariables", mock.Anything, mock.Anything).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.UpdateEnvironmentVariablesInput{
				EnvironmentName: "not-found",
				Variables:       map[string]string{"API_KEY": "key"},
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return errs.ErrEnvironmentNotFound
			},
			expectedErr: nil,
		},
		{
			name: "invalid variable key",
			input: dto.UpdateEnvironmentVariablesInput{
				EnvironmentName: "prod-env",
				Variables:       map[string]string{"invalid key": "value"},
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				accessControl.On("CanView", mock.Anything, env, mock.Anything).Return(true, nil).Once()
				return errs.ErrInvalidVariableKey
			},
			expectedErr: nil,
		},
		{
			name: "access denied",
			input: dto.UpdateEnvironmentVariablesInput{
				EnvironmentName: "prod-env",
				Variables:       map[string]string{"API_KEY": "key"},
				AccessKey:       stringPtr("invalid-key"),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				accessControl.On("CanView", mock.Anything, env, stringPtr("invalid-key")).Return(false, nil).Once()
				return errs.ErrAccessDenied
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			varsRepo := &mocks.MockEnvironmentVariablesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			expectedErr := tc.mockSetup(envRepo, varsRepo, accessControl)

			service := variables.New(envRepo, varsRepo, publisher, accessControl)

			err := service.UpdateEnvironmentVariables(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			varsRepo.AssertExpectations(t)
			accessControl.AssertExpectations(t)
		})
	}
}

func TestService_UpdateVariablesByClient(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.UpdateVariablesByClientInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockEnvironmentVariablesRepository, *mocks.MockAccessControl) error
		expectedErr error
	}{
		{
			name: "successful update with change access",
			input: dto.UpdateVariablesByClientInput{
				EnvironmentName: "prod-env",
				Variables:       map[string]string{"API_KEY": "new-key"},
				AccessKey:       "valid-key",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				accessControl.On("CanChange", mock.Anything, env, stringPtr("valid-key")).Return(true, nil).Once()
				varsRepo.On("UpdateVariables", mock.Anything, mock.Anything).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.UpdateVariablesByClientInput{
				EnvironmentName: "not-found",
				Variables:       map[string]string{"API_KEY": "key"},
				AccessKey:       "key",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return errs.ErrEnvironmentNotFound
			},
			expectedErr: nil,
		},
		{
			name: "access denied",
			input: dto.UpdateVariablesByClientInput{
				EnvironmentName: "prod-env",
				Variables:       map[string]string{"API_KEY": "key"},
				AccessKey:       "invalid-key",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository, accessControl *mocks.MockAccessControl) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				accessControl.On("CanChange", mock.Anything, env, stringPtr("invalid-key")).Return(false, nil).Once()
				return errs.ErrAccessDenied
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			varsRepo := &mocks.MockEnvironmentVariablesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			expectedErr := tc.mockSetup(envRepo, varsRepo, accessControl)

			service := variables.New(envRepo, varsRepo, publisher, accessControl)

			err := service.UpdateVariablesByClient(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			varsRepo.AssertExpectations(t)
			accessControl.AssertExpectations(t)
		})
	}
}

func TestService_RemoveVariableFromEnvironment(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.RemoveVariableFromEnvironmentInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockEnvironmentVariablesRepository) error
		expectedErr error
	}{
		{
			name: "successful remove",
			input: dto.RemoveVariableFromEnvironmentInput{
				EnvironmentName: "prod-env",
				VariableKey:     "API_KEY",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				varsRepo.On("DeleteVariable", mock.Anything, env.ID, mock.Anything).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.RemoveVariableFromEnvironmentInput{
				EnvironmentName: "not-found",
				VariableKey:     "API_KEY",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return errs.ErrEnvironmentNotFound
			},
			expectedErr: nil,
		},
		{
			name: "invalid variable key",
			input: dto.RemoveVariableFromEnvironmentInput{
				EnvironmentName: "prod-env",
				VariableKey:     "invalid key",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				return errs.ErrInvalidVariableKey
			},
			expectedErr: nil,
		},
		{
			name: "delete error",
			input: dto.RemoveVariableFromEnvironmentInput{
				EnvironmentName: "prod-env",
				VariableKey:     "API_KEY",
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, varsRepo *mocks.MockEnvironmentVariablesRepository) error {
				env, _ := ag.NewEnvironment("prod-env", "prod", entities.NewVariables())
				envRepo.On("FindByName", mock.Anything, "prod-env").Return(env, nil).Once()
				deleteErr := errors.New("delete failed")
				varsRepo.On("DeleteVariable", mock.Anything, env.ID, mock.Anything).Return(deleteErr).Once()
				return deleteErr
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			varsRepo := &mocks.MockEnvironmentVariablesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			expectedErr := tc.mockSetup(envRepo, varsRepo)

			service := variables.New(envRepo, varsRepo, publisher, accessControl)

			err := service.RemoveVariableFromEnvironment(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			varsRepo.AssertExpectations(t)
		})
	}
}
