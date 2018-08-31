package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// Pipe interface for flowgraph nodes that are connected by Connector edges
type Pipe interface {
	// Tracef for debug trace printing.  Use atomic log mechanism.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.  Use atomic log mechanism.
	LogError(format string, v ...interface{})

	// Name returns the pipe name
	Name() string

        // Source returns upstream connector by index
	Source(n int) Connector

	// Destination returns the nth outgoing connector
	Destination(n int) Connector

	// NumSource returns the number of upstream connectors
	NumSource() int

        // NumDestination returns the number of upstream connectors
	NumDestination() int

	// Auxiliary returns auxiliary storage used by
	// underlying implementation for storing state
	// (only required for external debug of fgbase)
	Auxiliary() interface{}
}

// implementation of Pipe
type pipe struct {
	node *fgbase.Node
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (p pipe) Tracef(format string, v ...interface{}) {
	p.node.Tracef(format, v)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (p pipe) LogError(format string, v ...interface{}) {
	p.node.LogError(format, v)
}

// Name returns the pipe name
func (p pipe) Name() string {
	return p.node.Name
}

// Source returns upstream connector by index
func (p pipe) Source(i int) Connector {
	return conn{p.node.Srcs[i]}
}

// Destination returns downstream connector by index
func (p pipe) Destination(i int) Connector {
	return conn{p.node.Dsts[i]}
}

// NumSource returns the number of upstream connectors
func (p pipe) NumSource() int {
	return len(p.node.Srcs)
}

// NumDestination returns the number of downstream connectors
func (p pipe) NumDestination() int {
	return len(p.node.Dsts)
}

// Auxiliary returns auxiliary storage for this pipe used by
// the underlying implementation for storing state
func (p pipe) Auxiliary() interface{} {
	return p.node.Aux
}
