package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Stream interface for flowgraph streams that connect flowgraph hubs
type Stream interface {

	// Name returns the stream name
	Name() string

	// SetName sets the stream name
	SetName(name string)

	// Upstream returns upstream hub by index
	Upstream(i int) Hub

	// Downstream returns downstream hub by index
	Downstream(i int) Hub

	// NumUpstream returns the number of upstream hubs
	NumUpstream() int

	// NumDownstream returns the number of downstream hubs
	NumDownstream() int

	// Init sets an initial value for flow
	Init(v interface{}) Stream

	// Const sets a value for continual flow
	Const(v interface{}) Stream

	// Sink sets a stream to be a sink
	Sink() Stream

	// IsConst returns true if stream is a constant
	IsConst() bool

	// IsSink returns true if stream is a sink
	IsSink() bool

	// Same returns true if two streams are the same underneath
	Same(Stream) bool

	// Empty returns true if the underlying implementation is nil
	Empty() bool

	// Flowgraph returns associated flowgraph
	Flowgraph() Flowgraph

	// Base returns value of underlying implementation
	Base() interface{}
}

// Stream implementation
type stream struct {
	base *fgbase.Edge
	fg   *flowgraph
}

// Name returns the stream name
func (s *stream) Name() string {
	return s.base.Name
}

// SetName sets the stream name
func (s *stream) SetName(name string) {
	s.base.SetName(name)
}

// Upstream returns upstream hub by index
func (s *stream) Upstream(i int) Hub {
	if s.base != nil && s.base.SrcNode(i) != nil {
		return s.base.SrcNode(i).Owner.(Hub)
	}
	return nil
}

// Downstream returns upstream hub by index
func (s *stream) Downstream(i int) Hub {
	if s.base != nil && s.base.DstNode(i) != nil {
		return s.base.DstNode(i).Owner.(Hub)
	}
	return nil
}

// NumUpstream returns the number of upstream hubs
func (s *stream) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream hubs
func (s *stream) NumDownstream() int {
	return s.base.DstCnt()
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

// Same returns true if two streams are the same underneath
func (s *stream) Same(s2 Stream) bool {
	checkInternalStream(s.fg, s2)

	if s.Base().(*fgbase.Edge) == nil {
		return s2.Base().(*fgbase.Edge) == nil
	}
	if s2.Base().(*fgbase.Edge) == nil {
		return s.Base().(*fgbase.Edge) == nil
	}
	return s.Base().(*fgbase.Edge).Same(s2.Base().(*fgbase.Edge))
}

// Empty returns true if the underlying implementation is nil
func (s *stream) Empty() bool {
	return s.base == nil
}

// Flowgraph returns associate flowgraph interface
func (s *stream) Flowgraph() Flowgraph {
	return s.fg
}

// Base returns value of underlying implementation
func (s *stream) Base() interface{} {
	return s.base
}
