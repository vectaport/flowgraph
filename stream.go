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

	// Downstream returns upstream hub by index
	Downstream(i int) Hub

	// NumUpstream returns the number of upstream hubs
	NumUpstream() int

	// NumDownstream returns the number of downstream hubs
	NumDownstream() int

	// Empty returns true if the underlying implementation is nil
	Empty() bool

	// Init sets an initial value for flow
	Init(v interface{}) Stream

	// Const sets a value for continual flow
	Const(v interface{}) Stream

	// Sink sets a stream to be a sink
	Sink() Stream

	// IsConst returns rtue if stream is a constant
	IsConst() bool

	// IsSink returns true if stream is a sink
	IsSink() bool

	// Base returns value of underlying implementation
	Base() interface{}
}

// Stream implementation
type stream struct {
	base *fgbase.Edge
}

// Name returns the stream name
func (s *stream) Name() string {
	return s.base.Name
}

// Upstream returns upstream hub by index
func (s *stream) Upstream(i int) Hub {
	return &hub{s.base.SrcNode(i)}
}

// Downstream returns upstream hub by index
func (s *stream) Downstream(i int) Hub {
	return &hub{s.base.DstNode(i)}
}

// NumUpstream returns the number of upstream hubs
func (s *stream) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream hubs
func (s *stream) NumDownstream() int {
	return s.base.DstCnt()
}

// Empty returns true if the underlying implementation is nil
func (s *stream) Empty() bool {
	return s.base == nil
}

// Init sets an initial value for flow
func (s *stream) Init(v interface{}) Stream {
	s.Base().(*fgbase.Edge).Val = v
	return s
}

// Const sets a value for continual flow
func (s *stream) Const(v interface{}) Stream {
	s.Base().(*fgbase.Edge).Const(v)
	return s
}

// Sink sets a stream to be a sink
func (s *stream) Sink() Stream {
	s.Base().(*fgbase.Edge).Sink()
	return s
}

// IsConst returns true if stream is a constant
func (s *stream) IsConst() bool {
	return s.Base().(*fgbase.Edge).IsConst()
}

// IsSink returns true if stream is a sink
func (s *stream) IsSink() bool {
	return s.Base().(*fgbase.Edge).IsSink()
}

// Base returns value of underlying implementation
func (s *stream) Base() interface{} {
	return s.base
}
