module github.com/dizzrt/ellie/cmd/protoc-gen-ellie-go-errors

go 1.25.3

replace github.com/dizzrt/ellie => ../..

require (
	github.com/dizzrt/ellie v0.0.0
	golang.org/x/text v0.29.0
	google.golang.org/protobuf v1.36.9
)

require (
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/grpc v1.75.0 // indirect
)
