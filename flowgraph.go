// Package flowgraph for scalable asynchronous system development
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"fmt"
	"github.com/vectaport/fgbase"
)

/*=====================================================================*/

// Getter receives an empty interface with the Get method, Pipe is for tracing
type Getter interface {
	Get(p Pipe) (interface{}, error)
}

// Putter transmits an empty interface with the Put method, Pipe is for tracing
type Putter interface {
	Put(p Pipe, v interface{}) error
}

// Transformer transforms a variadic list of empty interfaces into an array
// of empty interfaces, Pipe is for tracing 
type Transformer interface {
	Transform(p Pipe, c ...interface{}) ([]interface{}, error)
}

/*=====================================================================*/

// Flowgraph interface
type Flowgraph interface {

	// Name returns the name of a Flowgraph
	Name() string

	// Pipe returns the nth pipe in a Flowgraph
	Pipe(i int) Pipe

	// Connector returns the nth connector in a Flowgraph
	Connector(i int) Connector

	// NumPipe returns the number of pipes in this graph
	NumPipe() int
	// NumConnector returns the number of pipes in this graph
	NumConnector() int

	// FindPipe finds a Pipe by name
	FindPipe(name string) Pipe
	// FindConnector finds a Connector by name
	FindConnector(name string) Connector

	// InsertIncoming adds a single input source to a flowgraph that uses a Getter
	InsertIncoming(name string, getter Getter)
	// InsertOutgoing adds a single output source to a flowgraph that uses a Putter
	InsertOutgoing(name string, putter Putter)

	// InsertConst adds an input constant as an incoming source.
	InsertConst(name string, v interface{})
	// InsertArray adds an array as an incoming source.
	InsertArray(name string, arr []interface{})
	// InsertSink adds a output sink on the latest edge
	InsertSink(name string)

	// InsertAllOf adds a transform that waits for all inputs before producing outputs
	InsertAllOf(name string, transformer Transformer)

	// Run runs the flowgraph
	Run()
}

// implementation of Flowgraph
type graph struct {
	name  string
	nodes []fgbase.Node
	edges []fgbase.Edge
}

// New returns a named flowgraph
func New(nm string) Flowgraph {
	return &graph{nm, nil, nil}
}

// Name returns the name of a Flowgraph
func (fg *graph) Name() string {
	return fg.Name()
}

// Pipe returns the nth pipe in a Flowgraph
func (fg *graph) Pipe(n int) Pipe {
	return pipe{&fg.nodes[n]}
}

// Connector returns the nth connector in a Flowgraph
func (fg *graph) Connector(n int) Connector {
	return conn{&fg.edges[n]}
}

// NumPipe returns the number of pipes in this graph
func (fg *graph) NumPipe() int {
	return len(fg.nodes)
}

// NumConnector returns the number of pipes in this graph
func (fg *graph) NumConnector() int {
	return len(fg.edges)
}

// FindPipe finds a Pipe by name
func (fg *graph) FindPipe(name string) Pipe {
	// simple search for now
	for i, v := range fg.nodes {
		if fg.nodes[i].Name == name {
			return pipe{&v}
		}
	}
	return nil
}

// FindConnector finds a Connector by name
func (fg *graph) FindConnector(name string) Connector {
	// simple search for now
	for i, v := range fg.edges {
		if fg.edges[i].Name == name {
			return conn{&v}
		}
	}
	return nil
}

// InsertIncoming adds a single input source to a flowgraph that uses a Getter
func (fg *graph) InsertIncoming(name string, getter Getter) {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := funcIncoming(e, getter)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertOutgoing adds a single output source to a flowgraph that uses a Putter
func (fg *graph) InsertOutgoing(name string, putter Putter) {
	node := funcOutgoing(fg.edges[len(fg.edges)-1], putter)
	fg.nodes = append(fg.nodes, node)
	fg.nodes[len(fg.nodes)-1].Owner = pipe{&node}
}

// InsertConst adds an input constant as an incoming source.
func (fg *graph) InsertConst(name string, v interface{}) {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := fgbase.FuncConst(e, v)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertArray adds an array as an incoming source.
func (fg *graph) InsertArray(name string, arr []interface{}) {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.edges)), nil)
	fg.edges = append(fg.edges, e)
	node := fgbase.FuncArray(e, arr)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertSink adds a output sink on the latest edge
func (fg *graph) InsertSink(name string) {
	i := len(fg.edges) - 1
	node := fgbase.FuncSink(fg.edges[i])
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// InsertAllOf adds a transform that waits for all inputs before producing outputs
func (fg *graph) InsertAllOf(name string, transformer Transformer) {
	node := funcAllOf([]fgbase.Edge{fg.edges[0]}, []fgbase.Edge{fg.edges[len(fg.edges)-1]},
		name, transformer)
	fg.nodes = append(fg.nodes, node)
	node.Owner = pipe{&node}
}

// Run runs the flowgraph
func (fg *graph) Run() {
	fgbase.RunAll(fg.nodes)
}
