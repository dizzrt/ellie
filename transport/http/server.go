package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dizzrt/ellie/errors"
	"github.com/dizzrt/ellie/internal/endpoint"
	"github.com/dizzrt/ellie/internal/host"
	"github.com/dizzrt/ellie/log"
	"github.com/dizzrt/ellie/transport"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
	_ http.Handler         = (*Server)(nil)
)

type Server struct {
	*http.Server

	err error
	lis net.Listener

	engine                *gin.Engine
	noRouteHandlers       []gin.HandlerFunc
	NoMethodHandler       []gin.HandlerFunc
	redirectTrailingSlash bool

	tlsConf  *tls.Config
	endpoint *url.URL
	network  string
	address  string
	timeout  time.Duration
	filters  []FilterFunc
	// TODO middleware

	defaultSuccessCode    int
	defaultSuccessMessage string
	responseEncoder       HTTPResponseEncoder
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:               "tcp",
		address:               ":0",
		timeout:               1 * time.Second,
		defaultSuccessCode:    0,
		defaultSuccessMessage: "ok",
		responseEncoder:       DefaultResponseEncoder,
		engine:                gin.Default(),
		redirectTrailingSlash: true,
	}

	if len(srv.noRouteHandlers) > 0 {
		srv.engine.NoRoute(srv.noRouteHandlers...)
	}

	if len(srv.NoMethodHandler) > 0 {
		srv.engine.NoMethod(srv.NoMethodHandler...)
	}

	for _, opt := range opts {
		opt(srv)
	}

	srv.engine.RedirectTrailingSlash = srv.redirectTrailingSlash
	srv.Server = &http.Server{
		TLSConfig: srv.tlsConf,
		Handler:   FilterChain(srv.filters...)(srv.engine),
	}

	return srv
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) initializeListenerAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}

		s.lis = lis
	}

	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}

		s.endpoint = endpoint.New(endpoint.Scheme("http", s.tlsConf != nil), addr)
	}

	return s.err
}

func (s *Server) WrapHTTPResponse(data any, err error) gin.H {
	code := s.defaultSuccessCode
	message := s.defaultSuccessMessage

	if err != nil {
		if se, ok := err.(*errors.StandardError); ok {
			// standard error
			code = int(se.Code())
			message = se.Message()
		} else if st, ok := status.FromError(err); ok {
			// grpc error
			code = int(st.Code())
			message = st.Message()
		} else {
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

func (s *Server) EncodeResponse(ctx *gin.Context, data any, err error) {
	code, r := s.responseEncoder(data, err, s)
	ctx.Render(code, r)
}

// region interfaces impl

func (s *Server) Start(ctx context.Context) error {
	if err := s.initializeListenerAndEndpoint(); err != nil {
		return err
	}

	s.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}

	log.Infof("[HTTP] server listening on %s", s.lis.Addr().String())

	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.lis, "", "")
	} else {
		err = s.Serve(s.lis)
	}

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[HTTP] server stopping")

	err := s.Shutdown(ctx)
	if err != nil {
		if ctx.Err() != nil {
			log.Warn("[HTTP] server couldn't stop gracefully in time, forcing stop")
			fmt.Println("[HTTP] server couldn't stop gracefully in time, forcing stop")
			err = s.Close()
		}
	}

	return err
}

func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.initializeListenerAndEndpoint(); err != nil {
		return nil, s.err
	}

	return s.endpoint, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler.ServeHTTP(w, r)
}

// endregion
