package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// Pipe interface
type Pipe interface {
	// Tracef for debug trace printing.  Use atomic log mechanism.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.  Use atomic log mechanism.
	LogError(format string, v ...interface{})

	// Name returns the pipe name
	Name() string

	// Source returns the nth incoming connector
	Source(n int) Connector

	// Destination returns the nth outgoing connector
	Destination(n int) Connector

	// Auxiliary returns auxiliary storage for this pipe used by
	// the underlying implementation for storing state
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

// Source returns the nth incoming connector
func (p pipe) Source(n int) Connector {
	return conn{p.node.Srcs[n]}
}

// Destination returns the nth outgoing connector
func (p pipe) Destination(n int) Connector {
	return conn{p.node.Dsts[n]}
}

// Auxiliary returns auxiliary storage for this pipe used by
// the underlying implementation for storing state
func (p pipe) Auxiliary() interface{} {
	return p.node.Aux
}
