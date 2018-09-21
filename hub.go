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
	SetSource(port string, s Stream) error

	// SetDestination sets the stream for a named destination port
	SetDestination(port string, s Stream) error

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
	h.base.Tracef(format, v)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (h hub) LogError(format string, v ...interface{}) {
	h.base.LogError(format, v)
}

// Name returns the hub name
func (h hub) Name() string {
	return h.base.Name
}

// Source returns source stream by index
func (h hub) Source(i int) Stream {
	return stream{h.base.Srcs[i]}
}

// Destination returns destination stream by index
func (h hub) Destination(i int) Stream {
	return stream{h.base.Dsts[i]}
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
		h.base.Srcs = append(h.base.Srcs, sv.Base().(*fgbase.Edge))
	}
}

// AddDestination adds a destination port for each stream
func (h hub) AddDestination(s ...Stream) {
	for _, sv := range s {
		h.base.Dsts = append(h.base.Dsts, sv.Base().(*fgbase.Edge))
	}
}

// NumSource returns the number of source ports
func (h hub) NumSource() int {
	return len(h.base.Srcs)
}

// NumDestination returns the number of destination ports
func (h hub) NumDestination() int {
	return len(h.base.Dsts)
}

// SetSourceNames names the source ports
func (h hub) SetSourceNames(nm ...string) {
	h.base.SetSrcNames(nm...)
}

// SetDestinationNames names the destination ports
func (h hub) SetDestinationNames(nm ...string) {
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
func (h hub) SetSource(port string, s Stream) error {
	i, ok := h.base.FindSrcIndex(port)
	if !ok {
		return fmt.Errorf("source port %s not found on hub %s\n", port, h.Name())
	}
	h.base.Srcs[i] = s.Base().(*fgbase.Edge)
	return nil
}

// SetDestination sets the stream for a named destination port
func (h hub) SetDestination(port string, s Stream) error {
	i, ok := h.base.FindDstIndex(port)
	if !ok {
		return fmt.Errorf("destination port %s not found on hub %s\n", port, h.Name())
	}
	h.base.Dsts[i] = s.Base().(*fgbase.Edge)
	return nil
}

// Base returns the value that implements this hub
// The type of this value identifies the implementation.
func (h hub) Base() interface{} {
	return h.base
}
