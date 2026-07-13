package flowgraph

import (
	"github.com/vectaport/fgbase"
)

import ()

// Renamed Stream to Pipe: honors Doug McIlroy's Unix pipes, names what
// actually connects to a hub (not the stream flowing within it), and marks
// real backpressure -- unlike streams, which can overflow and cause havoc.

// Pipe interface for flowgraph pipes that connect flowgraph hubs
type Pipe interface {

	// Name returns the pipe name
	Name() string

	// SetName sets the pipe name
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
	Init(v interface{}) Pipe

	// Const sets a value for continual flow
	Const(v interface{}) Pipe

	// Sink sets a pipe to be a sink
	Sink() Pipe

	// IsConst returns true if pipe is a constant
	IsConst() bool

	// IsSink returns true if pipe is a sink
	IsSink() bool

	// Same returns true if two pipes are the same underneath
	Same(Pipe) bool

	// Empty returns true if the underlying implementation is nil
	Empty() bool

	// Flowgraph returns associated flowgraph
	Flowgraph() Flowgraph

	// Base returns value of underlying implementation
	Base() interface{}
}

// Pipe implementation
type pipe struct {
	base *fgbase.Edge
	fg   *flowgraph
}

// Name returns the pipe name
func (s *pipe) Name() string {
	return s.base.Name
}

// SetName sets the pipe name
func (s *pipe) SetName(name string) {
	s.base.SetName(name)
}

// Upstream returns upstream hub by index
func (s *pipe) Upstream(i int) Hub {
	if s.base != nil && s.base.SrcNode(i) != nil {
		return s.base.SrcNode(i).Owner.(Hub)
	}
	return nil
}

// Downstream returns upstream hub by index
func (s *pipe) Downstream(i int) Hub {
	if s.base != nil && s.base.DstNode(i) != nil {
		return s.base.DstNode(i).Owner.(Hub)
	}
	return nil
}

// NumUpstream returns the number of upstream hubs
func (s *pipe) NumUpstream() int {
	return s.base.SrcCnt()
}

// NumDownstream returns the number of downstream hubs
func (s *pipe) NumDownstream() int {
	return s.base.DstCnt()
}

// Init sets an initial value for flow
func (s *pipe) Init(v interface{}) Pipe {
	s.Base().(*fgbase.Edge).Val = v
	return s
}

// Const sets a value for continual flow
func (s *pipe) Const(v interface{}) Pipe {
	s.Base().(*fgbase.Edge).Const(v)
	return s
}

// Sink sets a pipe to be a sink
func (s *pipe) Sink() Pipe {
	s.Base().(*fgbase.Edge).Sink()
	return s
}

// IsConst returns true if pipe is a constant
func (s *pipe) IsConst() bool {
	return s.Base().(*fgbase.Edge).IsConst()
}

// IsSink returns true if pipe is a sink
func (s *pipe) IsSink() bool {
	return s.Base().(*fgbase.Edge).IsSink()
}

// Same returns true if two pipes are the same underneath
func (s *pipe) Same(s2 Pipe) bool {
	checkInternalPipe(s.fg, s2)

	if s.Base().(*fgbase.Edge) == nil {
		return s2.Base().(*fgbase.Edge) == nil
	}
	if s2.Base().(*fgbase.Edge) == nil {
		return s.Base().(*fgbase.Edge) == nil
	}
	return s.Base().(*fgbase.Edge).Same(s2.Base().(*fgbase.Edge))
}

// Empty returns true if the underlying implementation is nil
func (s *pipe) Empty() bool {
	return s.base == nil
}

// Flowgraph returns associate flowgraph interface
func (s *pipe) Flowgraph() Flowgraph {
	return s.fg
}

// Base returns value of underlying implementation
func (s *pipe) Base() interface{} {
	return s.base
}
