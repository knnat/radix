package radix

import (
	"sort"
)

// Node is a node of a radix tree.
type Node struct {
	Value    interface{}
	edges    []*edge
	priority int
	depth    int
}

// Depth returns the node's depth.
func (n *Node) Depth() int {
	return n.depth
}

// IsLeaf returns whether the node is a leaf.
func (n *Node) IsLeaf() bool {
	length := len(n.edges)
	return length == 0
}

// Priority returns the node's priority.
func (n *Node) Priority() int {
	return n.priority
}

func (n *Node) clone() *Node {
	c := *n
	c.incrDepth()
	return &c
}

func (n *Node) incrDepth() {
	n.depth++
	for _, e := range n.edges {
		e.n.incrDepth()
	}
}

// sort sorts the node and its children recursively.
func (n *Node) sort(st SortingTechnique) {
	s := &sorter{
		n:  n,
		st: st,
	}
	sort.Sort(s)
	for _, e := range n.edges {
		e.n.sort(st)
	}
}

func (n *Node) writeTo(bd *builder) {
	for i, e := range n.edges {
		e.writeTo(bd, []bool{i == len(n.edges)-1})
	}
}
