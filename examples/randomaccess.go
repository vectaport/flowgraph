/* randomaccess Flowgraph HDL

adder()(.X(e0))
accessor(.A(e0))(.X(e1))
sink(.A(e1))()

One hub grows a shared flowgraph.PropList by adding a new key on every
firing; a second hub, downstream of the first so the two never touch the
PropList at the same instant, reads back a random existing entry and bumps
a per-key "reads" field -- usage stats living right in the PropList, as
GetField/SetField intend.
*/

package main

import (
	"fmt"
	"math/rand"

	"github.com/vectaport/fgbase"
	"github.com/vectaport/flowgraph"
)

const nadd = 20

type adder struct {
	store flowgraph.PropList
}

func (a *adder) Transform(h flowgraph.Hub, source []any) (result []any, err error) {
	key, ok := source[0].(string)
	if !ok {
		return []any{nil}, nil
	}
	a.store.Set(key, key)
	a.store.SetField(key, "reads", 0)
	return []any{key}, nil
}

type accessor struct {
	store flowgraph.PropList
}

func (a *accessor) Transform(h flowgraph.Hub, source []any) (result []any, err error) {
	if _, ok := source[0].(string); !ok {
		return []any{nil}, nil
	}

	n := a.store.Len()
	if n == 0 {
		return []any{nil}, nil
	}

	idx := rand.Intn(n)
	k, v, _ := a.store.At(idx)
	reads, _ := a.store.GetField(k, "reads")
	r := reads.(int) + 1
	a.store.SetField(k, "reads", r)

	h.Tracef("random read: %s=%v (reads=%d, store now holds %d keys)\n", k, v, r, n)
	return []any{k}, nil
}

func main() {
	fgbase.TraceStyle = fgbase.New
	fgbase.ConfigByFlag(map[string]any{"trace": "V"})

	fg := flowgraph.New("randomaccess")

	store := flowgraph.NewPropList()

	names := make([]any, nadd)
	for i := range names {
		names[i] = fmt.Sprintf("a%d", i)
	}

	driver := fg.NewHub("driver", flowgraph.Array, names).
		SetResultNames("X")

	adderHub := fg.NewHub("adder", flowgraph.AllOf, &adder{store}).
		SetSourceNames("A").
		SetResultNames("X")

	accessorHub := fg.NewHub("accessor", flowgraph.AllOf, &accessor{store}).
		SetSourceNames("A").
		SetResultNames("X")

	sink := fg.NewHub("sink", flowgraph.Sink, nil).
		SetSourceNames("A")

	fg.Connect(driver, "X", adderHub, "A")
	fg.Connect(adderHub, "X", accessorHub, "A")
	fg.Connect(accessorHub, "X", sink, "A")

	fg.Run()

	fmt.Printf("\n-- final reads per key (%d keys) --\n", store.Len())
	for _, k := range store.Keys() {
		reads, _ := store.GetField(k, "reads")
		fmt.Printf("  %s: %v reads\n", k, reads)
	}
}
