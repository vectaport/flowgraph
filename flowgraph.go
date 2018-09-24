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

/*=====================================================================*/

// Transformer transforms a slice of source values into a slice
// of result values with the Transform method. Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h Hub, source []interface{}) (result []interface{}, err error)
}

/*=====================================================================*/

// Flowgraph interface for flowgraphs assembled out of hubs and streams
type Flowgraph interface {

	// Name returns the name of this flowgraph
	Name() string

	// Hub returns a hub by index
	Hub(i int) Hub
	// Stream returns a stream by index
	Stream(i int) Stream

	// NumHub returns the number of hubs
	NumHub() int
	// NumStream returns the number of streams
	NumStream() int

	// FindHub finds a hub by name
	FindHub(name string) Hub
	// FindStream finds a stream by name
	FindStream(name string) Stream

	// NewHub returns a new unconnected hub
	NewHub(name, code string, init interface{}) Hub
	// NewStream returns a new unconnected stream
	NewStream(name string) Stream

	// Connect connects two hubs via named (string) or indexed (int) ports
	Connect(
		upstream Hub, upstreamPort interface{},
		dnstream Hub, dnstreamPort interface{}) Stream

	// Run runs the flowgraph
	Run()
}

// implementation of Flowgraph
type graph struct {
	name         string
	hubs         []*fgbase.Node
	streams      []*fgbase.Edge
	nameToHub    map[string]*fgbase.Node
	nameToStream map[string]*fgbase.Edge
}

// New returns a named flowgraph implemented with the fgbase package
func New(nm string) Flowgraph {
	nameToHub := make(map[string]*fgbase.Node)
	nameToStream := make(map[string]*fgbase.Edge)
	return &graph{nm, nil, nil, nameToHub, nameToStream}
}

// Name returns the name of this flowgraph
func (fg *graph) Name() string {
	return fg.Name()
}

// Hub returns a hub by index
func (fg *graph) Hub(n int) Hub {
	return hub{fg.hubs[n]}
}

// Stream returns a stream by index
func (fg *graph) Stream(n int) Stream {
	return stream{fg.streams[n]}
}

// NumHub returns the number of hubs
func (fg *graph) NumHub() int {
	return len(fg.hubs)
}

// NumStream returns the number of hubs
func (fg *graph) NumStream() int {
	return len(fg.streams)
}

// NewHub returns a new uninitialized hub
//
// code is "command srcnames;dstnames"
//
// command is one of a set of predefined commands
//
// srcnames is:
// 	    "name[...,name]" for AllOf
// 	    "name[...|name]" for OneOf
// 	    "(name),name[...,name]" for Select
// 	    "<name>,name[...,name]" for Steer
// 	    "" for no sources
//
// dstnames is "name[...,name]" ("" for no results)
//
func (fg *graph) NewHub(name, code string, init interface{}) Hub {

	var n fgbase.Node

	switch code {

	// User Supplied Hubs
	case "ALLOF":
		n = fgbase.MakeNode(name, nil, nil, nil, allOfFire)

		// Math Hubs
	case "ADD":
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.AddFire)
	case "SUB":
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.SubFire)
	case "MUL":
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.MulFire)
	case "DIV":
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.DivFire)

		// General Purpose Hubs
	case "ARRAY":
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ArrayFire)
	case "CONST":
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ConstFire)
	case "SINK":
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, nil, nil, fgbase.SinkFire)
		n.Aux = fgbase.SinkStats{0, 0}
	default:
		log.Panicf("Unexpected Hub code:  %s\n", code)
	}
	if n.Aux == nil {
		n.Aux = init
	}

	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{&n}
}

// NewStream returns a new uninitialized stream
func (fg *graph) NewStream(name string) Stream {
	if name == "" {
		name = fmt.Sprintf("e%d", len(fg.streams))
	}
	e := fgbase.MakeEdge(name, nil)
	fg.streams = append(fg.streams, &e)
	fg.nameToStream[name] = &e
	return stream{&e}
}

// FindHub finds a hub by name
func (fg *graph) FindHub(name string) Hub {
	return hub{fg.nameToHub[name]}
}

// FindStream finds a Stream by name
func (fg *graph) FindStream(name string) Stream {
	return stream{fg.nameToStream[name]}
}

// Connect connects two hubs via named (string) or indexed (int) ports
func (fg *graph) Connect(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{}) Stream {

	var usEdge *fgbase.Edge
	var usok bool
	switch v := upstreamPort.(type) {
	case string:
		usEdge, usok = upstream.Base().(*fgbase.Node).FindDst(v)
		if !usok {
			upstream.Base().(*fgbase.Node).Panicf("No result port \"%s\" found on Hub \"%s\"\n", v, upstream.Name())
		}
	case int:
		usok = v >= 0 && v < upstream.Base().(*fgbase.Node).DstCnt()
		if !usok {
			upstream.Base().(*fgbase.Node).Panicf("No result port %d found on Hub \"%s\"\n", v, dnstream.Name())
		}
		if usok {
			usEdge = upstream.Base().(*fgbase.Node).Dst(v)
		}
	default:
		upstream.Base().(*fgbase.Node).Panicf("Need string or int to specify port on upstream Hub \"%s\"\n", upstream.Name())
	}

	var dsEdge *fgbase.Edge
	var dsok bool
	switch v := dnstreamPort.(type) {
	case string:
		dsEdge, dsok = dnstream.Base().(*fgbase.Node).FindSrc(v)
		if !dsok {
			dnstream.Base().(*fgbase.Node).Panicf("No source port \"%s\" found on Hub \"%s\"\n", v, dnstream.Name())
		}
	case int:
		dsok = v >= 0 && v < dnstream.Base().(*fgbase.Node).SrcCnt()
		if !usok {
			upstream.Base().(*fgbase.Node).Panicf("No source port %d found on Hub \"%s\"\n", v, dnstream.Name())
		}
		if dsok {
			dsEdge = dnstream.Base().(*fgbase.Node).Src(v)
		}
	default:
		dnstream.Base().(*fgbase.Node).Panicf("Need string or int to specify port on downstream Hub \"%s\"\n", dnstream.Name())
	}

	if usEdge == nil && dsEdge == nil {
		e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.streams)), nil)
		fg.streams = append(fg.streams, &e)
		eup := e
		edn := e
		upstream.SetResult(upstreamPort, stream{&eup})
		dnstream.SetSource(dnstreamPort, stream{&edn})
		return stream{&e}
	}
	return nil
}

// Run runs the flowgraph
func (fg *graph) Run() {
	fgbase.RunGraph(fg.hubs)
}
