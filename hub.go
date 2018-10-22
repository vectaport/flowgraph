package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// Hub interface for flowgraph hubs that are connected by flowgraph streams
type Hub interface {

	// Name returns the hub name
	Name() string

	// Tracef for debug trace printing.  Uses atomic log mechanism.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.  Uses atomic log mechanism.
	LogError(format string, v ...interface{})

	// Panicf for logging of panic messages.  Uses atomic log mechanism.
	Panicf(format string, v ...interface{})

	// Source returns source stream by index
	Source(i int) Stream

	// Result returns result stream by index
	Result(i int) Stream

	// SetSource sets a stream on a source port selected by string or int
	SetSource(port interface{}, s Stream) Hub

	// SetResult sets a stream on a result port selected by string or int
	SetResult(port interface{}, s Stream) Hub

	// FindSource returns source stream by port name
	FindSource(port interface{}) (s Stream, portok bool)

	// FindResult returns result stream by port name
	FindResult(port interface{}) (s Stream, portok bool)

	// AddSources adds a source port for each stream
	AddSources(s ...Stream) Hub

	// AddResults adds a result port for each stream
	AddResults(s ...Stream) Hub

	// NumSource returns the number of source ports
	NumSource() int

	// NumResult returns the number of result ports
	NumResult() int

	// SetNumSource sets the number of source ports
	SetNumSource(n int) Hub

	// SetNumResult sets the number of result ports
	SetNumResult(n int) Hub

	// SourceNames returns the names of the source ports
	SourceNames() []string

	// ResultNames returns the names of the result ports
	ResultNames() []string

	// SetSourceNames names the source ports
	SetSourceNames(nm ...string) Hub

	// SetResultNames names the result ports
	SetResultNames(nm ...string) Hub

	// SourceIndex returns the index of a named source port, -1 if not found
	SourceIndex(port string) int

	// ResultIndex returns the index of a named result port, -1 if not found
	ResultIndex(port string) int

	// ConnectSources connects a list of source Streams to this Hub
	ConnectSources(source ...Stream) Hub

	// ConnectResults connects a list of result Streams to this Hub
	ConnectResults(result ...Stream) Hub

	// Flowgraph returns associated flowgraph
	Flowgraph() Flowgraph

	// Empty returns true if the underlying implementation is nil
	Empty() bool

	// Base returns value of underlying implementation
	Base() interface{}
}

// Hub implementation
type hub struct {
	base *fgbase.Node
	fg   *flowgraph
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (h *hub) Tracef(format string, v ...interface{}) {
	h.base.Tracef(format, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (h *hub) LogError(format string, v ...interface{}) {
	h.base.LogError(format, v...)
}

// Panicf for logging of panic messages.  Uses atomic log mechanism.
func (h *hub) Panicf(format string, v ...interface{}) {
	h.base.Panicf(format, v...)
}

// Name returns the hub name
func (h *hub) Name() string {
	return h.base.Name
}

// Source returns source stream by index
func (h *hub) Source(i int) Stream {
	return &stream{h.base.Src(i), h.fg}
}

// Result returns result stream by index
func (h *hub) Result(i int) Stream {
	return &stream{h.base.Dst(i), h.fg}
}

// SetSource sets a stream on a source port selected by string or int
func (h *hub) SetSource(port interface{}, s Stream) Hub {
	checkfgStream(h.Flowgraph(), s)
	
	var i int
	var ok bool
	switch v := port.(type) {
	case string:
		i = h.SourceIndex(v)
		ok = i >= 0
	case int:
		ok = v >= 0 && v < h.NumSource()
		i = v
	default:
		h.Panicf("Need string or int to select port on Hub %s\n", h.Name())
	}

	if !ok {
		h.Panicf("Source port %v not found on Hub %v\n", port, h.Name())
	}

	if s != nil && s.Base() != nil {
		e := *s.Base().(*fgbase.Edge)
		h.base.SrcSet(i, &e)
	}
	return h

}

// SetResult sets a stream on a result port selected by string or int
func (h *hub) SetResult(port interface{}, s Stream) Hub {
	checkfgStream(h.Flowgraph(), s)
	
	var i int
	var ok bool
	switch v := port.(type) {
	case string:
		i = h.ResultIndex(v)
		ok = i >= 0
	case int:
		ok = v >= 0 && v < h.NumResult()
		i = v
	default:
		h.Panicf("Need string or int to select result port on hub %s\n", h.Name())
	}

	if !ok {
		h.Panicf("Result port %v not found on Hub %v\n", port, h.Name())
	}

	e := *s.Base().(*fgbase.Edge)

	h.base.DstSet(i, &e)
	return h
}

// FindSource returns source stream by port name
func (h *hub) FindSource(port interface{}) (s Stream, portok bool) {
	var e *fgbase.Edge
	var ok bool
	switch v := port.(type) {
	case string:
		e, ok = h.base.FindSrc(v)
	case int:
		ok = v >= 0 && v < h.NumSource()
		if ok {
			e = h.base.Src(v)
		}
	default:
		h.Panicf("Need string or int to select source port on hub %s\n", h.Name())
	}
	return &stream{e, h.fg}, ok
}

// FindResult returns result stream by port name
func (h *hub) FindResult(port interface{}) (s Stream, portok bool) {
	var e *fgbase.Edge
	var ok bool
	switch v := port.(type) {
	case string:
		e, ok = h.base.FindDst(v)
	case int:
		ok = v >= 0 && v < h.NumSource()
		if ok {
			e = h.base.Dsts[v]
		}
	default:
		h.Panicf("Need string or int to select result port on hub %s\n", h.Name())
	}
	return &stream{e, h.fg}, ok
}

// AddSources adds a source port for each stream
func (h *hub) AddSources(s ...Stream) Hub {
	for _, sv := range s {
		checkfgStream(h.Flowgraph(), sv)
		h.Base().(*fgbase.Node).SrcAppend(sv.Base().(*fgbase.Edge))
	}
	return h
}

// AddResults adds a result port for each stream
func (h *hub) AddResults(s ...Stream) Hub {
	for _, sv := range s {
		checkfgStream(h.Flowgraph(), sv)
		h.Base().(*fgbase.Node).DstAppend(sv.Base().(*fgbase.Edge))
	}
	return h
}

// NumSource returns the number of source ports
func (h *hub) NumSource() int {
	return h.base.SrcCnt()
}

// NumResult returns the number of result ports
func (h *hub) NumResult() int {
	return h.base.DstCnt()
}

// SetNumSource sets the number of source ports
func (h *hub) SetNumSource(n int) Hub {
	h.base.SetSrcNum(n)
	return h
}

// SourceNames returns the names of the source ports
func (h *hub) SourceNames() []string {
	return h.base.SrcNames()
}

// ResultNames returns the names of the result ports
func (h *hub) ResultNames() []string {
	return h.base.DstNames()
}


// SetNumResult sets the number of result ports
func (h *hub) SetNumResult(n int) Hub {
	h.base.SetDstNum(n)
	return h
}

// SetSourceNames names the source ports
func (h *hub) SetSourceNames(nm ...string) Hub {
	h.base.SetSrcNames(nm...)
	return h
}

// SetResultNames names the result ports
func (h *hub) SetResultNames(nm ...string) Hub {
	h.base.SetDstNames(nm...)
	return h
}

// SourceIndex returns the index of a named source port, -1 if not found
func (h *hub) SourceIndex(port string) int {
	i, ok := h.base.FindSrcIndex(port)
	if !ok {
		i = -1
	}
	return i
}

// ResultIndex returns the index of a named result port, -1 if not found
func (h *hub) ResultIndex(port string) int {
	i, ok := h.base.FindDstIndex(port)
	if !ok {
		i = -1
	}
	return i
}

// ConnectSources connects a list of source Streams to this Hub
func (h *hub) ConnectSources(source ...Stream) Hub {
	h.SetNumSource(len(source))
	for i, v := range source {
		h.SetSource(i, v)
	}
	return h
}

// ConnectResults connects a list of result Streams to this Hub
func (h *hub) ConnectResults(result ...Stream) Hub {
	h.SetNumResult(len(result))
	for i, v := range result {
		h.SetResult(i, v)
	}
	return h
}

// Flowgraph returns associate flowgraph interface
func (h *hub) Flowgraph() Flowgraph {
	return h.fg
}

// Empty returns true if the underlying implementation is nil
func (h *hub) Empty() bool {
	return h.base == nil
}

// Base returns value of underlying implementation
func (h *hub) Base() interface{} {
	return h.base
}
