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

// largestPredecessor is the opposite of smallestSuccessor
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

// Insert inserts the value into the tree with the specified key. It returns the radix node
// it just inserted, r must the root of the radix tree.
func (r *Radix) Insert(key string, value interface{}) *Radix {
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
		return child.Insert(key[prefixEnd:], value)
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
		newChild.Insert(key[prefixEnd:], value)
	} else {
		newChild.Value = value
	}
	return newChild
}

// Find returns the node associated with key,
// r must be the root of the Radix tree, although this is not enforced. If the node is located
// it is returned and exact is set to true. If the node is not found, the immediate predecessor
// is returned and exact is set to false. Is it up to the caller to call Up to get a real
// value. Note that find does return nodes where Value is nil.
func (r *Radix) Find(key string) (node *Radix, exact bool) {
	if key == "" {
		return nil, false
	}
	child, ok := r.children[key[0]]
	if !ok {
		return r, false
	}

	if key == child.key {
		return child, true
	}

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)

	// if child.key is not completely contained in key, abort [e.g. trying to find "ab" in "abc"]
	if child.key != commonPrefix {
		return r, false
	}

	// find the key left of key in child
	return child.Find(key[prefixEnd:])
}

// Next returns the next node in the tree. For non-leaf nodes this is the left most
// child node. For leaf nodes this is the first neighbor to the right. If no such
// neighbor is found, it's the first existing neighbor of a parent. This finally
// terminates the root of the tree. Next does not return nodes with Value is nil,
// so the caller is guaranteed to get a node with data, unless we hit the root node.
func (r *Radix) Next() *Radix {
	switch len(r.children) {
	case 0: // leaf-node, 
		// Look in my parent to get a list of my peers
		// r.parent is never nil?
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
// its parent is tried. This finishes at root, at which point nil 
// is returned.
func (r *Radix) next() *Radix {
	if r.parent == nil {
		return nil
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

// Predecessor returns the node who's key is the largest, but always smaller than t
// If nothing is found nil is returned. If Predecessor returns a node, its
// value will be non-nil.
func (r *Radix) Predecessor(key string) *Radix {
       child, ok := r.children[key[0]]
       if !ok {
               for r.Value == nil {
                       if r.parent == nil {
                               return nil // Root node
                       }
                       r = r.parent
               }
               return r
       }
       // Ok, we found the node... 
       if key == child.key {
               for r.Value == nil {
                       if r.parent == nil {
                               return nil // Root node
                       }
                       r = r.parent
               }
               return r
       }

       commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)

       // if child.key is not completely contained in key, return the parent
       if child.key != commonPrefix {
               for r.Value == nil {
                       if r.parent == nil {
                               return nil // Root node
                       }
                       r = r.parent
               }
               return r
       }
       // find the key left of key in child
       return child.Predecessor(key[prefixEnd:])
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
			// remove child from current node if child has no children on its own
			delete(r.children, key[0])
		case 1:
			// since len(child.children) == 1, there is only one subchild; we have to use range to get the value, though, since we do not know the key
			for _, subchild := range child.children {
				// essentially moves the subchild up one level to replace the child we want to delete, while keeping the key of child
				child.key = child.key + subchild.key
				child.Value = subchild.Value
				child.children = subchild.children
				child.parent = r
			}
		default:
			// if there are >= 2 subchilds, we can only set the value to nil, thus delete any value set to key
			child.Value = nil
		}
		return child
	}

	// Node has not been foundJ, key != child.keys

	commonPrefix, prefixEnd := longestCommonPrefix(key, child.key)
	// if child.key is not completely contained in key, abort [e.g. trying to delete "ab" from "abc"]
	if child.key != commonPrefix {
		return nil
	}
	// else: cut off common prefix and delete left string in child
	return child.Remove(key[prefixEnd:])
}

// Do calls function f on each node with Value != nil in the tree. f's parameter will be r.Value. The behavior of Do is              
// undefined if f changes r.                                                       
func (r *Radix) Do(f func(interface{})) {
	if r != nil {
		f(r.Value)
		for _, child := range r.children {
			child.Do(f)
		}
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

// Keys return all the keys from the node r and downwards
func (r *Radix) Keys() (s []string) {
	// get the full key for this node and use that to get all the other keys
	fullkey := r.key
	for p := r.parent; p != nil; p = p.parent {
		fullkey = p.key + fullkey
	}
	return r.keys(fullkey)
}

func (r *Radix) keys(fullkey string) (s []string) {
	if fullkey != "" { // root
		s = append(s, fullkey)
	}
	for _, c := range r.children {
		s = append(s, c.keys(fullkey+c.key)...)
	}
	return s
}
