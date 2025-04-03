package provider

var (
	MainnetGRPC = "base.blxrbdn.com:443"
)

type RPCOpts struct {
	Endpoint    string
	DisableAuth bool
	UseTLS      bool
	AuthHeader  string
}
