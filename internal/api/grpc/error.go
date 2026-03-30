package grpc

import (
	"errors"

	errs "envmn/internal/service/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

type grpcHandlerError struct {
	Code    codes.Code
	Message string
	Details []protoadapt.MessageV1
}

func (e *grpcHandlerError) Error() string {
	return e.Message
}

func (e *grpcHandlerError) GRPCStatus() *status.Status {
	st := status.New(e.Code, e.Message)
	if e.Details != nil {
		// Ошибки при добавлении деталей игнорируем, так как это не должно мешать отправке основного сообщения об ошибке
		st, _ = st.WithDetails(e.Details...)
	}
	return st
}

func toGRPCError(err error) (handlerError *grpcHandlerError, isInternal bool) {
	if err == nil {
		return nil, false
	}

	switch {
	case errors.Is(err, errs.ErrEnvironmentNotFound),
		errors.Is(err, errs.ErrAccessPolicyNotFound):
		return &grpcHandlerError{
			Code:    codes.NotFound,
			Message: err.Error(),
		}, false

	case errors.Is(err, errs.ErrAccessDenied):
		return &grpcHandlerError{
			Code:    codes.PermissionDenied,
			Message: err.Error(),
		}, false

	case errors.Is(err, errs.ErrInvalidAccessKey),
		errors.Is(err, errs.ErrInvalidAccessPolicy),
		errors.Is(err, errs.ErrInvalidVariableKey):
		return &grpcHandlerError{
			Code:    codes.InvalidArgument,
			Message: err.Error(),
		}, false

	case errors.Is(err, errs.ErrEnvironmentIsReserved):
		return &grpcHandlerError{
			Code:    codes.FailedPrecondition,
			Message: err.Error(),
		}, false

	default:
		// Всё неизвестное — Internal
		return &grpcHandlerError{
			Code:    codes.Internal,
			Message: "internal server error",
		}, true
	}
}
