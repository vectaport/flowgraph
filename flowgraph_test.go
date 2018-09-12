package flowgraph_test

import (
	"fmt"
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
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

func (g *getter) Get(node flowgraph.Node) (interface{}, error) {
	i := g.cnt
	g.cnt++
	return i, nil
}

func TestInsertIncoming(t *testing.T) {

	fmt.Printf("BEGIN:  TestInsertIncoming\n")

	fg := flowgraph.New("TestInsertIncoming")
	// n := fg.NewIncoming(&getter{})
	// fg.InsertNode("incoming", n)
	fg.InsertIncoming("incoming", &getter{})
	fg.InsertSink("sink")
	// n := fg.NewSink("sink")
	// fg.InsertNode(n)

	fg.Run()

	fmt.Printf("END:    TestInsertIncoming\n")
}

/*=====================================================================*/

type putter struct {
	sum int
}

func (p *putter) Put(node flowgraph.Node, v interface{}) error {
	p.sum += v.(int)
	return nil
}

func TestInsertOutgoing(t *testing.T) {

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

func (t *transformer) Transform(node flowgraph.Node, v ...interface{}) ([]interface{}, error) {
	xv := v[0].(int) * 2
	return []interface{}{xv}, nil
}

func TestInsertAllOf(t *testing.T) {

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

	fmt.Printf("BEGIN:  TestInsertArray\n")

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fg := flowgraph.New("TestInsertArray")
	fg.InsertArray("array", arr)
	fg.InsertSink("sink")

	fg.Run()

	s := fg.FindNode("sink").Auxiliary().(fgbase.SinkStats)

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

func (p pass) Transform(n flowgraph.Node, c ...interface{}) ([]interface{}, error) {
	return c, nil
}

var p pass

func TestInsertChain(t *testing.T) {
	now := time.Now().UTC()
	fmt.Printf("BEGIN:  TestInsertChain\n")

	arr := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	fg := flowgraph.New("TestInsertChain")
	fg.InsertArray("array", arr)
	fg.InsertAllOf("pass", p)
	for i := 0; i < 1024*1024; i++ {
		fg.InsertAllOf("pass", p)
	}
	fg.InsertSink("sink")

	fmt.Printf("time since:  %v\n", time.Since(now))
	fg.Run()

	s := fg.FindNode("sink").Auxiliary().(fgbase.SinkStats)

	if s.Cnt != len(arr) {
		t.Fatalf("SinkStats.Cnt %d != len(arr)\n", s.Cnt)
	}
	if s.Sum != 45 {
		t.Fatalf("SinkStats.Sum %d != sum(arr)\n", s.Sum)
	}

	fmt.Printf("END:    TestInsertChain\n")
}
