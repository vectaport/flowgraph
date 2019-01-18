package flowgraph_test

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

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

/*=====================================================================*/

func TestMain(m *testing.M) {
	fgbase.ConfigByFlag(map[string]interface{}{"trace": "V"})
	fgbase.TraceStyle = fgbase.New
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
	oldTraceLevel := fgbase.TraceLevel
	fgbase.TraceLevel = fgbase.V

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

	fgbase.TraceLevel = oldTraceLevel
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

	fg.Connect(h0, "XYZ", h1, "ABC")

	if h0.Result(0).Empty() {
		t.Fatalf("ERROR Unable to find stream at result port numbered 0 on hub %s\n", h0.Name())
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

type sinkAdd struct {
	t *testing.T
}

func (st *sinkAdd) Sink(source []interface{}) {
	if source[0].(int) != 101 {
		st.t.Fatalf("ERROR Add result is not 101 as expected %d.\n", source[0].(int))
	}
}

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

	sink := fg.NewHub("sink", flowgraph.Sink, &sinkAdd{t}).
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

type sinkLineEq struct {
	t  *testing.T
	gt []interface{}
}

func (st *sinkLineEq) Sink(source []interface{}) {
	if source[0].(int) != st.gt[0] {
		st.t.Fatalf("ERROR LineEq result is not %d as expected (%d).\n", source[0].(int), st.gt[0])
	}
	st.gt = st.gt[1:]
}

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
	yarr := make([]interface{}, len(marr))
	for i := range marr {
		yarr[i] = marr[i].(int)*xarr[i].(int) + barr[i].(int)
	}

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

	sink := fg.NewHub("sink", flowgraph.Sink, &sinkLineEq{t, yarr}).
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

/* TestIterator1 Flowgraph HDL *

tbten()(.X(firstval))
wait(.A(firstval),.B(lastval=0))(.X(oldval))
sub(.A(oldval),.B(1))(.X(newval))
steer(.A(newval))(.X(lastval),.Y(oldval))
sink(.A(lastval))()

*/

type tbten struct{}

func (t *tbten) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	return 10, nil
}

type sinkIterator1 struct {
	t *testing.T
}

func (st *sinkIterator1) Sink(source []interface{}) {
	if source[0].(int) != 0 {
		st.t.Fatalf("ERROR Iterator1 FAILED\n")
	}
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

	tbten := fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		SetResultNames("X")

	wait := fg.NewHub("wait", flowgraph.Wait, true).
		SetSourceNames("A", "B").
		SetResultNames("X")

	one := fg.NewHub("one", flowgraph.Constant, 1).
		SetResultNames("X")

	sub := fg.NewHub("sub", flowgraph.Subtract, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	steer := fg.NewHub("steer", flowgraph.Steer, 1).
		SetSourceNames("A"). // steer condition.
		SetResultNames("X", "Y")

	sink := fg.NewHub("sink", flowgraph.Sink, &sinkIterator1{t}).
		SetSourceNames("A")

	fg.Connect(tbten, "X", wait, "A").SetName("firstval")
	fg.ConnectInit(steer, "X", wait, "B", 0).SetName("lastval")

	fg.Connect(wait, "X", sub, "A").SetName("oldval")
	fg.Connect(steer, "Y", sub, "A") // oldval
	fg.Connect(one, "X", sub, "B").SetName("oneval")

	fg.Connect(sub, "X", steer, "A").SetName("newval")

	fg.Connect(steer, "X", sink, "A") // lastval

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator1\n")
}

/*=====================================================================*/

/* TestIterator2 Flowgraph HDL *

tbten()(firstval)
wait(firstval,lastval=0)(oldval)
sub(oldval,1)(newval)
steer(newval)(lastval,oldval)
sink(lastval)()

*/

type sinkIterator2 struct {
	t *testing.T
	i int
}

func (st *sinkIterator2) Sink(source []interface{}) {
	(*st).i = (*st).i - 1
	if (*st).i != source[0].(int) {
		st.t.Fatalf("ERROR Iterator2 FAILED\n")
	}
	if (*st).i == 0 {
		(*st).i = 10
	}
}

func TestIterator2(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator2\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator2")

	firstval := fg.NewStream("firstval")
	lastval := fg.NewStream("lastval").Init(0)
	oldval := fg.NewStream("oldval")
	oneval := fg.NewStream("oneval").Const(1)
	newval := fg.NewStream("newval")

	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval)

	fg.NewHub("wait", flowgraph.Wait, true).
		ConnectSources(firstval, lastval).
		ConnectResults(oldval)

	fg.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(oldval, oneval).
		ConnectResults(newval)

	fg.NewHub("steer", flowgraph.Steer, nil).
		ConnectSources(newval).
		ConnectResults(lastval, oldval)

	fg.NewHub("sink", flowgraph.Sink, &sinkIterator2{t, 10}).
		ConnectSources(newval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator2\n")
}

/*=====================================================================*/

/* TestIterator3 Flowgraph HDL *

tbten()(firstval)
while(firstval)(lastval) {
        sub(firstval, 1)(lastval)
}
sink(lastval)()

*/

type sinkIterator3 struct {
	t *testing.T
}

func (st *sinkIterator3) Sink(source []interface{}) {
	if source[0].(int) != 0 {
		st.t.Fatalf("ERROR Iterator3 FAILED\n")
	}
}

func TestIterator3(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator3\n")
	oldRunTime := fgbase.RunTime
	oldTracePorts := fgbase.TracePorts
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator3")

	firstval := fg.NewStream("firstval")
	lastval := fg.NewStream("lastval")

	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(firstval).
		ConnectResults(lastval)

	oneval := while.NewStream("oneval").Const(1)
	while.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, &sinkIterator3{t}).
		ConnectSources(lastval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator3\n")
}

/*=====================================================================*/

/* TestIterator4 Flowgraph HDL *

tbten()(firstval1)
while(firstval1)(lastval1) {
        sub(firstval1, 1)(lastval1)
}
wait(10,lastval1)(firstval2)
while(firstval2)(lastval2) {
        sub(firstval2, 1)(lastval2)
}
sink(lastval2)()

*/

func TestIterator4(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator4\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator4")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval1)

	lastval1 := fg.NewStream("lastval1")
	while1 := fg.NewGraphHub("while1", flowgraph.While)
	while1.ConnectSources(firstval1).
		ConnectResults(lastval1)

	oneval1 := while1.NewStream("oneval1").Const(1)
	while1.NewHub("sub1", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval1)
	while1.Loop()

	tenval := fg.NewStream("tenval").Const(10)
	firstval2 := fg.NewStream("firstval2")
	fg.NewHub("wait", flowgraph.Wait, 1).
		ConnectSources(tenval, lastval1).
		ConnectResults(firstval2)

	lastval2 := fg.NewStream("lastval2")
	while2 := fg.NewGraphHub("while2", flowgraph.While)
	while2.ConnectSources(firstval2).
		ConnectResults(lastval2)

	oneval2 := while2.NewStream("oneval2").Const(1)
	while2.NewHub("sub2", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval2)
	while2.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval2)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator4\n")
}

/*=====================================================================*/

/* TestIterator5 Flowgraph HDL *

tbten()(firstval)
while(firstval)(lastval) {
        while(firstval)(lastval) {
                sub(firstval, 1)(lastval)
	}
}
sink(lastval)()

*/

func TestIterator5(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator5\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.VVV

	fg := flowgraph.New("TestIterator5")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval1)

	lastval1 := fg.NewStream("lastval1")
	while1 := fg.NewGraphHub("while1", flowgraph.While)
	while1.ConnectSources(firstval1).
		ConnectResults(lastval1)

	while2 := while1.NewGraphHub("while2", flowgraph.While)
	while2.SetNumSource(1).SetNumResult(1) // could be detected
	while1.Loop()

	oneval := while2.NewStream("oneval").Const(1)
	while2.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while2.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval1)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator5\n")
}

/*=====================================================================*/

/* TestIterator6 Flowgraph HDL *

tbten()(firstval)
while(firstval)(lastval) {
        while(firstval)(lastval) {
        	while(firstval)(lastval) {
                	sub(firstval, 1)(lastval)
		}
	}
}
sink(lastval)()

*/

func TestIterator6(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator6\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator6")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval1)

	lastval1 := fg.NewStream("lastval1")
	while1 := fg.NewGraphHub("while1", flowgraph.While)
	while1.ConnectSources(firstval1).
		ConnectResults(lastval1)

	while2 := while1.NewGraphHub("while2", flowgraph.While)
	while2.SetNumSource(1).SetNumResult(1) // could be detected
	while1.Loop()

	while3 := while2.NewGraphHub("while3", flowgraph.While)
	while3.SetNumSource(1).SetNumResult(1) // could be detected
	while2.Loop()

	oneval := while3.NewStream("oneval").Const(1)
	while3.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while3.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval1)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator6\n")
}

/*=====================================================================*/

/* TestIterator7 Flowgraph HDL *

tbten()(firstval)
while(firstval)(lastval) {
        while(firstval)(lastval) {
        	while(firstval)(lastval) {
        		while(firstval)(lastval) {
                		sub(firstval, 1)(lastval)
			}
		}
	}
}
sink(lastval)()

*/

func TestIterator7(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator7\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator7")

	firstval := fg.NewStream("firstval")
	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval)

	lastval := fg.NewStream("lastval")
	while1 := fg.NewGraphHub("while1", flowgraph.While)
	while1.ConnectSources(firstval).
		ConnectResults(lastval)

	while2 := while1.NewGraphHub("while2", flowgraph.While)
	while2.SetNumSource(1).SetNumResult(1) // could be detected
	while1.Loop()

	while3 := while2.NewGraphHub("while3", flowgraph.While)
	while3.SetNumSource(1).SetNumResult(1) // could be detected
	while2.Loop()

	while4 := while3.NewGraphHub("while4", flowgraph.While)
	while4.SetNumSource(1).SetNumResult(1) // could be detected
	while3.Loop()

	oneval := while4.NewStream("oneval").Const(1)
	while4.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while4.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator7\n")
}

/*=====================================================================*/

/* TestIterator8 Flowgraph HDL *

tbten()(firstval)
while(firstval)(lastval) {
        while(firstval)(lastval) {
	        add(firstval,1)(bumpval)
        	while(bumpval)(lastval) {
        		while(bumpval)(lastval) {
                		sub(bumpval, 1)(lastval)
			}
		}
	}
}
sink(lastval)()

*/

func TestIterator8(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator8\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator8")

	firstval := fg.NewStream("firstval")
	fg.NewHub("tbten", flowgraph.Retrieve, &tbten{}).
		ConnectResults(firstval)

	lastval := fg.NewStream("lastval")
	while1 := fg.NewGraphHub("while1", flowgraph.While)
	while1.ConnectSources(firstval).
		ConnectResults(lastval)

	while2 := while1.NewGraphHub("while2", flowgraph.While)
	while2.SetNumSource(1).SetNumResult(1) // could be detected
	oneval := while1.NewStream("oneval").Const(1)
	add := while1.NewHub("add", flowgraph.Add, nil).
		ConnectSources(nil, oneval)
	while1.Connect(add, 0, while2, 0).SetName("bumpval")
	while1.Loop()

	oneval2 := while2.NewStream("oneval2").Const(1)
	while2.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval2)
	while2.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator8\n")
}

/*=====================================================================*/

/* TestIterator9 Flowgraph HDL *

tbtens()(firstval)
while(firstval)(lastval) {
        sub(firstval, 1)(lastval)
}
sink(lastval)()

*/

type valIterator9 struct {
	Count int
	ID    int
}

func (vi *valIterator9) Break() bool {
	return vi.Count == 0
}

type tbtens struct{}

var tbtensID int

func (t *tbtens) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	id := tbtensID
	tbtensID++
	return &valIterator9{10, id}, nil
}

type sinkIterator9 struct {
	t *testing.T
	c int
}

func (st *sinkIterator9) Sink(source []interface{}) {
	if source[0].(*valIterator9).Count != 0 {
		st.t.Fatalf("ERROR Iterator9 FAILED\n")
	}
	st.c++
}

type subber struct{}

func (s *subber) Transform(n flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	result = make([]interface{}, 1)
	a := source[0].(*valIterator9).Count
	b := source[1].(int)
	x := a - b
	result[0] = &valIterator9{x, source[0].(*valIterator9).ID}
	return
}

func TestIterator9(t *testing.T) {
	fmt.Printf("BEGIN:  TestIterator9\n")

	fg := flowgraph.New("TestIterator9")

	firstval := fg.NewStream("firstval")
	lastval := fg.NewStream("lastval")

	fg.NewHub("tbtens", flowgraph.Retrieve, &tbtens{}).
		ConnectResults(firstval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(firstval).
		ConnectResults(lastval)

	oneval := while.NewStream("oneval").Const(1)
	while.NewHub("subber", flowgraph.AllOf, &subber{}).
		ConnectSources(nil, oneval).
		SetNumResult(1)
	while.Loop()

	si := &sinkIterator9{t, 0}
	fg.NewHub("sink", flowgraph.Sink, si).
		ConnectSources(lastval)

	fg.Run()

	fmt.Printf("// SINKS: %d\n", si.c)
	fmt.Printf("END:    TestIterator9\n")
}

/*=====================================================================*/

/* TestGCD Flowgraph HDL *

tbrand()(mval)
tbrand()(nval)
while(mval, nval)(tcond, gcd) {
        sub(firstval, 1)(lastval)
}
sink(gcd)()

*/

type tbrand struct {
	init bool
}

func (t *tbrand) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	if !(*t).init {
		// (*t).init = true
		return rand.Intn(100) + 1, nil
	}
	return flowgraph.EOF, flowgraph.EOF
}

func TestGCD(t *testing.T) {
	fmt.Printf("BEGIN:  TestGCD\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.VVV

	fg := flowgraph.New("TestGCD")

	mval := fg.NewStream("mval")
	nval := fg.NewStream("nval")
	tcond := fg.NewStream("tcond")
	gcd := fg.NewStream("gcd")

	fg.NewHub("tbrand", flowgraph.Retrieve, &tbrand{}).
		ConnectResults(mval)
	fg.NewHub("tbrand", flowgraph.Retrieve, &tbrand{}).
		ConnectResults(nval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(mval, nval).
		ConnectResults(tcond, gcd)

	passm := while.NewHub("passm", flowgraph.Pass, nil)
	mod := while.NewHub("mod", flowgraph.Modulo, nil)
	while.ExposeResult(while.Connect(passm, 0, mod, 1))

	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(gcd)
	fg.NewHub("sink2", flowgraph.Sink, nil).
		ConnectSources(tcond)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestGCD\n")
}

/*=====================================================================*/

/* TestGoRound Flowgraph HDL *

ival,jval,kval,lval,mval,nval=tbcar(),tbcar(),tbcar(),tbcar(),tbcar(),tbcar()
while(ival,jval,kval,lval,mval,nval)(iout,jout,kout,lout,mout,nout) {
        iout,jout,kout,lout,mout,nout=onelap(ival,jval,kval,lval,mval,nval)
}
sink(iout)
sink(jout)
sink(kout)
sink(lout)
sink(mout)
sink(nout)

*/

type car struct {
	number   int64
	distance int
	winner   int
}

func (c *car) Break() bool {
	return c.winner != -1
}

type tbcar struct {
}

func (t *tbcar) Retrieve(h flowgraph.Hub) (result interface{}, err error) {
	return &car{h.Base().(*fgbase.Node).ID, rand.Intn(100) + 1, -1}, nil
}

type onelap struct{}

func (l *onelap) Transform(n flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	result = make([]interface{}, len(source))
	for i := range source {
		cprev := source[i].(*car)
		cnext := &car{cprev.number, max(0, cprev.distance-rand.Intn(6)+1), -1}
		result[i] = cnext
		if cnext.distance == 0 {
			czero := result[0].(*car)
			czero.winner = i
		}
	}
	return
}

func TestGoRound(t *testing.T) {
	fmt.Printf("BEGIN:  TestGoRound\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestGoRound")

	ival := fg.NewStream("ival")
	jval := fg.NewStream("jval")
	kval := fg.NewStream("kval")
	lval := fg.NewStream("lval")
	mval := fg.NewStream("mval")
	nval := fg.NewStream("nval")

	iout := fg.NewStream("iout")
	jout := fg.NewStream("jout")
	kout := fg.NewStream("kout")
	lout := fg.NewStream("lout")
	mout := fg.NewStream("mout")
	nout := fg.NewStream("nout")

	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(ival)
	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(jval)
	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(kval)
	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(lval)
	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(mval)
	fg.NewHub("tbcar", flowgraph.Retrieve, &tbcar{}).
		ConnectResults(nval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(ival, jval, kval, lval, mval, nval).
		ConnectResults(iout, jout, kout, lout, mout, nout)

	while.NewHub("onelap", flowgraph.AllOf, &onelap{}).
		SetNumSource(6).
		SetNumResult(6)

	while.Loop()

	fg.NewHub("sinki", flowgraph.Sink, nil).
		ConnectSources(iout)
	fg.NewHub("sinkj", flowgraph.Sink, nil).
		ConnectSources(jout)
	fg.NewHub("sinkk", flowgraph.Sink, nil).
		ConnectSources(kout)
	fg.NewHub("sinkl", flowgraph.Sink, nil).
		ConnectSources(lout)
	fg.NewHub("sinkm", flowgraph.Sink, nil).
		ConnectSources(mout)
	fg.NewHub("sinkn", flowgraph.Sink, nil).
		ConnectSources(nout)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestGoRound\n")
}

/*=====================================================================*/

/* TestCross Flowgraph HDL *

ival,jval=tbtoggle(),tbrand()
cross(ival,jval)(iout,jout)
sink(iout)
sink(kout)

*/

type tbtoggle struct {
	last bool
}

func (t *tbtoggle) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	f := t.last
	t.last = !f
	return f, nil
}

func TestCross(t *testing.T) {
	fmt.Printf("BEGIN:  TestCross\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestCross")

	ival := fg.NewStream("ival")
	jval := fg.NewStream("jval")
	kval := fg.NewStream("kval")
	lval := fg.NewStream("lval")
	iout := fg.NewStream("iout")
	jout := fg.NewStream("jout")
	kout := fg.NewStream("kout")
	lout := fg.NewStream("lout")

	fg.NewHub("tbtoggle", flowgraph.Retrieve, &tbtoggle{}).
		ConnectResults(ival)
	fg.NewHub("tbrand", flowgraph.Retrieve, &tbrand{}).
		ConnectResults(jval)
	fg.NewHub("tbtoggle", flowgraph.Retrieve, &tbtoggle{}).
		ConnectResults(kval)
	fg.NewHub("tbrand", flowgraph.Retrieve, &tbrand{}).
		ConnectResults(lval)

	fg.NewHub("cross", flowgraph.Cross, nil).
		ConnectSources(ival, jval, kval, lval).
		ConnectResults(iout, jout, kout, lout)

	fg.NewHub("sinki", flowgraph.Sink, nil).
		ConnectSources(iout)
	fg.NewHub("sinkj", flowgraph.Sink, nil).
		ConnectSources(jout)
	fg.NewHub("sinkk", flowgraph.Sink, nil).
		ConnectSources(kout)
	fg.NewHub("sinkl", flowgraph.Sink, nil).
		ConnectSources(lout)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestCross\n")
}

/*=====================================================================*/

/* DuckPondA Flowgraph HDL *

nest()(newduck)
while(newduck)(oldduck) {
        pond(newduck)(oldduck)
}
sink(oldduck)()

*/

var duckCnt int64 = 0

type duck struct {
	ID    int64
	Loops int
	Orig  string
	exit  bool
}

func (d duck) Break() bool {
	return d.exit
}

type nest struct{}

func (n *nest) Retrieve(h flowgraph.Hub) (result interface{}, err error) {
	d := duck{duckCnt, 0, h.Name()[0:1], false}
	atomic.AddInt64(&duckCnt, 1)
	return d, nil
}

type swimA struct {
	Count int
}

func (s *swimA) Transform(h flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	d := source[0].(duck)
	if d.Loops == 0 {
		s.Count++
	}
	d.exit = rand.Intn(100) >= 99
	if d.exit {
		s.Count--
	}
	d.Loops++
	result = []interface{}{d}
	return
}

type sink struct{}

func (s *sink) Sink(source []interface{}) {
	fmt.Printf("Duck %+v leaving pond\n", source[0])
}

func TestDuckPondA(t *testing.T) {
	fmt.Printf("BEGIN:  TestDuckPondA\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second * 30
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestDuckPondA")

	newduck := fg.NewStream("newduck")
	fg.NewHub("nest", flowgraph.Retrieve, &nest{}).
		ConnectResults(newduck)

	oldduck := fg.NewStream("oldduck")
	pond := fg.NewGraphHub("pond", flowgraph.While)
	pond.ConnectSources(newduck).ConnectResults(oldduck)

	pond.NewHub("swim", flowgraph.AllOf, &swimA{}).
		SetNumSource(1).SetNumResult(1)
	pond.Loop()

	fg.NewHub("sink", flowgraph.Sink, &sink{}).
		ConnectSources(oldduck)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestDuckPondA\n")
}

/*=====================================================================*/

/* DuckPondB Flowgraph HDL *

nestN()(duckNW)
nestS()(duckNW)
while(duckNW)(duckSE) {
        pond(duckNW)(duckSE)
}
steerSE(duckSE)(duckS, duckE)
sink_s(duckS)()
sink_e(duckE)()

*/

type swimB struct {
	Count int
}

func (s *swimB) Transform(h flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	d := source[0].(duck)
	if d.Loops == 0 {
		s.Count++
	}
	d.exit = rand.Intn(100) >= 99
	if d.exit {
		s.Count--
	}
	d.Loops++
	result = []interface{}{d}
	return
}

type steerSE struct {
}

func (s *steerSE) Transform(h flowgraph.Hub, source []interface{}) (result []interface{}, err error) {
	d := source[0].(duck)
	if d.Loops%2 == 0 {
		result = []interface{}{d, nil}
	} else {
		result = []interface{}{nil, d}
	}
	return
}

func TestDuckPondB(t *testing.T) {
	fmt.Printf("BEGIN:  TestDuckPondB\n")
	oldRunTime := fgbase.RunTime
	oldTraceLevel := fgbase.TraceLevel
	fgbase.RunTime = time.Second * 30
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestDuckPondB")

	duckNW := fg.NewStream("duckNW")
	fg.NewHub("nestN", flowgraph.Retrieve, &nest{}).
		ConnectResults(duckNW)
	fg.NewHub("nestW", flowgraph.Retrieve, &nest{}).
		ConnectResults(duckNW)

	duckSE := fg.NewStream("duckSE")
	pond := fg.NewGraphHub("pond", flowgraph.While)
	pond.ConnectSources(duckNW).ConnectResults(duckSE)

	pond.NewHub("swim", flowgraph.AllOf, &swimB{}).
		SetNumSource(1).SetNumResult(1)
	pond.Loop()

	duckS := fg.NewStream("duckS")
	duckE := fg.NewStream("duckE")
	fg.NewHub("steer", flowgraph.AllOf, &steerSE{}).
		ConnectSources(duckSE).ConnectResults(duckS, duckE)

	fg.NewHub("sinkS", flowgraph.Sink, &sink{}).
		ConnectSources(duckS)
	fg.NewHub("sinkE", flowgraph.Sink, &sink{}).
		ConnectSources(duckE)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestDuckPondB\n")
}
