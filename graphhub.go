package flowgraph

import (
)

// GraphHub struct for flowgraph hub made out of a graph of hubs
type GraphHub interface {
	Hub
	Flowgraph
	Loop(h Hub) GraphHub
}

// GraphHub implementation
type graphhub struct {
	hb Hub
	fg Flowgraph
}

// Title returns the title of this flowgraph
func (gh *graphhub) Title() string {
	return gh.fg.Title()
}

// Name returns the name of this hub
func (gh *graphhub) Name() string {
	return gh.hb.Name()
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
	gh.hb.Tracef(format, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (gh *graphhub) LogError(format string, v ...interface{}) {
	gh.hb.LogError(format, v...)
}

// Panicf for logging of panic messages.  Uses atomic log mechanism.
func (gh *graphhub) Panicf(format string, v ...interface{}) {
	gh.hb.Panicf(format, v...)
}

// Source returns source stream by index
func (gh *graphhub) Source(i int) Stream {
	return gh.hb.Source(i)
}

// Result returns result stream by index
func (gh *graphhub) Result(i int) Stream {
	return gh.hb.Result(i)
}

// FindSource returns source stream by port name
func (gh *graphhub) FindSource(port interface{}) (s Stream, portok bool) {
	return gh.hb.FindSource(port)
}

// FindResult returns result stream by port name
func (gh *graphhub) FindResult(port interface{}) (s Stream, portok bool) {
	return gh.hb.FindResult(port)
}

// AddSources adds a source port for each stream
func (gh *graphhub) AddSources(s ...Stream) Hub {
	return gh.hb.AddSources(s...)
}

// AddResults adds a result port for each stream
func (gh *graphhub) AddResults(s ...Stream) Hub {
	return gh.hb.AddResults(s...)
}

// NumSource returns the number of source ports
func (gh *graphhub) NumSource() int {
	return gh.hb.NumSource()
}

// NumResult returns the number of result ports
func (gh *graphhub) NumResult() int {
	return gh.hb.NumResult()
}

// SetSourceNum sets the number of source ports
func (gh *graphhub) SetSourceNum(n int) Hub {
	return gh.hb.SetSourceNum(n)
}

// SetResultNum sets the number of result ports
func (gh *graphhub) SetResultNum(n int) Hub {
	return gh.hb.SetResultNum(n)
}

// SetSourceNames names the source ports
func (gh *graphhub) SetSourceNames(nm ...string) Hub {
	return gh.hb.SetSourceNames(nm...)
}

// SetResultNames names the result ports
func (gh *graphhub) SetResultNames(nm ...string) Hub {
	return gh.hb.SetResultNames(nm...)
}

// SourceNames returns the names of the source ports
func (gh *graphhub) SourceNames() []string {
	return gh.hb.SourceNames()
}

// ResultNames returns the names of the result ports
func (gh *graphhub) ResultNames() []string {
	return gh.hb.ResultNames()
}

// SetSource sets a stream on a source port selected by string or int
func (gh *graphhub) SetSource(port interface{}, s Stream) Hub {
	return gh.hb.SetSource(port, s)
}

// SetResult sets a stream on a result port selected by string or int
func (gh *graphhub) SetResult(port interface{}, s Stream) Hub {
	return gh.hb.SetResult(port, s)
}

// SourceIndex returns the index of a named source port, -1 if not found
func (gh *graphhub) SourceIndex(port string) int {
	return gh.hb.SourceIndex(port)
}

// ResultIndex returns the index of a named result port, -1 if not found
func (gh *graphhub) ResultIndex(port string) int {
	return gh.hb.ResultIndex(port)
}

// Empty returns true if the underlying implementation is nil
func (gh *graphhub) Empty() bool {
	return gh.fg == nil && gh.hb == nil
}

// ConnectSources connects a list of source Streams to this Hub
func (gh *graphhub) ConnectSources(source ...Stream) Hub {
	return gh.hb.ConnectSources(source...)
}

// ConnectResults connects a list of result Streams to this Hub
func (gh *graphhub) ConnectResults(result ...Stream) Hub {
	return gh.hb.ConnectResults(result...)

}

// Base returns value of underlying implementation
func (gh *graphhub) Base() interface{} {
	return gh.hb.Base()
}

func (gh *graphhub) Loop(h Hub) GraphHub {

	ns := h.NumSource()
	nsmax := ns
	for i := 0; i < nsmax; i++ {
		s := h.Source(i)
		if s.Empty() { continue }
		if s.IsConst() {
			ns--
		}
	}

	nr := h.NumResult()
	if ns != nr {
		panic("handle ns != nr\n")
	}

	subfg := New(h.Name() + "_fg")
	wait := subfg.NewHub(h.Name()+"_wait", Wait, nil).
		SetSourceNum(ns).
		SetResultNum(nr)

	steer := subfg.NewHub(h.Name()+"_steer", Steer, nil).
		SetSourceNum(ns).
		SetResultNum(nr)

	for i := 0; i < ns; i++ {
		subfg.Connect(wait, i, h, i)
		subfg.Connect(h, i, steer, i)
	}

	return gh
}
