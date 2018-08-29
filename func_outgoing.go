package flowgraph

import (
	"github.com/vectaport/fgbase"
)

func outgoingFire(n *fgbase.Node) {
	a := n.Srcs[0]
	d := n.Aux.(Putter)
	err := d.Put(pipe{n}, a.SrcGet())
	if err != nil {
		n.LogError(err.Error())
	}

}

// funcOutgoing accepts one output value from the flowgraph and exports it using a Putter
func funcOutgoing(a fgbase.Edge, putter Putter) fgbase.Node {

	node := fgbase.MakeNode("outgoing", []*fgbase.Edge{&a}, nil, nil, outgoingFire)
	node.Aux = putter
	return node
}
