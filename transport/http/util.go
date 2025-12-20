package http

import (
	"net/http"

	"github.com/dizzrt/ellie/errors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

func HTTPStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if se, ok := err.(*errors.StandardError); ok {
		return HTTPStatusFromGRPCCode(se.Status())
	}

	if st, ok := status.FromError(err); ok {
		return HTTPStatusFromGRPCCode(st.Code())
	}

	return http.StatusOK
}

func WrapHTTPResponse(code int, message string, data any, err error) gin.H {
	if err != nil {
		if ee, ok := err.(*errors.StandardError); ok {
			// ellie error
			code = int(ee.Code())
			message = ee.Message()
		} else if st, ok := status.FromError(err); ok {
			// grpc error
			code = int(st.Code())
			message = st.Message()
		} else {
			// unknown error type
			code = -1
			message = err.Error()
		}
	}

	return gin.H{
		"data":    data,
		"status":  code,
		"message": message,
	}
}
