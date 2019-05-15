/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

type BaseNode struct {
	_depth int16
	idx    int
	parent *Node
}

func (n *BaseNode) getDepth() int16 {
	return n._depth
}

func (n *BaseNode) setParentNode(parent *Node) {
	if parent != nil {
		n.parent = parent
		n.idx = parent.getLength()
		n._depth = parent._depth + 1
	}
}

func (n *BaseNode) GetParent() *Node {
	if n.parent != nil {
		return n.parent
	}
	return nil
}
