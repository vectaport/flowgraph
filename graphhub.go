package flowgraph

import (
	"github.com/vectaport/fgbase"
)

// GraphHub interface for flowgraph hub made out of a graph of hubs
type GraphHub interface {
	Hub
	Flowgraph

	// Loop builds a conditional iterator for a while or during loop
	Loop(h Hub)

	// Link links an internal stream to an external stream
	Link(in, ex Stream)
}

// GraphHub implementation
type graphhub struct {
	hub      Hub
	fg       Flowgraph
	isources []Stream
	iresults []Stream
}

// Title returns the title of this flowgraph
func (gh *graphhub) Title() string {
	return gh.fg.Title()
}

// Name returns the name of this hub
func (gh *graphhub) Name() string {
	return gh.hub.Name()
}

// Hub returns a hub by index
func (gh *graphhub) Hub(n int) Hub {
	return gh.fg.Hub(n)
}

// Stream returns a stream by index
func (gh *graphhub) Stream(n int) Stream {
	return gh.fg.Stream(n)
}

// NumHub returns the number of hubs
func (gh *graphhub) NumHub() int {
	return gh.fg.NumHub()
}

// NumStream returns the number of hubs
func (gh *graphhub) NumStream() int {
	return gh.fg.NumStream()
}

// NewHub returns a new unconnected hub
func (gh *graphhub) NewHub(name string, code Code, init interface{}) Hub {
	return gh.fg.NewHub(name, code, init)
}

// NewStream returns a new unconnected stream
func (gh *graphhub) NewStream(name string) Stream {
	return gh.fg.NewStream(name)
}

// NewGraphHub returns a hub with a sub-graph
func (gh *graphhub) NewGraphHub(name string, code Code) GraphHub {
	return gh.fg.NewGraphHub(name, code)
}

// FindHub finds a hub by name
func (gh *graphhub) FindHub(name string) Hub {
	return gh.fg.FindHub(name)
}

// FindStream finds a Stream by name
func (gh *graphhub) FindStream(name string) Stream {
	return gh.fg.FindStream(name)
}

// Connect connects two hubs via named (string) or indexed (int) ports
func (gh *graphhub) Connect(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{}) Stream {
	return gh.fg.Connect(upstream, upstreamPort, dnstream, dnstreamPort)
}

// ConnectInit connects two hubs via named (string) or indexed (int) ports
// and sets an initial value for flow
func (gh *graphhub) ConnectInit(
	upstream Hub, upstreamPort interface{},
	dnstream Hub, dnstreamPort interface{},
	init interface{}) Stream {
	return gh.fg.ConnectInit(upstream, upstreamPort, dnstream, dnstreamPort, init)
}

// Run runs the flowgraph
func (gh *graphhub) Run() {
	gh.fg.Run()
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (gh *graphhub) Tracef(format string, v ...interface{}) {
	gh.hub.Tracef(format, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (gh *graphhub) LogError(format string, v ...interface{}) {
	gh.hub.LogError(format, v...)
}

// Panicf for logging of panic messages.  Uses atomic log mechanism.
func (gh *graphhub) Panicf(format string, v ...interface{}) {
	gh.hub.Panicf(format, v...)
}

// Source returns source stream by index
func (gh *graphhub) Source(port interface{}) Stream {
	return gh.hub.Source(port)
}

// Result returns result stream by index
func (gh *graphhub) Result(port interface{}) Stream {
	return gh.hub.Result(port)
}

// SetSource sets a stream on a source port selected by string or int
func (gh *graphhub) SetSource(port interface{}, s Stream) Hub {
	return gh.hub.SetSource(port, s)
}

// SetResult sets a stream on a result port selected by string or int
func (gh *graphhub) SetResult(port interface{}, s Stream) Hub {
	return gh.hub.SetResult(port, s)
}

// AddSources adds a source port for each stream
func (gh *graphhub) AddSources(s ...Stream) Hub {
	return gh.hub.AddSources(s...)
}

// AddResults adds a result port for each stream
func (gh *graphhub) AddResults(s ...Stream) Hub {
	return gh.hub.AddResults(s...)
}

// NumSource returns the number of source ports
func (gh *graphhub) NumSource() int {
	return gh.hub.NumSource()
}

// NumResult returns the number of result ports
func (gh *graphhub) NumResult() int {
	return gh.hub.NumResult()
}

// SetNumSource sets the number of source ports
func (gh *graphhub) SetNumSource(n int) Hub {
	return gh.hub.SetNumSource(n)
}

// SetNumResult sets the number of result ports
func (gh *graphhub) SetNumResult(n int) Hub {
	return gh.hub.SetNumResult(n)
}

// SourceNames returns the names of the source ports
func (gh *graphhub) SourceNames() []string {
	return gh.hub.SourceNames()
}

// ResultNames returns the names of the result ports
func (gh *graphhub) ResultNames() []string {
	return gh.hub.ResultNames()
}

// SetSourceNames names the source ports
func (gh *graphhub) SetSourceNames(nm ...string) Hub {
	return gh.hub.SetSourceNames(nm...)
}

// SetResultNames names the result ports
func (gh *graphhub) SetResultNames(nm ...string) Hub {
	return gh.hub.SetResultNames(nm...)
}

// SourceIndex returns the index of a source port matched by name or Stream
func (gh *graphhub) SourceIndex(port interface{}) int {
	return gh.hub.SourceIndex(port)
}

// ResultIndex returns the index of a source port matched by name or Stream
func (gh *graphhub) ResultIndex(port interface{}) int {
	return gh.hub.ResultIndex(port)
}

// ConnectSources connects a list of source Streams to this Hub
func (gh *graphhub) ConnectSources(source ...Stream) Hub {
	return gh.hub.ConnectSources(source...)
}

// ConnectResults connects a list of result Streams to this Hub
func (gh *graphhub) ConnectResults(result ...Stream) Hub {
	return gh.hub.ConnectResults(result...)

}

// Code returns code associated with a Hub
func (gh *graphhub) Code() Code {
	return gh.hub.Code()
}

// Flowgraph returns associate flowgraph interface
func (gh *graphhub) Flowgraph() Flowgraph {
	return gh.fg
}

// Empty returns true if the underlying implementation is nil
func (gh *graphhub) Empty() bool {
	return gh.fg == nil && gh.hub == nil
}

// Base returns value of underlying implementation
func (gh *graphhub) Base() interface{} {
	return gh.hub.Base()
}

// Loop builds a conditional iterator for a while or during loop
func (gh *graphhub) Loop(h Hub) {

	checkInternalHub(gh.fg, h)

	ns := h.NumSource()
	nsmax := ns
	for i := 0; i < nsmax; i++ {
		s := h.Source(i)
		if s.Empty() {
			continue
		}
		if s.IsConst() {
			ns--
		}
	}

	nr := h.NumResult()
	if ns != nr {
		panic("handle ns != nr\n")
	}

	wait := gh.NewHub(h.Name()+"_wait", Wait, nil).
		SetNumSource(ns + 1).
		SetNumResult(ns)

	steer := gh.NewHub(h.Name()+"_steer", Steer, nil).
		SetNumSource(ns).
		SetNumResult(ns * 2)

	if ns != 1 {
		panic("need to support more than ns==1\n")
	}

	for i := 0; i < ns; i++ {
		gh.Connect(wait, i, h, i)
		gh.Connect(h, i, steer, i)
		gh.Connect(steer, i+1, h, i)
	}
	termc := gh.ConnectInit(steer, 0, wait, ns, 0) // termination condition recycled but also needs to be output
	termc.Base().(*fgbase.Edge).Val = nil          // remove initialization condition from termination condition
	gh.iresults = append(gh.iresults, termc)

}

// flatten connects GraphHub external ports to internal dangling streams
func (gh *graphhub) flatten(nodes []*fgbase.Node) []*fgbase.Node {
	ns, nr := 0, 0
	for _, v := range gh.fg.(*flowgraph).hubs {
		if gv, ok := v.(GraphHub); ok {
			nodes = gv.(*graphhub).flatten(nodes)
			if fgbase.DotOutput {
				nodes = append(nodes, v.Base().(*fgbase.Node))
			}

		} else {
			nodes = append(nodes, v.Base().(*fgbase.Node))
		}
		for i := 0; i < v.NumSource(); i++ {
			s := v.Source(i)
			if s.Empty() {
				s = gh.NewStream("")
				v.SetSource(i, s)
				// gh.Tracef("Making stub source stream %q with *Edge %p\n", s.Name(), s.Base().(*fgbase.Edge))
				gh.isources = append(gh.isources, v.Source(i))
				ns++
			}
		}
		for i := 0; i < v.NumResult(); i++ {
			r := v.Result(i)
			if r.Empty() {
				r = gh.NewStream("")
				v.SetResult(i, r)
				// gh.Tracef("Making stub result stream %q with *Edge %p\n", r.Name(), r.Base().(*fgbase.Edge))
				gh.iresults = append(gh.iresults, v.Result(i))
				nr++
			}
		}
	}
	if len(gh.isources) != gh.NumSource() {
		gh.Panicf("# of GraphHub sources (%d) does not match # of dangling internal inputs (%d)\n", gh.NumSource(), len(gh.isources))
	}
	if len(gh.iresults) != gh.NumResult() {
		gh.Panicf("# of GraphHub results (%d) does not match # of dangling internal outputs (%d)\n", gh.NumResult(), len(gh.iresults))
	}

	// dangling inputs
	for i, s := range gh.isources {
		gh.Tracef("Source stream %q on outer hub \"%s\"\n", gh.Source(i).Name(), gh.Name())
		for j := 0; j < s.NumDownstream(); j++ {
			gh.Tracef("\tlinked by source stream %q (*fbase.Edge=%p) that ends at hub %q port %v\n", s.Name(), s.Base().(*fgbase.Edge), s.Downstream(j).Name(), s.Downstream(j).SourceIndex(s))
		}
		jmax := s.NumDownstream()
		for j := 0; j < jmax; j++ {
			gh.Link(s, gh.Source(i))
		}

	}

	// dangling or designated outputs
	for i, r := range gh.iresults {
		gh.Tracef("Result stream %q that starts at hub %q port %v\n", r.Name(), r.Upstream(0).Name(), r.Upstream(0).ResultIndex(r))
		gh.Tracef("\tlinked by result stream %q (*fgbase.Edge=%p) on outer hub %q\n", gh.Result(i).Name(), gh.Result(i).Base().(*fgbase.Edge), gh.Name())
		jmax := r.NumUpstream()
		for j := 0; j < jmax; j++ {
			gh.Link(r, gh.Result(i))
		}
	}
	return nodes
}

// Link links an internal stream to an external stream
func (gh *graphhub) Link(in, ex Stream) {

	checkInternalStream(gh.fg, in)
	checkExternalStream(gh.fg, ex)

	ein := in.Base().(*fgbase.Edge)
	eex := ex.Base().(*fgbase.Edge)
	gh.Base().(*fgbase.Node).Link(ein, eex)
}
