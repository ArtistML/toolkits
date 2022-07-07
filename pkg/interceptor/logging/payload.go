package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gogf/gf/frame/g"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	rawJSONKeys = map[string]interface{}{}

	// SystemTag is tag representing an event inside gRPC call.
	SystemTag = []string{"protocol", "grpc"}
	// ComponentFieldKey is a tag representing the client/server that is calling.
	ComponentFieldKey    = "grpc.component"
	KindServerFieldValue = "server"
	KindClientFieldValue = "client"
	ServiceFieldKey      = "grpc.service"
	MethodFieldKey       = "grpc.method"
	MethodTypeFieldKey   = "grpc.method_type"
)

// extractFields returns all fields from tags.
func extractFields(tags tags.Tags) logrus.Fields {
	var fields = logrus.Fields{}
	for k, v := range tags.Values() {
		fields[k] = v
	}
	return fields
}

func commonFields(kind string, typ interceptors.GRPCType, service string, method string) logrus.Fields {
	return logrus.Fields{
		SystemTag[0]:       SystemTag[1],
		ComponentFieldKey:  kind,
		ServiceFieldKey:    service,
		MethodFieldKey:     method,
		MethodTypeFieldKey: string(typ),
	}
}

type serverPayloadReporter struct {
	ctx   context.Context
	entry *logrus.Entry
}

func (c *serverPayloadReporter) PostCall(error, time.Duration) {}

func (c *serverPayloadReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	entry := c.entry.WithFields(extractFields(tags.Extract(c.ctx)))
	p, ok := req.(proto.Message)
	if !ok {
		entry.WithField("req.type", fmt.Sprintf("%T", req)).Error("req is not a google.golang.org/protobuf/proto.Message; programmatic error?")

		return
	}
	// For server recv message is the request.
	entry.WithFields(logrus.Fields{
		"grpc.send.duration":    duration.String(),
		"grpc.response.content": p,
	}).Info("response")
}

func (c *serverPayloadReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}

	entry := c.entry.WithFields(extractFields(tags.Extract(c.ctx)))
	p, ok := reply.(proto.Message)
	if !ok {
		entry.WithField("reply.type", fmt.Sprintf("%T", reply)).Error("reply is not a google.golang.org/protobuf/proto.Message; programmatic error?")
		return
	}
	// For server recv message is the request.
	entry.WithFields(logrus.Fields{
		"grpc.recv.duration":   duration.String(),
		"grpc.request.content": p,
	}).Info("request")
}

type payloadReportable struct {
	serverDecider   logging.ServerPayloadLoggingDecider
	logger          *logrus.Logger
	timestampFormat string
}

func (r *payloadReportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.serverDecider(ctx, interceptors.FullMethod(service, method), req) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(logging.KindServerFieldValue, typ, service, method)
	fields["grpc.start_time"] = time.Now().Format(r.timestampFormat)
	if d, ok := ctx.Deadline(); ok {
		fields["grpc.request.deadline"] = d.Format(r.timestampFormat)
	}
	return &serverPayloadReporter{
		ctx:   ctx,
		entry: r.logger.WithFields(fields),
	}, ctx
}

func logWriter() io.Writer {
	logDir := g.Cfg().GetString("logger.Path", "/var/log/nebula")
	fr := FileRotator{FileName: filepath.Join(logDir, "audit.log")}
	if g.Cfg().GetBool("logger.StdoutPrint", false) {
		// NOTE: should use stderr for default log.
		return io.MultiWriter(&fr, os.Stderr)
	}
	return &fr
}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadUnaryServerInterceptor(formatter logrus.Formatter, decider logging.ServerPayloadLoggingDecider, timestampFormat string) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&payloadReportable{
		logger: &logrus.Logger{
			Formatter: formatter,
			Level:     logrus.InfoLevel,
			Out:       logWriter(),
		},
		serverDecider:   decider,
		timestampFormat: timestampFormat,
	})
}

// PayloadStreamServerInterceptor returns a new stream server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadStreamServerInterceptor(formatter logrus.Formatter, decider logging.ServerPayloadLoggingDecider, timestampFormat string) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&payloadReportable{
		logger: &logrus.Logger{
			Formatter: formatter,
			Level:     logrus.InfoLevel,
			Out:       logWriter(),
		},
		serverDecider:   decider,
		timestampFormat: timestampFormat,
	})
}
