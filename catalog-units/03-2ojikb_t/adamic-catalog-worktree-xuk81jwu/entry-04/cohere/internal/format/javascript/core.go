// Package javascript is Prettier's JavaScript and TypeScript printer, src/language-js in the fork,
// ported to Go over the ESTree tree internal/format/estree builds.
//
// # Reading this package against upstream
//
// Each file ports one upstream file and is named for it: print_call_arguments.go is
// print/call-arguments.js, utility_comments.go is utilities/comments.js, parentheses.go is
// parentheses/needs-parentheses.js. Functions keep upstream's names. The JavaScript idioms translate
// one way throughout, so the Go reads line for line against its source:
//
//	path.node, path.parent, path.key       node(path), parentOf(path), keyOf(path)
//	path.call(print, "body")               print("body", nil)
//	path.map(print, "params")              printAll(path, print, "params")
//	[a, "b", c]                            concat(a, "b", c)
//	group(x, { shouldBreak: true })        groupWith(x, doc.GroupOptions{ShouldBreak: true})
//	node.body?.type === "X"                node.Child("body").Is("X")
//
// Offsets are bytes into the UTF-8 source (see internal/format/estree), and every text helper here
// works in bytes, decoding a rune only where JavaScript's notion of whitespace reaches past ASCII.
package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/printing"
)

// Node is an ESTree node or comment.
type Node = *estree.Node

// Path is upstream's AstPath over ESTree nodes.
type Path = printing.AstPath[*estree.Node]

// Options is upstream's options object for one format: the printer, the original text, the comments.
type Options = printing.Options[*estree.Node]

// PrintFunc is upstream's print(selector, args).
type PrintFunc = printing.PrintFunc

// Doc is a document.
type Doc = doc.Doc

// settings is what upstream reads off `options` beyond the core's fields: the resolved Prettier
// options, the file path, and the caches upstream keeps in WeakMaps keyed on nodes.
type settings struct {
	formatoptions.Options
	FilePath string

	// docs is where this format's concatIn, groupIn and indentIn take their docs from, released with the
	// tree once the doc is laid out, or nil for the heap (#fyw36kf).
	docs *docMemory

	// Parser is upstream's options.parser: "typescript", "babel" for JavaScript, or "json" or
	// "json-stringify" for JSON files.
	// Only print/key.js and utilities/print-string.js read it.
	Parser string

	// ParentParser is upstream's options.parentParser: the parser of the file this text is embedded in,
	// set by main/multiparser.js's textToDoc on every embed, "" for a file of its own.
	ParentParser string

	callArguments      map[Node][]Node
	functionParameters map[Node][]Node
	// strippedText is the text without its comments, shared with the parse's postprocess when it ran one.
	strippedText *estree.StrippedText

	// groupIDs holds createGroupIdMapper's ids, by the mapper's description and then the node.
	groupIDs map[string]map[Node]*doc.GroupID

	// inJestEach is upstream's options.__inJestEach, set by print/template-literal.js while it
	// prints a jest `each` table and read by print/array.js.
	inJestEach bool
}

// settingsOf is the per-format state carried in options.Settings.
func settingsOf(options *Options) *settings {
	return options.Settings.(*settings)
}

// printArguments are upstream's `args` to print: an object a parent passes to shape a child's print.
// A nil *printArguments is passed to the core as an untyped nil, because the core caches only when
// args is nil and a typed nil in an interface is not.
type printArguments struct {
	expandFirstArg   bool
	expandLastArg    bool
	assignmentLayout string
}

func argsOf(args any) *printArguments {
	typed, _ := args.(*printArguments)
	return typed
}

// printWith calls print with args, keeping a missing args an untyped nil.
func printWith(print PrintFunc, selector any, args *printArguments) Doc {
	if args == nil {
		return print(selector, nil)
	}
	return print(selector, args)
}

// node is path.node.
func node(path *Path) Node {
	current, _ := path.Node()
	return current
}

// parentOf is path.parent.
func parentOf(path *Path) Node {
	current, _ := path.Parent()
	return current
}

// grandparentOf is path.grandparent.
func grandparentOf(path *Path) Node {
	current, _ := path.Grandparent()
	return current
}

// nodeAt is path.getNode(count).
func nodeAt(path *Path, count int) Node {
	current, _ := path.GetNode(count)
	return current
}

// keyOf is path.key, "" at the root.
func keyOf(path *Path) string {
	key, _ := path.Key()
	return key
}

// indexOf is path.index, -1 outside an array.
func indexOf(path *Path) int {
	index, inArray := path.Index()
	if !inArray {
		return -1
	}
	return index
}

// siblingsOf is path.siblings as nodes, nil outside an array.
func siblingsOf(path *Path) []Node {
	siblings, _ := path.Siblings().([]Node)
	return siblings
}

// nextOf is path.next.
func nextOf(path *Path) Node {
	next, _ := path.Next()
	return next
}

// previousOf is path.previous.
func previousOf(path *Path) Node {
	previous, _ := path.Previous()
	return previous
}

// call is path.call(callback, ...names).
func call[R any](path *Path, callback func(*Path) R, names ...any) R {
	return printing.Call(path, callback, names...)
}

// callParent is path.callParent(callback, count).
func callParent[R any](path *Path, callback func(*Path) R, count int) R {
	return printing.CallParent(path, callback, count)
}

// each is path.each(callback, ...names).
func each(path *Path, callback func(path *Path, index int), names ...any) {
	path.Each(func(path *Path, index int, _ any) { callback(path, index) }, names...)
}

// mapPath is path.map(callback, ...names).
func mapPath[R any](path *Path, callback func(path *Path, index int) R, names ...any) []R {
	return printing.Map(path, func(path *Path, index int, _ any) R { return callback(path, index) }, names...)
}

// printAll is path.map(print, ...names).
func printAll(path *Path, print PrintFunc, names ...any) []Doc {
	return mapPath(path, func(*Path, int) Doc { return print(nil, nil) }, names...)
}

// originalText is options.originalText.
func originalText(options *Options) string {
	return options.OriginalText
}

// locStart and locEnd are the JavaScript printer's location functions.
func locStart(node Node) int { return estree.LocStart(node) }
func locEnd(node Node) int   { return estree.LocEnd(node) }
