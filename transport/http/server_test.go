package http_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	nhttp "net/http"

	"github.com/dizzrt/ellie/internal/mock/ping"
	"github.com/dizzrt/ellie/log"
	"github.com/dizzrt/ellie/middleware/tracing"
	"github.com/dizzrt/ellie/transport/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	trace_sdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pingServer struct {
	ping.UnimplementedPingServiceServer
}

func (s *pingServer) Ping(ctx context.Context, req *ping.PingRequest) (*ping.PingResponse, error) {
	status.New(codes.Unknown, "unknown error")
	return &ping.PingResponse{
		Message: "pong",
	}, nil
}

func (s *pingServer) Hello(ctx context.Context, req *ping.HelloRequest) (*ping.HelloResponse, error) {
	log.CtxInfof(ctx, "get request from %s", req.GetName())

	return &ping.HelloResponse{
		Message: fmt.Sprintf("hello %s, type is %s", req.GetName(), req.GetType()),
	}, nil
}

func TestHTTPServer(t *testing.T) {
	ctx := context.Background()

	var opts = []http.ServerOption{
		http.DefaultSuccessCode(10000),
		http.DefaultSuccessMessage("success"),
		// http.ResponseEncoder(func(data any, err error, s *http.Server) (int, render.Render) {
		// 	code := http.HTTPStatusCodeFromError(err)
		// 	r := render.JSON{Data: gin.H{
		// 		"data": data,
		// 		"err":  err,
		// 		"ext":  "custom response encoder",
		// 	}}
		// 	return code, r
		// }),
	}
	srv := http.NewServer(opts...)

	ping.RegisterPingServiceHTTPServer(srv, &pingServer{})
	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	//
	e, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}

	url := e.String() + "/hello/ellie?type=mock"
	// resp, err := nhttp.Post(url, "application/json", strings.NewReader(``))
	resp, err := nhttp.Post(url, "application/json", strings.NewReader(`{"name": "ellieFromBody","type": "mockFromBody"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// if resp.statusCode != http.StatusOK {
	// 	t.Fatal(resp.statusCode)
	// }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(body))
	_ = srv.Stop(ctx)
}

func TestHTTPServerWithTracing(t *testing.T) {
	ctx := context.Background()

	// init tracing provider
	tp, err := tracing.Initialize(ctx,
		tracing.ServiceName("transport-test"),
		tracing.ServiceVersion("v0.0.1-dev"),
		tracing.Endpoint("localhost:4318"),
		tracing.EndpointType(tracing.EndpointType_HTTP),
		tracing.Insecure(true),
		tracing.Metadata(map[string]string{
			"ip":  "127.0.0.1",
			"env": "dev",
		}),
	)
	if err != nil {
		t.Fatalf("init tracing provider failed: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if sdkTp, ok := tp.(*trace_sdk.TracerProvider); ok {
			sdkTp.Shutdown(ctx)
		}
	}()

	var opts = []http.ServerOption{
		http.DefaultSuccessCode(10000),
		http.DefaultSuccessMessage("success"),
		http.ResponseEncoder(func(data any, err error, s *http.Server) (int, render.Render) {
			code := http.HTTPStatusCodeFromError(err)
			r := render.JSON{Data: gin.H{
				"data": data,
				"err":  err,
				"ext":  "custom response encoder",
			}}
			return code, r
		}),
		http.Middleware(
			tracing.TracingMiddleware(),
		),
	}

	srv := http.NewServer(opts...)
	ping.RegisterPingServiceHTTPServer(srv, &pingServer{})
	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	e, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}

	url := e.String() + "/hello/ellie?type=mock"
	reqBody := strings.NewReader(`{"name": "ellieFromBody","type": "mockFromBody"}`)
	req, err := nhttp.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	// req.Header.Set("X-Log-ID", "123test111abc")
	// req.Header.Set("log.id", "123test111")

	client := &nhttp.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(body))
	_ = srv.Stop(ctx)

	log.Sync()
}
