package echo

//go:generate go run "github.com/artistml/toolkits/cmd/gen-grpc" -n echo -d github.com/artistml/toolkits/pkg/proto -v v1 -v v2 -s v1.EchoService -s v2.EchoService -i github.com/artistml/toolkits/pkg/interceptor/logging gateway rpc
