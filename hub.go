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

	// Destination returns destination stream by index
	Destination(n int) Stream

	// FindSource returns source stream by port name
	FindSource(name string) Stream

	// FindDestination returns destination stream by port name
	FindDestination(name string) Stream

	// NumSource returns the number of source ports
	NumSource() int

	// NumDestination returns the number of destination ports
	NumDestination() int

	// SourceNames returns the names of the source ports
	SourceNames() []string

	// DestinationNames returns the names of the destination ports
	DestinationNames() []string

	/* CONFIGURING */

	// AddSource adds a source port for each stream
	AddSource(s ...Stream)

	// AddDestination adds a destination port for each stream
	AddDestination(s ...Stream)

	// SetSourceNames names the source ports
	SetSourceNames(nm ...string)

	// SetDestinationNames names the destination ports
	SetDestinationNames(nm ...string)

	// SetSource sets the stream for a named source port
	SetSource(port interface{}, s Stream) error

	// SetDestination sets the stream for a named destination port
	SetDestination(port interface{}, s Stream) error

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

// Destination returns destination stream by index
func (h hub) Destination(i int) Stream {
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

// FindDestination returns destination stream by port name
func (h hub) FindDestination(name string) Stream {
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

// AddDestination adds a destination port for each stream
func (h hub) AddDestination(s ...Stream) {
	for _, sv := range s {
		h.base.DstAppend(sv.Base().(*fgbase.Edge))
	}
}

// NumSource returns the number of source ports
func (h hub) NumSource() int {
	return h.base.SrcCnt()
}

// NumDestination returns the number of destination ports
func (h hub) NumDestination() int {
	return h.base.DstCnt()
}

// SetSourceNames names the source ports
func (h hub) SetSourceNames(nm ...string) {
	if len(nm) != h.base.SrcCnt() {
		h.base.Panicf("Number of source ports names %v does not match number of source ports (%d)\n", nm, h.base.SrcCnt())
	}
	h.base.SetSrcNames(nm...)
}

// SetDestinationNames names the destination ports
func (h hub) SetDestinationNames(nm ...string) {
	if len(nm) != h.base.DstCnt() {
		h.base.Panicf("Number of destination port names %v does not match number of destination ports (%d)\n", nm, h.base.DstCnt())
	}
	h.base.SetDstNames(nm...)
}

// SourceNames returns the names of the source ports
func (h hub) SourceNames() []string {
	return h.base.SrcNames()
}

// DestinationNames returns the names of the destination ports
func (h hub) DestinationNames() []string {
	return h.base.DstNames()
}

// SetSource sets the stream for a named source port
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
		h.Base().(*fgbase.Node).Panicf("Need string or int to specify port on hub %s\n", h.Name())
	}

	if !ok {
		return fmt.Errorf("source port %s not found on hub %v\n", port, h.Name())
	}

	h.base.SrcSet(i, s.Base().(*fgbase.Edge))
	return nil
}

// SetDestination sets the stream for a named destination port
func (h hub) SetDestination(port interface{}, s Stream) error {
	var i int
	var ok bool
	switch v := port.(type) {
	case string:
		i, ok = h.base.FindDstIndex(v)
	case int:
		ok = v >= 0 && v < h.NumSource()
		i = v
	default:
		h.Base().(*fgbase.Node).Panicf("Need string or int to specify destination port on hub %s\n", h.Name())
	}

	if !ok {
		return fmt.Errorf("destination port %s not found on hub %v\n", port, h.Name())
	}

	h.base.DstSet(i, s.Base().(*fgbase.Edge))
	return nil
}

// Base returns the value that implements this hub
// The type of this value identifies the implementation.
func (h hub) Base() interface{} {
	return h.base
}
