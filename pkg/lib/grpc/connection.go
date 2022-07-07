package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(ctx context.Context, server string) (*grpc.ClientConn, error) {
	receiveMsgSize, sendMsgSize := GetReceiveAndSendMsgSize()
	return grpc.DialContext(ctx, server,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(receiveMsgSize),
			grpc.MaxCallSendMsgSize(sendMsgSize),
		),
	)
}
