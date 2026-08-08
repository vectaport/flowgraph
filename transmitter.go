package flowgraph

// Transmitter transmits one value using a Transmit method.
// Provide as init arg to NewHub with Transmit HubCode.
// Use Hub.Tracef for tracing.
type Transmitter interface {
	Transmit(h Hub, source any) (err error)
}
