package flowgraph

import ()

// GraphHub struct for flowgraph hub made out of a graph of hubs
type GraphHub interface {
	Hub
	Flowgraph
	Loop(h Hub)
}

// GraphHub implementation
type graphhub struct {
	hub Hub
	fg Flowgraph
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
func (gh *graphhub) NewGraphHub(name string, graphCode GraphCode) GraphHub {
	return gh.fg.NewGraphHub(name, graphCode)
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
func (gh *graphhub) Source(i int) Stream {
	return gh.hub.Source(i)
}

// Result returns result stream by index
func (gh *graphhub) Result(i int) Stream {
	return gh.hub.Result(i)
}

// SetSource sets a stream on a source port selected by string or int
func (gh *graphhub) SetSource(port interface{}, s Stream) Hub {
	return gh.hub.SetSource(port, s)
}

// SetResult sets a stream on a result port selected by string or int
func (gh *graphhub) SetResult(port interface{}, s Stream) Hub {
	return gh.hub.SetResult(port, s)
}

// FindSource returns source stream by port name
func (gh *graphhub) FindSource(port interface{}) (s Stream, portok bool) {
	return gh.hub.FindSource(port)
}

// FindResult returns result stream by port name
func (gh *graphhub) FindResult(port interface{}) (s Stream, portok bool) {
	return gh.hub.FindResult(port)
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

// SourceIndex returns the index of a named source port, -1 if not found
func (gh *graphhub) SourceIndex(port string) int {
	return gh.hub.SourceIndex(port)
}

// ResultIndex returns the index of a named result port, -1 if not found
func (gh *graphhub) ResultIndex(port string) int {
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

// Flowgraph returns associate flowgraph interface
func (gh *graphhub) Flowgraph() Flowgraph {
	return gh.hub.Flowgraph()
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

        checkfgHub(gh.fg, h)

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

	for i := 0; i < ns; i++ {
		gh.Connect(wait, i, h, i)
		gh.Connect(h, i, steer, i)
	}
}

// flatten connects GraphHub external ports to internal dangling streams
func (gh *graphhub) flatten() {
	gh.Tracef("FLATTEN %T(%+v) nhub, nstream %d,%d %s\n", gh, gh, gh.NumHub(), gh.NumStream(), gh.Name())
	var sources []Stream
	var results []Stream
	for _, v := range gh.fg.(*flowgraph).hubs {
		if gv, ok := v.(GraphHub); ok {
			gv.(*graphhub).flatten()
		}
		gh.Tracef("FLATTEN ns,nr %d,%d on %s %T(%+v) %s\n", v.NumSource(), v.NumResult(), v.Name(), v, v, v.Name())
		for i:=0; i<v.NumSource(); i++ {
		    s := v.Source(i)
		    if s.Empty() {
		       s = gh.NewStream("")
		       v.SetSource(i, s)
		       sources = append(sources, s)
		    }
		}
		for i:=0; i<v.NumResult(); i++ {
		    r := v.Result(i)
		    if r.Empty() {
		       r = gh.NewStream("")
		       v.SetResult(i, r)
		       results = append(results, r)
		    }
		}
	}
	for i,v := range sources {
	    gh.Tracef("sources[%d] = %+v, nu,nd %d,%d\n", i, v, v.NumUpstream(), v.NumDownstream())
	    gh.Tracef("v.Downstream(0) is %+v\n", v.Downstream(0))
        }
	for i,v := range results {
	    gh.Tracef("results[%d] = %+v  nu,nd %d,%d\n", i, v, v.NumUpstream(), v.NumDownstream())
	    gh.Tracef("v.Upstream(0) is %+v\n", v.Upstream(0))
        }
}
