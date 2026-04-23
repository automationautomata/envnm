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
	Name string
	Key  string
}

type RemovePolicyInput struct {
	ID uuid.UUID
}

type ListPolicyEnvironmentsInput struct {
	ID uuid.UUID
}

type GetPolicyByNameInput struct {
	Name string
}

type AddPolicyToEnvironmentInput struct {
	PolicyID          uuid.UUID
	EnvironmentName   string
	ChangesPermission bool
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

type EnvironmentPolicyItem struct {
	Name              string
	Key               string
	ChangesPermission bool
}

type PolicyEnvironmentsItem struct {
	EnvironmentName   string
	ChangesPermission bool
}

type PolicyDTO struct {
	ID   uuid.UUID
	Name string
	Key  string
}

type AuthByPasswordInput struct {
	Password string
}
