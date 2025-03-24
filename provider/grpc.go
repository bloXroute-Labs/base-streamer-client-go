package provider

import (
	"context"
	"os"

	"github.com/bloXroute-Labs/base-streamer-client-go/connections"

	streamerapi "github.com/bloXroute-Labs/base-streamer-proto/streamer_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Version = "v0"
	Name    = "base-streamer"
)

type GRPCClient struct {
	streamerapi.UnimplementedApiServer

	apiClient streamerapi.ApiClient
}

type blxrCredentials struct {
	authorization string
}

func (bc blxrCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": bc.authorization,
		"x-sdk":         Name,
		"x-sdk-version": Version,
	}, nil
}

func (bc blxrCredentials) RequireTransportSecurity() bool {
	return false
}

// NewCustomGRPCClient connects to custom provider
func NewCustomGRPCClient(endpoint string) (*GRPCClient, error) {
	opts := DefaultRPCOpts(endpoint)
	return NewGRPCClientWithOpts(opts)
}

// NewGRPCClient connects to main provider
func NewGRPCClient() (*GRPCClient, error) {
	return NewCustomGRPCClient(MainnetGRPC)
}

// NewGRPCLocal connects to local provider
func NewGRPCLocal() (*GRPCClient, error) {
	return NewCustomGRPCClient(LocalGRPC)
}

func DefaultRPCOpts(endpoint string) RPCOpts {
	return RPCOpts{
		Endpoint:   endpoint,
		AuthHeader: os.Getenv("AUTH_HEADER"),
	}
}

// NewGRPCClientWithOpts connects to custom provider
func NewGRPCClientWithOpts(opts RPCOpts, dialOpts ...grpc.DialOption) (*GRPCClient, error) {
	var (
		conn     grpc.ClientConnInterface
		err      error
		grpcOpts = make([]grpc.DialOption, 0)
	)

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	// if opts.UseTLS {
	// 	transportOption = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	// }
	grpcOpts = append(grpcOpts, transportOption)

	if !opts.DisableAuth {
		grpcOpts = append(grpcOpts, grpc.WithPerRPCCredentials(blxrCredentials{authorization: opts.AuthHeader}))
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(&grpc.MaxRecvMsgSizeCallOption{MaxRecvMsgSize: 1024 * 1024 * 16}))
	grpcOpts = append(grpcOpts, dialOpts...)
	conn, err = grpc.Dial(opts.Endpoint, grpcOpts...)
	if err != nil {
		return nil, err
	}

	client := &GRPCClient{
		apiClient: streamerapi.NewApiClient(conn),
	}
	return client, nil
}

// GetNewBlockStream subscribes to newly created block stream
func (g *GRPCClient) GetNewBlockStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetNewBlockStreamResponse], error) {
	stream, err := g.apiClient.GetNewBlockStream(ctx, &streamerapi.GetNewBlockStreamRequest{})
	if err != nil {
		return nil, err
	}

	return connections.GRPCStream[streamerapi.GetNewBlockStreamResponse](stream, ""), nil
}
