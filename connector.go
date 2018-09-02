package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Connector interface for flowgraph edges that connect pipe nodes
type Connector interface {

	// Name returns the connector name
	Name() string

	// Connect connects an upstream pipe to a downstream pipe
	Connect(upstream, dnstream Pipe, upname, dnname string)

	// Value returns the connector's current value
	Value() interface{}

	// Source returns upstream pipe by index
	Source(i int) Pipe

	// Destination return downstream pipe by index
	Destination(i int) Pipe

	// NumSource returns the number of upstream pipes
	NumSource() int

	// NumDestination returns the number of downstream pipes
	NumDestination() int
}

// implementation of Connector
type conn struct {
	edge *fgbase.Edge
}

// Name returns the connector name
func (c conn) Name() string {
	return c.edge.Name
}

// Connect connects an upstream pipe to a downstream pipe
func (c conn) Connect(upstream, dnstream Pipe, upname, dnname string) {
}

// Value returns the connector's current value
func (c conn) Value() interface{} {
	return c.edge.Val
}

// Source returns upstream pipe by index
func (c conn) Source(i int) Pipe {
	return pipe{c.edge.SrcNode(i)}
}

// Destination return upstream pipe by index
func (c conn) Destination(i int) Pipe {
	return pipe{c.edge.DstNode(i)}
}

// NumSource returns the number of upstream pipes
func (c conn) NumSource() int {
	return c.edge.SrcCnt()
}

// NumDestination returns the number of downstream pipes
func (c conn) NumDestination() int {
	return c.edge.DstCnt()
}
