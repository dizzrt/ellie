package errors

import (
	"errors"
	"fmt"
	"maps"

	"github.com/bytedance/sonic"
	"github.com/dizzrt/ellie/pkg/ptrconv"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ AdvancedError = (*StandardError)(nil)

const GRPC_STATUS_MAX_CODE = 17

const _STANDARD_ERROR_TYPE = "standard_error"

type StandardError struct {
	cause error
	core  *ErrorCore
}

func NewStandardError(status *codes.Code, code int, reason, message string) *StandardError {
	var statusPtr *int32 = nil
	if status != nil {
		temp := int32(*status)
		statusPtr = &temp
	}

	return &StandardError{
		core: &ErrorCore{
			Status:  statusPtr,
			Code:    int32(code),
			Reason:  reason,
			Message: message,
		},
	}
}

func NewStandardErrorf(status *codes.Code, code int, reason, format string, a ...any) *StandardError {
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
		scode := st.Code()
		se := NewStandardError(&scode, -1, "CONVERT_FROM_GRPC_STATUS", st.Message())
		for _, detail := range st.Details() {
			switch ty := detail.(type) {
			case *errdetails.ErrorInfo:
				se.WithReason(ty.Reason)
				se.WithMetadata(ty.Metadata)
				return se
			}
		}
	}

	scode := codes.Unknown
	se := NewStandardError(&scode, -1, "CONVERT_FROM_ERROR", err.Error())
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

	if se.core.GetCode() != temp.core.GetCode() {
		return false
	}

	if se.core.GetReason() != temp.core.GetReason() {
		return false
	}

	return true
}

func (se *StandardError) Status() *codes.Code {
	if se.core.Status == nil {
		return nil
	}

	var temp codes.Code
	v := se.core.GetStatus()
	if v < 0 || v >= GRPC_STATUS_MAX_CODE {
		temp = codes.Unknown
	} else {
		temp = codes.Code(v)
	}

	return &temp
}

func (se *StandardError) Code() int32 {
	return se.core.GetCode()
}

func (se *StandardError) Reason() string {
	return se.core.GetReason()
}

func (se *StandardError) Message() string {
	return se.core.GetMessage()
}

func (se *StandardError) Metadata() map[string]string {
	return se.core.GetMetadata()
}

func (se *StandardError) Cause() error {
	return se.cause
}

func (se *StandardError) WithStatus(status codes.Code) AdvancedError {
	temp := int32(status)
	se.core.Status = &temp
	return se
}

func (se *StandardError) WithCode(code int32) AdvancedError {
	se.core.Code = code
	return se
}

func (se *StandardError) WithReason(reason string) AdvancedError {
	se.core.Reason = reason
	return se
}

func (se *StandardError) WithMessage(format string, a ...any) AdvancedError {
	se.core.Message = fmt.Sprintf(format, a...)
	return se
}

func (se *StandardError) WithMetadata(metadata map[string]string) AdvancedError {
	see := se.Clone()
	if see == nil {
		return nil
	}

	see.core.Metadata = metadata
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
	return sonic.Marshal(se.core)
}

func (se *StandardError) MapError() map[string]any {
	mp := map[string]any{
		"code":   se.Code(),
		"reason": se.Reason(),
	}

	// Add status if it's not nil
	if status := se.Status(); status != nil {
		mp["status"] = status.String()
	}

	// Add message if it's not empty
	if message := se.Message(); message != "" {
		mp["message"] = message
	}

	// Add metadata if it's not nil and not empty
	if metadata := se.Metadata(); len(metadata) > 0 {
		mp["metadata"] = metadata
	}

	// Add cause if it's not nil
	if se.cause != nil {
		err := se.cause
		if ee, ok := err.(*StandardError); ok {
			mp["cause"] = ee.MapError()
		} else {
			mp["cause"] = se.cause.Error()
		}
	}

	return mp
}

func (se *StandardError) Error() string {
	emap := se.MapError()

	errBytes, err := sonic.Marshal(emap)
	if err != nil {
		// fallback to simple format if JSON marshaling fails
		msg := fmt.Sprintf("[%d][%s]", se.Code(), se.Reason())
		if m := se.Message(); m != "" {
			msg += m
		}

		return msg
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
		core: &ErrorCore{
			Status:   ptrconv.Ptr(se.core.GetStatus()),
			Code:     se.core.GetCode(),
			Reason:   se.core.GetReason(),
			Message:  se.core.GetMessage(),
			Metadata: metadata,
		},
		cause: se.cause,
	}
}

func standardErrorChainableUnmarshal(_ string, data []byte) error {
	core := &ErrorCore{}
	if err := sonic.Unmarshal(data, core); err != nil {
		return fmt.Errorf("failed to unmarshal standard error with data: %s, error: %w", string(data), err)
	}

	return &StandardError{
		cause: nil,
		core:  core,
	}
}
