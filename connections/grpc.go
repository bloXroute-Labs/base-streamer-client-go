package connections

import (
	"fmt"
	"io"

	"google.golang.org/grpc"
)

// GRPCStream creates a Streamer that reads from a gRPC client stream
// The input parameter is used for logging and debugging purposes
func GRPCStream[T any](stream grpc.ClientStream, input string) Streamer[*T] {
	var generator Streamer[*T] = func() (*T, error) {
		m := new(T)
		err := stream.RecvMsg(m)
		if err == io.EOF {
			return nil, fmt.Errorf("stream for %s ended successfully", input)
		} else if err != nil {
			return nil, fmt.Errorf("stream for %s error: %v", input, err)
		}
		return m, nil
	}

	return generator
}
