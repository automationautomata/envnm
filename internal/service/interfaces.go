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
	RemovePolicy(ctx context.Context, input dto.RemovePolicyInput) error
	AddPolicyToEnvironment(ctx context.Context, input dto.AddPolicyToEnvironmentInput) error
	RemovePolicyFromEnvironment(ctx context.Context, input dto.RemovePolicyFromEnvironmentInput) error
}

type ManageEnvironmentVariables interface {
	UpdateEnvironmentVariables(ctx context.Context, input dto.UpdateEnvironmentVariablesInput) error
	RemoveVariableFromEnvironment(ctx context.Context, input dto.RemoveVariableFromEnvironmentInput) error
}

type Client struct {
	ClientVariables
	ClientSubscribtion
}

func NewClient(vars ClientVariables, subs ClientSubscribtion) *Client {
	return &Client{
		ClientVariables:    vars,
		ClientSubscribtion: subs,
	}
}

type Management struct {
	ManageEnvironment
	ManageAccessPolicy
	ManageEnvironmentVariables
}

func NewManagement(
	env ManageEnvironment,
	policy ManageAccessPolicy,
	envVar ManageEnvironmentVariables,
) *Management {
	return &Management{
		ManageEnvironment:          env,
		ManageAccessPolicy:         policy,
		ManageEnvironmentVariables: envVar,
	}
}
