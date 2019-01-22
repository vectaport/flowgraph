// Package flowgraph for scalable asynchronous development.
// Build systems out of hubs interconnected by streams of data.
// https://github.com/vectaport/flowgraph/wiki
package flowgraph

import (
	"github.com/vectaport/fgbase"

	"errors"
	"flag"
	"fmt"
	"log"
)

// End of flow. Transmitted when end-of-file occurs, and promises no more
// data to follow.
var EOF = errors.New("EOF")

// ParseFlags parses the command line flags for this package
var flatDot = false

func ParseFlags() {
	flag.BoolVar(&flatDot, "flatdot", false, "flatten dot output")
	fgbase.ConfigByFlag(map[string]interface{}{"trace": "V"})
	fgbase.TraceStyle = fgbase.New
}

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

	// NumStream returns the number of streams
	NumStream() int

	// NewHub returns a new unconnected hub
	NewHub(name string, code HubCode, init interface{}) Hub

	// NewStream returns a new unconnected stream
	NewStream(name string) Stream

	// NewGraphHub returns a hub with a flowgraph inside
	NewGraphHub(name string, code HubCode) GraphHub

	// FindHub finds a hub by name
	FindHub(name string) Hub

	// FindStream finds a stream by name
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

func (f *fgTransformer) String() string {
	return fgbase.String(f.t)
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

	// User Hubs
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
		init = &fgTransformer{fg, init.(Transformer)}

	// Control Hubs
	case Wait:
		n = fgbase.MakeNode(name, nil, nil, waitRdy, waitFire)

	case Select:
		n = fgbase.MakeNode(name, nil, nil, fgbase.SteervRdy, selectFire)

	case Steer:
		n = fgbase.MakeNode(name, nil, nil, fgbase.SteervRdy, fgbase.SteervFire)

	case Cross:
		n = fgbase.MakeNode(name, nil, nil, crossRdy, crossFire)

	// Data Hubs
	case Array:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ArrayFire)

	case Constant:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, fgbase.ConstFire)

	case Pass:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, []*fgbase.Edge{nil}, nil, nil)

	case Split:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, splitFire)

	case Join:
		n = fgbase.MakeNode(name, nil, []*fgbase.Edge{nil}, nil, joinFire)

	case Sink:
		if init != nil {
			if _, ok := init.(Sinker); !ok {
				panic(fmt.Sprintf("Hub with Sink code not given Sinker for init %T(%+v)", init, init))
			}
		}
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, nil, nil, fgbase.SinkFire)

	// Math and Logic Hubs
	case Add:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.AddFire)

	case Subtract:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.SubFire)

	case Multiply:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.MulFire)

	case Divide:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.DivFire)

	case Modulo:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, fgbase.ModFire)

	case And:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, andFire)

	case Or:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, orFire)

	case Not:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil}, []*fgbase.Edge{nil}, nil, notFire)

	case Shift:
		n = fgbase.MakeNode(name, []*fgbase.Edge{nil, nil}, []*fgbase.Edge{nil}, nil, shiftFire)

	default:
		log.Panicf("Unexpected Hub code NewHub:  %s\n", code)
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
		return ds
	}
	if ds.Empty() {
		dnstream.SetSource(dnstreamPort, us)
		return us
	}
	/*
		upstream.Panicf("Unexpected that with two ports to connect (%s:%v(%s) and %s:%v(%s() that they both are already connected to another stream as well",
			upstream.Name(), upstreamPort, us.Name(), dnstream.Name(), dnstreamPort, ds.Name())
	*/
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
		panic(fmt.Sprintf("Hub %q created by flowgraph %q (expected it to be created by flowgraph %q)",
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
		panic(fmt.Sprintf("Stream %q created by flowgraph %q (expected it to be created by flowgraph %q)",
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
	if flatDot {
		fgbase.DotOutput = true
	}
	nodes := make([]*fgbase.Node, 0)
	for _, v := range fg.hubs {
		if gv, ok := v.(GraphHub); ok && (flatDot || !fgbase.DotOutput) {
			nodes = gv.(*graphhub).flatten(nodes)
			if fgbase.DotOutput {
				nodes = append(nodes, v.Base().(*fgbase.Node))
				v.Base().(*fgbase.Node).SetDotAttr("style=\"dashed\"")
			} else {
				for i := 0; i < v.NumSource(); i++ {
					v.Source(i).Base().(*fgbase.Edge).Disconnect(v.Base().(*fgbase.Node))
				}
				for i := 0; i < v.NumResult(); i++ {
					v.Result(i).Base().(*fgbase.Edge).Disconnect(v.Base().(*fgbase.Node))
				}
			}
		} else {
			nodes = append(nodes, v.Base().(*fgbase.Node))
		}
	}
	fmt.Printf("\n")
	for _, v := range nodes {
		fmt.Printf("// %s\n", v)
	}
	fmt.Printf("\n")
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
			if x[i] != nil {
				n.Dsts[i].DstPut(x[i])
			}
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
			if x[i] != nil {
				n.Dsts[i].DstPut(x[i])
			}
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

type waitStruct struct {
	Request int
}

func waitRdy(n *fgbase.Node) bool {
	ns := n.SrcCnt()
	ws, init := n.Aux.(waitStruct)
	if !init {
		ws = waitStruct{Request: fgbase.ChannelSize - 1}
		elocal := n.Srcs[ns-1]
		usnode := elocal.SrcNode(0)
		for i := 0; i < len(usnode.Dsts); i++ {
			if usnode.Dsts[i].Same(elocal) {
				usnode.Dsts[i].RdyCnt += fgbase.ChannelSize
				break
			}
		}
		n.Aux = ws
	}

	for i := 0; i < ns-1; i++ {
		if !n.Srcs[i].SrcRdy(n) {
			return false
		}
	}
	if ws.Request > 0 {
		ws.Request--
		n.Srcs[ns-1].Flow = false
		n.Aux = ws
		return true
	}
	n.Aux = ws

	rdy := n.Srcs[ns-1].SrcRdy(n)
	if rdy {
		n.Srcs[ns-1].Flow = true
	}
	return rdy
}

func waitFire(n *fgbase.Node) error {
	ns := n.SrcCnt()
	for i := 0; i < ns-1; i++ {
		n.Dsts[i].DstPut(n.Srcs[i].SrcGet())
	}
	return nil
}

func selectFire(n *fgbase.Node) error {
	n.Panicf("Select HubCode still needs implementation.")
	return nil
}

func ranksz(n *fgbase.Node) int {
	return n.SrcCnt() / 2
}

type crossStruct struct {
	in  int
	out int
}

func crossRdy(n *fgbase.Node) bool {
	cs, init := n.Aux.(crossStruct)
	if !init {
		cs = crossStruct{}
	}

	numrank := ranksz(n)

	f := func(offset int) bool {
		for i := 0; i < numrank; i++ {
			if !n.Srcs[i+offset].SrcRdy(n) {
				return false
			}
		}
		return true
	}
	left := f(0)
	right := f(numrank)

	steerDir := func() int {
		v := n.Srcs[numrank*cs.in].SrcGet()
		if b, ok := v.(Breaker); ok {
			if b.Break() {
				return 0
			}
			return 1
		}
		if fgbase.IsZero(v) {
			return 0
		}
		return 1
	}

	if left || right {
		if cs.in == 0 {
			if right {
				cs.in = 1
			} else {
				cs.in = 0
			}
		} else {
			if left {
				cs.in = 0
			} else {
				cs.in = 1
			}
		}
		cs.out = steerDir()
		rdy := n.Dsts[cs.out].DstRdy(n)
		if false && rdy {
			n.Tracef("STEERDIR IS %d\n", cs.out)
		}
		n.Aux = cs
		return rdy
	}
	return false
}

func crossFire(n *fgbase.Node) error {
	numrank := ranksz(n)
	cs := n.Aux.(crossStruct)

	if cs.out == 0 {
		for i := 0; i < numrank; i++ {
			n.Dsts[i].DstPut(n.Srcs[cs.in*numrank+i].SrcGet())
		}
	} else {
		for i := 0; i < numrank; i++ {
			n.Dsts[numrank+i].DstPut(n.Srcs[cs.in*numrank+i].SrcGet())
		}
	}
	return nil
}

func splitFire(n *fgbase.Node) error {
	n.Panicf("Split HubCode still needs implementation.")
	return nil
}

func joinFire(n *fgbase.Node) error {
	n.Panicf("Join HubCode still needs implementation.")
	return nil
}

func whileFire(n *fgbase.Node) error {
	n.Panicf("While loop still needs flattening.")
	return nil
}

func duringFire(n *fgbase.Node) error {
	n.Panicf("During loop still needs flattening.")
	return nil
}

func graphFire(n *fgbase.Node) error {
	n.Panicf("Graph still needs flattening.")
	return nil
}

func andFire(n *fgbase.Node) error {
	n.Panicf("And HubCode still needs implementation.")
	return nil
}

func orFire(n *fgbase.Node) error {
	n.Panicf("Or HubCode still needs implementation.")
	return nil
}

func notFire(n *fgbase.Node) error {
	n.Panicf("Not HubCode still needs implementation.")
	return nil
}

func shiftFire(n *fgbase.Node) error {
	n.Panicf("Shift HubCode still needs implementation.")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
