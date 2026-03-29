package entities

import (
	vo "envmn/internal/domain/environment/valueobjects"
)

type Variables map[vo.VariableKey]vo.VariableValue

func NewVariables() Variables {
	return make(Variables)
}

func (vars Variables) Keys() []vo.VariableKey {
	keys := make([]vo.VariableKey, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	return keys
}

func (vars Variables) Copy() Variables {
	copyVars := make(Variables, len(vars))
	for k, v := range vars {
		copyVars[k] = v
	}
	return copyVars
}
