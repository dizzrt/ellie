package http

import (
	"github.com/gin-gonic/gin/render"
)

type HTTPResponseEncoder = func(data any, err error, s *Server) (int, render.Render)

func DefaultResponseEncoder(data any, err error, s *Server) (int, render.Render) {
	code := HTTPStatusCodeFromError(err)
	r := render.JSON{Data: s.WrapHTTPResponse(data, err)}

	return code, r
}
