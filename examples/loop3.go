/* loop3 Flowgraph HDL *

ten()(firstval)
while(firstval)(lastval) {
        sub(firstval, 1)(lastval)
}
sink(lastval)()

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

type sinkIterator3 struct {
}

func (st *sinkIterator3) Sink(source []interface{}) {
	if source[0].(int) != 0 {
		fmt.Printf("ERROR Iterator3 FAILED\n")
	}
}

func main() {
	fgbase.RunTime = time.Second * 10
	fgbase.TracePorts = true
	fgbase.TraceLevel = fgbase.V
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("loop3")

	firstval := fg.NewStream("firstval")
	lastval := fg.NewStream("lastval")

	fg.NewHub("ten", flowgraph.Retrieve, &ten{}).
		ConnectResults(firstval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(firstval).
		ConnectResults(lastval)

	oneval := while.NewStream("oneval").Const(1)
	while.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, &sinkIterator3{}).
		ConnectSources(lastval)

	fg.Run()
}
