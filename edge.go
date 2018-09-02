package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Edge interface for flowgraph edges that connect flowgraph nodes
type Edge interface {

	// Name returns the edge name
	Name() string

	// Connect connects an upstream node to a downstream node
	Connect(upstream, dnstream Node, upname, dnname string)

	// Value returns the edge's current value
	Value() interface{}

	// Source returns upstream node by index
	Source(i int) Node

	// Destination return downstream node by index
	Destination(i int) Node

	// NumSource returns the number of upstream nodes
	NumSource() int

	// NumDestination returns the number of downstream nodes
	NumDestination() int
}

// implementation of Edge
type edge struct {
	base *fgbase.Edge
}

// Name returns the edge name
func (e edge) Name() string {
	return e.base.Name
}

// Connect connects an upstream node to a downstream node
func (e edge) Connect(upstream, dnstream Node, upname, dnname string) {
}

// Value returns the edge's current value
func (e edge) Value() interface{} {
	return e.base.Val
}

// Source returns upstream node by index
func (e edge) Source(i int) Node {
	return node{e.base.SrcNode(i)}
}

// Destination return upstream node by index
func (e edge) Destination(i int) Node {
	return node{e.base.DstNode(i)}
}

// NumSource returns the number of upstream nodes
func (e edge) NumSource() int {
	return e.base.SrcCnt()
}

// NumDestination returns the number of downstream nodes
func (e edge) NumDestination() int {
	return e.base.DstCnt()
}
