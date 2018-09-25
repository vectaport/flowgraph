// Package flowgraph for scalable asynchronous system development
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"github.com/vectaport/fgbase"

	"errors"
	"fmt"
	"log"
)

/*=====================================================================*/

// End of flow
var EOF = errors.New("EOF")

// Hub code
type Code int

const (
	AllOf Code = iota	// pick this doc up
	OneOf
	Steer
	Select
	Add
	Sub
	Mul
	Div
	Const
	Array
	Sink
)

/*=====================================================================*/

// Transformer transforms a slice of source values into a slice
// of result values with the Transform method. Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h *Hub, source []interface{}) (result []interface{}, err error)
}

/*=====================================================================*/

// Flowgraph struct for flowgraphs assembled out of hubs and streams
type Flowgraph struct {
	name         string
	hubs         []*Hub
	streams      []*Stream
	nameToHub    map[string]*Hub
	nameToStream map[string]*Stream
}

// New returns a named flowgraph
func New(name string) *Flowgraph {
	nameToHub := make(map[string]*Hub)
	nameToStream := make(map[string]*Stream)
	fg := Flowgraph{name, nil, nil, nameToHub, nameToStream}
	return &fg
}

// Name returns the name of this flowgraph
func (fg *Flowgraph) Name() string {
	return fg.Name()
}

// Hub returns a hub by index
func (fg *Flowgraph) Hub(n int) *Hub {
	return fg.hubs[n]
}

// Stream returns a stream by index
func (fg *Flowgraph) Stream(n int) *Stream {
	return fg.streams[n]
}

// NumHub returns the number of hubs
func (fg *Flowgraph) NumHub() int {
	return len(fg.hubs)
}

// NumStream returns the number of hubs
func (fg *Flowgraph) NumStream() int {
	return len(fg.streams)
}

// NewHub returns a new unconnected hub
func (fg *Flowgraph) NewHub(name string, code Code, init interface{}) *Hub {

	var n fgbase.Node

	switch code {

	// User Supplied Hubs
	case AllOf:
		n = fgbase.MakeNode(name, nil, nil, nil, allOfFire)

	// Math Hubs
	case Add:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.AddFire)
	case Sub:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.SubFire)
	case Mul:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.MulFire)
	case Div:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.DivFire)

	// General Purpose Hubs
	case Array:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ArrayFire)
	case Const:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ConstFire)
	case Sink:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, nil, nil, fgbase.SinkFire)
		n.Aux = fgbase.SinkStats{0, 0}
	default:
		log.Panicf("Unexpected Hub code:  %s\n", code)
	}
	if n.Aux == nil {
		n.Aux = init
	}

	h := Hub{&n}
	fg.hubs = append(fg.hubs, &h)
	fg.nameToHub[name] = &h
	return &h
}

// NewStream returns a new unconnected stream
func (fg *Flowgraph) NewStream(name string) *Stream {
	if name == "" {
		name = fmt.Sprintf("e%d", len(fg.streams))
	}
	e := fgbase.MakeEdge(name, nil)
	s := Stream{&e}
	fg.streams = append(fg.streams, &s)
	fg.nameToStream[name] = &s
	return &s
}

// FindHub finds a hub by name
func (fg *Flowgraph) FindHub(name string) *Hub {
	return fg.nameToHub[name]
}

// FindStream finds a Stream by name
func (fg *Flowgraph) FindStream(name string) *Stream {
	return fg.nameToStream[name]
}

// Connect connects two hubs via named (string) or indexed (int) ports
func (fg *Flowgraph) Connect(
	upstream *Hub, upstreamPort interface{},
	dnstream *Hub, dnstreamPort interface{}) *Stream {

	var us *Stream
	var usok bool
	switch v := upstreamPort.(type) {
	case string:
		us, usok = upstream.FindResult(v)
		if !usok {
			upstream.Panicf("No result port \"%s\" found on Hub \"%s\"\n", v, upstream.Name())
		}
	case int:
		usok = v >= 0 && v < upstream.NumResult()
		if !usok {
			upstream.Panicf("No result port %d found on Hub \"%s\"\n", v, dnstream.Name())
		}
		us = upstream.Result(v)
	default:
		upstream.Panicf("Need string or int to specify port on upstream Hub \"%s\"\n", upstream.Name())
	}

	var ds *Stream
	var dsok bool
	switch v := dnstreamPort.(type) {
	case string:
		ds, dsok = dnstream.FindSource(v)
		if !dsok {
			dnstream.Panicf("No source port \"%s\" found on Hub \"%s\"\n", v, dnstream.Name())
		}
	case int:
		dsok = v >= 0 && v < dnstream.NumSource()
		if !dsok {
			upstream.Panicf("No source port %d found on Hub \"%s\"\n", v, dnstream.Name())
		}
		ds = dnstream.Source(v)
	default:
		dnstream.Panicf("Need string or int to specify port on downstream Hub \"%s\"\n", dnstream.Name())
	}

	if us == nil && ds == nil {
		s := fg.NewStream("")
		fg.streams = append(fg.streams, s)
		upstream.SetResult(upstreamPort, s)
		dnstream.SetSource(dnstreamPort, s)
		return s
	}
	return nil
}

// Run runs the flowgraph
func (fg *Flowgraph) Run() {
	var nodes = make([]*fgbase.Node, len(fg.hubs))
	for i := range nodes {
		nodes[i] = fg.hubs[i].base
	}
	fgbase.RunGraph(nodes)
}
