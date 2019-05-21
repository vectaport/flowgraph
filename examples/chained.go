/* chained Flowgraph HDL *

ten()(firstval1)
while1(firstval1)(lastval1) {
        sub1(firstval1, 1)(lastval1)
}
wait(10,lastval1)(firstval2)
while2(firstval2)(lastval2) {
        sub2(firstval2, 1)(lastval2)
}
sink(lastval2)()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"time"
)

type ten struct{}

func (t *ten) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	return 10, nil
}

func main() {
	fgbase.RunTime = time.Second
	fgbase.TraceLevel = fgbase.V
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("chained")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("ten", flowgraph.Retrieve, &ten{}).
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
}
