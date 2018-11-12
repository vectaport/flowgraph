package flowgraph_test

import (
	"errors"
	"fmt"
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
	"io"
	"math/rand"
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

/* TestIterator1 Flowgraph HDL *

tbi()(.X(firstval))
wait(.A(firstval),.B(lastval=0))(.X(oldval))
sub(.A(oldval),.B(1))(.X(newval))
steer(.A(newval))(.X(lastval),.Y(oldval))
sink(.A(lastval))()

*/

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

	steer := fg.NewHub("steer", flowgraph.Steer, 1).
		SetSourceNames("A"). // steer condition.
		SetResultNames("X", "Y")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
		SetSourceNames("A")

	fg.Connect(tbi, "X", wait, "A").SetName("firstval")
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

tbi()(firstval)
wait(firstval,lastval=0)(oldval)
sub(oldval,1)(newval)
steer(newval)(lastval,oldval)
sink(lastval)()

*/

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

	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator2\n")
}

/*=====================================================================*/

/* TestIterator3 Flowgraph HDL *

tbi()(firstval)
while(firstval)(lastval) {
        sub(firstval, 1)(lastval)
}
sink(lastval)()

*/

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

	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
		ConnectResults(firstval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(firstval).
		ConnectResults(lastval)

	oneval := while.NewStream("oneval").Const(1)
	while.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(lastval)

	fg.Run()

	fgbase.RunTime = oldRunTime
	fgbase.TracePorts = oldTracePorts
	fgbase.TraceLevel = oldTraceLevel
	fmt.Printf("END:    TestIterator3\n")
}

/*=====================================================================*/

/* TestIterator4 Flowgraph HDL *

tbi()(firstval1)
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
	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

tbi()(firstval)
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
	fgbase.TraceLevel = fgbase.V

	fg := flowgraph.New("TestIterator5")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

tbi()(firstval)
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
	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

tbi()(firstval)
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
	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

tbi()(firstval)
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
	fg.NewHub("tbi", flowgraph.Retrieve, &tbi{}).
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

/* TestGCD Flowgraph HDL *

tbrand()(mval)
tbrand()(nval)
while(mval, nval)(tcond, gcd) {
        sub(firstval, 1)(lastval)
}
sink(gcd)()

*/

type tbrand struct{}

func (t *tbrand) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	return rand.Intn(100) + 1, nil
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
