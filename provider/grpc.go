package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/bloXroute-Labs/base-streamer-client-go/connections"

	streamerapi "github.com/bloXroute-Labs/base-streamer-proto/streamer_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Version = "v0"
	Name    = "base-streamer"
)

type GRPCClient struct {
	streamerapi.UnimplementedApiServer

	apiClient streamerapi.ApiClient
	Endpoint  string
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
func NewCustomGRPCClient(endpoint string, secure bool) (*GRPCClient, error) {
	authHeader, exists := os.LookupEnv("AUTH_HEADER")
	if !exists || authHeader == "" {
		return nil, fmt.Errorf("auth header not found")
	}

	opts := DefaultRPCOpts(endpoint, authHeader, secure)
	return NewGRPCClientWithOpts(opts)
}

// NewGRPCClient connects to main provider
func NewGRPCClient() (*GRPCClient, error) {
	return NewCustomGRPCClient(MainnetGRPC, true)
}

func DefaultRPCOpts(endpoint string, authHeader string, secure bool) RPCOpts {
	return RPCOpts{
		Endpoint:   endpoint,
		AuthHeader: authHeader,
		UseTLS:     secure,
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
	if opts.UseTLS {
		transportOption = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
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
		Endpoint:  opts.Endpoint,
	}
	return client, nil
}

// GetBdnBlockStream subscribes to BDN block stream
func (g *GRPCClient) GetBdnBlockStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetBdnBlockStreamResponse], error) {
	stream, err := g.apiClient.GetBdnBlockStream(ctx, &streamerapi.GetBdnBlockStreamRequest{})
	if err != nil {
		return nil, err
	}

	return connections.GRPCStream[streamerapi.GetBdnBlockStreamResponse](stream, ""), nil
}

// GetBdnFlashBlockStream subscribes to BDN flash block stream
func (g *GRPCClient) GetBdnFlashBlockStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetBdnFlashBlockStreamResponse], error) {
	stream, err := g.apiClient.GetBdnFlashBlockStream(ctx, &streamerapi.GetBdnFlashBlockStreamRequest{})
	if err != nil {
		return nil, err
	}

	return connections.GRPCStream[streamerapi.GetBdnFlashBlockStreamResponse](stream, ""), nil
}

// GetParsedBdnFlashBlockStream subscribes to parsed BDN flash block stream
func (g *GRPCClient) GetParsedBdnFlashBlockStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetParsedBdnFlashBlockStreamResponse], error) {
	stream, err := g.apiClient.GetParsedBdnFlashBlockStream(ctx, &streamerapi.GetParsedBdnFlashBlockStreamRequest{})
	if err != nil {
		return nil, err
	}

	return connections.GRPCStream[streamerapi.GetParsedBdnFlashBlockStreamResponse](stream, ""), nil
}

// GetBdnFBTxnStateDiffStream subscribes to BDN flash block transaction state diff stream
func (g *GRPCClient) GetBdnFBTxnStateDiffStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetBdnFBTxnStateDiffStreamResponse], error) {
	stream, err := g.apiClient.GetBdnFBTxnStateDiffStream(ctx, &streamerapi.GetBdnFBTxnStateDiffStreamRequest{})
	if err != nil {
		return nil, err
	}

	return connections.GRPCStream[streamerapi.GetBdnFBTxnStateDiffStreamResponse](stream, ""), nil
}
