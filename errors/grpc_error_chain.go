package errors

import (
	"errors"
	"fmt"
	"maps"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

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
		Type:    ErrorChainNodeType_GOLANG_ERROR,
	}

	var ee *Error
	if As(err, &ee) {
		node.Type = ErrorChainNodeType_ELLIE_ERROR
		node.Code = ee.Code
		node.Reason = ee.Reason
		node.BizMessage = ee.Message

		node.Metadata = make(map[string]string, len(ee.Metadata))
		maps.Copy(node.Metadata, ee.Metadata)
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
		if c, ok := detail.(*ErrorChain); ok {
			chain = c
			break
		}
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
	if node.Type == ErrorChainNodeType_ELLIE_ERROR {
		ee := New(int(node.Code), node.Reason, node.BizMessage)

		ee.Metadata = make(map[string]string, len(node.Metadata))
		maps.Copy(ee.Metadata, node.Metadata)

		err = ee
	} else {
		err = errors.New(node.Message)
	}

	wrappedErr := recursiveUnmarshal(node.Wrapped)
	if wrappedErr != nil {
		if ee, ok := err.(*Error); ok {
			err = ee.WithCause(wrappedErr)
		} else {
			err = fmt.Errorf("%w", wrappedErr)
		}
	}

	return err
}
