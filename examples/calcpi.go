/* calcpi Flowgraph HDL *

randA()(.X(e0))
randB()(.X(e1))
check(.A(e0),.B(e1))(.X(e2))
pi(.A(e2))()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"math/rand"
	"time"
)

type randOne struct{}

func (r *randOne) Retrieve(hub flowgraph.Hub) (result interface{}, err error) {
	return rand.Float64(), nil
}

type checkSumSquare struct{}

func (e *checkSumSquare) Transform(n flowgraph.Hub, source []interface{}) (
	result []interface{}, err error) {

	result = make([]interface{}, 1)
	a := source[0].(float64)
	b := source[1].(float64)
	x := a*a + b*b
	result[0] = x <= 1.0
	return
}

type piCalc struct {
	cnt int
	sum int
	Pi  float64
}

func (p *piCalc) Transmit(hub flowgraph.Hub, source interface{}) error {
	p.cnt++

	t := source.(bool)
	if t {
		p.sum++
	}

	p.Pi = 4.0 * float64(p.sum) / float64(p.cnt)

	/*
		if p.cnt > 1 {
			fmt.Printf("\b\b\b\b\b\b\b\b")
		}
		fmt.Printf("%8.6f", p.Pi)
	*/

	return nil
}

func main() {
	fgbase.RunTime = time.Second / 10
	fgbase.TraceLevel = fgbase.V
	fgbase.TracePorts = true
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("calcpi")

	randA := fg.NewHub("randA", flowgraph.Retrieve, &randOne{}).
		SetResultNames("X")

	randB := fg.NewHub("randB", flowgraph.Retrieve, &randOne{}).
		SetResultNames("X")

	check := fg.NewHub("check", flowgraph.AllOf, &checkSumSquare{}).
		SetSourceNames("A", "B").
		SetResultNames("X")

	pi := fg.NewHub("pi", flowgraph.Transmit, &piCalc{}).
		SetSourceNames("A")

	fg.Connect(randA, "X", check, "A")
	fg.Connect(randB, "X", check, "B")
	fg.Connect(check, "X", pi, "A")

	fg.Run()
}
