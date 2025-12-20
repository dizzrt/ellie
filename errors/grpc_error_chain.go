package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const CHAINABLE_ERROR_TYPE_GOLANG = "errorString"
const CHAINABLE_ERROR_TYPE_STANDARD = _STANDARD_ERROR_TYPE

var chainableErrorTypeMap = map[string]func(string, []byte) error{
	CHAINABLE_ERROR_TYPE_GOLANG:   nil,
	CHAINABLE_ERROR_TYPE_STANDARD: standardErrorChainableUnmarshal,
}

type Chainable interface {
	error

	Type() string
	Wrap(error) error
	Marshal() ([]byte, error)
}

func RegisterChainableErrorType(ty string, fn func(string, []byte) error) error {
	if _, ok := chainableErrorTypeMap[ty]; ok {
		return fmt.Errorf("chainable error type '%s' already registered", ty)
	}

	chainableErrorTypeMap[ty] = fn
	return nil
}

func Marshal(code codes.Code, err error) error {
	if err == nil {
		return nil
	}

	rootNode := recursiveMarshal(err)
	chain := &ErrorChain{
		Root: rootNode,
	}

	anyChain, ee := anypb.New(chain)
	if ee != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to chain error: %v, raw error: %v", ee, err))
	}

	st := status.New(code, err.Error())
	st, ee = st.WithDetails(anyChain)
	if ee != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to chain error: %v, raw error: %v", ee, err))
	}

	return st.Err()
}

func recursiveMarshal(err error) *ErrorChainNode {
	if err == nil {
		return nil
	}

	node := &ErrorChainNode{
		Message: err.Error(),
		Type:    CHAINABLE_ERROR_TYPE_GOLANG,
	}

	ce, ok := err.(Chainable)
	if ok {
		node.Type = ce.Type()
		data, ee := ce.Marshal()
		if ee != nil {
			ee = fmt.Errorf("failed to marshal chainable error: %v, raw error: %v", ee, err)
			node.Message = ee.Error()
		} else {
			node.Data = data
		}
	}

	wrappedErr := errors.Unwrap(err)
	if wrappedErr != nil {
		node.Wrapped = recursiveMarshal(wrappedErr)
	}

	return node
}

func Unmarshal(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	var chain *ErrorChain
	for _, detail := range st.Details() {
		anyDetail, ok := detail.(*anypb.Any)
		if !ok {
			continue
		}

		tmpChain := &ErrorChain{}
		if ee := anypb.UnmarshalTo(anyDetail, tmpChain, proto.UnmarshalOptions{}); ee != nil {
			continue
		}

		chain = tmpChain
		break
	}

	if chain == nil || chain.Root == nil {
		return st.Err()
	}

	return recursiveUnmarshal(chain.Root)
}

func recursiveUnmarshal(node *ErrorChainNode) error {
	if node == nil {
		return nil
	}

	var err error
	ty := node.GetType()
	if ty != "" {
		if fn, ok := chainableErrorTypeMap[ty]; ok && fn != nil {
			err = fn(node.GetMessage(), node.GetData())
		}
	}

	if err == nil {
		err = errors.New(node.GetMessage())
	}

	wrappedErr := recursiveUnmarshal(node.Wrapped)
	if wrappedErr != nil {
		if ce, ok := err.(Chainable); ok {
			err = ce.Wrap(wrappedErr)
		} else {
			err = fmt.Errorf("%w", wrappedErr)
		}
	}

	return err
}

func WrpGRPCResponse[T any](data T, err error) (T, error) {
	if _, ok := status.FromError(err); ok {
		return data, err
	}

	if se, ok := err.(*StandardError); ok {
		status := se.Status()
		return data, Marshal(status, err)
	}

	return data, err
}

func UnwrapGRPCResponse[T any](data T, err error) (T, error) {
	return data, Unmarshal(err)
}
