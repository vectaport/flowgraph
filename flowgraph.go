// Package flowgraph for scalable asynchronous development.
// Build systems out of hubs interconnected by streams of data.
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

// HubCode is code arg to NewHub()
type HubCode int

// HubCode constants for NewHub() code arg.
// Comment fields are init arg for NewHub, number of source and number of results
// (if n or m given thru init arg), and description.
const (
	Retrieve HubCode = iota // Retriever	0,1	retrieve one value with Retriever
	Transmit                // Transmitter	1,0	transmit one value with Transmitter
	AllOf                   // Transformer	n,m	waiting for all sources
	OneOf                   // Transformer	n,m	waiting for one source

	Array    // []interface{}	0,1	produce array of values then EOF
	Constant // interface{}	0,1	produce constant values forever
	Sink     // [Sinker]	1,0	consume values forever
	Split    // n		1,n	split into separate values
	Join     // n		n,1	join into one value
	Wait     // n		n+1,1 	wait for last source to pass rest
	Pass     // n		n,n	pass all values at once
	Steer    // n,m		1+n,n*m	steer rest by m ways from first source
	Select   // n		1+n,1   select from rest by first source

	Graph  // nil		n,m     hub with general purpose internals
	While  // nil		n,n	hub with internal wait-body-steer loop
	During // nil		n,n	hub with while loop with continuous results

	Add      // [Transformer]	2,1	add numbers, concat strings
	Subtract // [Transformer]	2,1	subtract numbers
	Multiply // [Transformer]	2,1	multiply numbers
	Divide   // [Transformer]	2,1	divide numbers
	Modulo   // [Transformer]	2,1	modulate numbers
	And      // [Transformer]	2,1	AND bool or bit-wise AND integers
	Or       // [Transformer]	2,1	OR bool or bit-wise OR integers
	Not      // [Transformer]	1,1	negate bool, invert integers
	Shift    // ShiftCode|Transformer	2,1	shift first by second, Arith,Barrel,Signed
)

// ShiftCode is the subcode for the Shift HubCode
type ShiftCode int

const (
	Arith ShiftCode = iota
	Barrel
	Signed
)

/*=====================================================================*/

// Transformer transforms a slice of source values into a slice
// of result values with the Transform method.  Used by Hub with AllOf or OneOf HubCode.
// Use Hub.Tracef for tracing.
type Transformer interface {
	Transform(h Hub, source []interface{}) (result []interface{}, err error)
}

// Retriever retrieves one value using the Retrieve method.
// Used by Hub with Retrieve HubCode.  Use Hub.Tracef for tracing.
type Retriever interface {
	Retrieve(h Hub) (result interface{}, err error)
}

// Transmitter transmits one value using a Transmit method.
// Used by Hub with Transmit HubCode to transmit values. Use Hub.Tracef for tracing.
type Transmitter interface {
	Transmit(h Hub, source interface{}) (err error)
}

// Sinker consumes wavefronts of values one at a time forever
// Optionally used by Hub with Sink HubCode.
type Sinker interface {
	Sink(source []interface{})
}

/*=====================================================================*/

// Flowgraph interface for flowgraphs assembled out of hubs and streams
type Flowgraph interface {

	// Title returns the title of this flowgraph
	Title() string

	// Hub returns a hub by index
	Hub(n int) Hub

	// Stream returns a stream by index
	Stream(n int) Stream

	// NumHub returns the number of hubs
	NumHub() int

	// NumStream returns the number of hubs
	NumStream() int

	// NewHub returns a new unconnected hub
	NewHub(name string, code HubCode, init interface{}) Hub

	// NewStream returns a new unconnected stream
	NewStream(name string) Stream

	// NewGraphHub returns a hub with a flowgraph inside
	NewGraphHub(name string, code HubCode) GraphHub

	// FindHub finds a hub by name
	FindHub(name string) Hub

	// FindStream finds a Stream by name
	FindStream(name string) Stream

	// Connect connects two hubs via named (string) or indexed (int) ports
	Connect(
		upstream Hub, upstreamPort interface{},
		dnstream Hub, dnstreamPort interface{}) Stream

	// ConnectInit connects two hubs via named (string) or indexed (int) ports
	// and sets an initial value for flow
	ConnectInit(
		upstream Hub, upstreamPort interface{},
		dnstream Hub, dnstreamPort interface{},
		init interface{}) Stream

	// Run runs the flowgraph
	Run()
}

type flowgraph struct {
	title        string
	hubs         []Hub
	streams      []Stream
	nameToHub    map[string]Hub
	nameToStream map[string]Stream
}

// New returns a titled flowgraph
func New(title string) Flowgraph {
	nameToHub := make(map[string]Hub)
	nameToStream := make(map[string]Stream)
	fg := flowgraph{title, nil, nil, nameToHub, nameToStream}
	return &fg
}

// Title returns the titleof this flowgraph
func (fg *flowgraph) Title() string {
	return fg.title
}

// Hub returns a hub by index
func (fg *flowgraph) Hub(n int) Hub {
	return fg.hubs[n]
}

// Stream returns a stream by index
func (fg *flowgraph) Stream(n int) Stream {
	return fg.streams[n]
}

// NumHub returns the number of hubs
func (fg *flowgraph) NumHub() int {
	return len(fg.hubs)
}

// NumStream returns the number of hubs
func (fg *flowgraph) NumStream() int {
	return len(fg.streams)
}

type fgTransformer struct {
	fg *flowgraph
	t  Transformer
}

type fgRetriever struct {
	fg *flowgraph
	r  Retriever
}

type fgTransmitter struct {
	fg *flowgraph
	t  Transmitter
}

// NewHub returns a new unconnected hub
func (fg *flowgraph) NewHub(name string, code HubCode, init interface{}) Hub {

	var n fgbase.Node

	switch code {

	// User Supplied Hubs
	case Retrieve:
		if _, ok := init.(Retriever); !ok {
			panic(fmt.Sprintf("Hub with Retrieve code not given Retriever for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, retrieveFire)
		init = &fgRetriever{fg, init.(Retriever)}
	case Transmit:
		if _, ok := init.(Transmitter); !ok {
			panic(fmt.Sprintf("Hub with Transmit code not given Transmitter for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, transmitFire)
		init = &fgTransmitter{fg, init.(Transmitter)}
	case AllOf:
		if _, ok := init.(Transformer); !ok {
			panic(fmt.Sprintf("Hub with AllOf code not given Transformer for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, nil, allOfFire)
		init = &fgTransformer{fg, init.(Transformer)}

	case OneOf:
		if _, ok := init.(Transformer); !ok {
			panic(fmt.Sprintf("Hub with OneOf code not given Transformer for init %T(%+v)", init, init))
		}
		n = fgbase.MakeNode(name, nil, nil, oneOfRdy, oneOfFire)

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
		n = fgbase.MakeNode(name, nil, nil, fgbase.SteervRdy, fgbase.SteervFire)

	default:
		log.Panicf("Unexpected Hub code NewHub:  %v\n", code)
	}
	if n.Aux == nil {
		n.Aux = init
	}

	h := &hub{&n, fg, code}
	n.Owner = h
	fg.hubs = append(fg.hubs, h)
	fg.nameToHub[name] = h
	return h
}

// NewStream returns a new unconnected stream
func (fg *flowgraph) NewStream(name string) Stream {
	e := fgbase.MakeEdge(name, nil)
	s := &stream{&e, fg}
	fg.streams = append(fg.streams, s)
	fg.nameToStream[name] = s
	return s
}

// NewGraphHub returns a hub with an internal flowgraph
func (fg *flowgraph) NewGraphHub(name string, code HubCode) GraphHub {
	newfg := New(name + "_fg")

	var n fgbase.Node

	switch code {
	case While:
		n = fgbase.MakeNode(name, nil, nil, nil, whileFire)
	case During:
		n = fgbase.MakeNode(name, nil, nil, nil, duringFire)
	case Graph:
		n = fgbase.MakeNode(name, nil, nil, nil, graphFire)
	default:
		log.Panicf("Unexpected HubCode for NewGraphHub:  %v\n", code)
	}
	gh := &graphhub{&hub{&n, fg, code}, newfg, nil, nil}
	n.Owner = gh
	fg.hubs = append(fg.hubs, gh)
	fg.nameToHub[name] = gh
	return gh
}

// FindHub finds a hub by name
func (fg *flowgraph) FindHub(name string) Hub {
	return fg.nameToHub[name]
}

// FindStream finds a Stream by name
func (fg *flowgraph) FindStream(name string) Stream {
	return fg.nameToStream[name]
}

// Connect connects two hubs via named (string) or indexed (int) ports
func (fg *flowgraph) Connect(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{}) Stream {
	return fg.connectInit(upstream, upstreamPort, dnstream, dnstreamPort, nil)
}

// ConnectInit connects two hubs via named (string) or indexed (int) ports
// and sets an initial value for flow
func (fg *flowgraph) ConnectInit(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{},
	init interface{}) Stream {
	return fg.connectInit(upstream, upstreamPort, dnstream, dnstreamPort, init)
}

// connectInit connects two hubs via named (string) or indexed (int) ports
// and sets an initial value for flow
func (fg *flowgraph) connectInit(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{},
	init interface{}) Stream {

	checkInternalHub(fg, upstream)
	checkInternalHub(fg, dnstream)

	var us Stream
	var usok bool
	switch v := upstreamPort.(type) {
	case string:
		us, usok = upstream.Result(v), !upstream.Empty()
		if !usok {
			upstream.Panicf("No result port \"%s\" found on Hub \"%s\"\n", v, upstream.Name())
		}
	case int:
		usok = v >= 0 && v < upstream.NumResult()
		if !usok {
			upstream.Panicf("No result port %d found on Hub \"%s\"\n", v, upstream.Name())
		}
		us = upstream.Result(v)
	default:
		upstream.Panicf("Need string or int to specify port on upstream Hub \"%s\"\n", upstream.Name())
	}

	var ds Stream
	var dsok bool
	switch v := dnstreamPort.(type) {
	case string:
		ds, dsok = dnstream.Source(v), !dnstream.Empty()
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

	if us.Empty() && ds.Empty() {
		s := fg.NewStream("")
		s.Init(init)
		fg.streams = append(fg.streams, s)
		upstream.SetResult(upstreamPort, s)
		dnstream.SetSource(dnstreamPort, s)
		return s
	}
	if us.Empty() {
		upstream.SetResult(upstreamPort, ds)
	} else if ds.Empty() {
		dnstream.SetSource(dnstreamPort, us)
	} else {
		panic("Unexpected with two ports to connect that they both are already connected to another stream as well")
	}
	return nil
}

// Run runs the flowgraph
func (fg *flowgraph) Run() {
	fg.run()
}

// checkInternalHub checks that the flowgraph associated with a Hub matches
func checkInternalHub(fg Flowgraph, h Hub) {
	if h == nil {
		return
	}
	fgknown := fg
	fgtest := h.Flowgraph()
	if fgknown != fgtest {
		panic(fmt.Sprintf("Hub %q created by flowgraph %q (expected flowgraph %q)",
			h.Name(), fgtest.Title(), fgknown.Title()))
	}
}

// checkExternalHub checks that the flowgraph associated with a Hub doesn't match
func checkExternalHub(fg Flowgraph, h Hub) {
	if h == nil {
		return
	}
	fgknown := fg
	fgtest := h.Flowgraph()
	if fgknown == fgtest {
		panic(fmt.Sprintf("External Hub %q created by same flowgraph %q",
			h.Name(), fgtest.Title()))
	}
}

// checkInternalStream checks that the flowgraph associated with a Stream matches
func checkInternalStream(fg Flowgraph, s Stream) {
	if s == nil {
		return
	}
	fgknown := fg
	fgtest := s.Flowgraph()
	if fgknown != fgtest {
		panic(fmt.Sprintf("Stream %q created by flowgraph %q (expected flowgraph %q)",
			s.Name(), fgtest.Title(), fgknown.Title()))
	}
}

// checkExternalStream checks that the flowgraph associated with a Stream doesn't match
func checkExternalStream(fg Flowgraph, s Stream) {
	if s == nil {
		return
	}
	fgknown := fg
	fgtest := s.Flowgraph()
	if fgknown == fgtest {
		panic(fmt.Sprintf("External Stream %q created by same flowgraph %q",
			s.Name(), fgtest.Title()))
	}
}

// flatten connects GraphHub external ports to internal dangling streams
func (fg *flowgraph) flatten() []*fgbase.Node {
	nodes := make([]*fgbase.Node, 0)
	for _, v := range fg.hubs {
		if gv, ok := v.(GraphHub); ok {
			nodes = gv.(*graphhub).flatten(nodes)
			if fgbase.DotOutput {
				nodes = append(nodes, v.Base().(*fgbase.Node))
			}
		} else {
			nodes = append(nodes, v.Base().(*fgbase.Node))
		}
	}
	return nodes
}

// run runs the flowgraph
func (fg *flowgraph) run() {

	nodes := fg.flatten()
	fgbase.RunGraph(nodes)
}

func allOfFire(n *fgbase.Node) error {
	var a []interface{}
	a = make([]interface{}, len(n.Srcs))
	t := n.Aux.(*fgTransformer).t
	fg := n.Aux.(*fgTransformer).fg
	eofflag := false
	for i, _ := range a {
		a[i] = n.Srcs[i].SrcGet()
		if v, ok := a[i].(error); ok && v.Error() == "EOF" {
			n.Srcs[i].Flow = false
			eofflag = true
		}
	}
	x, _ := t.Transform(&hub{n, fg, AllOf}, a)
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
	t := n.Aux.(*fgTransformer).t
	fg := n.Aux.(*fgTransformer).fg
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
	x, _ := t.Transform(&hub{n, fg, OneOf}, a)
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
	retriever := n.Aux.(*fgRetriever).r
	fg := n.Aux.(*fgRetriever).fg
	v, err := retriever.Retrieve(&hub{n, fg, Retrieve})
	n.Dsts[0].DstPut(v)
	return err
}

func transmitFire(n *fgbase.Node) error {
	transmitter := n.Aux.(*fgTransmitter).t
	fg := n.Aux.(*fgTransmitter).fg
	err := transmitter.Transmit(&hub{n, fg, Transmit}, n.Srcs[0].SrcGet())
	return err
}

func whileFire(n *fgbase.Node) error {
	n.Panicf("while loop still needs flattening.")
	return nil
}

func duringFire(n *fgbase.Node) error {
	n.Panicf("during loop still needs flattening.")
	return nil
}

func graphFire(n *fgbase.Node) error {
	n.Panicf("graph still needs expanding.")
	return nil
}
