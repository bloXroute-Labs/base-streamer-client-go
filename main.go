package main

import (
	"github.com/bloXroute-Labs/base-streamer-client-go/examples"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	err := examples.ListenForBdnBlocks(10)
	if err != nil {
		panic(err)
	}
}
