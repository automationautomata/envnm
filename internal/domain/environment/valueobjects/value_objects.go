package valueobjects

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidVariableKey       = errors.New("invalid variable key")
	ErrInvalidVariableKeyFormat = errors.New("variable key must be UPPER_SNAKE_CASE")
)

// Должен быть UPPER_SNAKE_CASE и меньше 255 байт
type VariableKey string

func NewVariableKey(key string) (VariableKey, error) {
	if key == "" || len(key) > 255 {
		return "", ErrInvalidVariableKey
	}
	if matched, _ := regexp.MatchString(`^[A-Z][A-Z0-9_]*$`, key); !matched {
		return "", ErrInvalidVariableKeyFormat
	}
	return VariableKey(key), nil
}

func (key VariableKey) String() string { return string(key) }

type VariableValue string

func NewVariableValue(value string) VariableValue {
	return VariableValue(value)
}

func (val VariableValue) String() string { return string(val) }
