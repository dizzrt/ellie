package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ AdvancedError = (*StandardError)(nil)

const _STANDARD_ERROR_TYPE = "standard_error"

type StandardError struct {
	cause       error
	errorStatus ErrorStatus
}

func NewStandardError(status codes.Code, code int, reason, message string) *StandardError {
	return &StandardError{
		errorStatus: ErrorStatus{
			Status:  uint32(status),
			Code:    int32(code),
			Reason:  reason,
			Message: message,
		},
	}
}

func NewStandardErrorf(status codes.Code, code int, reason, format string, a ...any) *StandardError {
	return NewStandardError(status, code, reason, fmt.Sprintf(format, a...))
}

func NewStandardErrorFromError(err error) *StandardError {
	if err == nil {
		return nil
	}

	if se, ok := err.(*StandardError); ok {
		return se
	}

	if st, ok := status.FromError(err); ok {
		se := NewStandardError(st.Code(), -1, "CONVERT_FROM_GRPC_STATUS", st.Message())
		for _, detail := range st.Details() {
			switch ty := detail.(type) {
			case *errdetails.ErrorInfo:
				se.WithReason(ty.Reason)
				se.WithMetadata(ty.Metadata)
				return se
			}
		}
	}

	se := NewStandardError(codes.Unknown, -1, "CONVERT_FROM_ERROR", err.Error())
	return se
}

func (se *StandardError) Unwrap() error {
	return se.cause
}

func (se *StandardError) Is(target error) bool {
	temp := new(StandardError)
	if !errors.As(target, &temp) {
		return false
	}

	if se.errorStatus.GetCode() != temp.errorStatus.GetCode() {
		return false
	}

	if se.errorStatus.GetReason() != temp.errorStatus.GetReason() {
		return false
	}

	return true
}

func (se *StandardError) Status() codes.Code {
	return codes.Code(se.errorStatus.GetStatus())
}

func (se *StandardError) Code() int32 {
	return se.errorStatus.GetCode()
}

func (se *StandardError) Reason() string {
	return se.errorStatus.GetReason()
}

func (se *StandardError) Message() string {
	return se.errorStatus.GetMessage()
}

func (se *StandardError) Metadata() map[string]string {
	return se.errorStatus.GetMetadata()
}

func (se *StandardError) Cause() error {
	return se.cause
}

func (se *StandardError) WithStatus(status codes.Code) AdvancedError {
	se.errorStatus.Status = uint32(status)
	return se
}

func (se *StandardError) WithCode(code int32) AdvancedError {
	se.errorStatus.Code = code
	return se
}

func (se *StandardError) WithReason(reason string) AdvancedError {
	se.errorStatus.Reason = reason
	return se
}

func (se *StandardError) WithMessage(message string) AdvancedError {
	se.errorStatus.Message = message
	return se
}

func (se *StandardError) WithMetadata(metadata map[string]string) AdvancedError {
	see := se.Clone()
	if see == nil {
		return nil
	}

	see.errorStatus.Metadata = metadata
	return see
}

func (se *StandardError) WithCause(cause error) AdvancedError {
	see := se.Clone()
	if see == nil {
		return nil
	}

	see.cause = cause
	return see
}

func (se *StandardError) Type() string {
	return _STANDARD_ERROR_TYPE
}

func (se *StandardError) Wrap(err error) error {
	return se.WithCause(err)
}

func (se *StandardError) Marshal() ([]byte, error) {
	return json.Marshal(se)
}

func (se *StandardError) Error() string {
	st := codes.Code(se.errorStatus.GetStatus()).String()
	errorInfo := map[string]any{
		"status": st,
		"code":   se.errorStatus.GetCode(),
		"reason": se.errorStatus.GetReason(),
	}

	// Add message if it's not empty
	if message := se.errorStatus.GetMessage(); message != "" {
		errorInfo["message"] = message
	}

	// Add metadata if it's not nil and not empty
	if metadata := se.errorStatus.GetMetadata(); len(metadata) > 0 {
		errorInfo["metadata"] = metadata
	}

	// Add cause if it's not nil
	if se.cause != nil {
		errorInfo["cause"] = se.cause.Error()
	}

	// Marshal to JSON
	errBytes, err := json.Marshal(errorInfo)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		baseMsg := "standard_error"
		if msg := se.errorStatus.GetMessage(); msg != "" {
			baseMsg = msg
		} else if reason := se.errorStatus.GetReason(); reason != "" {
			baseMsg = reason
		}
		return baseMsg
	}

	return string(errBytes)
}

func (se *StandardError) Clone() *StandardError {
	if se == nil {
		return nil
	}

	seMetadata := se.Metadata()
	metadata := make(map[string]string, len(seMetadata))
	maps.Copy(metadata, seMetadata)

	return &StandardError{
		errorStatus: ErrorStatus{
			Status:   se.errorStatus.Status,
			Code:     se.errorStatus.Code,
			Reason:   se.errorStatus.Reason,
			Message:  se.errorStatus.Message,
			Metadata: metadata,
		},
		cause: se.cause,
	}
}

func standardErrorChainableUnmarshal(_ string, data []byte) error {
	se := &StandardError{}
	if err := json.Unmarshal(data, &se); err != nil {
		return fmt.Errorf("failed to unmarshal standard error with data: %s, error: %w", string(data), err)
	}

	return se
}
