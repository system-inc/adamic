package compose

// Ported from eemeli/yaml 2.9.0, dist/doc/Document.js: the constructor as the composer calls it (no
// value, so no createNode) and setSchema. The rest of the class is the document API and stringify.

import "github.com/system-inc/cohere/internal/format/arena"

// Options is the part of upstream's options composing reads, with the Document's defaults already
// applied. UnistParserOptions is the configuration yaml-unist-parser composes with.
type Options struct {
	// Nodes is where a parse's scalars, pairs and collections come from, nil to allocate each (#v6ksqg3).
	// A caller that sets it owns their lifetime: unist.Parse resets it once its tree is built, since no
	// unist node keeps a compose node.
	Nodes *arena.Arena[Node]
	// KeepSourceTokens sets each node's srcToken.
	KeepSourceTokens bool
	// UniqueKeys reports duplicate map keys, comparing keys as upstream's default uniqueKeys does.
	UniqueKeys bool
	// Merge adds the << merge key tag to the schema.
	Merge bool
	// Strict is upstream's strict, true by default.
	Strict bool
	// StringKeys requires every map key to be a string.
	StringKeys bool
	// Version is the YAML version a stream starts in, "1.2" by default.
	Version string
}

// UnistParserOptions is yaml-unist-parser's `new Composer({ keepSourceTokens: true, uniqueKeys: false,
// lineCounter, merge: true })` with the Document defaults it leaves as they are (strict: true,
// stringKeys: false, version: '1.2'). The lineCounter rides along in the options unread.
func UnistParserOptions() Options {
	return Options{KeepSourceTokens: true, UniqueKeys: false, Merge: true, Strict: true, Version: "1.2"}
}

// Document is a composed YAML document.
type Document struct {
	// CommentBefore is a comment before this Document, "" for none (upstream's null).
	CommentBefore string
	// Comment is a comment immediately after this Document, "" for none.
	Comment string
	// Errors are the errors encountered during composing.
	Errors []*YAMLError
	// Warnings are the warnings encountered during composing.
	Warnings []*YAMLError

	Options    *Options
	Directives *Directives
	Schema     *Schema
	// Contents is the document's root node, nil for the empty document the composer forces at the
	// end of a stream with no documents.
	Contents *Node
	// Range is [start, contents end, end] in UTF-16 code units.
	Range []int
}

// newDocument is `new Document(undefined, Object.assign({ _directives: directives }, options))`.
func newDocument(directives *Directives, options Options) *Document {
	document := &Document{Errors: []*YAMLError{}, Warnings: []*YAMLError{}}
	document.Options = &options
	version := options.Version
	document.Directives = directives.atDocument()
	if document.Directives.YAML.Explicit {
		version = document.Directives.YAML.Version
	}
	document.setSchema(version, options)
	return document
}

// setSchema is upstream's for the two versions a directive can set.
func (document *Document) setSchema(version string, options Options) {
	var schema schemaOptions
	switch version {
	case "1.1":
		document.Directives.YAML.Version = "1.1"
		schema = schemaOptions{resolveKnownTags: false, schema: "yaml-1.1"}
	case "1.2", "next":
		document.Directives.YAML.Version = version
		schema = schemaOptions{resolveKnownTags: true, schema: "core"}
	default:
		panic("compose: Expected '1.1', '1.2' or null as first argument, but found: " + version)
	}
	schema.merge = options.Merge
	document.Schema = newSchema(schema)
}
