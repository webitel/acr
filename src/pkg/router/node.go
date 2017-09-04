/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router


type Node struct {
	BaseNode
	position int
	children []App
}

func (n *Node) getLength() int  {
	return len(n.children)
}

func (n *Node) Add (newNode App)  {
	n.children = append(n.children, newNode)
}

func (n *Node) setFirst ()  {
	n.position = 0
}

func (n *Node) Next() (newNode App)  {
	if n.position < n.getLength() {
		newNode = n.children[n.position]
		n.position++
	}
	return
}

func NewNode(parent *Node) *Node {
	n := &Node{}
	n.setParentNode(parent)

	return n
}