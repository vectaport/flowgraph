package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Stream interface for flowgraph streams that connect flowgraph nodes
type Stream interface {

	// Name returns the stream name
	Name() string

	// Connect connects an upstream node to a downstream node
	Connect(upstream, dnstream Node, upname, dnname string)

	// Upstream returns upstream node by index
	Upstream(i int) Node

	// Downstream return downstream node by index
	Downstream(i int) Node

	// NumUpstream returns the number of upstream nodes
	NumUpstream() int

	// NumDownstream returns the number of downstream nodes
	NumDownstream() int

	// Base returns the value that implements this stream
	// The type of this value identifies the implementation.
	Base() interface{}
}

// implementation of Stream
type stream struct {
	base *fgbase.Edge
}

// Name returns the stream name
func (s stream) Name() string {
	return s.base.Name
}

// Connect connects an upstream node to a downstream node
func (s stream) Connect(upstream, dnstream Node, upname, dnname string) {
}

// Upstream returns upstream node by index
func (s stream) Upstream(i int) Node {
	return node{s.base.SrcNode(i)}
}

// Downstream returns upstream node by index
func (s stream) Downstream(i int) Node {
	return node{s.base.DstNode(i)}
}

// NumUpstream returns the number of upstream nodes
func (s stream) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream nodes
func (s stream) NumDownstream() int {
	return s.base.DstCnt()
}

// Base returns the value that implements this stream
// The type of this value identifies the implementation.
func (s stream) Base() interface{} {
	return s.base
}
