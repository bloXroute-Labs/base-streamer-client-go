package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bloXroute-Labs/base-streamer-client-go/provider"
	streamerapi "github.com/bloXroute-Labs/base-streamer-proto/streamer_api"
	"github.com/ethereum/go-ethereum/core/types"
)

func ListenForBdnBlocks(numberOfBlocks uint64) error {
	grpcClient, err := provider.NewGRPCClient()
	if err != nil {
		return err
	}
	blocksChan := make(chan *streamerapi.GetBdnBlockStreamResponse)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.GetBdnBlockStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream with: %v", err)
	}
	stream.Into(blocksChan)

	fmt.Println("waiting on blocks channel")
	for range numberOfBlocks {
		bdnBlock, ok := <-blocksChan
		if !ok {
			// channel closed
			return fmt.Errorf("bdn blocks channel closed")
		}
		updateTime := time.Now()

		blockHeader := &types.Header{}
		err := blockHeader.UnmarshalJSON(bdnBlock.BlockHeader)
		if err != nil {
			return fmt.Errorf("failed to unmarshal block header with: %v", err)
		}

		var blockBody *types.Body
		err = json.Unmarshal(bdnBlock.BlockBody, &blockBody)
		if err != nil {
			return fmt.Errorf("failed to unmarshal block body with: %v", err)
		}

		fmt.Printf("bdn block: %v, %v txns at %v\n", blockHeader.Number.Uint64(), len(blockBody.Transactions), updateTime.UTC())
	}
	return nil
}
