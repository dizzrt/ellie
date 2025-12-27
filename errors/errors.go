package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type AdvancedError interface {
	Unwrap() error
	Is(error) bool

	// getters
	Status() *codes.Code
	Code() int32
	Reason() string
	Message() string
	Metadata() map[string]string
	Cause() error

	// setters
	WithStatus(codes.Code) AdvancedError
	WithCode(int32) AdvancedError
	WithReason(string) AdvancedError
	WithMessage(string, ...any) AdvancedError
	WithMetadata(map[string]string) AdvancedError
	WithCause(error) AdvancedError

	Chainable
}

func New(code int, reason, message string) error {
	return NewStandardError(nil, code, reason, message)
}

func Newf(code int, reason, format string, a ...any) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

func Errorf(code int, reason, format string, a ...any) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

func StatusPtrFromInt(code int) *codes.Code {
	if code < 0 || code >= GRPC_STATUS_MAX_CODE {
		return nil
	}

	temp := codes.Code(code)
	return &temp
}
