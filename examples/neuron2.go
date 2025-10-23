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

type input struct{
}

func (w *input) Retrieve(hub flowgraph.Hub) (result interface{}, err error) {
        m := float64(rand.Intn(3))
	return m, nil
}

type neuron struct{
     bias float64
     weights []float64
     currval float64
     actfunc func(float64) float64
}

func (e *neuron) Transform(n flowgraph.Hub, source []interface{}) (
	result []interface{}, err error) {

	result = make([]interface{}, 1)
	var sum float64
	for i, v := range source {
	        val, valok := v.(float64)
		n.Tracef("I %v, VAL %v, VALOK %v\n", i, val, valok)
		if !valok {
		        val = e.weights[i]
		} else {
		        e.weights[i] = val
		}
		sum += val
	}

	x := sum * e.bias
	
        if e.actfunc != nil {
	        x = e.actfunc(x)
	}

	if x != e.currval {
		e.currval = x
		result[0] = x
	n.Tracef("CURRVAL %v\n", e.currval)
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
	fgbase.RunTime = time.Second / 10
	fgbase.TraceLevel = fgbase.V
	fgbase.TracePorts = true
	fgbase.TraceStyle = fgbase.New
	fgbase.ConfigByFlag(nil)

	fg := flowgraph.New("neuralnet")

	in0 := fg.NewHub("input0", flowgraph.Retrieve, &input{}).
		SetResultNames("X")

	in1 := fg.NewHub("input1", flowgraph.Retrieve, &input{}).
		SetResultNames("X")

	nA := fg.NewHub("neuronA", flowgraph.OneOf, &neuron{1.0/3.0, make([]float64, 2), 0.0,
	        func(f float64) float64 { if f>0 {return f} else {return 0.0} }}).
		SetSourceNames("A", "B").
		SetResultNames("X")

	nB := fg.NewHub("neuronA", flowgraph.OneOf, &neuron{1.0/3.0, make([]float64, 2), 0.0,
	        func(f float64) float64 { if f>0 {return f} else {return 0.0} }}).
		SetSourceNames("A", "B").
		SetResultNames("X")

	nC := fg.NewHub("neuronA", flowgraph.OneOf, &neuron{1.0/3.0, make([]float64, 2), 0.0,
	        func(f float64) float64 { if f>0 {return f} else {return 0.0} }}).
		SetSourceNames("A", "B").
		SetResultNames("X")

	nD := fg.NewHub("neuronA", flowgraph.OneOf, &neuron{1.0/3.0, make([]float64, 2), 0.0,
	        func(f float64) float64 { if f>0 {return f} else {return 0.0} }}).
		SetSourceNames("A", "B").
		SetResultNames("X")

	r := fg.NewHub("result", flowgraph.Transmit, &result{}).
		SetSourceNames("A")

	fg.Connect(in0, "X", nA, "A")
	fg.Connect(in1, "X", nA, "B")
	fg.Connect(in0, "X", nB, "A")
	fg.Connect(in1, "X", nB, "B")
	fg.Connect(in0, "X", nC, "A")
	fg.Connect(in1, "X", nC, "B")
	fg.Connect(in0, "X", nD, "A")
	fg.Connect(in1, "X", nD, "B")
	fg.Connect(nA, "X", r, "A")
	fg.Connect(nB, "X", r, "A")
	fg.Connect(nC, "X", r, "A")
	fg.Connect(nD, "X", r, "A")

	fg.Run()
}
