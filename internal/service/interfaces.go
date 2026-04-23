package service

import (
	"context"
	"envmn/internal/service/dto"

	"github.com/google/uuid"
)

type ClientVariables interface {
	GetClientVariables(ctx context.Context, input dto.GetClientVariablesInput) (map[string]string, error)
	UpdateVariablesByClient(ctx context.Context, input dto.UpdateVariablesByClientInput) error
}

type ClientSubscribtion interface {
	SubscribeOnUpdates(ctx context.Context, input dto.SubscribeOnUpdatesInput) (key string, err error)
}

type ManageEnvironment interface {
	CreateEnvironment(ctx context.Context, input dto.CreateEnvironmentInput) (uuid.UUID, error)
	GetAllEnvironments(ctx context.Context, input dto.GetAllEnvironmentsInput) ([]*dto.EnvironmentDTO, error)
	DeleteEnvironment(ctx context.Context, input dto.DeleteEnvironmentInput) error
	UpdateEnvironmentInfo(ctx context.Context, input dto.UpdateEnvironmentInfoInput) error
}

type ManageAccessPolicy interface {
	CreateAccessPolicy(ctx context.Context, input dto.CreateAccessPolicyInput) (uuid.UUID, error)
	ListPolicyEnvironments(ctx context.Context, input dto.ListPolicyEnvironmentsInput) ([]*dto.PolicyEnvironmentsItem, error)
	GetPolicyByName(ctx context.Context, input dto.GetPolicyByNameInput) (*dto.PolicyDTO, error)
	RemovePolicy(ctx context.Context, input dto.RemovePolicyInput) error
	AddPolicyToEnvironment(ctx context.Context, input dto.AddPolicyToEnvironmentInput) error
	RemovePolicyFromEnvironment(ctx context.Context, input dto.RemovePolicyFromEnvironmentInput) error
}

type ManageEnvironmentVariables interface {
	UpdateEnvironmentVariables(ctx context.Context, input dto.UpdateEnvironmentVariablesInput) error
	RemoveVariableFromEnvironment(ctx context.Context, input dto.RemoveVariableFromEnvironmentInput) error
}

type DistributionServices struct {
	ClientVariables
	ClientSubscribtion
}

func NewDistributionServices(vars ClientVariables, subs ClientSubscribtion) *DistributionServices {
	return &DistributionServices{
		ClientVariables:    vars,
		ClientSubscribtion: subs,
	}
}

type ManagementServices struct {
	ManageEnvironment
	ManageAccessPolicy
	ManageEnvironmentVariables
}

func NewManagementServices(
	env ManageEnvironment,
	policy ManageAccessPolicy,
	envVar ManageEnvironmentVariables,
) *ManagementServices {
	return &ManagementServices{
		ManageEnvironment:          env,
		ManageAccessPolicy:         policy,
		ManageEnvironmentVariables: envVar,
	}
}
