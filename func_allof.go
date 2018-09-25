package flowgraph

import (
	"github.com/vectaport/fgbase"
)

func allOfFire(n *fgbase.Node) error {
	var a []interface{}
	a = make([]interface{}, len(n.Srcs))
	t := n.Aux.(Transformer)
	eofflag := false
	for i, _ := range a {
		a[i] = n.Srcs[i].SrcGet()
		if v, ok := a[i].(error); ok && v.Error() == "EOF" {
			n.Srcs[i].Flow = false
			eofflag = true
		}
	}
	x, _ := t.Transform(&Hub{n}, a)
	for i, _ := range x {
		if eofflag {
			n.Dsts[i].DstPut(EOF)
		} else {
			n.Dsts[i].DstPut(x[i])
		}
	}
	if eofflag {
		return EOF
	} else {
		return nil
	}

}

// funcAllOf waits for all inputs to be ready before transforming them into all outputs
func funcAllOf(a, x []fgbase.Edge, name string, transformer Transformer) fgbase.Node {

	var abuf []*fgbase.Edge
	for i, _ := range a {
		abuf = append(abuf, &a[i])
	}
	var xbuf []*fgbase.Edge
	for i, _ := range x {
		xbuf = append(xbuf, &x[i])
	}
	node := fgbase.MakeNode(name, abuf, xbuf, nil, allOfFire)
	node.Aux = transformer
	return node
}
