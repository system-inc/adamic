package printing

import (
	"fmt"
	"reflect"
)

// AstPath is upstream's AstPath, src/common/ast-path.js: the printer's position in the tree.
//
// The stack alternates names and values exactly as upstream's does: root value, then name, value,
// name, value. A value is a node, a slice of nodes, or whatever a property held; a name is a property
// name, or an index when the value above it is a slice. That shape is load-bearing. siblings, index,
// key, next and previous all read it, and a stack that dropped the slice entries would answer them
// differently.
type AstPath[N Node[N]] struct {
	stack []any

	// Settings is the print's Options.Settings, set by PrintAstToDoc, so a printer function given only the
	// path reaches its language's per-print state. Nothing in this package reads it.
	Settings any
}

// NewAstPath is upstream's `new AstPath(value)`.
func NewAstPath[N Node[N]](root N) *AstPath[N] {
	return &AstPath[N]{stack: []any{root}}
}

// at is JavaScript's Array.prototype.at, with a missing element as nil.
func (path *AstPath[N]) at(index int) any {
	if index < 0 {
		index += len(path.stack)
	}
	if index < 0 || index >= len(path.stack) {
		return nil
	}
	return path.stack[index]
}

// asNode reports whether a stack value is a node.
func asNode[N Node[N]](value any) (N, bool) {
	node, isNode := value.(N)
	return node, isNode
}

// isSlice is upstream's Array.isArray on a stack value.
func isSlice(value any) bool {
	if value == nil {
		return false
	}
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

// sliceElement returns element index of a slice value, and its length.
func sliceElement(value any, index int) (any, int) {
	reflected := reflect.ValueOf(value)
	if index < 0 || index >= reflected.Len() {
		return nil, reflected.Len()
	}
	return reflected.Index(index).Interface(), reflected.Len()
}

func sliceLength(value any) int {
	if !isSlice(value) {
		return 0
	}
	return reflect.ValueOf(value).Len()
}

// Key is upstream's path.key: the property name holding the current node, or the slice holding it.
func (path *AstPath[N]) Key() (string, bool) {
	offset := -2
	if path.Siblings() != nil {
		offset = -4
	}
	name, isName := path.at(offset).(string)
	return name, isName
}

// Index is upstream's path.index: the position in the sibling slice, when there is one.
func (path *AstPath[N]) Index() (int, bool) {
	if path.Siblings() == nil {
		return 0, false
	}
	index, isIndex := path.at(-2).(int)
	return index, isIndex
}

// Stack is upstream's path.stack, for the printers that read it directly (the YAML printer's
// isLastDescendantNode walks it). The slice is the path's own: read it, never modify it.
func (path *AstPath[N]) Stack() []any { return path.stack }

// Value is upstream's path.node and getValue(): the current value, typed as the stack holds it.
func (path *AstPath[N]) Value() any { return path.at(-1) }

// Node is upstream's path.node as a node, and whether the current value is one.
func (path *AstPath[N]) Node() (N, bool) { return asNode[N](path.at(-1)) }

// Parent is upstream's path.parent.
func (path *AstPath[N]) Parent() (N, bool) { return path.GetNode(1) }

// Grandparent is upstream's path.grandparent.
func (path *AstPath[N]) Grandparent() (N, bool) { return path.GetNode(2) }

// IsInArray is upstream's path.isInArray.
func (path *AstPath[N]) IsInArray() bool { return path.Siblings() != nil }

// Siblings is upstream's path.siblings: the slice holding the current value, or nil.
func (path *AstPath[N]) Siblings() any {
	maybe := path.at(-3)
	if isSlice(maybe) {
		return maybe
	}
	return nil
}

// Next is upstream's path.next.
func (path *AstPath[N]) Next() (N, bool) {
	siblings := path.Siblings()
	index, present := path.Index()
	if siblings == nil || !present {
		var zero N
		return zero, false
	}
	element, _ := sliceElement(siblings, index+1)
	return asNode[N](element)
}

// Previous is upstream's path.previous.
func (path *AstPath[N]) Previous() (N, bool) {
	siblings := path.Siblings()
	index, present := path.Index()
	if siblings == nil || !present {
		var zero N
		return zero, false
	}
	element, _ := sliceElement(siblings, index-1)
	return asNode[N](element)
}

// IsFirst is upstream's path.isFirst.
func (path *AstPath[N]) IsFirst() bool {
	index, present := path.Index()
	return present && index == 0
}

// IsLast is upstream's path.isLast.
func (path *AstPath[N]) IsLast() bool {
	siblings := path.Siblings()
	index, present := path.Index()
	return siblings != nil && present && index == sliceLength(siblings)-1
}

// IsRoot is upstream's path.isRoot.
func (path *AstPath[N]) IsRoot() bool { return len(path.stack) == 1 }

// Root is upstream's path.root.
func (path *AstPath[N]) Root() (N, bool) { return asNode[N](path.stack[0]) }

// Ancestors is upstream's path.ancestors: every node above the current one, nearest first.
func (path *AstPath[N]) Ancestors() []N {
	var ancestors []N
	for index := len(path.stack) - 3; index >= 0; index -= 2 {
		if node, isNode := asNode[N](path.stack[index]); isNode {
			ancestors = append(ancestors, node)
		}
	}
	return ancestors
}

// Name is upstream's getName(): the property name or index of the current value.
func (path *AstPath[N]) Name() any {
	if len(path.stack) > 1 {
		return path.at(-2)
	}
	return nil
}

// GetNode is upstream's getNode(count): the count-th node up from the current one.
func (path *AstPath[N]) GetNode(count int) (N, bool) {
	stackIndex := path.nodeStackIndex(count)
	if stackIndex == -1 {
		var zero N
		return zero, false
	}
	return asNode[N](path.stack[stackIndex])
}

// GetParentNode is upstream's getParentNode(count).
func (path *AstPath[N]) GetParentNode(count int) (N, bool) { return path.GetNode(count + 1) }

// nodeStackIndex is upstream's #getNodeStackIndex: walk value positions, skipping slices.
func (path *AstPath[N]) nodeStackIndex(count int) int {
	for index := len(path.stack) - 1; index >= 0; index -= 2 {
		if !isSlice(path.stack[index]) {
			count--
			if count < 0 {
				return index
			}
		}
	}
	return -1
}

// field reads one property or index from a value, upstream's `value?.[name]`.
func field[N Node[N]](value any, name any) any {
	switch key := name.(type) {
	case string:
		if node, isNode := asNode[N](value); isNode {
			return node.Field(key)
		}
		return nil
	case int:
		if isSlice(value) {
			element, _ := sliceElement(value, key)
			return element
		}
		return nil
	default:
		panic(fmt.Sprintf("printing: a path name must be a string or an int, got %T", name))
	}
}

// Call is upstream's path.call(callback, ...names): descend through names, run callback, restore.
//
// Names are property names or slice indexes, in upstream's order. The stack is restored even if the
// callback panics, as upstream's finally restores it.
func Call[N Node[N], R any](path *AstPath[N], callback func(*AstPath[N]) R, names ...any) R {
	length := len(path.stack)
	value := path.at(-1)
	for _, name := range names {
		value = field[N](value, name)
		path.stack = append(path.stack, name, value)
	}
	defer func() { path.stack = path.stack[:length] }()
	return callback(path)
}

// CallParent is upstream's path.callParent(callback, count): run callback with the path at an ancestor.
func CallParent[N Node[N], R any](path *AstPath[N], callback func(*AstPath[N]) R, count int) R {
	stackIndex := path.nodeStackIndex(count + 1)
	parentValues := append([]any(nil), path.stack[stackIndex+1:]...)
	path.stack = path.stack[:stackIndex+1]
	defer func() { path.stack = append(path.stack, parentValues...) }()
	return callback(path)
}

// Each is upstream's path.each(callback, ...names): visit every element of the slice at names.
func (path *AstPath[N]) Each(callback func(path *AstPath[N], index int, value any), names ...any) {
	length := len(path.stack)
	value := path.at(-1)
	for _, name := range names {
		value = field[N](value, name)
		path.stack = append(path.stack, name, value)
	}
	defer func() { path.stack = path.stack[:length] }()

	count := sliceLength(value)
	for index := 0; index < count; index++ {
		element, _ := sliceElement(value, index)
		path.stack = append(path.stack, index, element)
		callback(path, index, value)
		path.stack = path.stack[:len(path.stack)-2]
	}
}

// Map is upstream's path.map(callback, ...names): Each, collecting results by index.
func Map[N Node[N], R any](path *AstPath[N], callback func(path *AstPath[N], index int, value any) R, names ...any) []R {
	var results []R
	path.Each(func(path *AstPath[N], index int, value any) {
		for len(results) <= index {
			var zero R
			results = append(results, zero)
		}
		results[index] = callback(path, index, value)
	}, names...)
	return results
}

// Predicate is one step of Match: the node at that depth, its name, and its index when in a slice.
type Predicate[N Node[N]] func(node any, name any, index int, hasIndex bool) bool

// Match is upstream's path.match(...predicates): test the current node and its ancestors in turn. A nil
// predicate matches anything at its depth, as upstream's undefined does.
func (path *AstPath[N]) Match(predicates ...Predicate[N]) bool {
	pointer := len(path.stack) - 1
	var name any
	var node any = path.stack[pointer]
	pointer--

	pop := func() any {
		if pointer < 0 {
			pointer--
			return nil
		}
		value := path.stack[pointer]
		pointer--
		return value
	}

	for _, predicate := range predicates {
		if node == nil {
			return false
		}
		index, hasIndex := 0, false
		if number, isNumber := name.(int); isNumber {
			index, hasIndex = number, true
			name = pop()
			node = pop()
		}
		if predicate != nil && !predicate(node, name, index, hasIndex) {
			return false
		}
		name = pop()
		node = pop()
	}
	return true
}

// FindAncestor is upstream's path.findAncestor(predicate). It walks the stack as Ancestors does, nearest
// first, without building the list: markdown asks it of every node it prints (#vbjv3d6).
func (path *AstPath[N]) FindAncestor(predicate func(N) bool) (N, bool) {
	for index := len(path.stack) - 3; index >= 0; index -= 2 {
		if node, isNode := asNode[N](path.stack[index]); isNode && predicate(node) {
			return node, true
		}
	}
	var zero N
	return zero, false
}

// HasAncestor is upstream's path.hasAncestor(predicate).
func (path *AstPath[N]) HasAncestor(predicate func(N) bool) bool {
	_, found := path.FindAncestor(predicate)
	return found
}
