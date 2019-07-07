package radix

import "bytes"

const tabSize = 4

type edge struct {
	label string
	node  *Node
}

func (e *edge) writeTo(bd *builder, tabList []bool) {
	length := len(tabList)
	isLast, tlist := tabList[length-1], tabList[:length-1]
	for _, hasTab := range tlist {
		if hasTab {
			bd.Write(bytes.Repeat([]byte(" "), tabSize))
			continue
		}
		bd.WriteRune('│')
		bd.Write(bytes.Repeat([]byte(" "), tabSize-1))
	}
	if !isLast {
		bd.WriteRune('├')
	} else {
		bd.WriteRune('└')
	}
	bd.WriteString("── ")
	bd.WriteString(bd.colors[colorBold].Wrap(e.label))
	if bd.debug {
		if e.node.IsLeaf() {
			bd.WriteString(bd.colors[colorGreen].Wrap(" 🍂"))
		}
		bd.WriteString(bd.colors[colorMagenta].Wrapf(" → %#v", e.node.Value))
	}
	bd.WriteByte('\n')
	for i, next := range e.node.edges {
		if len(tabList) < next.node.depth { // runs only for the first edge
			tabList = append(tabList, i == len(e.node.edges)-1)
		} else {
			tabList[next.node.depth-1] = i == len(e.node.edges)-1
		}
		next.writeTo(bd, tabList)
	}
}
