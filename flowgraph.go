// Package flowgraph for scalable asynchronous system development
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"fmt"
	"github.com/vectaport/fgbase"
)

/*=====================================================================*/

// Getter receives a value with the Get method. Use Node.Tracef for tracing.
type Getter interface {
	Get(n Node) (interface{}, error)
}

// Putter transmits a value with the Put method. Use Node.Tracef
// for tracing
type Putter interface {
	Put(n Node, v interface{}) error
}

// Transformer transforms a variadic list of values into a slice
// of values with the Transform method. Use Node.Tracef for tracing.
type Transformer interface {
	Transform(n Node, c ...interface{}) ([]interface{}, error)
}

/*=====================================================================*/

// Flowgraph interface for flowgraphs assembled out of node nodes and connector edges
type Flowgraph interface {

	// Name returns the name of this flowgraph
	Name() string

	// Node returns a node by index
	Node(i int) Node
	// Edge returns a connector by index
	Edge(i int) Edge

	// NumNode returns the number of nodes
	NumNode() int
	// NumEdge returns the number of nodes
	NumEdge() int

	// FindNode finds a node by name
	FindNode(name string) Node
	// FindEdge finds a connector by name
	FindEdge(name string) Edge

	// NewNode returns a new uninitialized node
	NewNode(nm string) Node
	// NewEdge returns a new uninitialized connector
	NewEdge(nm string) Edge

	// InsertNode adds a Node to the flowgraph, connecting inputs to existing
	// dangling edges as available and creating dangling output edges as needed.
	InsertNode(n Node)
	
	// InsertIncoming adds an input source that uses a Getter
	InsertIncoming(name string, getter Getter) Node
	// InsertOutgoing adds an output destination that uses a Putter
	InsertOutgoing(name string, putter Putter) Node

	// InsertConst adds an input constant as an incoming source.
	InsertConst(name string, v interface{}) Node
	// InsertArray adds an array as an incoming source.
	InsertArray(name string, arr []interface{}) Node
	// InsertSink adds an output sink
	InsertSink(name string) Node

	// InsertAllOf adds a transform that waits for all inputs before producing outputs
	InsertAllOf(name string, transformer Transformer) Node

	// Run runs the flowgraph
	Run()
}

// implementation of Flowgraph
type graph struct {
	name  string
	nodes []fgbase.Node
	edges []fgbase.Edge
}

// New returns a named flowgraph implemented with the fgbase package
func New(nm string) Flowgraph {
	return &graph{nm, nil, nil}
}

// Name returns the name of this flowgraph
func (fg *graph) Name() string {
	return fg.Name()
}

// Node returns a node by index
func (fg *graph) Node(n int) Node {
	return node{&fg.nodes[n]}
}

// Edge returns a connector by index
func (fg *graph) Edge(n int) Edge {
	return edge{&fg.edges[n]}
}

// NumNode returns the number of nodes
func (fg *graph) NumNode() int {
	return len(fg.nodes)
}

// NumEdge returns the number of nodes
func (fg *graph) NumEdge() int {
	return len(fg.edges)
}

// NewNode returns a new uninitialized node
func (fg *graph) NewNode(nm string) Node {
	n := fgbase.MakeNode(nm, nil, nil, nil, nil)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// NewEdge returns a new uninitialized edge
func (fg *graph) NewEdge(nm string) Edge {
	if nm == "" {
		nm = fmt.Sprintf("e%d", len(fg.edges))
	}
	e := fgbase.MakeEdge(nm, nil)
	fg.edges = append(fg.edges, e)
	return edge{&fg.edges[len(fg.nodes)-1]}
}

// FindNode finds a node by name
func (fg *graph) FindNode(name string) Node {
	// simple search for now
	for i, v := range fg.nodes {
		if fg.nodes[i].Name == name {
			return node{&v}
		}
	}
	return nil
}

// FindEdge finds a Edge by name
func (fg *graph) FindEdge(name string) Edge {
	// simple search for now
	for i, v := range fg.edges {
		if fg.edges[i].Name == name {
			return edge{&v}
		}
	}
	return nil
}

// InsertNode adds a Node to the flowgraph, connecting inputs to existing
// dangling edges as available and creating dangling output edges as needed.
InsertNode(n Node) {
}
2
// InsertIncoming adds an input source that uses a Getter
func (fg *graph) InsertIncoming(name string, getter Getter) Node {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	n := funcIncoming(e, getter)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// InsertOutgoing adds a destination that uses a Putter
func (fg *graph) InsertOutgoing(name string, putter Putter) Node {
	n := funcOutgoing(fg.edges[len(fg.edges)-1], putter)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// InsertConst adds an input constant as an incoming source.
func (fg *graph) InsertConst(name string, v interface{}) Node {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	n := fgbase.FuncConst(e, v)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// InsertArray adds an array as an incoming source.
func (fg *graph) InsertArray(name string, arr []interface{}) Node {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	n := fgbase.FuncArray(e, arr)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// InsertSink adds a output sink on the latest edge
func (fg *graph) InsertSink(name string) Node {
	i := len(fg.edges) - 1
	n := fgbase.FuncSink(fg.edges[i])
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// InsertAllOf adds a transform that waits for all inputs before producing outputs
func (fg *graph) InsertAllOf(name string, transformer Transformer) Node {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	n := funcAllOf([]fgbase.Edge{fg.edges[len(fg.edges)-2]}, []fgbase.Edge{fg.edges[len(fg.edges)-1]},
		name, transformer)
	fg.nodes = append(fg.nodes, n)
	return node{&fg.nodes[len(fg.nodes)-1]}
}

// Run runs the flowgraph
func (fg *graph) Run() {
	fgbase.RunAll(fg.nodes)
}
