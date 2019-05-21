/*  simple Flowgraph HDL

const100()(.X(e0))
const1()(.X(e1))
add(.A(e0),.B(e1))(.X(e2))
sink(.A(e2))()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"time"
)

func main() {
	fgbase.RunTime = time.Second / 1000
	fgbase.TraceLevel = fgbase.V
	fgbase.TracePorts = true
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("simple")

	const100 := fg.NewHub("const100", flowgraph.Constant, 100).
		SetResultNames("X")

	const1 := fg.NewHub("const1", flowgraph.Constant, 1).
		SetResultNames("X")

	add := fg.NewHub("add", flowgraph.Add, nil).
		SetSourceNames("A", "B").
		SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
		SetSourceNames("A")

	fg.Connect(const100, "X", add, "A")
	fg.Connect(const1, "X", add, "B")
	fg.Connect(add, "X", sink, "A")

	fg.Run()
}
