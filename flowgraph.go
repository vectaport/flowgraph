// Package flowgraph for scalable asynchronous system development
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"errors"
	"fmt"
	"github.com/vectaport/fgbase"
)

/*=====================================================================*/

// End of flow
var EOF = errors.New("EOF")

/*=====================================================================*/

// Getter receives a value with the Get method. Use Hub.Tracef for tracing.
type Getter interface {
	Get(h Hub) (interface{}, error)
}

// Putter transmits a value with the Put method. Use Hub.Tracef
// for tracing
type Putter interface {
	Put(h Hub, v interface{}) error
}

// Transformer transforms a variadic list of values into a slice
// of values with the Transform method. Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h Hub, c ...interface{}) ([]interface{}, error)
}

/*=====================================================================*/

// Flowgraph interface for flowgraphs assembled out of hubs and streams
type Flowgraph interface {

	// Name returns the name of this flowgraph
	Name() string

	// Hub returns a hub by index
	Hub(i int) Hub
	// Stream returns a connector by index
	Stream(i int) Stream

	// NumHub returns the number of hubs
	NumHub() int
	// NumStream returns the number of streams
	NumStream() int

	// FindHub finds a hub by name
	FindHub(name string) Hub
	// FindStream finds a stream by name
	FindStream(name string) Stream

	// NewHub returns a new uninitialized hub
	NewHub(name, code string) Hub
	// NewStream returns a new uninitialized connector
	NewStream(name string) Stream

	// InsertHub adds a Hub to the flowgraph, connecting inputs to existing
	// dangling streams as available and creating dangling output streams as needed.
	InsertHub(h Hub)

	// NewIncoming creates an input hub that uses a Getter
	NewIncoming(name string, getter Getter) Hub
	// InsertIncoming adds an input source that uses a Getter
	InsertIncoming(name string, getter Getter) Hub

	// InsertOutgoing adds an output destination that uses a Putter
	InsertOutgoing(name string, putter Putter) Hub

	// InsertConst adds an input constant as an incoming source.
	InsertConst(name string, v interface{}) Hub
	// InsertArray adds an array as an incoming source.
	InsertArray(name string, arr []interface{}) Hub

	// NewSink creates an output sink hub
	NewSink(name string) Hub
	// InsertSink adds an output sink
	InsertSink(name string) Hub

	// InsertAllOf adds a transform that waits for all inputs before producing outputs
	InsertAllOf(name string, transformer Transformer) Hub

	// Connect connects two hubs via named ports
	Connect(upstream Hub, dstport string, dnstream Hub, srcport string) Stream

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

// Stream returns a connector by index
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
func (fg *graph) NewHub(name, code string) Hub {
	n := fgbase.MakeNode(name, nil, nil, nil, nil)
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

// InsertHub adds a Hub to the flowgraph, connecting inputs to existing
// dangling streams as available and creating dangling output streams as needed.
func (fg *graph) InsertHub(h Hub) {
	fmt.Printf("CALL MADE TO INSERTHUB\n")

	i := 0

	nextDanglingSrcStream :=
		func() *fgbase.Edge {
			for ; i < len(fg.streams) && fg.streams[i].DstCnt() != 0; i++ {
			}

			if i == len(fg.streams) {
				return nil // or makeStream?
			} else {
				return fg.streams[i]
			}
		}

	// connect to input streams
	for j := 0; j < h.NumSource(); j++ {
		if h.Source(j) == nil {
			p := nextDanglingSrcStream()
			fmt.Printf("p is now %v\n", p)
		}
	}

	// create output streams

}

// NewIncoming adds an incoming source that uses a Getter
func (fg *graph) NewIncoming(name string, getter Getter) Hub {
	n := funcIncoming(fgbase.Edge{}, getter)
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertIncoming adds an incoming source that uses a Getter
func (fg *graph) InsertIncoming(name string, getter Getter) Hub {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.streams)), nil)
	fg.streams = append(fg.streams, &e)
	n := funcIncoming(e, getter)
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertOutgoing adds a destination that uses a Putter
func (fg *graph) InsertOutgoing(name string, putter Putter) Hub {
	n := funcOutgoing(*fg.streams[len(fg.streams)-1], putter)
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertConst adds an input constant as an incoming source.
func (fg *graph) InsertConst(name string, v interface{}) Hub {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.streams)), nil)
	fg.streams = append(fg.streams, &e)
	n := fgbase.FuncConst(e, v)
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertArray adds an array as an incoming source.
func (fg *graph) InsertArray(name string, arr []interface{}) Hub {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.streams)), nil)
	fg.streams = append(fg.streams, &e)
	n := fgbase.FuncArray(e, arr)
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// NewSink creates an output sink hub
func (fg *graph) NewSink(name string) Hub {
	n := fgbase.FuncSink(fgbase.Edge{})
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertSink adds a output sink on the latest stream
func (fg *graph) InsertSink(name string) Hub {
	i := len(fg.streams) - 1
	n := fgbase.FuncSink(*fg.streams[i])
	n.Name = name
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// InsertAllOf adds a transform that waits for all inputs before producing outputs
func (fg *graph) InsertAllOf(name string, transformer Transformer) Hub {
	e := fgbase.MakeEdge(fmt.Sprintf("e%d", len(fg.streams)), nil)
	fg.streams = append(fg.streams, &e)
	n := funcAllOf([]fgbase.Edge{*fg.streams[len(fg.streams)-2]}, []fgbase.Edge{*fg.streams[len(fg.streams)-1]},
		name, transformer)
	fg.hubs = append(fg.hubs, &n)
	fg.nameToHub[name] = &n
	return hub{fg.hubs[len(fg.hubs)-1]}
}

// Connect connects two hubs via named ports
func (fg *graph) Connect(upstream Hub, dstPort string, dnstream Hub, srcPort string) Stream {
	usEdge, srcok := upstream.Base().(*fgbase.Node).FindDst(dstPort)
	dsEdge, dstok := dnstream.Base().(*fgbase.Node).FindSrc(srcPort)
	if !srcok || !dstok {
		return stream{nil}
	}
	if usEdge == nil && dsEdge == nil {
		fmt.Printf("READY TO CONNECT\n")
	}
	return nil
}

// Run runs the flowgraph
func (fg *graph) Run() {
	fgbase.RunGraph(fg.hubs)
}
