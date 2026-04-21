package policy_test

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
	"envmn/internal/service/policy"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func chooseError(e1, e2 error) error {
	if e1 != nil {
		return e1
	}
	return e2
}

func TestService_CreateAccessPolicy(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.CreateAccessPolicyInput
		mockSetup   func(*mocks.MockAccessControl) error
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "successful creation",
			input: dto.CreateAccessPolicyInput{
				Name: "admin-policy",
			},
			mockSetup: func(accessControl *mocks.MockAccessControl) error {
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "admin-policy"}
				accessControl.On("CreatePolicy", mock.Anything, "admin-policy").Return(policy, nil).Once()
				return nil
			},
			expectedID:  uuid.Nil,
			expectedErr: nil,
		},
		{
			name: "create policy error",
			input: dto.CreateAccessPolicyInput{
				Name: "admin-policy",
			},
			mockSetup: func(accessControl *mocks.MockAccessControl) error {
				createErr := errors.New("create policy failed")
				accessControl.On("CreatePolicy", mock.Anything, "admin-policy").Return(nil, createErr).Once()
				return createErr
			},
			expectedID:  uuid.Nil,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			policyRepo := &mocks.MockAccessPolicyRepository{}
			envPolicisRepo := &mocks.MockEnvironmentPoliciesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			expectedErr := tc.mockSetup(accessControl)

			service := policy.New(envRepo, policyRepo, envPolicisRepo, publisher, accessControl)

			id, err := service.CreateAccessPolicy(context.Background(), tc.input)

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

			accessControl.AssertExpectations(t)
		})
	}
}

func TestService_AddPolicyToEnvironment(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.AddPolicyToEnvironmentInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockAccessPolicyRepository, *mocks.MockEnvironmentPoliciesRepository) error
		expectedErr error
	}{
		{
			name: "successful add",
			input: dto.AddPolicyToEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{ID: uuid.New()}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				envPolicies.On("AddToEnvironment", mock.Anything, env.ID, policy).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.AddPolicyToEnvironmentInput{
				EnvironmentName: "not-found",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return nil
			},
			expectedErr: errs.ErrEnvironmentNotFound,
		},
		{
			name: "policy not found",
			input: dto.AddPolicyToEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, errs.ErrAccessPolicyNotFound).Once()
				return nil
			},
			expectedErr: errs.ErrAccessPolicyNotFound,
		},
		{
			name: "add to environment error",
			input: dto.AddPolicyToEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				addErr := errors.New("add failed")
				envPolicies.On("AddToEnvironment", mock.Anything, env.ID, policy).Return(addErr).Once()
				return addErr
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			policyRepo := &mocks.MockAccessPolicyRepository{}
			envPolicies := &mocks.MockEnvironmentPoliciesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := event.NewPublisher()
			setupErr := tc.mockSetup(envRepo, policyRepo, envPolicies)

			expectedErr := chooseError(setupErr, tc.expectedErr)

			service := policy.New(envRepo, policyRepo, envPolicies, publisher, accessControl)

			err := service.AddPolicyToEnvironment(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			policyRepo.AssertExpectations(t)
			envPolicies.AssertExpectations(t)
		})
	}
}

func TestService_RemovePolicyFromEnvironment(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.RemovePolicyFromEnvironmentInput
		mockSetup   func(*mocks.MockEnvironmentRepository, *mocks.MockAccessPolicyRepository, *mocks.MockEnvironmentPoliciesRepository) error
		expectedErr error
	}{
		{
			name: "successful remove",
			input: dto.RemovePolicyFromEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				envPolicies.On("DeleteFromEnvironment", mock.Anything, env.ID, policy.ID).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "environment not found",
			input: dto.RemovePolicyFromEnvironmentInput{
				EnvironmentName: "not-found",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				envRepo.On("FindByName", mock.Anything, "not-found").Return(nil, errs.ErrEnvironmentNotFound).Once()
				return nil
			},
			expectedErr: errs.ErrEnvironmentNotFound,
		},
		{
			name: "policy not found",
			input: dto.RemovePolicyFromEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, errs.ErrAccessPolicyNotFound).Once()
				return nil
			},
			expectedErr: errs.ErrAccessPolicyNotFound,
		},
		{
			name: "delete from environment error",
			input: dto.RemovePolicyFromEnvironmentInput{
				EnvironmentName: "test-env",
				PolicyID:        uuid.New(),
			},
			mockSetup: func(envRepo *mocks.MockEnvironmentRepository, policyRepo *mocks.MockAccessPolicyRepository, envPolicies *mocks.MockEnvironmentPoliciesRepository) error {
				env := &ag.Environment{}
				env.Name = "test-env"
				envRepo.On("FindByName", mock.Anything, "test-env").Return(env, nil).Once()
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				deleteErr := errors.New("delete failed")
				envPolicies.On("DeleteFromEnvironment", mock.Anything, env.ID, policy.ID).Return(deleteErr).Once()
				return deleteErr
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			policyRepo := &mocks.MockAccessPolicyRepository{}
			envPolicies := &mocks.MockEnvironmentPoliciesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			setupErr := tc.mockSetup(envRepo, policyRepo, envPolicies)

			expectedErr := chooseError(setupErr, tc.expectedErr)

			service := policy.New(envRepo, policyRepo, envPolicies, publisher, accessControl)

			err := service.RemovePolicyFromEnvironment(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else {
				assert.NoError(t, err)
			}

			envRepo.AssertExpectations(t)
			policyRepo.AssertExpectations(t)
			envPolicies.AssertExpectations(t)
		})
	}
}

func TestService_RemovePolicy(t *testing.T) {
	testCases := []struct {
		name        string
		input       dto.RemovePolicyInput
		mockSetup   func(*mocks.MockAccessPolicyRepository) error
		expectedErr error
	}{
		{
			name: "successful remove",
			input: dto.RemovePolicyInput{
				ID: uuid.New(),
			},
			mockSetup: func(policyRepo *mocks.MockAccessPolicyRepository) error {
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				policyRepo.On("Delete", mock.Anything, policy.ID).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "policy not found",
			input: dto.RemovePolicyInput{
				ID: uuid.New(),
			},
			mockSetup: func(policyRepo *mocks.MockAccessPolicyRepository) error {
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, errs.ErrAccessPolicyNotFound).Once()
				return nil
			},
			expectedErr: errs.ErrAccessPolicyNotFound,
		},
		{
			name: "delete error",
			input: dto.RemovePolicyInput{
				ID: uuid.New(),
			},
			mockSetup: func(policyRepo *mocks.MockAccessPolicyRepository) error {
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "policy"}
				policyRepo.On("FindByID", mock.Anything, mock.Anything).Return(policy, nil).Once()
				deleteErr := errors.New("delete failed")
				policyRepo.On("Delete", mock.Anything, policy.ID).Return(deleteErr).Once()
				return deleteErr
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envRepo := &mocks.MockEnvironmentRepository{}
			policyRepo := &mocks.MockAccessPolicyRepository{}
			envPoliciesRepo := &mocks.MockEnvironmentPoliciesRepository{}
			accessControl := &mocks.MockAccessControl{}
			publisher := &event.Publisher{}
			setupErr := tc.mockSetup(policyRepo)

			expectedErr := chooseError(setupErr, tc.expectedErr)
			service := policy.New(envRepo, policyRepo, envPoliciesRepo, publisher, accessControl)

			err := service.RemovePolicy(context.Background(), tc.input)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else {
				assert.NoError(t, err)
			}

			policyRepo.AssertExpectations(t)
		})
	}
}
