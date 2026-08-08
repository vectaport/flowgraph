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
)

type ten struct{}

func (t *ten) Retrieve(n flowgraph.Hub) (result any, err error) {
	return 10, nil
}

type sinkIterator3 struct {
}

func (st *sinkIterator3) Sink(source []any) {
	if source[0].(int) != 0 {
		fmt.Printf("ERROR Iterator3 FAILED\n")
	}
}

func main() {
	fgbase.TraceStyle = fgbase.New
	fgbase.ConfigByFlag(map[string]any{
		"trace":  "V",
		"sec":    10,
		"trport": true,
	})

	fg := flowgraph.New("loop3")

	firstval := fg.NewPipe("firstval")
	lastval := fg.NewPipe("lastval")

	fg.NewHub("ten", flowgraph.Retrieve, &ten{}).
		ConnectResults(firstval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(firstval).
		ConnectResults(lastval)

	oneval := while.NewPipe("oneval").Const(1)
	while.NewHub("sub", flowgraph.Subtract, nil).
		ConnectSources(nil, oneval)
	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, &sinkIterator3{}).
		ConnectSources(lastval)

	fg.Run()
}
