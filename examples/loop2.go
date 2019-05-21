/* loop2 Flowgraph HDL *

tbten()(firstval)
wait(firstval,lastval=0)(oldval)
sub(oldval,1)(newval)
steer(newval)(lastval,oldval)
sink(lastval)()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"fmt"
	"time"
)

type sinkIterator2 struct {
	i int
}

func (st *sinkIterator2) Sink(source []interface{}) {
	(*st).i = (*st).i - 1
	if (*st).i != source[0].(int) {
		fmt.Printf("ERROR loop2 FAILED\n")
	}
	if (*st).i == 0 {
		(*st).i = 10
	}
}

func main() {
	fgbase.RunTime = time.Second
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("loop2")

	firstval := fg.NewStream("firstval")
	lastval := fg.NewStream("lastval").Init(0)
	oldval := fg.NewStream("oldval")
	oneval := fg.NewStream("oneval").Const(1)
	newval := fg.NewStream("newval")

	fg.NewHub("ten", flowgraph.Constant, 10).
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

	fg.NewHub("sink", flowgraph.Sink, &sinkIterator2{10}).
		ConnectSources(newval)

	fg.Run()

}
