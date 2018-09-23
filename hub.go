package flowgraph

import (
	"github.com/vectaport/fgbase"

	"fmt"
)

// Hub interface for flowgraph hubs that are connected by flowgraph streams
type Hub interface {

	/* TRACING */

	// Tracef for debug trace printing.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.
	LogError(format string, v ...interface{})

	/* PROBING */

	// Name returns the hub name
	Name() string

	// Source returns source stream by index
	Source(n int) Stream

	// Result returns result stream by index
	Result(n int) Stream

	// FindSource returns source stream by port name
	FindSource(name string) Stream

	// FindResult returns result stream by port name
	FindResult(name string) Stream

	// NumSource returns the number of source ports
	NumSource() int

	// NumResult returns the number of result ports
	NumResult() int

	// SourceNames returns the names of the source ports
	SourceNames() []string

	// ResultNames returns the names of the result ports
	ResultNames() []string

	/* CONFIGURING */

	// AddSource adds a source port for each stream
	AddSource(s ...Stream)

	// AddResult adds a result port for each stream
	AddResult(s ...Stream)

	// SetSourceNames names the source ports
	SetSourceNames(nm ...string)

	// SetResultNames names the result ports
	SetResultNames(nm ...string)

	// SetSource sets a stream on a source port selected by string or int
	SetSource(port interface{}, s Stream) error

	// SetResult sets a stream on a result port selected by string or int
	SetResult(port interface{}, s Stream) error

	/* IMPLEMENTATION */

	// Base returns the value that implements this hub
	// The type of this value identifies the implementation.
	Base() interface{}
}

// implementation of Hub
type hub struct {
	base *fgbase.Node
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (h hub) Tracef(format string, v ...interface{}) {
	h.base.Tracef(format, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (h hub) LogError(format string, v ...interface{}) {
	h.base.LogError(format, v...)
}

// Name returns the hub name
func (h hub) Name() string {
	return h.base.Name
}

// Source returns source stream by index
func (h hub) Source(i int) Stream {
	return stream{h.base.Src(i)}
}

// Result returns result stream by index
func (h hub) Result(i int) Stream {
	return stream{h.base.Dst(i)}
}

// FindSource returns source stream by port name
func (h hub) FindSource(name string) Stream {
	e, ok := h.base.FindSrc(name)
	if !ok {
		return nil
	}
	if e == nil {
		return stream{nil}
	}
	return stream{e}
}

// FindResult returns result stream by port name
func (h hub) FindResult(name string) Stream {
	e, ok := h.base.FindDst(name)
	if !ok {
		return nil
	}
	if e == nil {
		return stream{nil}
	}
	return stream{e}
}

// AddSource adds a source port for each stream
func (h hub) AddSource(s ...Stream) {
	for _, sv := range s {
		h.base.SrcAppend(sv.Base().(*fgbase.Edge))
	}
}

// AddResult adds a result port for each stream
func (h hub) AddResult(s ...Stream) {
	for _, sv := range s {
		h.base.DstAppend(sv.Base().(*fgbase.Edge))
	}
}

// NumSource returns the number of source ports
func (h hub) NumSource() int {
	return h.base.SrcCnt()
}

// NumResult returns the number of result ports
func (h hub) NumResult() int {
	return h.base.DstCnt()
}

// SetSourceNames names the source ports
func (h hub) SetSourceNames(nm ...string) {
	if len(nm) != h.base.SrcCnt() {
		h.base.Panicf("Number of source ports names %v does not match number of source ports (%d)\n", nm, h.base.SrcCnt())
	}
	h.base.SetSrcNames(nm...)
}

// SetResultNames names the result ports
func (h hub) SetResultNames(nm ...string) {
	if len(nm) != h.base.DstCnt() {
		h.base.Panicf("Number of result port names %v does not match number of result ports (%d)\n", nm, h.base.DstCnt())
	}
	h.base.SetDstNames(nm...)
}

// SourceNames returns the names of the source ports
func (h hub) SourceNames() []string {
	return h.base.SrcNames()
}

// ResultNames returns the names of the result ports
func (h hub) ResultNames() []string {
	return h.base.DstNames()
}

// SetSource sets a stream on a source port selected by string or int
func (h hub) SetSource(port interface{}, s Stream) error {
	var i int
	var ok bool
	switch v := port.(type) {
	case string:
		i, ok = h.base.FindSrcIndex(v)
	case int:
		ok = v >= 0 && v < h.NumSource()
		i = v
	default:
		h.Base().(*fgbase.Node).Panicf("Need string or int to select port on hub %s\n", h.Name())
	}

	if !ok {
		return fmt.Errorf("source port %s not found on hub %v\n", port, h.Name())
	}

	h.base.SrcSet(i, s.Base().(*fgbase.Edge))
	return nil
}

// SetResult sets a stream on a result port selected by string or int
func (h hub) SetResult(port interface{}, s Stream) error {
	var i int
	var ok bool
	switch v := port.(type) {
	case string:
		i, ok = h.base.FindDstIndex(v)
	case int:
		ok = v >= 0 && v < h.NumSource()
		i = v
	default:
		h.Base().(*fgbase.Node).Panicf("Need string or int to select result port on hub %s\n", h.Name())
	}

	if !ok {
		return fmt.Errorf("result port %s not found on hub %v\n", port, h.Name())
	}

	h.base.DstSet(i, s.Base().(*fgbase.Edge))
	return nil
}

// Base returns the value that implements this hub
// The type of this value identifies the implementation.
func (h hub) Base() interface{} {
	return h.base
}
