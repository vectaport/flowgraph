package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// Node interface for flowgraph nodes that are connected by flowgraph edges
type Node interface {
	// Tracef for debug trace printing.  Use atomic log mechanism.
	Tracef(format string, v ...interface{})

	// LogError for logging of error messages.  Use atomic log mechanism.
	LogError(format string, v ...interface{})

	// Name returns the node name
	Name() string

	// Source returns upstream edge by index
	Source(n int) Edge

	// Destination returns downstream edge by index
	Destination(n int) Edge

	// FindSource returns upstream edge by name
	FindSource(name string) Edge

	// FindDestination returns downstream edge by name
	FindDestination(name string) Edge

	// AddSource adds a list of source edges
	AddSource(e ...Edge)

	// AddDestination adds a list of destination edges
	AddDestination(e ...Edge)

	// NumSource returns the number of upstream edges
	NumSource() int

	// NumDestination returns the number of downstream edges
	NumDestination() int

	// SetSourceNames names the sources
	SetSourceNames(nm ...string)

	// SetDestinationNames names the destinations
	SetDestinationNames(nm ...string)

	// SourceNames returns the names of the sources
	SourceNames() []string

	// DestinationNames returns the names of the destinations
	DestinationNames() []string

	// Base returns the value that implements this node
	// The type of this value identifies the implementation.
	Base() interface{}
}

// implementation of Node
type node struct {
	base *fgbase.Node
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n node) Tracef(format string, v ...interface{}) {
	n.base.Tracef(format, v)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (n node) LogError(format string, v ...interface{}) {
	n.base.LogError(format, v)
}

// Name returns the node name
func (n node) Name() string {
	return n.base.Name
}

// Source returns upstream edge by index
func (n node) Source(i int) Edge {
	return edge{n.base.Srcs[i]}
}

// Destination returns downstream edge by index
func (n node) Destination(i int) Edge {
	return edge{n.base.Dsts[i]}
}

// FindSource returns upstream edge by name
func (n node) FindSource(name string) Edge {
	return edge{n.base.FindSrc(name)}
}

// FindDestination returns downstream edge by name
func (n node) FindDestination(name string) Edge {
	return edge{n.base.FindDst(name)}
}

// AddSource adds a list of source edges
func (n node) AddSource(e ...Edge) {
	for _, ev := range e {
		n.base.Srcs = append(n.base.Srcs, ev.Base().(*fgbase.Edge))
	}
}

// AddDestination adds a list of destination edges
func (n node) AddDestination(e ...Edge) {
	for _, ev := range e {
		n.base.Dsts = append(n.base.Dsts, ev.Base().(*fgbase.Edge))
	}
}

// NumSource returns the number of upstream edges
func (n node) NumSource() int {
	return len(n.base.Srcs)
}

// NumDestination returns the number of downstream edges
func (n node) NumDestination() int {
	return len(n.base.Dsts)
}

// SetSourceNames names the sources
func (n node) SetSourceNames(nm ...string) {
	n.base.SetSrcNames(nm...)
}

// SetDestinationNames names the destination
func (n node) SetDestinationNames(nm ...string) {
	n.base.SetDstNames(nm...)
}

// SourceNames returns the names of the sources
func (n node) SourceNames() []string {
	return n.base.SrcNames()
}

// DestinationNames returns the names of the destinatiopns
func (n node) DestinationNames() []string {
	return n.base.DstNames()
}

// Base returns the value that implements this node
// The type of this value identifies the implementation.
func (n node) Base() interface{} {
	return n.base
}
