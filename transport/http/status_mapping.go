package http

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

func GRPCCodeFromHTTPStatus(code int) codes.Code {
	switch code {
	case http.StatusOK:
		return codes.OK
	case 499:
		return codes.Canceled
	case http.StatusBadRequest, http.StatusLengthRequired, http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong, http.StatusUnsupportedMediaType, http.StatusNotAcceptable,
		http.StatusRequestHeaderFieldsTooLarge, http.StatusUnprocessableEntity:
		return codes.InvalidArgument
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotFound, http.StatusGone:
		return codes.NotFound
	case http.StatusConflict:
		// Note: HTTP 409 can map to both AlreadyExists and Aborted in gRPC
		// Using AlreadyExists as the primary mapping
		return codes.AlreadyExists
	case http.StatusForbidden, http.StatusUnavailableForLegalReasons:
		return codes.PermissionDenied
	case http.StatusUnauthorized, http.StatusProxyAuthRequired:
		return codes.Unauthenticated
	case http.StatusTooManyRequests, http.StatusTooEarly:
		return codes.ResourceExhausted
	case http.StatusPreconditionFailed, http.StatusPreconditionRequired, http.StatusUpgradeRequired:
		return codes.FailedPrecondition
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusInternalServerError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

func HTTPStatusFromGRPCCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		// Client Closed Request (Non-standard but widely used)
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
