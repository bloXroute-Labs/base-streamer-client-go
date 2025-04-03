package examples

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bloXroute-Labs/base-streamer-client-go/provider"
	streamerapi "github.com/bloXroute-Labs/base-streamer-proto/streamer_api"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// ListenForBdnBlocks connects to a BDN and listens for blocks indefinitely
// If numberOfBlocks is 0, it will listen indefinitely
func ListenForBdnBlocks(numberOfBlocks uint64) error {
	grpcClient, err := provider.NewGRPCClient()
	if err != nil {
		return err
	}

	blocksChan := make(chan *streamerapi.GetBdnBlockStreamResponse)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal, closing connections...")
		cancel()
	}()

	fmt.Println("Connecting to BDN for block streaming...")
	stream, err := grpcClient.GetBdnBlockStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream with: %v", err)
	}
	stream.Into(blocksChan)

	fmt.Println("Waiting for blocks...")
	var count uint64 = 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		case bdnBlock, ok := <-blocksChan:
			if !ok {
				// This should not happen with reconnection logic in place
				// but we'll handle it just in case
				fmt.Println("Block channel closed unexpectedly, reconnection should happen automatically...")
				time.Sleep(time.Second) // Brief pause to avoid tight loop
				continue
			}

			blockHeader := &types.Header{}
			err := rlp.DecodeBytes(bdnBlock.BlockHeader, blockHeader)
			if err != nil {
				return fmt.Errorf("failed to RLP decode block header with: %v", err)
			}

			blockBody := &types.Body{}
			err = rlp.DecodeBytes(bdnBlock.BlockBody, blockBody)
			if err != nil {
				return fmt.Errorf("failed to RLP decode block body with: %v", err)
			}

			fmt.Printf("Block #%d: %v, %v txns\n",
				count,
				blockHeader.Number.Uint64(),
				len(blockBody.Transactions))

			count++
			// If we've reached the requested number of blocks and it's not 0 (indefinite), exit
			if numberOfBlocks > 0 && count >= numberOfBlocks {
				return nil
			}
		}
	}
}
