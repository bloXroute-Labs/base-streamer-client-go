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
	return NewCustomGRPCClient(MainnetGRPC, false) // TODO: change to true when we have a secure connection for production
}

// NewGRPCLocal connects to local provider
func NewGRPCLocal() (*GRPCClient, error) {
	return NewCustomGRPCClient(LocalGRPC, false)
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
	}
	return client, nil
}

// GetBdnBlockStream subscribes to BDN block stream
func (g *GRPCClient) GetBdnBlockStream(
	ctx context.Context,
) (connections.Streamer[*streamerapi.GetBdnBlockStreamResponse], error) {
	factory := func() (connections.Streamer[*streamerapi.GetBdnBlockStreamResponse], error) {
		stream, err := g.apiClient.GetBdnBlockStream(ctx, &streamerapi.GetBdnBlockStreamRequest{})
		if err != nil {
			return nil, err
		}
		return connections.GRPCStream[streamerapi.GetBdnBlockStreamResponse](stream, "BDN block stream"), nil
	}

	options := connections.DefaultReconnectingOptions()
	reconnectingStreamer, err := connections.NewReconnectingStreamer(factory, options)
	if err != nil {
		return nil, err
	}

	return reconnectingStreamer.Streamer(), nil
}
