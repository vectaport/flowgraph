package flowgraph_test

import (
	"errors"
	"fmt"
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
	"io"
	"os"
	"testing"
	"time"
)

/*=====================================================================*/

func TestMain(m *testing.M) {
	fgbase.ConfigByFlag(map[string]interface{}{"trace": "QQ"})
	os.Exit(m.Run())
}

/*=====================================================================*/

func TestNewEqual(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestNewEqual\n")

	// Different allocations should not be equal.
	if flowgraph.New("abc") == flowgraph.New("abc") {
		t.Errorf(`New("abc") == New("abc")`)
	}
	if flowgraph.New("abc") == flowgraph.New("xyz") {
		t.Errorf(`New("abc") == New("xyz")`)
	}

	// Same allocation should be equal to itself (not crash).
	g := flowgraph.New("jkl")
	if g != g {
		t.Errorf(`graph != graph`)
	}

	fmt.Printf("END:    TestNewEqual\n")
}

/*=====================================================================*/

type getter struct {
	cnt int
}

func (g *getter) Get(hub flowgraph.Hub) (interface{}, error) {
	i := g.cnt
	g.cnt++
	return i, nil
}

func TestInsertIncoming(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestInsertIncoming\n")

	fg := flowgraph.New("TestInsertIncoming")
	// n := fg.NewIncoming(&getter{})
	// fg.InsertHub("incoming", n)
	fg.InsertIncoming("incoming", &getter{})
	fg.InsertSink("sink")
	// n := fg.NewSink("sink")
	// fg.InsertHub(n)

	fg.Run()

	fmt.Printf("END:    TestInsertIncoming\n")
}

/*=====================================================================*/

type putter struct {
	sum int
}

func (p *putter) Put(hub flowgraph.Hub, v interface{}) error {
	p.sum += v.(int)
	return nil
}

func TestInsertOutgoing(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestInsertOutgoing\n")

	fg := flowgraph.New("TestInsertOutgoing")
	fg.InsertConst("one", 1)
	fg.InsertOutgoing("outgoing", &putter{})

	fg.Run()

	fmt.Printf("END:    TestInsertOutgoing\n")
}

/*=====================================================================*/

type transformer struct {
}

func (t *transformer) Transform(hub flowgraph.Hub, v ...interface{}) ([]interface{}, error) {
	xv := v[0].(int) * 2
	return []interface{}{xv}, nil
}

func TestInsertAllOf(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestInsertAllOf\n")

	fg := flowgraph.New("TestInsertAllOf")
	fg.InsertConst("one", 1)
	fg.InsertAllOf("double", &transformer{})
	fg.InsertSink("sink")

	fg.Run()

	fmt.Printf("END:    TestInsertAllOf\n")
}

/*=====================================================================*/

func TestInsertArray(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestInsertArray\n")

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fg := flowgraph.New("TestInsertArray")
	fg.InsertArray("array", arr)
	fg.InsertSink("sink")

	fg.Run()

	s := fg.FindHub("sink").Base().(*fgbase.Node).Aux.(fgbase.SinkStats)

	if s.Cnt != len(arr) {
		t.Fatalf("SinkStats.Cnt %d != len(arr)\n", s.Cnt)
	}
	if s.Sum != 45 {
		t.Fatalf("SinkStats.Sum %d != sum(arr)\n", s.Sum)
	}

	fmt.Printf("END:    TestInsertArray\n")
}

/*=====================================================================*/

type pass int

func (p pass) Transform(n flowgraph.Hub, c ...interface{}) ([]interface{}, error) {
	return c, nil
}

var p pass

func TestInsertChain(t *testing.T) {
	// now := time.Now().UTC()
	fmt.Printf("BEGIN:  TestInsertChain\n")
	oldRunTime := fgbase.RunTime
	fgbase.RunTime = 0

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fg := flowgraph.New("TestInsertChain")
	fg.InsertArray("array", arr)
	fg.InsertAllOf("pass", p)
	for i := 0; i < 1024; i++ {
		fg.InsertAllOf("pass", p)
	}
	fg.InsertSink("sink")

	// fmt.Printf("time since:  %v\n", time.Since(now))
	fg.Run()

	s := fg.FindHub("sink").Base().(*fgbase.Node).Aux.(fgbase.SinkStats)

	if s.Cnt != len(arr) || s.Sum != 45 {
		t.Fatalf("ERROR SinkStats %+v\n", s)
	}

	fgbase.RunTime = oldRunTime
	fmt.Printf("END:    TestInsertChain\n")
}

/*=====================================================================*/

func TestDotNaming(t *testing.T) {
	fmt.Printf("BEGIN:  TestDotNaming\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second / 10000
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	io.EOF = errors.New("XXX")

	fg := flowgraph.New("TestDotNaming")

	h0 := fg.NewHub("name0", "CONST", 100)
	h0.SetDestinationNames("XYZ")

	h1 := fg.NewHub("name1", "SINK", nil)
	h1.SetSourceNames("ABC")

	s0 := h0.FindDestination("XYZ")
	if s0 == nil {
		t.Fatalf("ERROR Unable to find destination port named XYZ\n")
	}

	if s0.Base() == nil {
		t.Fatalf("ERROR Unable to find stream at destination port named XYZ\n")
	}

	s1 := h1.FindSource("ABC")
	if s1 == nil {
		t.Fatalf("ERROR Unable to find source port named ABC\n")
	}

	if s1.Base() == nil {
		t.Fatalf("ERROR Unable to find stream at source port named ABC\n")
	}

	fg.Connect(h0, "XYZ", h1, "ABC")

	if h0.Destination(0) == nil {
		t.Fatalf("ERROR Unable to find destination port numbered 0\n")
	}
	if h0.Destination(0).Base().(*fgbase.Edge) == nil {
		t.Fatalf("ERROR Unable to find stream at destination port numbered 0\n")
	}

	if h1.Source(0) == nil {
		t.Fatalf("ERROR Unable to find source port numbered 0\n")
	}
	if h1.Source(0).Base().(*fgbase.Edge) == nil {
		t.Fatalf("ERROR Unable to find stream at source port numbered 0\n")
	}

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestDotNaming\n")
}

/*=====================================================================*/

func TestAdd(t *testing.T) {
	fmt.Printf("BEGIN:  TestAdd\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second / 10000
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestAdd")

	const100 := fg.NewHub("const100", "CONST", 100)
	const100.SetDestinationNames("X")

	const1 := fg.NewHub("const1", "CONST", 1)
	const1.SetDestinationNames("X")

	add := fg.NewHub("add", "ADD", nil)
	add.SetSourceNames("A", "B")
	add.SetDestinationNames("X")

	sink := fg.NewHub("sink", "SINK", nil)
	sink.SetSourceNames("A")

	fg.Connect(const100, "X", add, "A")
	fg.Connect(const1, "X", add, "B")
	fg.Connect(add, "X", sink, "A")

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestAdd\n")
}
