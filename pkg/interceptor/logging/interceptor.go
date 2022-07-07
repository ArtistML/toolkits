package logging

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	"github.com/opentracing/basictracer-go"

	"github.com/artistml/toolkits/pkg/lib/grpc/interceptor"
)

// NOTE(leventeliu): VBarFormatter is currently used for BCS log collector. Need to add option to switch formatter later.

func init() {
	interceptor.RegisterUnary("tags", tags.UnaryServerInterceptor())
	interceptor.RegisterUnary("tracing", tracing.UnaryServerInterceptor(tracing.WithTracer(basictracer.New(&nopRecoder{}))))
	interceptor.RegisterUnary("payload-logging", PayloadUnaryServerInterceptor(&VBarFormatter{}, logAny, time.RFC3339Nano))
	interceptor.RegisterStream("tags", tags.StreamServerInterceptor())
	interceptor.RegisterStream("tracing", tracing.StreamServerInterceptor(tracing.WithTracer(basictracer.New(&nopRecoder{}))))
	interceptor.RegisterStream("payload-logging", PayloadStreamServerInterceptor(&VBarFormatter{}, logAny, time.RFC3339Nano))
}

var logAny logging.ServerPayloadLoggingDecider = func(_ context.Context, _ string, _ interface{}) bool {
	return true
}

type nopRecoder struct{}

// RecordSpan implements basictracer.SpanRecorder.
func (nopRecoder) RecordSpan(_ basictracer.RawSpan) {}
