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

// Code constants for NewHub, followed by init argument for Connect, number of sources, number of results, and description
const (
	Retrieve Code = iota // Retriever	0,1	retrieve one value with Retriever
	Transmit             // Transmitter	1,0	transmit one value with Transmitter
	AllOf                // Transformer	n,m	waiting for all sources
	OneOf                // Transformer	n,m	waiting for one source
	Array                // []interface{}	0,1	produce array of values then EOF
	Constant             // interface{}	0,1	produce constant values forever
	Sink                 // [Sinker]	1,0	consume values forever
	Split                // nil		1,n	split into separate values
	Join                 // nil		n,1	join into one value
	Wait                 // nil		2,1   	wait for second source to pass first
	Pass                 // nil		n,n	pass all values at once
	Steer                // nil		n+1,m	steer rest by first source (?)
	Select               // nil		n+1,1   select rest by first source
	Add                  // nil		2,1	add numbers, concat strings, or use Adder
	Subtract             // nil		2,1	subtract numbers or use Subtracter
	Multiply             // nil		2,1	multiply numbers or use Multiplier
	Divide               // nil		2,1	divide numbers or use Divider
	Modulo               // nil		2,1	modulato numbers or use Moduloer
	And                  // nil		2,1	and bool, bit-wise and integers, or use Ander
	Or                   // nil		2,1	or bool, bit-wise or integers, or use Orer
	Not                  // nil		1,1	negate bool, invert integers, or use Notter
	Shift                // ShiftCode	2,1	shift first by second, Arith,Barrel,Signed or use Shifter
)

// Shift Code
type ShiftCode int

const (
	Arith ShiftCode = iota
	Barrel
	Signed
)

// interfaces used by specific codes
type Adder interface {
	Add(interface{}) interface{}
}

type Subtracter interface {
	Subtract(interface{}) interface{}
}

type Multiplier interface {
	Multiply(interface{}) interface{}
}

type Divider interface {
	Divide(interface{}) interface{}
}

type Moduloer interface {
	Modulo(interface{}) interface{}
}

type Ander interface {
	And(interface{}) interface{}
}

type Orer interface {
	Or(interface{}) interface{}
}

type Notter interface {
	Not() interface{}
}

type Shifter interface {
	Shift(interface{}) interface{}
}

/*=====================================================================*/

// Transformer transforms a slice of source values into a slice
// of result values with the Transform method. Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h *Hub, source []interface{}) (result []interface{}, err error)
}

// Retriever retrieves one value using the Retrieve method. Use Hub.Tracef for tracing.
type Retriever interface {
	Retrieve(h *Hub) (result interface{}, err error)
}

// Transmitter transmits one value using a Transmit method. Use Hub.Tracef for tracing.
type Transmitter interface {
	Transmit(h *Hub, source interface{}) (err error)
}

// Sinker consumes wavefronts of values one at a time forever
type Sinker interface {
	Sink(source []interface{})
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
		if _, ok := init.(Transformer); !ok {
			panic(fmt.Sprintf("Hub with AllOf code not given Transformer for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, allOfFire)
	case OneOf:
		if _, ok := init.(Transformer); !ok {
			panic(fmt.Sprintf("Hub with OneOf code not given Transformer for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, oneOfRdy, oneOfFire)
	case Retrieve:
		if _, ok := init.(Retriever); !ok {
			panic(fmt.Sprintf("Hub with Retrieve code not given Retriever for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, retrieveFire)
	case Transmit:
		if _, ok := init.(Transmitter); !ok {
			panic(fmt.Sprintf("Hub with Transmit code not given Transmitter for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, transmitFire)

	// Math Hubs
	case Add:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.AddFire)
	case Subtract:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.SubFire)
	case Multiply:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.MulFire)
	case Divide:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.DivFire)

	// General Purpose Hubs
	case Wait:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.RdyFire)
	case Array:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ArrayFire)
	case Constant:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ConstFire)
	case Sink:
		if init != nil {
			if _, ok := init.(Sinker); !ok {
				panic(fmt.Sprintf("Hub with Sink code not given Sinker for init %T(%+v)", init, init))
			}
		}
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, nil, nil, fgbase.SinkFire)
	case Steer:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, nil, fgbase.SteervRdy, fgbase.SteervFire)
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
	return fg.connectInit(upstream, upstreamPort, dnstream, dnstreamPort, nil)
}

// ConnectInit connects two hubs via named (string) or indexed (int) ports
// and sets an initial value for flow
func (fg *Flowgraph) ConnectInit(
	upstream *Hub, upstreamPort interface{},
	dnstream *Hub, dnstreamPort interface{},
	init interface{}) *Stream {
	return fg.connectInit(upstream, upstreamPort, dnstream, dnstreamPort, init)
}

// connectInit connects two hubs via named (string) or indexed (int) ports
// and sets an initial value for flow
func (fg *Flowgraph) connectInit(
	upstream *Hub, upstreamPort interface{},
	dnstream *Hub, dnstreamPort interface{},
	init interface{}) *Stream {

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

	if us.base == nil && ds.base == nil {
		s := fg.NewStream("")
		s.Init(init)
		fg.streams = append(fg.streams, s)
		upstream.SetResult(upstreamPort, s)
		dnstream.SetSource(dnstreamPort, s)
		return s
	}
	if us.base == nil {
		upstream.SetResult(upstreamPort, ds)
	} else if ds.base == nil {
		dnstream.SetSource(dnstreamPort, us)
	} else {
		panic("Unexpected two ends to connect but they both are already connected")
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

func allOfFire(n *fgbase.Node) error {
	var a []interface{}
	a = make([]interface{}, len(n.Srcs))
	t := n.Aux.(Transformer)
	eofflag := false
	for i, _ := range a {
		a[i] = n.Srcs[i].SrcGet()
		if v, ok := a[i].(error); ok && v.Error() == "EOF" {
			n.Srcs[i].Flow = false
			eofflag = true
		}
	}
	x, _ := t.Transform(&Hub{n}, a)
	for i, _ := range x {
		if eofflag {
			n.Dsts[i].DstPut(EOF)
		} else {
			n.Dsts[i].DstPut(x[i])
		}
	}
	if eofflag {
		return EOF
	} else {
		return nil
	}

}

func oneOfRdy(n *fgbase.Node) bool {
	for _, v := range n.Srcs {
		if v.SrcRdy(n) {
			return true
		}
	}
	return false
}

func oneOfFire(n *fgbase.Node) error {
	var a []interface{}
	a = make([]interface{}, len(n.Srcs))
	t := n.Aux.(Transformer)
	eofflag := false
	for i, _ := range a {
		if n.Srcs[i].SrcRdy(n) {
			a[i] = n.Srcs[i].SrcGet()
			if v, ok := a[i].(error); ok && v.Error() == "EOF" {
				n.Srcs[i].Flow = false
				eofflag = true
			}
			break
		}
	}
	x, _ := t.Transform(&Hub{n}, a)
	for i, _ := range x {
		if eofflag {
			n.Dsts[i].DstPut(EOF)
		} else {
			n.Dsts[i].DstPut(x[i])
		}
	}
	if eofflag {
		return EOF
	} else {
		return nil
	}

}

func retrieveFire(n *fgbase.Node) error {
	retriever := n.Aux.(Retriever)
	v, err := retriever.Retrieve(&Hub{n})
	n.Dsts[0].DstPut(v)
	return err
}

func transmitFire(n *fgbase.Node) error {
	transmitter := n.Aux.(Transmitter)
	err := transmitter.Transmit(&Hub{n}, n.Srcs[0].SrcGet())
	return err
}
