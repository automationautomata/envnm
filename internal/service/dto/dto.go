package dto

import "github.com/google/uuid"

type GetClientVariablesInput struct {
	EnvironmentName string
	AccessKey       *string
}

type UpdateVariablesByClientInput struct {
	EnvironmentName string
	AccessKey       string
	Variables       map[string]string
}

type SubscribeOnUpdatesInput struct {
	EnvironmentName string
	AccessKey       string
}

type CreateEnvironmentInput struct {
	Name        string
	Description *string
	Variables   map[string]string
}

type GetAllEnvironmentsInput struct {
	Reserved bool
}

type DeleteEnvironmentInput struct {
	Name string
}

type UpdateEnvironmentInfoInput struct {
	OldName     string
	NewName     *string
	Description *string
}

type CreateAccessPolicyInput struct {
	Name           string
	Key            string
	ChangesAllowed bool
}

type RemovePolicyInput struct {
	ID uuid.UUID
}

type AddPolicyToEnvironmentInput struct {
	EnvironmentName string
	PolicyID        uuid.UUID
}

type RemovePolicyFromEnvironmentInput struct {
	EnvironmentName string
	PolicyID        uuid.UUID
}

type GetPolicyKeyInput struct {
	ID uuid.UUID
}

type UpdateEnvironmentVariablesInput struct {
	EnvironmentName string
	Variables       map[string]string
	AccessKey       *string
}

type RemoveVariableFromEnvironmentInput struct {
	EnvironmentName string
	VariableKey     string
}

type GetEnvironmentPoliciesInput struct {
	EnvironmentName string
}

type EnvironmentDTO struct {
	Name        string
	Variables   map[string]string
	Description string
	Reserved    bool
}

type PolicyDTO struct {
	Name           string
	Key            string
	ChangesAllowed bool
}

type AuthByPasswordInput struct {
	Password string
}
