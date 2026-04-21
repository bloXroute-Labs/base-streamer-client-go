module github.com/bloXroute-Labs/base-streamer-client-go

go 1.25.1

replace github.com/ethereum/go-ethereum => github.com/ethereum-optimism/op-geth v1.101503.1-dev.1

require (
	github.com/bloXroute-Labs/base-streamer-proto v1.1.3
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.71.0
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
