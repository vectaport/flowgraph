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

	// NumSource returns the number of upstream edges
	NumSource() int

	// NumDestination returns the number of upstream edges
	NumDestination() int

	// Auxiliary returns auxiliary storage used by
	// underlying implementation for storing state
	// (only required for external debug of fgbase)
	Auxiliary() interface{}
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

// NumSource returns the number of upstream edges
func (n node) NumSource() int {
	return len(n.base.Srcs)
}

// NumDestination returns the number of downstream edges
func (n node) NumDestination() int {
	return len(n.base.Dsts)
}

// Auxiliary returns auxiliary storage for this node used by
// the underlying implementation for storing state
func (n node) Auxiliary() interface{} {
	return n.base.Aux
}
