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

func (g *getter) Retrieve(hub flowgraph.Hub) (result interface{}, err error) {
	i := g.cnt
	g.cnt++
	return i, nil
}

func TestIncoming(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestIncoming\n")

	fg := flowgraph.New("TestIncoming")

	incoming := fg.NewHub("incoming", flowgraph.Retrieve, &getter{}).
		 SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
	     SetSourceNames("A")

	fg.Connect(incoming, "X", sink, "A")

	fg.Run()

	fmt.Printf("END:    TestIncoming\n")
}

/*=====================================================================*/

type putter struct {
	sum int
}

func (p *putter) Transmit(hub flowgraph.Hub, source interface{}) error {
	p.sum += source.(int)
	return nil
}

func TestOutgoing(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestOutgoing\n")

	fg := flowgraph.New("TestOutgoing")

	const1 := fg.NewHub("const1", flowgraph.Constant, 1).
	       SetResultNames("X")

	outgoing := fg.NewHub("outgoing", flowgraph.Transmit, &putter{}).
		 SetSourceNames("A")

	fg.Connect(const1, "X", outgoing, "A")

	fg.Run()

	fmt.Printf("END:    TestOutgoing\n")

}

/*=====================================================================*/

type transformer struct{}

func (t *transformer) Transform(hub flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	xv := source[0].(int) * 2
	return []interface{}{xv}, nil
}

func TestAllOf(t *testing.T) {

	t.Parallel()

	fmt.Printf("BEGIN:  TestAllOf\n")

	fg := flowgraph.New("TestAllOf")

	const1 := fg.NewHub("const1", flowgraph.Constant, 1).
	       SetResultNames("X")

	transformer := fg.NewHub("outgoing", flowgraph.AllOf, &transformer{}).
		    SetSourceNames("A").
		    SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
	     	    SetSourceNames("A")

	fg.Connect(const1, "X", transformer, "A")
	fg.Connect(transformer, "X", sink, "A")

	fg.Run()

	fmt.Printf("END:    TestAllOf\n")
}

/*=====================================================================*/

func TestArray(t *testing.T) {

	fmt.Printf("BEGIN:  TestArray\n")

	fg := flowgraph.New("TestArray")

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	array := fg.NewHub("array", flowgraph.Array, arr)
	array.SetResultNames("X")

	s := &fgbase.SinkStats{}
	sink := fg.NewHub("sink", flowgraph.Sink, s)
	sink.SetSourceNames("A")

	fg.Connect(array, "X", sink, "A")

	fg.Run()

	if s.Cnt != len(arr) {
		t.Fatalf("SinkStats.Cnt %d != len(arr) (%d)\n", s.Cnt, len(arr))
	}

	sum := 0
	for _, v := range arr {
		sum += v.(int)
	}
	if s.Sum != sum {
		t.Fatalf("SinkStats.Sum %d != sum(arr)\n", s.Sum)
	}

	fmt.Printf("END:  TestArray\n")
}

/*=====================================================================*/

type pass struct{}

func (p *pass) Transform(n flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	v := source[n.SourceIndex("A")]
	i := n.ResultIndex("X")
	r := make([]interface{}, i+1)
	r[i] = v
	return r, nil
}

func TestChain(t *testing.T) {
	fmt.Printf("BEGIN:  TestChain\n")
	oldRunTime := fgbase.RunTime
	fgbase.RunTime = 0

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fg := flowgraph.New("TestChain")

	array := fg.NewHub("array", flowgraph.Array, arr)
	array.SetResultNames("X")

	l := 1024
	p := make([]flowgraph.Hub, l)

	for i := 0; i < l; i++ {
		p[i] = fg.NewHub(fmt.Sprintf("t%04d", i), flowgraph.AllOf, &pass{})
		p[i].SetSourceNames("A")
		p[i].SetResultNames("X")
	}

	s := &fgbase.SinkStats{}
	sink := fg.NewHub("sink", flowgraph.Sink, s)
	sink.SetSourceNames("A")

	fg.Connect(array, "X", p[0], "A")
	for i := 0; i < l-1; i++ {
		fg.Connect(p[i], "X", p[i+1], "A")
	}
	fg.Connect(p[l-1], "X", sink, "A")

	fg.Run()

	if s.Cnt != len(arr) {
		t.Fatalf("SinkStats.Cnt %d != len(arr)\n", s.Cnt)
	}

	sum := 0
	for _, v := range arr {
		sum += v.(int)
	}
	if s.Sum != sum {
		t.Fatalf("SinkStats.Sum %d != sum(arr)\n", s.Sum)
	}

	fgbase.RunTime = oldRunTime
	fmt.Printf("END:    TestChain\n")
}

/*=====================================================================*/

func TestDotNaming(t *testing.T) {
	fmt.Printf("BEGIN:  TestDotNaming\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second / 100
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	io.EOF = errors.New("XXX")

	fg := flowgraph.New("TestDotNaming")

	h0 := fg.NewHub("name0", flowgraph.Constant, 100)
	h0.SetResultNames("XYZ")

	h1 := fg.NewHub("name1", flowgraph.Sink, nil)
	h1.SetSourceNames("ABC")

	_, s0ok := h0.FindResult("XYZ")
	if !s0ok {
		t.Fatalf("ERROR Unable to find result port named XYZ\n")
	}

	_, s1ok := h1.FindSource("ABC")
	if !s1ok {
		t.Fatalf("ERROR Unable to find source port named ABC\n")
	}

	fg.Connect(h0, "XYZ", h1, "ABC")

	if h0.Result(0) == nil {
		t.Fatalf("ERROR Unable to find result port numbered 0\n")
	}
	if h0.Result(0).Empty() {
		t.Fatalf("ERROR Unable to find stream at result port numbered 0 on hub %s\n", h0.Name())
	}

	if h1.Source(0) == nil {
		t.Fatalf("ERROR Unable to find source port numbered 0\n")
	}
	if h1.Source(0).Empty() {
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
	fgbase.RunTime = time.Second / 100
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestAdd")

	const100 := fg.NewHub("const100", flowgraph.Constant, 100).
		SetResultNames("X")

	const1 := fg.NewHub("const1", flowgraph.Constant, 1).
		SetResultNames("X")

	add := fg.NewHub("add", flowgraph.Add, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
		SetSourceNames("A")

	fg.Connect(const100, "X", add, "A")
	fg.Connect(const1, "X", add, "B")
	fg.Connect(add, "X", sink, "A")

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestAdd\n")
}

/*=====================================================================*/

func TestLineEq(t *testing.T) {
	fmt.Printf("BEGIN:  TestLineEq\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestLineEq")

	marr := []interface{}{-100, -10, -1, 0, 1, 10, 100}
	xarr := []interface{}{0, 10, 20, 30, -10, -20, -30}
	barr := []interface{}{40, 20, 10, 0, -10, -20, -40}

	m := fg.NewHub("m", flowgraph.Array, marr).
		SetResultNames("X")

	x := fg.NewHub("x", flowgraph.Array, xarr).
		SetResultNames("X")

	b := fg.NewHub("b", flowgraph.Array, barr).
		SetResultNames("X")

	mul := fg.NewHub("mul", flowgraph.Multiply, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	add := fg.NewHub("add", flowgraph.Add, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
		SetSourceNames("A")

	fg.Connect(m, "X", mul, "A")
	fg.Connect(x, "X", mul, "B")
	fg.Connect(mul, "X", add, "A")
	fg.Connect(b, "X", add, "B")
	fg.Connect(add, "X", sink, "A")

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestLineEq\n")
}

/*=====================================================================*/

type tbi struct{}

func (t *tbi) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	return 10, nil
}

func TestIterator1(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator1\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator1")

	tbi := fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
		SetResultNames("X")

	wait := fg.NewHub("wait", flowgraph.Wait, true).
		SetSourceNames("A", "B").
		SetResultNames("X")

	one := fg.NewHub("one", flowgraph.Constant, 1).
		SetResultNames("X")

	sub := fg.NewHub("sub", flowgraph.Subtract, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	steer := fg.NewHub("steer", flowgraph.Steer, nil).
		SetSourceNames("A"). // steer condition.
		SetResultNames("X", "Y")

	fg.Connect(tbi, "X", wait, "A")
	fg.ConnectInit(steer, "X", wait, "B", 0)

	fg.Connect(wait, "X", sub, "A")
	fg.Connect(steer, "Y", sub, "A")
	fg.Connect(one, "X", sub, "B")

	fg.Connect(sub, "X", steer, "A")

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator1\n")
}

/*=====================================================================*/

func TestIterator2(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator2\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator2")

	nextval := fg.NewStream("nextval")
	ready := fg.NewStream("ready").Init(0)
	oldval := fg.NewStream("oldval")
	oneval := fg.NewStream("oneval").Const(1)
	newval := fg.NewStream("newval")

	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
		ConnectResults(nextval)

	fg.NewHub("wait", flowgraph.Wait, true).
		ConnectSources(nextval, ready).
		ConnectResults(oldval)

	fg.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(oldval, oneval).
		ConnectResults(newval)

	fg.NewHub("steer", flowgraph.Steer, nil).
		ConnectSources(newval).
		ConnectResults(ready, oldval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator2\n")
}

/*=====================================================================*/

func TestIterator3(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator3\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator3")

	nextval := fg.NewStream("nextval")
	prevval := fg.NewStream("prevval")

	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
		ConnectResults(nextval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(nextval).
		ConnectResults(prevval)

	oneval := while.NewStream("oneval").Const(1)
	sub := while.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while.Loop(sub)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator3\n")
}
