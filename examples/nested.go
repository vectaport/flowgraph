/* nested Flowgraph HDL *

ten()(firstval)
while1(firstval)(lastval) {
        while2(firstval)(lastval) {
                sub(firstval, 1)(lastval)
	}
}
sink(lastval)()

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
	fgbase.TraceLevel = fgbase.VVV
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("TestIterator5")

	firstval1 := fg.NewStream("firstval1")
	fg.NewHub("ten", flowgraph.Retrieve, &ten{}).
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
}
