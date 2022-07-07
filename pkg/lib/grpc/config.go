package grpc

const ServerKey = "server.grpc"
const GatewayServerKey = "server.gateway"
const UnaryServerInterceptorsKey = "server.interceptors.unary"
const StreamServerInterceptorsKey = "server.interceptors.stream"

// GatewayServerConfig setting for http gateway.
type GatewayServerConfig struct {
	Name         string `json:"name"`
	Scheme       string `json:"scheme"`
	Address      string `json:"address"`
	GRPCEndpoint string `json:"grpcEndpoint"`
}

// ServerConfig setting for grpc server.
type ServerConfig struct {
	Name    string `json:"name"`
	Network string `json:"network"`
	Address string `json:"address"`
}
