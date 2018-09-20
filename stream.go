package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Stream interface for flowgraph streams that connect flowgraph hubs
type Stream interface {

	// Name returns the stream name
	Name() string

	// Upstream returns upstream hub by index
	Upstream(i int) Hub

	// Downstream return downstream hub by index
	Downstream(i int) Hub

	// NumUpstream returns the number of upstream hubs
	NumUpstream() int

	// NumDownstream returns the number of downstream hubs
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

// Upstream returns upstream hub by index
func (s stream) Upstream(i int) Hub {
	return hub{s.base.SrcNode(i)}
}

// Downstream returns upstream hub by index
func (s stream) Downstream(i int) Hub {
	return hub{s.base.DstNode(i)}
}

// NumUpstream returns the number of upstream hubs
func (s stream) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream hubs
func (s stream) NumDownstream() int {
	return s.base.DstCnt()
}

// Base returns the value that implements this stream
// The type of this value identifies the implementation.
func (s stream) Base() interface{} {
	return s.base
}
