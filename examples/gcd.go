/* gcd Flowgraph HDL *

rand100()(mval)
rand100()(nval)
while(mval, nval)(tcond, gcd) {
        pass(mval)(gcd)
        mod(nval, mval)(tcond)
}
sink(gcd)()
sink2(tcond)()

*/

package main

import (
	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"

	"math/rand"
	"time"
)

type rand100 struct {
	init bool
}

func (t *rand100) Retrieve(n flowgraph.Hub) (result interface{}, err error) {
	if !(*t).init {
		// (*t).init = true
		return rand.Intn(100) + 1, nil
	}
	return flowgraph.EOF, flowgraph.EOF
}

func main() {
	fgbase.RunTime = time.Second * 1
	fgbase.TraceLevel = fgbase.V
	fgbase.TraceStyle = fgbase.New

	fg := flowgraph.New("gcd")

	mval := fg.NewStream("mval")
	nval := fg.NewStream("nval")
	tcond := fg.NewStream("tcond")
	gcd := fg.NewStream("gcd")

	fg.NewHub("rand100", flowgraph.Retrieve, &rand100{}).
		ConnectResults(mval)
	fg.NewHub("rand100", flowgraph.Retrieve, &rand100{}).
		ConnectResults(nval)

	while := fg.NewGraphHub("while", flowgraph.While)
	while.ConnectSources(mval, nval).
		ConnectResults(tcond, gcd)

	passm := while.NewHub("passm", flowgraph.Pass, nil)
	mod := while.NewHub("mod", flowgraph.Modulo, nil)
	while.ExposeResult(while.Connect(passm, 0, mod, 1))

	while.Loop()

	fg.NewHub("sink", flowgraph.Sink, nil).
		ConnectSources(gcd)
	fg.NewHub("sink2", flowgraph.Sink, nil).
		ConnectSources(tcond)

	fg.Run()

}
