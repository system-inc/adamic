// Package estree is the ESTree tree Prettier's JavaScript printer reads, built from typescript-go's AST.
//
// Prettier formats TypeScript through a pipeline of three stages: the TypeScript parser, then
// typescript-estree's converter, then the printer. cohere already holds the first stage's output, so
// this package is the second stage ported to Go (convert.go follows typescript-estree's convert.js
// case for case) plus Prettier's own parse postprocess, and the printer reads what it produces.
//
// # Why the node is dynamic
//
// Prettier's printer reaches into nodes by property name, constantly and generically: path.call(print,
// "callee"), a visitor-key walk for comment attachment, `node[key]` comparisons between siblings. A
// struct per ESTree type would make every one of those a switch. So a Node is a type name, a range,
// and an ordered list of named properties, and the printer port reads like its source:
//
//	node.callee.type === "Identifier"     node.Child("callee").Is("Identifier")
//
// Every accessor is nil-safe on a nil *Node, because the JavaScript it ports leans on optional
// chaining everywhere and a nil dereference would turn `undefined` into a crash.
//
// # Offsets are bytes
//
// Ranges are byte offsets into the UTF-8 source, where Prettier's are UTF-16 indexes. The difference
// is invisible as long as every text operation agrees on the unit, and here they all do: slicing,
// searching for newlines, and skipping whitespace all walk bytes, decoding a rune only where a
// multi-byte character can be whitespace or a line terminator. Width is measured on the text itself,
// never derived from offsets.
package estree

import "github.com/system-inc/cohere/internal/format/printing"

// Node is one ESTree node, or one comment: comments are {type: "Line" | "Block", value} objects in
// Prettier and travel through the same paths, so they are Nodes too.
//
// Comment bookkeeping (node.comments, comment.leading, comment.printed) is the shared print core's
// CommentFields, embedded so the core reads and writes it the same way for every language. The core
// keys its caches on node identity, which is why a Node is always used through a pointer.
type Node struct {
	nodeType string
	Range    [2]int

	properties []property

	printing.CommentFields[*Node]

	// Parenthesized is Prettier's node.extra.parenthesized, set when the parser saw redundant parentheses.
	Parenthesized bool

	// ContentEnd is Prettier's node.__contentEnd, the end before a trailing semicolon, when HasContentEnd.
	ContentEnd    int
	HasContentEnd bool
}

type property struct {
	key   string
	value any
}

// TemplateValue is a TemplateElement's value: Cooked is nil where the escape is invalid in a tagged
// template, the way upstream's cooked is null.
type TemplateValue struct {
	Cooked *string
	Raw    string
}

// Regex is a regular expression literal's regex property.
type Regex struct {
	Pattern string
	Flags   string
}

// New makes a node with its properties in the order given, key then value.
//
// The properties are allocated at their final length, rather than grown one Set at a time: growing them
// was 3.5M of the conversion's 11.9M allocations on ahra's 3,926 program files (#k8cht0w).
func New(nodeType string, start int, end int, keysAndValues ...any) *Node {
	node := &Node{nodeType: nodeType, Range: [2]int{start, end}, properties: make([]property, 0, len(keysAndValues)/2)}
	for index := 0; index+1 < len(keysAndValues); index += 2 {
		node.Set(keysAndValues[index].(string), keysAndValues[index+1])
	}
	return node
}

// Is reports whether the node exists and has one of the types.
func (node *Node) Is(types ...string) bool {
	if node == nil {
		return false
	}
	for _, nodeType := range types {
		if node.nodeType == nodeType {
			return true
		}
	}
	return false
}

// Type is node.type, or "" for a missing node.
func (node *Node) Type() string {
	if node == nil {
		return ""
	}
	return node.nodeType
}

// SetType replaces node.type, which Prettier's postprocess and printers do in place.
func (node *Node) SetType(nodeType string) { node.nodeType = nodeType }

// Get is node[key]: the property's value, or nil when it is absent.
func (node *Node) Get(key string) any {
	if node == nil {
		return nil
	}
	for index := range node.properties {
		if node.properties[index].key == key {
			return node.properties[index].value
		}
	}
	return nil
}

// Field is node[name] for the shared print core. "comments" is the attached comments, which live in
// CommentFields rather than among the properties.
func (node *Node) Field(name string) any {
	if name == "comments" {
		if node == nil || len(node.Comments) == 0 {
			return nil
		}
		return node.Comments
	}
	return node.Get(name)
}

// CommentData is the shared print core's access to the comment bookkeeping.
func (node *Node) CommentData() *printing.CommentFields[*Node] { return &node.CommentFields }

// Has is `key in node`: whether the property exists, even with a nil value.
func (node *Node) Has(key string) bool {
	if node == nil {
		return false
	}
	for index := range node.properties {
		if node.properties[index].key == key {
			return true
		}
	}
	return false
}

// Set writes node[key], appending a new property at the end the way JavaScript does.
func (node *Node) Set(key string, value any) {
	// A typed nil stored in an interface is not == nil, which would make a missing child look present.
	switch typed := value.(type) {
	case *Node:
		if typed == nil {
			value = nil
		}
	case *TemplateValue:
		if typed == nil {
			value = nil
		}
	}
	for index := range node.properties {
		if node.properties[index].key == key {
			node.properties[index].value = value
			return
		}
	}
	node.properties = append(node.properties, property{key: key, value: value})
}

// Delete removes node[key].
func (node *Node) Delete(key string) {
	for index := range node.properties {
		if node.properties[index].key == key {
			node.properties = append(node.properties[:index], node.properties[index+1:]...)
			return
		}
	}
}

// Keys are the property names in insertion order.
func (node *Node) Keys() []string {
	if node == nil {
		return nil
	}
	keys := make([]string, len(node.properties))
	for index, property := range node.properties {
		keys[index] = property.key
	}
	return keys
}

// Child is node[key] when it is a node.
func (node *Node) Child(key string) *Node {
	child, _ := node.Get(key).(*Node)
	return child
}

// List is node[key] when it is an array of nodes. Elements may be nil, as array holes are.
func (node *Node) List(key string) []*Node {
	list, _ := node.Get(key).([]*Node)
	return list
}

// String is node[key] when it is a string.
func (node *Node) String(key string) string {
	text, _ := node.Get(key).(string)
	return text
}

// Bool is node[key] when it is a boolean, JavaScript's `node.key === true`.
func (node *Node) Bool(key string) bool {
	value, _ := node.Get(key).(bool)
	return value
}

// Truthy is JavaScript's truthiness of node[key].
func (node *Node) Truthy(key string) bool {
	return IsTruthy(node.Get(key))
}

// IsTruthy is JavaScript's truthiness for the values a node holds.
func IsTruthy(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case string:
		return typed != ""
	case float64:
		return typed != 0 && typed == typed
	case int:
		return typed != 0
	case *Node:
		return typed != nil
	default:
		return true
	}
}

// Start is the node's range start, before Prettier's decorator adjustment (see LocStart).
func (node *Node) Start() int { return node.Range[0] }

// End is the node's range end.
func (node *Node) End() int { return node.Range[1] }

// VisitorKeys is Prettier's getVisitorKeys(node): the properties that hold child nodes, in order.
func VisitorKeys(node *Node) []string {
	return visitorKeys[node.nodeType]
}

// HasVisitorKeys reports whether Prettier knows the type at all.
func HasVisitorKeys(nodeType string) bool {
	_, known := visitorKeys[nodeType]
	return known
}

// ChildNodes are the node's children in visitor-key order, flattening arrays and skipping holes.
func ChildNodes(node *Node) []*Node {
	var children []*Node
	for _, key := range VisitorKeys(node) {
		switch value := node.Get(key).(type) {
		case *Node:
			children = append(children, value)
		case []*Node:
			for _, child := range value {
				if child != nil {
					children = append(children, child)
				}
			}
		}
	}
	return children
}
