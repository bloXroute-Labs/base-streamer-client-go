package provider

var (
	MainnetGRPC = "base-poc.blxrbdn.com:8080"
	LocalGRPC   = "localhost:8080"
)

type RPCOpts struct {
	Endpoint    string
	DisableAuth bool
	// UseTLS         bool
	AuthHeader string
}
