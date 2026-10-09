package compose

// Ported from eemeli/yaml 2.9.0, dist/nodes/: identity.js, Node.js, Scalar.js, Pair.js, Alias.js,
// Collection.js, YAMLMap.js and YAMLSeq.js, as far as composing reaches them, plus the node classes of
// dist/schema/yaml-1.1/omap.js and set.js, which the core schema's known tags construct.

import "github.com/system-inc/cohere/internal/format/yaml/cst"

// Kind is what upstream's identity checks read off a node's NODE_TYPE symbol.
type Kind string

const (
	KindAlias  Kind = "Alias"
	KindMap    Kind = "Map"
	KindPair   Kind = "Pair"
	KindScalar Kind = "Scalar"
	KindSeq    Kind = "Seq"
)

// Scalar types, upstream's Scalar.BLOCK_FOLDED and friends, with upstream's spelling: they are the
// values yaml-unist-parser switches on.
const (
	BlockFolded  = "BLOCK_FOLDED"
	BlockLiteral = "BLOCK_LITERAL"
	Plain        = "PLAIN"
	QuoteDouble  = "QUOTE_DOUBLE"
	QuoteSingle  = "QUOTE_SINGLE"
)

// Node is any composed node: a Scalar, Alias, YAMLMap, YAMLSeq or Pair, one struct for all of them as in
// the cst and unist packages. Kind says which, and the identity functions below read it the way
// upstream's isScalar and friends read NODE_TYPE. A field exists upstream only on the kinds noted.
//
// JavaScript's absent and empty fields: Tag and Comment and CommentBefore are "" when upstream's are
// undefined or null, because upstream never assigns them an empty string. Anchor can be the empty string
// (`&` alone), so HasAnchor tells it from absent. Range is nil when upstream's is undefined, which only a
// pair key resolvePairs makes from nothing has.
type Node struct {
	Kind Kind
	// Class is the node's constructor name: Scalar, Alias, Pair, YAMLMap, YAMLSeq, or the YAML 1.1
	// collections YAMLOMap (a YAMLSeq) and YAMLSet (a YAMLMap).
	Class string

	// Range is [start, value end, node end] in UTF-16 code units: the end of the value, then the end
	// of the node including any comment on the same line.
	Range []int
	// SrcToken is the CST token the node was composed from (keepSourceTokens): every scalar, alias and
	// collection composed from a token, but not an empty node composed from nothing, nor the map a
	// flow sequence wraps around a pair.
	SrcToken *cst.Token
	// SrcItem is a Pair's srcToken: the collection item it was composed from.
	SrcItem *cst.CollectionItem

	// Anchor is the anchor's name without the &, set by composeNode; HasAnchor says whether it is set.
	Anchor        string
	anchorPresent bool
	// Tag is the resolved tag: a full tag like tag:yaml.org,2002:str, a local tag like !foo, or ! for
	// the non-specific tag.
	Tag string

	Comment       string
	CommentBefore string
	SpaceBefore   bool

	// Flow marks a flow collection, and the map a flow sequence wraps around a pair.
	Flow bool
	// Items is a collection's items: Pairs in a map; Nodes in a sequence, or Pairs once a !!pairs or
	// !!omap tag has resolved it.
	Items []*Node

	// Key and Value are a Pair's. Value is nil for a pair with no value (upstream's null).
	Key   *Node
	Value *Node

	// Source is a Scalar's string value before the tag resolved it, and an Alias's anchor name without
	// the *. A Go string, WTF-8 where the JavaScript string holds a lone surrogate. HasSource says
	// whether it is set: the pair key resolvePairs makes from nothing has none.
	Source    string
	sourceSet bool
	// Type is a Scalar's style: BlockFolded, BlockLiteral, Plain, QuoteDouble, QuoteSingle, or "" for
	// the empty scalar of a block scalar without a header.
	Type string
	// Format is a Scalar's number format from its tag: EXP, OCT, HEX, BIN or TIME.
	Format string
	// MinFractionDigits is set by the float tags for a value written with trailing zeros, 1.50.
	MinFractionDigits int
	// ScalarValue is a Scalar's resolved value: nil (null), bool, float64, string (WTF-8), MergeKey,
	// Date or Binary.
	ScalarValue any
}

// HasSource is whether the node has a source: every alias, and every scalar composeScalar made.
func (node *Node) HasSource() bool { return node.sourceSet }

// HasAnchor is whether composeNode gave the node an anchor, perhaps an empty one.
func (node *Node) HasAnchor() bool { return node.anchorPresent }

// SetAnchor is `node.anchor = name`.
func (node *Node) SetAnchor(name string) {
	node.Anchor = name
	node.anchorPresent = true
}

// MergeKey is the value the merge tag resolves << to, upstream's Symbol('<<'). Each one is a distinct
// symbol upstream, so it is a pointer here, to a type that is not zero-sized so distinct pointers stay
// distinct.
type MergeKey struct{ _ byte }

// Date is the value the timestamp tag resolves to: upstream's Date, as its time value in milliseconds,
// NaN for an invalid date.
type Date float64

// Binary is the value the binary tag resolves to: upstream's Buffer.
type Binary []byte

// newScalar is `new Scalar(value)`.
func newScalar(value any) *Node {
	return &Node{Kind: KindScalar, Class: "Scalar", ScalarValue: value}
}

// newAlias is `new Alias(source)`.
func newAlias(source string) *Node {
	return &Node{Kind: KindAlias, Class: "Alias", Source: source, sourceSet: true}
}

// newScalar, newPair and newCollection on a context take the node from the parse's arena when the
// options carry one (Options.Nodes), and allocate it otherwise (#v6ksqg3).
func (ctx *composeContext) newScalar(value any) *Node {
	return ctx.options.Nodes.New(Node{Kind: KindScalar, Class: "Scalar", ScalarValue: value})
}

func (ctx *composeContext) newPair(key *Node, value *Node) *Node {
	return ctx.options.Nodes.New(Node{Kind: KindPair, Class: "Pair", Key: key, Value: value})
}

func (ctx *composeContext) newCollection(class string) *Node {
	return ctx.options.Nodes.New(collectionNode(class))
}

// newPair is `new Pair(key, value)`.
func newPair(key *Node, value *Node) *Node {
	return &Node{Kind: KindPair, Class: "Pair", Key: key, Value: value}
}

// newCollection is `new NodeClass(schema)` for one of the four collection classes. YAMLOMap's and
// YAMLSet's constructors set the tag to their own; items start as an empty array.
func newCollection(class string) *Node {
	node := collectionNode(class)
	return &node
}

// collectionNode is the node newCollection makes, as a value.
func collectionNode(class string) Node {
	switch class {
	case "YAMLMap":
		return Node{Kind: KindMap, Class: class, Items: []*Node{}}
	case "YAMLSet":
		return Node{Kind: KindMap, Class: class, Items: []*Node{}, Tag: setTagName}
	case "YAMLSeq":
		return Node{Kind: KindSeq, Class: class, Items: []*Node{}}
	case "YAMLOMap":
		return Node{Kind: KindSeq, Class: class, Items: []*Node{}, Tag: omapTagName}
	}
	panic("compose: unknown collection class " + class)
}

// collectionTagName is the static tagName of a collection's class: YAMLSet inherits YAMLMap's and
// YAMLOMap inherits YAMLSeq's.
func collectionTagName(node *Node) string {
	if node.Kind == KindMap {
		return mapTagName
	}
	return seqTagName
}

// assignCollection is `Object.assign(new Class(), collection)`: a new node of the class with every own
// enumerable property of the collection copied over it. Upstream's collection has no anchor or srcToken
// yet at this point, and a property the collection does not have keeps the new node's value (the tag
// the constructor set). The range and items arrays are shared, as Object.assign shares them.
func assignCollection(class string, collection *Node) *Node {
	node := newCollection(class)
	node.Items = collection.Items
	node.Range = collection.Range
	if collection.Tag != "" {
		node.Tag = collection.Tag
	}
	node.Flow = collection.Flow
	node.Comment = collection.Comment
	node.CommentBefore = collection.CommentBefore
	node.SpaceBefore = collection.SpaceBefore
	if collection.anchorPresent {
		node.SetAnchor(collection.Anchor)
	}
	node.SrcToken = collection.SrcToken
	return node
}

// IsAlias is upstream's isAlias.
func IsAlias(node *Node) bool { return node != nil && node.Kind == KindAlias }

// IsMap is upstream's isMap.
func IsMap(node *Node) bool { return node != nil && node.Kind == KindMap }

// IsPair is upstream's isPair.
func IsPair(node *Node) bool { return node != nil && node.Kind == KindPair }

// IsScalar is upstream's isScalar.
func IsScalar(node *Node) bool { return node != nil && node.Kind == KindScalar }

// IsSeq is upstream's isSeq.
func IsSeq(node *Node) bool { return node != nil && node.Kind == KindSeq }

// IsCollection is upstream's isCollection.
func IsCollection(node *Node) bool { return IsMap(node) || IsSeq(node) }

// IsNode is upstream's isNode: anything but a pair.
func IsNode(node *Node) bool { return node != nil && node.Kind != KindPair }

// hasAllNullValues is Collection.hasAllNullValues.
func (node *Node) hasAllNullValues(allowScalar bool) bool {
	for _, item := range node.Items {
		if !IsPair(item) {
			return false
		}
		n := item.Value
		// `n == null || (allowScalar && isScalar(n) && n.value == null && !n.commentBefore &&
		// !n.comment && !n.tag)`
		if !(n == nil || allowScalar && IsScalar(n) && n.ScalarValue == nil && n.CommentBefore == "" &&
			n.Comment == "" && n.Tag == "") {
			return false
		}
	}
	return true
}
