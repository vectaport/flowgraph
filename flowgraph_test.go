package flowgraph_test

import (
	"fmt"
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
	"os"
	"testing"
	// "time"
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
		t.Fatalf("SinkStats.Cnt %d != len(arr) %d\n", s.Cnt)
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
		t.Fatalf("ERROR SinkStats %+V\n", s)
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
	fgbase.RunTime = 0
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.VVVV

	fg := flowgraph.New("TestDotNaming")

	n0 := fg.NewHub("name0", "")
	n0.SetDestinationNames("xyz")

	n1 := fg.NewHub("name1", "")
	n1.SetSourceNames("abc")

	fg.Connect(n0, "xyz", n1, "abc")

	e0 := n0.FindDestination("xyz")
	if e0 == nil {
		t.Fatalf("ERROR Unable to find destination port named xyz\n")
	}

	if e0.Base() == nil {
		t.Fatalf("ERROR Unable to find edge at destination port named xyz\n")
	}

	e1 := n1.FindSource("abc")
	if e1 == nil {
		t.Fatalf("ERROR Unable to find source port named abc\n")
	}

	if e1.Base() == nil {
		t.Fatalf("ERROR Unable to find edge at source port named abc\n")
	}

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestDotNaming\n")
}
