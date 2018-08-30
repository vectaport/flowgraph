package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Connector interface for flowgraph edges that connect pipe nodes
type Connector interface {

	// Name returns the connector name
	Name() string

	// Value returns the connector's current value
	Value() interface{}

	// Source returns the nth downstream pipe for this connector
	Source(i int) Pipe

	// Destination returns the nth upstream pipe for this connector
	Destination(i int) Pipe
}

// implementation of Connector
type conn struct {
	edge *fgbase.Edge
}

// Name returns the connector name
func (c conn) Name() string {
	return c.edge.Name
}

// Value returns the connector's current value
func (c conn) Value() interface{} {
	return c.edge.Val
}

// Source returns the nth downstream pipe for this connector
func (c conn) Source(n int) Pipe {
	return pipe{c.edge.SrcNode(n)}
}

// Destination returns the nth upstream pipe for this connector
func (c conn) Destination(n int) Pipe {
	return pipe{c.edge.DstNode(n)}
}
