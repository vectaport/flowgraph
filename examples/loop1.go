/* loop1 Flowgraph HDL *

aten()(.X(firstval))
wait(.A(firstval),.B(lastval=0))(.X(oldval))
sub(.A(oldval),.B(1))(.X(newval))
steer(.A(newval))(.X(lastval),.Y(oldval))
sink(.A(lastval))()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"fmt"
	"time"
)

type ten struct{}

func (t *ten) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	return 10, nil
}

type sinkIterator1 struct {
}

func (st *sinkIterator1) Sink(source []interface{}) {
	if source[0].(int) != 0 {
		fmt.Printf("ERROR loop1 FAILED\n")
	}
}

func main() {
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("loop1")

	ten := fg.NewHub("ten", flowgraph.Constant, 10).
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

	sink := fg.NewHub("sink", flowgraph.Sink, &sinkIterator1{}).
		SetSourceNames("A")

	fg.Connect(ten, "X", wait, "A").SetName("firstval")
	fg.ConnectInit(steer, "X", wait, "B", 0).SetName("lastval")

	fg.Connect(wait, "X", sub, "A").SetName("oldval")
	fg.Connect(steer, "Y", sub, "A") // oldval
	fg.Connect(one, "X", sub, "B").SetName("oneval")

	fg.Connect(sub, "X", steer, "A").SetName("newval")

	fg.Connect(steer, "X", sink, "A") // lastval

	fg.Run()
}
