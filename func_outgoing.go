package flowgraph

import (
	"github.com/vectaport/fgbase"
)

func outgoingFire(n *fgbase.Node) error {
	a := n.Srcs[0]
	d := n.Aux.(Putter)
	err := d.Put(hub{n}, a.SrcGet())
	if err != nil {
		n.LogError(err.Error())
		return err
	}
	return nil

}

// funcOutgoing accepts one output value from the flowgraph and exports it using a Putter
func funcOutgoing(a fgbase.Edge, putter Putter) fgbase.Node {

	node := fgbase.MakeNode("outgoing", []*fgbase.Edge{&a}, nil, nil, outgoingFire)
	node.Aux = putter
	return node
}
