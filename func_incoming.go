package flowgraph

import (
	"github.com/vectaport/fgbase"

	"io"
)

func incomingFire(n *fgbase.Node) error {
	x := n.Dsts[0]
	r := n.Aux.(Getter)
	v, err := r.Get(node{n})
	if err != nil {
		if err != io.EOF {
			n.LogError(err.Error())
		}
		return err
	}
	x.DstPut(v)
	return nil
}

// funcIncoming imports one input value using a Getter and feeds it to the flowgraph
func funcIncoming(x fgbase.Edge, receiver Getter) fgbase.Node {

	node := fgbase.MakeNode("incoming", nil, []*fgbase.Edge{&x}, nil, incomingFire)
	node.Aux = receiver
	return node
}
