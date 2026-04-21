package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/bloXroute-Labs/base-streamer-client-go/provider"
	streamerapi "github.com/bloXroute-Labs/base-streamer-proto/streamer_api"
)

func ListenForParsedBdnFlashBlocks(numberOfBlocks uint64) error {
	grpcClient, err := provider.NewGRPCClient()
	if err != nil {
		return err
	}
	flashblocksChan := make(chan *streamerapi.GetParsedBdnFlashBlockStreamResponse)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.GetParsedBdnFlashBlockStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream with: %v", err)
	}
	stream.Into(flashblocksChan)

	fmt.Println("waiting on parsed bdn flashblocks channel")
	for range numberOfBlocks {
		flashblock, ok := <-flashblocksChan
		if !ok {
			// channel closed
			return fmt.Errorf("parsed bdn flashblocks channel closed")
		}
		updateTime := time.Now()
		fmt.Printf("bdn flash block: %v, index: %v, at %v\n", flashblock.Metadata.BlockNumber, flashblock.Index, updateTime.UTC())
	}
	return nil
}
