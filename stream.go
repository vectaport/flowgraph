package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Stream struct for flowgraph streams that connect flowgraph hubs
type Stream struct {
	base *fgbase.Edge
}

// Name returns the stream name
func (s *Stream) Name() string {
	return s.base.Name
}

// Upstream returns upstream hub by index
func (s *Stream) Upstream(i int) *Hub {
	return &Hub{s.base.SrcNode(i)}
}

// Downstream returns upstream hub by index
func (s *Stream) Downstream(i int) *Hub {
	return &Hub{s.base.DstNode(i)}
}

// NumUpstream returns the number of upstream hubs
func (s *Stream) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream hubs
func (s *Stream) NumDownstream() int {
	return s.base.DstCnt()
}

// Empty returns true if the underlying implementation is nil
func (s *Stream) Empty() bool {
	return s.base == nil
}

// Init sets an initial value for flow
func (s *Stream) Init(v interface{}) {
	s.Edge().Val = v
}

// Edge returns pointer to underlying Edge
func (s *Stream) Edge() *fgbase.Edge {
	return s.base
}
