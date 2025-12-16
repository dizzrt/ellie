package errors

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

var _ error = (*Error)(nil)
var _ Chainable = (*Error)(nil)

const (
	UnknownCode   = 500
	UnknownReason = ""
)

type Error struct {
	Status
	cause error
}

func New(code int, reason, message string) *Error {
	return &Error{
		Status: Status{
			Code:     int32(code),
			Reason:   reason,
			Message:  message,
			Metadata: make(map[string]string),
		},
	}
}

func Newf(code int, reason, format string, a ...any) *Error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

func Errorf(code int, reason, format string, a ...any) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

func Code(err error) int {
	if err == nil {
		return 200
	}

	return int(FromError(err).Code)
}

func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}

	return FromError(err).Reason
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Is(target error) bool {
	if ee := new(Error); errors.As(target, &ee) {
		return ee.Code == e.Code && ee.Reason == e.Reason
	}

	return false
}

func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

func (e *Error) WithMetadata(metadata map[string]string) *Error {
	err := Clone(e)
	err.Metadata = metadata
	return err
}

func (e *Error) GRPCStatus() *status.Status {
	// TODO
	return nil
}

func (e *Error) Type() string {
	return CHAINABLE_ERROR_TYPE_ELLIE
}

func (e *Error) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Error) Wrap(err error) error {
	return e.WithCause(err)
}

func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}

	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}

	return &Error{
		Status: Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
		cause: err.cause,
	}
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if ee := new(Error); errors.As(err, &ee) {
		return ee
	}

	grpcStatus, ok := status.FromError(err)
	if !ok {
		return New(UnknownCode, UnknownReason, err.Error())
	}

	ret := New(UnknownCode, UnknownReason, grpcStatus.Message())
	for _, detail := range grpcStatus.Details() {
		switch temp := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Reason = temp.Reason
			return ret.WithMetadata(temp.Metadata)
		}
	}

	return ret
}
