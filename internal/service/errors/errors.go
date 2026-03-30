package errors

import "errors"

var (
	ErrEnvironmentNotFound   = errors.New("environment not found")
	ErrAccessDenied          = errors.New("access denied")
	ErrInvalidAccessKey      = errors.New("invalid access key")
	ErrInvalidAccessPolicy   = errors.New("invalid access policy")
	ErrAccessPolicyNotFound  = errors.New("access policy not found")
	ErrInvalidVariableKey    = errors.New("invalid variable key")
	ErrEnvironmentIsReserved = errors.New("cannot change reserved ernvironment")
)
