/* neuron Flowgraph HDL *

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

type weight struct{
        w float64
	d float64
}

func (w *weight) Retrieve(hub flowgraph.Hub) (result interface{}, err error) {
        f := w.w
        m := rand.Intn(3) + 1
	switch m {
	       case 0: f += w.d
	       case 1: f -= w.d
	       case 2:
	}
	if f == w.w {
	        return nil, nil
	}
	w.w = f
	return f, nil
}

type neuron struct{
     bias float64
     weights []float64
     currval float64 
}

func (e *neuron) Transform(n flowgraph.Hub, source []interface{}) (
	result []interface{}, err error) {

	result = make([]interface{}, 1)
	a, aok := source[0].(float64)
	b, bok := source[1].(float64)
	c, cok := source[2].(float64)
	if !aok { a = e.weights[0] } else { e.weights[0] = a }
	if !bok { b = e.weights[1] } else { e.weights[1] = b }
	if !cok { c = e.weights[2] } else { e.weights[2] = c }
	x := (a + b + c)*e.bias
	n.Tracef("AOK %v BOK %v COK %v\n", aok, bok, cok)
	n.Tracef("X %v  CURR %v\n", x, e.currval)
	if x != e.currval {
		result[0] = x
		e.currval = x
	}
	return
}

type result struct {
}

func (p *result) Transmit(hub flowgraph.Hub, source interface{}) error {

	t := source.(float64)
        hub.Tracef("RESULT %v\n", t)

        return nil
}

func main() {
	fgbase.RunTime = time.Second / 1
	fgbase.TraceLevel = fgbase.V
	fgbase.TracePorts = true
	fgbase.TraceStyle = fgbase.New
	fgbase.ConfigByFlag(nil)

	fg := flowgraph.New("neuralnet")

	w0 := fg.NewHub("weight0", flowgraph.Retrieve, &weight{.3, .01}).
		SetResultNames("X")

	w1 := fg.NewHub("weight1", flowgraph.Retrieve, &weight{.5, .01}).
		SetResultNames("X")

	w2 := fg.NewHub("weight2", flowgraph.Retrieve, &weight{.8, .01}).
		SetResultNames("X")

	nA := fg.NewHub("neuronA", flowgraph.OneOf, &neuron{1.0/3.0, make([]float64, 3), 0.0}).
		SetSourceNames("A", "B", "C").
		SetResultNames("X")

	r := fg.NewHub("result", flowgraph.Transmit, &result{}).
		SetSourceNames("A")

	fg.Connect(w0, "X", nA, "A")
	fg.Connect(w1, "X", nA, "B")
	fg.Connect(w2, "X", nA, "C")
	fg.Connect(nA, "X", r, "A")

	fg.Run()
}
