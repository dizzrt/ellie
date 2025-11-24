package tracing

type _EndpointType string

const (
	EndpointType_GRPC _EndpointType = "grpc"
	EndpointType_HTTP _EndpointType = "http"
)

type config struct {
	serviceName    string
	serviceVersion string
	metadata       map[string]string
	endpoint       string
	endpointType   _EndpointType
	insecure       bool
}

type Option func(*config)

func ServiceName(serviceName string) Option {
	return func(opts *config) {
		opts.serviceName = serviceName
	}
}

func ServiceVersion(serviceVersion string) Option {
	return func(opts *config) {
		opts.serviceVersion = serviceVersion
	}
}

func Metadata(metadata map[string]string) Option {
	return func(opts *config) {
		opts.metadata = metadata
	}
}

func Endpoint(endpoint string) Option {
	return func(opts *config) {
		opts.endpoint = endpoint
	}
}

func EndpointType(endpointType _EndpointType) Option {
	return func(opts *config) {
		opts.endpointType = endpointType
	}
}

func Insecure(insecure bool) Option {
	return func(opts *config) {
		opts.insecure = insecure
	}
}

func ParseEndpointType(endpointType string) _EndpointType {
	switch endpointType {
	case "http":
		return EndpointType_HTTP
	default:
		return EndpointType_GRPC
	}
}
