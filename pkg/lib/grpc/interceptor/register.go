package interceptor

import (
	"github.com/gogf/gf/frame/g"
	"google.golang.org/grpc"
)

var (
	unaryInterceptors  = make(map[string]grpc.UnaryServerInterceptor)
	streamInterceptors = make(map[string]grpc.StreamServerInterceptor)
)

// RegisterUnary registers a grpc.UnaryServerInterceptor.
// It is non-thread-safe, should be called in main goroutine.
func RegisterUnary(name string, unary grpc.UnaryServerInterceptor) {
	if _, ok := unaryInterceptors[name]; ok {
		g.Log().Warningf("overwriting an unary interceptor: name=%s", name)
	}
	unaryInterceptors[name] = unary
}

// RegisterStream registers a grpc.StreamServerInterceptor.
// It is non-thread-safe, should be called in main goroutine.
func RegisterStream(name string, stream grpc.StreamServerInterceptor) {
	if _, ok := streamInterceptors[name]; ok {
		g.Log().Warningf("overwriting an stream interceptor: name=%s", name)
	}
	streamInterceptors[name] = stream
}

// GetUnary returns a registered grpc.UnaryServerInterceptor with name.
func GetUnary(name string) (grpc.UnaryServerInterceptor, bool) {
	unary, ok := unaryInterceptors[name]
	return unary, ok
}

// GetStream returns a registered grpc.StreamServerInterceptor with name.
func GetStream(name string) (grpc.StreamServerInterceptor, bool) {
	stream, ok := streamInterceptors[name]
	return stream, ok
}
