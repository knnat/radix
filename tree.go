package radix

import (
	"errors"
	"strings"
	"sync"

	"github.com/gbrlsnchs/color"
)

const (
	// Tsafe activates thread safety.
	Tsafe = 1 << iota
	// Tdebug adds more information to the tree's string representation.
	Tdebug
	// Tnocolor disables colorful output.
	Tnocolor
)

// Tree is a radix tree.
type Tree struct {
	root      *Node
	length    int // total number of nodes
	size      int // total byte size
	safe      bool
	escapeAt  byte // default '@'
	escapeEnd byte // default '*'
	delim     byte // default '/'
	mu        *sync.RWMutex
	bd        *builder
}

// Settings ...
type Settings struct {
	Flags     int
	EscapeAt  byte
	EscapeEnd byte
	Delimiter byte
}

var defaults = &Settings{
	Flags:     0,
	EscapeAt:  '@',
	EscapeEnd: '*',
	Delimiter: '/',
}

var (
	// ErrEscape indicates conflicting escape symbol.
	ErrEscape = errors.New("escape symbols conflict")
)

// New creates a named radix tree with a single node (its root).
func (s *Settings) New() *Tree {
	tr := &Tree{
		root:      &Node{},
		length:    1,
		escapeAt:  s.EscapeAt,
		escapeEnd: s.EscapeEnd,
		delim:     s.Delimiter,
	}
	if s.Flags&Tsafe > 0 {
		tr.mu = &sync.RWMutex{}
		tr.safe = true
	}
	tr.bd = &builder{
		Builder: &strings.Builder{},
		debug:   s.Flags&Tdebug > 0,
	}
	tr.bd.colors[colorRed] = color.New(color.CodeFgRed)
	tr.bd.colors[colorGreen] = color.New(color.CodeFgGreen)
	tr.bd.colors[colorMagenta] = color.New(color.CodeFgMagenta)
	tr.bd.colors[colorBold] = color.New(color.CodeBold)
	for _, c := range tr.bd.colors {
		c.SetDisabled(s.Flags&Tnocolor > 0)
	}
	return tr
}

// New creates a named radix tree with a single node (its root).
func New() *Tree {
	return defaults.New()
}

// Add adds a new node to the tree.
func (tr *Tree) Add(label string, v interface{}) error {
	// No empty strings or interfaces allowed.
	if label == "" || v == nil {
		return nil
	}
	if tr.safe {
		defer tr.mu.Unlock()
		tr.mu.Lock()
	}
	// Sanitize label
	var inEscapeAt bool
	var inEscapeEnd bool
	for i := range label {
		switch label[i] {
		case tr.escapeAt:
			if inEscapeAt || inEscapeEnd {
				return ErrEscape
			}
			inEscapeAt = true
		case tr.escapeEnd:
			if inEscapeAt || inEscapeEnd {
				return ErrEscape
			}
			inEscapeEnd = true
		case tr.delim:
			if inEscapeEnd {
				return ErrEscape
			}
			inEscapeAt = false
		}
	}
	inEscapeAt = false
	inEscapeEnd = false
	tnode := tr.root
	for {
		var next *edge
		var slice string
		for _, edge := range tnode.edges {
			var found int
			slice = edge.label
			for i := range slice {
				if i < len(label) && slice[i] == label[i] {
					switch slice[i] {
					case tr.escapeAt:
						inEscapeAt = true
					case tr.escapeEnd:
						inEscapeAt = true
					case tr.delim:
						inEscapeAt = false
					}
					found++
					continue
				}
				break
			}
			if found > 0 {
				label = label[found:]
				if label[0] == tr.escapeAt {
					inEscapeAt = true
				}
				if label[0] == tr.escapeEnd {
					inEscapeEnd = true
				}
				if label[0] == tr.delim {
					return ErrEscape
				}
				slice = slice[found:]
				next = edge
				break
			}
		}
		if next != nil {
			tnode = next.node
			// Match the whole word.
			if len(label) == 0 {
				// The label is exactly the same as the edge's label,
				// so just replace its node's value.
				//
				// Example:
				// 	(root) -> tnode("tomato", v1)
				// 	becomes
				// 	(root) -> tnode("tomato", v2)
				if len(slice) == 0 {
					tnode.Value = v
					return nil
				}
				// The label is a prefix of the edge's label.
				//
				// Example:
				// 	(root) -> tnode("tomato", v1)
				// 	then add "tom"
				// 	(root) -> ("tom", v2) -> ("ato", v1)
				if inEscapeAt || inEscapeEnd {
					return ErrEscape
				}
				next.label = next.label[:len(next.label)-len(slice)]
				c := tnode.clone()
				tnode.edges = []*edge{
					&edge{
						label: slice,
						node:  c,
					},
				}
				tnode.Value = v
				tr.length++
				return nil
			}
			// Add a new node but break its parent into prefix and
			// the remaining slice as a new edge.
			//
			// Example:
			// 	(root) -> ("tomato", v1)
			// 	then add "tornado"
			// 	(root) -> ("to", nil) -> ("mato", v1)
			// 	                      +> ("rnado", v2)
			if len(slice) > 0 {
				if inEscapeAt || inEscapeEnd {
					return ErrEscape
				}
				c := tnode.clone()
				tnode.edges = []*edge{
					&edge{ // the suffix that is clone into a new node
						label: slice,
						node:  c,
					},
					&edge{ // the new node
						label: label,
						node: &Node{
							Value: v,
							depth: tnode.depth + 1,
						},
					},
				}
				next.label = next.label[:len(next.label)-len(slice)]
				tnode.Value = nil
				tr.length += 2
				tr.size += len(label)
				return nil
			}
			continue
		}
		// Make sure the edge with escape prefixed label is placed
		// on the last edges and check for escape symbol conflict.
		//
		// Example:
		//  (root) -> ("users", v1)
		//         -> ("@uid", v2)
		//  then add ("all", v3)
		//  (root) -> ("users", v1)
		//         -> ("all", v3)
		//         -> ("@uid", v2)
		var e byte
		if l := len(tnode.edges); l > 0 {
			e = tnode.edges[len(tnode.edges)-1].label[0]
		}
		if e == tr.escapeAt || e == tr.escapeEnd {
			if label[0] == tr.escapeAt || label[0] == tr.escapeEnd {
				return ErrEscape
			}
			// Insert new edge before the last edge.
			tnode.edges = append(tnode.edges, &edge{})
			copy(tnode.edges[len(tnode.edges)-1:], tnode.edges[len(tnode.edges)-2:])
			tnode.edges[len(tnode.edges)-2] = &edge{
				label: label,
				node: &Node{
					Value: v,
					depth: tnode.depth + 1,
				},
			}
		} else {
			tnode.edges = append(tnode.edges, &edge{
				label: label,
				node: &Node{
					Value: v,
					depth: tnode.depth + 1,
				},
			})
		}
		tr.length++
		tr.size += len(label)
		return nil
	}
}

// Del deletes a node.
//
// If a parent node that holds no value ends up holding only one edge
// after a deletion of one of its edges, it gets merged with the remaining edge.
func (tr *Tree) Del(label string) {
	if string(label) == "" {
		return
	}
	if tr.safe {
		defer tr.mu.Unlock()
		tr.mu.Lock()
	}
	tnode := tr.root
	var edgex int
	var parent *edge
	for tnode != nil && label != "" {
		var next *edge
		// Look for exact matches.
		for i, e := range tnode.edges {
			if strings.HasPrefix(label, e.label) {
				next = e
				edgex = i
				break
			}
		}
		if next != nil {
			tnode = next.node
			label = label[len(next.label):]
			// While not the exact match, set the tnode's parent.
			if label != "" {
				parent = next
			}
			continue
		}
		// No matches.
		parent = nil
		tnode = nil
	}
	if tnode != nil {
		pnode := tr.root // in case label matched in the first try
		if parent != nil {
			pnode = parent.node
		}
		// Merge tnode's edges with the parent's.
		pnode.edges = append(pnode.edges, tnode.edges...)
		// Remove tnode from the parent, leaving only its edges behind.
		pnode.edges = append(pnode.edges[:edgex], pnode.edges[edgex+1:]...)
		// When only one edge remained in pnode and its value is nil, they can be merged.
		if len(pnode.edges) == 1 && pnode.Value == nil && parent != nil {
			e := pnode.edges[0]
			parent.label += e.label
			pnode.Value = e.node.Value
			pnode.edges = e.node.edges
			tr.length--
		}
		tr.length--
	}
}

// Get retrieves a node.
func (tr *Tree) Get(label string) (*Node, map[string]string) {
	if label == "" {
		return nil, nil
	}
	if tr.safe {
		defer tr.mu.RUnlock()
		tr.mu.RLock()
	}
	tnode := tr.root
	var params map[string]string
	var escapeEnd bool
	for tnode != nil && label != "" {
		var next *edge
	Walk:
		for _, edge := range tnode.edges {
			slice := edge.label
			for {
				phIndex := len(slice)
				// Check if there are any placeholders.
				// If there are none, then use the whole word for comparison.
				// Only one of :
				if i := strings.IndexByte(slice, tr.escapeAt); i >= 0 {
					phIndex = i
				}
				if i := strings.IndexByte(slice, tr.escapeEnd); i >= 0 {
					phIndex = i
					escapeEnd = true
				}
				prefix := slice[:phIndex]
				// If "slice" (until placeholder) is not prefix of
				// "label", then keep walking.
				if !strings.HasPrefix(label, prefix) {
					continue Walk
				}
				label = label[len(prefix):]
				// If "slice" is the whole label,
				// then the match is complete and the algorithm
				// is ready to go to the next edge.
				if len(prefix) == len(slice) {
					next = edge
					break Walk
				}
				// Check whether there is a delimiter.
				// If there isn'tr, then use the whole world as parameter.
				var delimIndex int
				slice = slice[phIndex:]
				if delimIndex = strings.IndexByte(slice[1:], tr.delim) + 1; delimIndex <= 0 {
					delimIndex = len(slice)
				}
				key := slice[1:delimIndex] // remove the placeholder from the map key
				slice = slice[delimIndex:]
				if escapeEnd {
					delimIndex = len(label)
				} else {
					if delimIndex = strings.IndexByte(label[1:], tr.delim) + 1; delimIndex <= 0 {
						delimIndex = len(label)
					}
				}
				if len(key) > 0 {
					if params == nil {
						params = make(map[string]string)
					}
					params[key] = label[:delimIndex]
				}
				label = label[delimIndex:]
				if slice == "" && label == "" {
					next = edge
					break Walk
				}
			}
		}
		if next != nil {
			tnode = next.node
			continue
		}
		tnode = nil
	}
	return tnode, params
}

// Len returns the total numbers of nodes,
// including the tree's root.
func (tr *Tree) Len() int {
	if tr.safe {
		defer tr.mu.RUnlock()
		tr.mu.RLock()
	}
	return tr.length
}

// Size returns the total byte size stored in the tree.
func (tr *Tree) Size() int {
	return tr.size
}

// Sort sorts the tree nodes and its children recursively
// according to their priority lengther.
func (tr *Tree) Sort(st SortingTechnique) {
	if tr.safe {
		defer tr.mu.Unlock()
		tr.mu.Lock()
	}
	tr.root.sort(st)
}

// String returns a string representation of the tree structure.
func (tr *Tree) String() string {
	if tr.safe {
		defer tr.mu.RUnlock()
		tr.mu.RLock()
	}
	bd := tr.bd
	bd.Reset()
	bd.WriteString(bd.colors[colorBold].Wrap("\n."))
	if tr.bd.debug {
		mag := bd.colors[colorMagenta]
		bd.WriteString(mag.Wrapf(" (%d node", tr.length))
		if tr.length != 1 {
			bd.WriteString(mag.Wrap("s")) // avoid writing "1 nodes"
		}
		bd.WriteString(mag.Wrap(")"))
	}
	tr.bd.WriteByte('\n')
	tr.root.writeTo(tr.bd)
	return tr.bd.String()
}
