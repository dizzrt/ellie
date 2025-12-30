package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func StatusCodeFromError(err error) codes.Code {
	if err == nil {
		// if error is nil, return ok code
		return codes.OK
	}

	// if error is grpc status error, return its code directly
	if st, ok := status.FromError(err); ok {
		return st.Code()
	}

	// if error is advanced error, return its status code when it is set
	if ae, ok := err.(AdvancedError); ok {
		sptr := ae.Status()
		if sptr != nil {
			return *sptr
		}
	}

	// default to unknown code
	return codes.Unknown
}
