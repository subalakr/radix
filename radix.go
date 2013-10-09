// Package radix implements a radix tree.
//
// A radix tree is defined in:
//    Donald R. Morrison. "PATRICIA -- practical algorithm to retrieve
//    information coded in alphanumeric". Journal of the ACM, 15(4):514-534,
//    October 1968
//
// Also see http://en.wikipedia.org/wiki/Radix_tree for more information.
//
package radix

import (
	"io"
)

var MaxKeySize int

func readKey(r io.Reader) string {
	if MaxKeySize == 0 {
		MaxKeySize = 512
	}

	b := make([]byte, MaxKeySize)

	if n, err := r.Read(b); err == nil && n > 0 {
		return string(b)
	}
	return ""
}

// longestCommonPrefix returns the longest prefiex key and bar have
// in common.
func longestCommonPrefix(key, bar string) (string, int) {
	if key == "" || bar == "" {
		return "", 0
	}
	x := 0
	for key[x] == bar[x] {
		x = x + 1
		if x == len(key) || x == len(bar) {
			break
		}
	}
	return key[:x], x // == bar[:x]
}

// smallestSuccessor walks the keys of the map and returns the smallest
// successor for key and true. Or if key is the largest key, it will return
// false, the value of successor isn't specified in that case.
// We need this function because a map isn't sorted and for the Next() function
// we *do* need to sort this.
func smallestSuccessor(m map[byte]*Radix, key byte) (successor byte, found bool) {
	guard := 256
	for k, _ := range m {
		if k > key && int(k) < guard {
			guard = int(k)
			successor = k
			found = true
		}
	}
	return
}

// leftMostChild returns the smallest child of the current node.
func leftMostChild(m map[byte]*Radix) (left byte) {
	left = 255
	for k, _ := range m {
		if k < left {
			left = k
		}
	}
	return
}

// largestPredecessor is the opposite of smallestSuccessor.
func largestPredecessor(m map[byte]*Radix, key byte) (pred byte, found bool) {
	guard := -1
	for k, _ := range m {
		if k < key && int(k) > guard {
			guard = int(k)
			pred = k
			found = true
		}
	}
	return
}

// rightMostChild returns the largest child of the current node.
func rightMostChild(m map[byte]*Radix) (right byte) {
	right = 0
	for k, _ := range m {
		if k > right {
			right = k
		}
	}
	return
}

// Radix represents a radix tree.
type Radix struct {
	// children maps the first letter of each child to the child.
	children map[byte]*Radix
	key      string
	parent   *Radix // a pointer back to the parent

	// The contents of the radix node.
	Value interface{}
}

// New returns an initialized radix tree.
func New() *Radix {
	return &Radix{make(map[byte]*Radix), "", nil, nil}
}

func (r *Radix) String() string {
	return r.stringHelper("")
}

func (r *Radix) stringHelper(indent string) (s string) {
	s = indent + r.Key() + " '" + r.key + "'" + ":"
	if r.Value == nil {
		s = indent + "<nil>:"
	}
	for i, _ := range r.children {
		s += string(i)
	}
	s += "\n"
	for i, r1 := range r.children {
		s += indent + string(i) + ":" + r1.stringHelper("  "+indent)
	}
	return s
}

// Key returns the full (from r down to this node) key under which r is stored.
func (r *Radix) Key() (s string) {
	for p := r; p != nil; p = p.parent {
		s = p.key + s
	}
	return
}

// Up returns the first node above r which has a non-nil Value.
// It terminates at the root and returns nil if that happens.
func (r *Radix) Up() *Radix {
	if r.parent == nil {
		return nil
	}
	for r = r.parent; r != nil && r.Value == nil; r = r.parent {
		// ...
	}
	return r
}

func (r *Radix) Insert(reader io.Reader, value interface{}) *Radix {
	key := readKey(reader)
	if key != "" {
        return nil
	}
    return r.insert(key, value)
}

// Insert inserts the value into the tree with the specified key. It returns the radix node
// it just inserted, r must the root of the radix tree.
func (r *Radix) insert(key string, value interface{}) *Radix {

	// look up the child starting with the same letter as key
	// if there is no child with the same starting letter, insert a new one
	child, ok := r.children[key[0]]
	if !ok {
		r.children[key[0]] = &Radix{make(map[byte]*Radix), key, r, value}
		return r.children[key[0]]
	}

	if key == child.key {
		child.Value = value
		return child
	}

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)

	if commonPrefix == child.key {
		return child.insert(key[prefixEnd:], value)
	}

	// create new child node to replace current child
	newChild := &Radix{make(map[byte]*Radix), commonPrefix, r, nil}

	// replace child of current node with new child: map first letter of common prefix to new child
	r.children[commonPrefix[0]] = newChild

	// shorten old key to the non-shared part
	child.key = child.key[prefixEnd:]

	// map old child's new first letter to old child as a child of the new child
	newChild.children[child.key[0]] = child
	child.parent = newChild

	// if there are key left of key, insert them into our new child
	if key != newChild.key {
		newChild.insert(key[prefixEnd:], value)
	} else {
		newChild.Value = value
	}
	return newChild
}

// Find returns the node associated with key,
// r must be the root of the Radix tree, although this is not enforced. If the node is located
// it is returned and exact is set to true. If the node found has a nil Value, Find will go
// up in the tree to look for a non-nil Value. If this happens exact is set to false.
// Also if the node is not found, the immediate predecessor
// is returned and exact is set to false. If this node also has a nil Value the same thing
// happens: the tree is search upwards, until the first non-nil Value node is found.

func (r *Radix) Find(reader io.Reader) (node *Radix, exact bool) {
	key := readKey(reader)

    return r.find(key) 
}

func (r *Radix) find(key string) (node *Radix, exact bool) {

	if key == "" {
		return nil, false
	}
	child, ok := r.children[key[0]]
	if !ok {
		if r.Value != nil {
			return r, false
		}
		for r.Value == nil {
			if r.parent == nil {
				return nil, false // Root
			}
			r = r.parent
		}
		return r, false
	}

	if key == child.key {
		if child.Value != nil {
			return child, true
		}
		r := child
		for r.Value == nil {
			if r.parent == nil {
				return nil, false // Root
			}
			r = r.parent
		}
		return r, false
	}

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)

	// if child.key is not completely contained in key, abort [e.g. trying to find "ab" in "abc"]
	if child.key != commonPrefix {
		if r.Value != nil {
			return r, false
		}
		for r.Value == nil {
			if r.parent == nil {
				return nil, false
			}
			r = r.parent
		}
		return r, false
	}

	// find the key left of key in child
	return child.find(key[prefixEnd:])
}

// FindFunc works just like Find, but each non-nil Value of each node traversed during
// the search is given to the function f. Is this function returns true, that node is returned
// and the search stops, exact is set to false and funcfound to true. If during the search f does
// not return true FindFunc behaves just as Find.
func (r *Radix) FindFunc(reader io.Reader, f func(interface{}) bool) (node *Radix, exact bool, funcfound bool) {
    key := readKey(reader)
    return r.findFunc(key, f)
}

func (r *Radix) findFunc(key string, f func(interface{}) bool) (node *Radix, exact bool, funcfound bool) {

	if key == "" {
		return nil, false, false
	}
	if r.Value != nil && f(r.Value) {
		return r, false, true
	}

	child, ok := r.children[key[0]]
	if !ok {
		if r.Value != nil {
			return r, false, false
		}
		for r.Value == nil {
			if r.parent == nil {
				return nil, false, false // Root
			}
			r = r.parent
		}
		return r, false, false
	}

	if key == child.key {
		if child.Value != nil {
			return child, true, false
		}
		r := child
		for r.Value == nil {
			if r.parent == nil {
				return nil, false, false // Root
			}
			r = r.parent
		}
		return r, false, false
	}

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)

	// if child.key is not completely contained in key, abort [e.g. trying to find "ab" in "abc"]
	if child.key != commonPrefix {
		if r.Value != nil {
			return r, false, false
		}
		for r.Value == nil {
			if r.parent == nil {
				return nil, false, false
			}
			r = r.parent
		}
		return r, false, false
	}

	// find the key left of key in child
    return child.findFunc(key[prefixEnd:], f)
}

// Next returns the next node in the tree. For non-leaf nodes this is the left most
// child node. For leaf nodes this is the first neighbor to the right. If no such
// neighbor is found, it's the first existing neighbor of a parent. This finally
// terminates the root of the tree. Next does not return nodes with Value is nil,
// so the caller is guaranteed to get a node with data.
func (r *Radix) Next() *Radix {
	if len(r.key) == 0 {
		return nil // Empty tree
	}
	if r.parent == nil {
		// The root node should have one child, which is the
		// apex of the zone, return that
		for _, x := range r.children { // only one
			return x
		}
	}
	switch len(r.children) {
	case 0: // leaf-node
		// Look in my parent to get a list of my peers
		neighbor, found := smallestSuccessor(r.parent.children, r.key[0])
		if found {
			ret := r.parent.children[neighbor]
			for ret.Value == nil {
				ret = ret.children[leftMostChild(ret.children)]
			}
			return ret
		}
		// There are no neighbors left, loop up
		return r.next()
	default: // non-leaf node
		// Skip <nil> value nodes, because those have no data
		ret := r.children[leftMostChild(r.children)]
		for ret.Value == nil {
			ret = ret.children[leftMostChild(ret.children)]
		}
		return ret
	}
	panic("dns: not reached")
}

// next goes up in the tree to look for nodes with a neighbor.
// if found that neighbor is returned. If a parent has no neighbor
// its parent is tried. This finishes at first non-nil Value node
// in the tree: the shortest key added.
func (r *Radix) next() *Radix {
	if r.parent == nil {
		// The root node should have one child, which is the
		// apex of the zone, return that
		for _, x := range r.children { // only one
			return x
		}
	}
	neighbor, found := smallestSuccessor(r.parent.children, r.key[0])
	if found {
		ret := r.parent.children[neighbor]
		if ret.Value == nil {
			ret = ret.children[leftMostChild(ret.children)]
		}
		return ret
	}
	return r.parent.next()
}

// Prev returns the previous node in the tree, it is the opposite of Next.
// The following holds true: r.Next().Prev().Key() = r.Key()
func (r *Radix) Prev() *Radix {
	if len(r.key) == 0 {
		return nil // Empty tree
	}
	if r.parent == nil {
		// The root node should have one child, which is the
		// apex of the zone, return that
		for _, x := range r.children { // only one
			return x
		}
	}
	neighbor, found := largestPredecessor(r.parent.children, r.key[0])
	if found {
		ret := r.parent.children[neighbor]
		return ret.prev()
	}
	// leaf-node, but no left neighbor, go up...
	r = r.parent
	for r.Value == nil {
		if r.parent == nil {
			// return largest right leaf node
			for len(r.children) != 0 {
				r = r.children[rightMostChild(r.children)]
			}
			return r
		}
		r = r.parent
	}
	return r
}

// prev does down in the tree and selected the right most child until a leaf
// node is hit.
func (r *Radix) prev() *Radix {
	if len(r.children) == 0 {
		return r
	}
	r = r.children[rightMostChild(r.children)]
	return r.prev()
}

// Remove removes any value set to key. It returns the removed node or nil if the
// node cannot be found.
func (r *Radix) Remove(key string) *Radix {
	child, ok := r.children[key[0]]
	if !ok {
		return nil
	}

	// if the correct end node is found...
	if key == child.key {
		switch len(child.children) {
		case 0:
			delete(r.children, key[0])
		case 1:
			for _, subchild := range child.children {
				// essentially moves the subchild up one level to replace the child we want to delete, while keeping the key of child
				child.key = child.key + subchild.key
				child.Value = subchild.Value
				child.children = subchild.children
				child.parent = r
			}
		default:
			child.Value = nil
		}
		return child
	}

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)
	if child.key != commonPrefix {
		return nil
	}
	return child.Remove(key[prefixEnd:])
}

// Do traverses the tree r in an unordered fashion and calls function f on each (non-nil) node,
// f's parameter is r.Value.
func (r *Radix) Do(f func(interface{})) {
	if r == nil {
		return
	}
	if r.Value != nil {
		f(r.Value)
	}
	for _, child := range r.children {
		child.Do(f)
	}
}

// NextDo traverses the tree r in Next-order and calls function f on each node,
// f's parameter is be r.Value.
func (r *Radix) NextDo(f func(interface{})) {
	if r == nil || len(r.children) == 0 {
		return
	}
	if r.parent == nil {
		// root of the tree descend to the first node
		for _, x := range r.children { // only one
			r = x
		}

	}
	k := r.Key()
	f(r.Value)
	r = r.Next()
	for r.Key() != k {
		f(r.Value)
		r = r.Next()
	}
}

// PrevDo traverses the tree r in Prev-order and calls function f on each node,
// f's parameter is be r.Value.
func (r *Radix) PrevDo(f func(interface{})) {
	if r == nil || len(r.children) == 0 {
		return
	}
	if r.parent == nil {
		// root of the tree descend to the first node
		for _, x := range r.children { // only one
			r = x
		}
	}
	k := r.Key()
	f(r.Value)
	r = r.Prev()
	for r.Key() != k {
		f(r.Value)
		r = r.Prev()
	}
}

// Len computes the number of nodes in the radix tree r.
func (r *Radix) Len() int {
	i := 0
	if r != nil {
		if r.Value != nil {
			i++
		}
		for _, child := range r.children {
			i += child.Len()
		}
	}
	return i
}
