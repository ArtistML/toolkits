package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// IncomingToOutgoingContext fetches metadata from incoming context and creates a new outgoing.
func IncomingToOutgoingContext(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
