package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// Hub interface for flowgraph hubs that are connected by flowgraph streams
type Hub interface {
	// Tracef for debug trace printing.  Use atomic log mechanism.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.  Use atomic log mechanism.
	LogError(format string, v ...interface{})

	// Name returns the hub name
	Name() string

	// Source returns upstream stream by index
	Source(n int) Stream

	// Destination returns downstream stream by index
	Destination(n int) Stream

	// FindSource returns upstream stream by port name
	FindSource(name string) Stream

	// FindDestination returns downstream stream by port name
	FindDestination(name string) Stream

	// AddSource adds a source port for each stream
	AddSource(e ...Stream)

	// AddDestination adds a destination port for each stream
	AddDestination(e ...Stream)

	// NumSource returns the number of upstream ports
	NumSource() int

	// NumDestination returns the number of downstream ports
	NumDestination() int

	// SetSourceNames names the source ports
	SetSourceNames(nm ...string)

	// SetDestinationNames names the destination ports
	SetDestinationNames(nm ...string)

	// SourceNames returns the names of the source ports
	SourceNames() []string

	// DestinationNames returns the names of the destination ports
	DestinationNames() []string

	// Base returns the value that implements this hub
	// The type of this value identifies the implementation.
	Base() interface{}
}

// implementation of Hub
type hub struct {
	base *fgbase.Node
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n hub) Tracef(format string, v ...interface{}) {
	n.base.Tracef(format, v)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (n hub) LogError(format string, v ...interface{}) {
	n.base.LogError(format, v)
}

// Name returns the hub name
func (n hub) Name() string {
	return n.base.Name
}

// Source returns upstream stream by index
func (n hub) Source(i int) Stream {
	return stream{n.base.Srcs[i]}
}

// Destination returns downstream stream by index
func (n hub) Destination(i int) Stream {
	return stream{n.base.Dsts[i]}
}

// FindSource returns upstream stream by port name
func (n hub) FindSource(name string) Stream {
	e, ok := n.base.FindSrc(name)
	if !ok {
		return nil
	}
	if e == nil {
		return stream{nil}
	}
	return stream{e}
}

// FindDestination returns downstream stream by port name
func (n hub) FindDestination(name string) Stream {
	e, ok := n.base.FindDst(name)
	if !ok {
		return nil
	}
	if e == nil {
		return stream{nil}
	}
	return stream{e}
}

// AddSource adds a source port for each stream
func (n hub) AddSource(e ...Stream) {
	for _, ev := range e {
		n.base.Srcs = append(n.base.Srcs, ev.Base().(*fgbase.Edge))
	}
}

// AddDestination adds a destination port for each stream
func (n hub) AddDestination(e ...Stream) {
	for _, ev := range e {
		n.base.Dsts = append(n.base.Dsts, ev.Base().(*fgbase.Edge))
	}
}

// NumSource returns the number of upstream ports
func (n hub) NumSource() int {
	return len(n.base.Srcs)
}

// NumDestination returns the number of downstream ports
func (n hub) NumDestination() int {
	return len(n.base.Dsts)
}

// SetSourceNames names the source ports
func (n hub) SetSourceNames(nm ...string) {
	n.base.SetSrcNames(nm...)
}

// SetDestinationNames names the destination ports
func (n hub) SetDestinationNames(nm ...string) {
	n.base.SetDstNames(nm...)
}

// SourceNames returns the names of the source ports
func (n hub) SourceNames() []string {
	return n.base.SrcNames()
}

// DestinationNames returns the names of the destination ports
func (n hub) DestinationNames() []string {
	return n.base.DstNames()
}

// Base returns the value that implements this hub
// The type of this value identifies the implementation.
func (n hub) Base() interface{} {
	return n.base
}
