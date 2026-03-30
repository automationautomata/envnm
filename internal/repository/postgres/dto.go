package postgres

import (
	"envmn/internal/domain/environment/aggregates"
	"envmn/internal/domain/environment/entities"
	vo "envmn/internal/domain/environment/valueobjects"
	"time"

	"github.com/google/uuid"
)

type variableDTO struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

func variablesToDomain(variables []variableDTO) (entities.Variables, error) {
	varsEntity := entities.NewVariables()
	for _, variable := range variables {
		varKey, err := vo.NewVariableKey(variable.Key)
		if err != nil {
			return nil, err
		}
		varsEntity[varKey] = vo.NewVariableValue(variable.Value)
	}
	return varsEntity, nil
}

type environmentPolicyDTO struct {
	ID             uuid.UUID `db:"id"`
	ChangesAllowed bool      `db:"changes_allowed"`
}

type environmentDTO struct {
	ID                  uuid.UUID              `db:"id"`
	Name                string                 `db:"name"`
	Description         *string                `db:"description"`
	Variables           []variableDTO          `db:"variables"`
	AccessPolicies      []environmentPolicyDTO `db:"policies"`
	LastVariablesUpdate time.Time              `db:"last_variables_update"`
	CreatedAt           time.Time              `db:"created_at"`
}

func (dto environmentDTO) toDomain() (*aggregates.Environment, error) {
	varsEntity, err := variablesToDomain(dto.Variables)
	if err != nil {
		return nil, err
	}

	envEntity, err := aggregates.NewEnvironment(dto.Name, "", varsEntity)
	if err != nil {
		return nil, err
	}
	if dto.Description != nil {
		envEntity.Description = *dto.Description
	}
	for _, policy := range dto.AccessPolicies {
		envEntity.AddPolicy(policy.ID, policy.ChangesAllowed)
	}

	envEntity.ID = dto.ID
	envEntity.LastVariablesUpdate = dto.LastVariablesUpdate
	return envEntity, nil
}

type policyDTO struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
	Key  string    `db:"key"`
}
