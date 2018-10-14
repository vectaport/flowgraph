package flowgraph

import (
	"github.com/vectaport/fgbase"

	"fmt"
)

// Hub struct for flowgraph hubs that are connected by flowgraph streams
type Hub struct {
	base *fgbase.Node
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (h *Hub) Tracef(format string, v ...interface{}) {
	h.base.Tracef(format, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (h *Hub) LogError(format string, v ...interface{}) {
	h.base.LogError(format, v...)
}

// Panicf for logging of panic messages.  Uses atomic log mechanism.
func (h *Hub) Panicf(format string, v ...interface{}) {
	h.base.Panicf(format, v...)
}

// Name returns the hub name
func (h *Hub) Name() string {
	return h.base.Name
}

// Source returns source stream by index
func (h *Hub) Source(i int) *Stream {
	return &Stream{h.base.Src(i)}
}

// Result returns result stream by index
func (h *Hub) Result(i int) *Stream {
	return &Stream{h.base.Dst(i)}
}

// FindSource returns source stream by port name
func (h *Hub) FindSource(port interface{}) (s *Stream, portok bool) {
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
	return &Stream{e}, ok
}

// FindResult returns result stream by port name
func (h *Hub) FindResult(port interface{}) (s *Stream, portok bool) {
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
	return &Stream{e}, ok
}

// AddSource adds a source port for each stream
func (h *Hub) AddSource(s ...*Stream) {
	for _, sv := range s {
		h.base.SrcAppend(sv.base)
	}
}

// AddResult adds a result port for each stream
func (h *Hub) AddResult(s ...*Stream) {
	for _, sv := range s {
		h.base.DstAppend(sv.base)
	}
}

// NumSource returns the number of source ports
func (h *Hub) NumSource() int {
	return h.base.SrcCnt()
}

// NumResult returns the number of result ports
func (h *Hub) NumResult() int {
	return h.base.DstCnt()
}

// SetSourceNum sets the number of source ports
func (h *Hub) SetSourceNum(n int) {
	h.base.SetSrcNum(n)
}

// SetResultNum sets the number of result ports
func (h *Hub) SetResultNum(n int) {
	h.base.SetDstNum(n)
}

// SetSourceNames names the source ports
func (h *Hub) SetSourceNames(nm ...string) {
	h.base.SetSrcNames(nm...)
}

// SetResultNames names the result ports
func (h *Hub) SetResultNames(nm ...string) {
	h.base.SetDstNames(nm...)
}

// SourceNames returns the names of the source ports
func (h *Hub) SourceNames() []string {
	return h.base.SrcNames()
}

// ResultNames returns the names of the result ports
func (h *Hub) ResultNames() []string {
	return h.base.DstNames()
}

// SetSource sets a stream on a source port selected by string or int
func (h *Hub) SetSource(port interface{}, s *Stream) error {
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
		h.Panicf("Need string or int to select port on hub %s\n", h.Name())
	}

	if !ok {
		return fmt.Errorf("source port %s not found on hub %v\n", port, h.Name())
	}

	e := *s.base

	h.base.SrcSet(i, &e)
	return nil
}

// SetResult sets a stream on a result port selected by string or int
func (h *Hub) SetResult(port interface{}, s *Stream) error {
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
		return fmt.Errorf("result port %s not found on hub %v\n", port, h.Name())
	}

	e := *s.base

	h.base.DstSet(i, &e)
	return nil
}

// SourceIndex returns the index of a named source port, -1 if not found
func (h *Hub) SourceIndex(port string) int {
	i, ok := h.base.FindSrcIndex(port)
	if !ok {
		i = -1
	}
	return i
}

// ResultIndex returns the index of a named result port, -1 if not found
func (h *Hub) ResultIndex(port string) int {
	i, ok := h.base.FindDstIndex(port)
	if !ok {
		i = -1
	}
	return i
}

// Empty returns true if the underlying implementation is nil
func (h *Hub) Empty() bool {
	return h.base == nil
}

// Node returns pointer to underlying Node
func (h *Hub) Node() *fgbase.Node {
	return h.base
}
