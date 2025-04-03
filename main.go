package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bloXroute-Labs/base-streamer-client-go/examples"
	"github.com/joho/godotenv"
)

func main() {
	// Define and parse command line flags
	numBlocks := flag.Int("blocks", 10, "Number of blocks to listen for")
	quiet := flag.Bool("quiet", false, "Suppress log output")
	flag.Parse()

	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	// Display startup information
	if !*quiet {
		fmt.Printf("Listening for blocks (limit: %d)...\n", *numBlocks)
	}

	// Start the listener (convert int to uint64 as required by the function)
	err := examples.ListenForBdnBlocks(uint64(*numBlocks))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
