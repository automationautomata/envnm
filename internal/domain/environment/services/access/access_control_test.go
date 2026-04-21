package access_test

import (
	"context"
	"errors"
	"testing"

	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	"envmn/internal/domain/environment/services/access"
	serverrors "envmn/internal/domain/environment/services/errors"
	"envmn/internal/domain/environment/services/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccessControlService_CreatePolicy(t *testing.T) {
	testCases := []struct {
		name        string
		nameInput   string
		mockSetup   func(*mocks.MockAccessPolicyFinderSaver, *mocks.MockKeyGenerator) error
		expectedErr error
	}{
		{
			name:      "success",
			nameInput: "test-policy",
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, mockKeyGen *mocks.MockKeyGenerator) error {
				mockKeyGen.On("Generate").Return("generated-key").Once()
				mockPolicyStor.On("Save", mock.Anything, mock.AnythingOfType("*entities.AccessPolicy")).Return(nil).Once()
				return nil
			},
			expectedErr: nil,
		},
		{
			name:      "save error",
			nameInput: "test-policy",
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, mockKeyGen *mocks.MockKeyGenerator) error {
				mockKeyGen.On("Generate").Return("generated-key").Once()
				saveErr := errors.New("save failed")
				mockPolicyStor.On("Save", mock.Anything, mock.AnythingOfType("*entities.AccessPolicy")).Return(saveErr).Once()
				return saveErr
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPolicyStor := &mocks.MockAccessPolicyFinderSaver{}
			mockKeyGen := &mocks.MockKeyGenerator{}
			service := access.New(mockPolicyStor, mockKeyGen)

			ctx := context.Background()
			expectedErr := tc.mockSetup(mockPolicyStor, mockKeyGen)

			policy, err := service.CreatePolicy(ctx, tc.nameInput)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
				assert.Nil(t, policy)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, policy)
				assert.Equal(t, tc.nameInput, policy.Name)
				assert.NotEqual(t, uuid.Nil, policy.ID)
			}

			mockKeyGen.AssertExpectations(t)
			mockPolicyStor.AssertExpectations(t)
		})
	}
}

func TestAccessControlService_CanView(t *testing.T) {
	testCases := []struct {
		name            string
		policyID        *uuid.UUID
		envSetup        func(policyID *uuid.UUID) *aggregates.Environment
		providedKey     *string
		mockSetup       func(*mocks.MockAccessPolicyFinderSaver, *uuid.UUID) error
		expectedCanView bool
	}{
		{
			name:     "no policies, no key provided",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey:     nil,
			mockSetup:       func(*mocks.MockAccessPolicyFinderSaver, *uuid.UUID) error { return nil },
			expectedCanView: true,
		},
		{
			name:     "no policies, key provided",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey:     stringPtr("some-key"),
			mockSetup:       func(*mocks.MockAccessPolicyFinderSaver, *uuid.UUID) error { return nil },
			expectedCanView: true,
		},
		{
			name:     "has policies, no key provided",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				env.AddPolicy(uuid.New(), true)
				return env
			},
			providedKey:     nil,
			mockSetup:       func(*mocks.MockAccessPolicyFinderSaver, *uuid.UUID) error { return nil },
			expectedCanView: false,
		},
		{
			name:     "has policies, invalid key",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				env.AddPolicy(uuid.New(), true)
				return env
			},
			providedKey: stringPtr("invalid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID *uuid.UUID) error {
				mockPolicyStor.On("FindByKey", mock.Anything, "invalid-key").Return(nil, nil).Once()
				return serverrors.ErrInvalidAccessKey
			},
			expectedCanView: false,
		},
		{
			name:     "has policies, valid key",
			policyID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				if policyID != nil {
					env.AddPolicy(*policyID, true)
				}
				return env
			},
			providedKey: stringPtr("valid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID *uuid.UUID) error {
				policy := &entities.AccessPolicy{ID: *policyID, Name: "test", Key: "valid-key"}
				mockPolicyStor.On("FindByKey", mock.Anything, "valid-key").Return(policy, nil).Once()
				return nil
			},
			expectedCanView: true,
		},
		{
			name:     "has policies, valid key but not attached",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				env.AddPolicy(uuid.New(), true)
				return env
			},
			providedKey: stringPtr("valid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID *uuid.UUID) error {
				policyIDNotAttached := uuid.New()
				policy := &entities.AccessPolicy{ID: policyIDNotAttached, Name: "test", Key: "valid-key"}
				mockPolicyStor.On("FindByKey", mock.Anything, "valid-key").Return(policy, nil).Once()
				return nil
			},
			expectedCanView: false,
		},
		{
			name:     "find policy error",
			policyID: nil,
			envSetup: func(policyID *uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				env.AddPolicy(uuid.New(), true)
				return env
			},
			providedKey: stringPtr("some-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID *uuid.UUID) error {
				findErr := errors.New("db error")
				mockPolicyStor.On("FindByKey", mock.Anything, "some-key").Return(nil, findErr).Once()
				return findErr
			},
			expectedCanView: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPolicyStor := &mocks.MockAccessPolicyFinderSaver{}
			mockKeyGen := &mocks.MockKeyGenerator{}
			service := access.New(mockPolicyStor, mockKeyGen)

			ctx := context.Background()
			env := tc.envSetup(tc.policyID)
			expectedErr := tc.mockSetup(mockPolicyStor, tc.policyID)

			canView, err := service.CanView(ctx, env, tc.providedKey)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
				assert.Equal(t, tc.expectedCanView, canView)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCanView, canView)
			}

			mockPolicyStor.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestAccessControlService_CanChange(t *testing.T) {
	testCases := []struct {
		name              string
		policyID          uuid.UUID
		envSetup          func(policyID uuid.UUID) *aggregates.Environment
		providedKey       *string
		mockSetup         func(*mocks.MockAccessPolicyFinderSaver, uuid.UUID) error
		expectedCanChange bool
		expectedPolicy    interface{} // nil or not nil
	}{
		{
			name:     "no key provided",
			policyID: uuid.Nil,
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey:       nil,
			mockSetup:         func(*mocks.MockAccessPolicyFinderSaver, uuid.UUID) error { return nil },
			expectedCanChange: false,
			expectedPolicy:    nil,
		},
		{
			name:     "invalid key",
			policyID: uuid.Nil,
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey: new("invalid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID uuid.UUID) error {
				mockPolicyStor.On("FindByKey", mock.Anything, "invalid-key").Return(nil, nil).Once()
				return serverrors.ErrInvalidAccessKey
			},
			expectedCanChange: false,
			expectedPolicy:    serverrors.ErrInvalidAccessKey,
		},
		{
			name:     "find policy error",
			policyID: uuid.Nil,
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey: new("some-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID uuid.UUID) error {
				findErr := errors.New("db error")
				mockPolicyStor.On("FindByKey", mock.Anything, "some-key").Return(nil, findErr).Once()
				return findErr
			},
			expectedCanChange: false,
			expectedPolicy:    nil,
		},
		{
			name:     "no policies, valid key",
			policyID: uuid.Nil,
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				return env
			},
			providedKey: new("valid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID uuid.UUID) error {
				policy := &entities.AccessPolicy{ID: uuid.New(), Name: "test", Key: "valid-key"}
				mockPolicyStor.On("FindByKey", mock.Anything, "valid-key").Return(policy, nil).Once()
				return nil
			},
			expectedCanChange: true,
			expectedPolicy:    "not nil",
		},
		{
			name:     "has policies, valid key with change permission",
			policyID: uuid.New(),
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				if policyID != uuid.Nil {
					env.AddPolicy(policyID, true)
				}
				return env
			},
			providedKey: new("valid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID uuid.UUID) error {
				policy := &entities.AccessPolicy{ID: policyID, Name: "test", Key: "valid-key"}
				mockPolicyStor.On("FindByKey", mock.Anything, "valid-key").Return(policy, nil).Once()
				return nil
			},
			expectedCanChange: true,
			expectedPolicy:    "not nil",
		},
		{
			name:     "has policies, valid key without change permission",
			policyID: uuid.New(),
			envSetup: func(policyID uuid.UUID) *aggregates.Environment {
				env, _ := aggregates.NewEnvironment("test", "", entities.NewVariables())
				if policyID != uuid.Nil {
					env.AddPolicy(policyID, false)
				}
				return env
			},
			providedKey: new("valid-key"),
			mockSetup: func(mockPolicyStor *mocks.MockAccessPolicyFinderSaver, policyID uuid.UUID) error {
				policy := &entities.AccessPolicy{ID: policyID, Name: "test", Key: "valid-key"}
				mockPolicyStor.On("FindByKey", mock.Anything, "valid-key").Return(policy, nil).Once()
				return nil
			},
			expectedCanChange: false,
			expectedPolicy:    "not nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPolicyStor := &mocks.MockAccessPolicyFinderSaver{}
			mockKeyGen := &mocks.MockKeyGenerator{}
			service := access.New(mockPolicyStor, mockKeyGen)

			ctx := context.Background()
			env := tc.envSetup(tc.policyID)
			expectedErr := tc.mockSetup(mockPolicyStor, tc.policyID)

			canChange, err := service.CanChange(ctx, env, tc.providedKey)

			if expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedCanChange, canChange)
			mockPolicyStor.AssertExpectations(t)
		})
	}
}
