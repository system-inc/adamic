package printing

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/system-inc/cohere/internal/format/doc"
)

// PrintFunc is upstream's mainPrint as a printer receives it: print the node at selector, relative to
// the current path. A nil selector prints the current node. A selector is one name or a slice of names,
// exactly as upstream's print(selector, args).
type PrintFunc func(selector any, args any) doc.Doc

// CommentContext is upstream's decorated comment, what the ownLine, endOfLine and remaining handlers
// receive when the printer sets AvoidAstMutation, and what attachment builds either way.
type CommentContext[N Node[N]] struct {
	Comment       N
	Text          string
	Options       *Options[N]
	Ast           N
	IsLastComment bool
	Placement     string

	Enclosing, Preceding, Following          N
	HasEnclosing, HasPreceding, HasFollowing bool
}

// CommentHandlers are upstream's printer.handleComments. Each returns true when it attached the
// comment itself, and false to let attachment's default placement run.
type CommentHandlers[N Node[N]] struct {
	OwnLine   func(*CommentContext[N]) bool
	EndOfLine func(*CommentContext[N]) bool
	Remaining func(*CommentContext[N]) bool
}

// Embed is upstream's printer.embed result, made synchronous: given the path, a printer either returns
// nil (nothing embedded here) or a function that produces the embedded doc. The function receives
// TextToDoc to format another language's text.
type Embed[N Node[N]] func(path *AstPath[N], options *Options[N]) func(textToDoc TextToDoc, print PrintFunc, path *AstPath[N], options *Options[N]) (doc.Doc, error)

// TextToDoc formats text in another language and returns its doc, upstream's textToDoc. The parser
// name is upstream's (typescript, babel, json, graphql, css, markdown). Supplied by whoever knows every
// language, which is internal/format/native, because the print core cannot import the printers that
// import it.
type TextToDoc func(text string, parser string) (doc.Doc, error)

// Printer is upstream's printer object: what a language supplies to the core. Every field but Print,
// LocStart, LocEnd and VisitorKeys is optional, and a nil one means what upstream's absent one means.
type Printer[N Node[N]] struct {
	Print func(path *AstPath[N], options *Options[N], print PrintFunc, args any) doc.Doc

	// LocStart and LocEnd are upstream's options.locStart and locEnd: byte offsets, with the
	// language's own overrides.
	LocStart func(node N) int
	LocEnd   func(node N) int

	// VisitorKeys is upstream's getVisitorKeys: the child properties of a node, in upstream's order.
	VisitorKeys func(node N) []string

	// MayHoldEmbed, when set, says whether a tree could hold a node Embed answers for. A tree that cannot
	// skips the embed walk, which visits every node through the path and was 7.6M of the JavaScript
	// printer's allocations on ahra (#fyw36kf). It may answer true for a tree with none; never false for one.
	MayHoldEmbed func(root N) bool

	Embed                Embed[N]
	EmbedVisitorKeys     func(node N) []string
	Preprocess           func(ast N, options *Options[N]) N
	HasPrettierIgnore    func(path *AstPath[N]) bool
	PrintPrettierIgnored func(path *AstPath[N], options *Options[N], print PrintFunc, args any) doc.Doc
	WillPrintOwnComments func(path *AstPath[N], options *Options[N]) bool

	CanAttachComment     func(node N, ancestors []N) bool
	IsBlockComment       func(comment N) bool
	PrintComment         func(path *AstPath[N], options *Options[N]) doc.Doc
	GetCommentChildNodes func(node N, options *Options[N]) ([]N, bool)
	IsGap                func(gap string, options *Options[N]) (bool, bool)
	HandleComments       CommentHandlers[N]

	// AvoidAstMutation is upstream's features.experimental_avoidAstMutation, which decides whether the
	// comment handlers receive a CommentContext. Here they always do; the flag is kept so a printer port
	// can say which upstream mode it is porting.
	AvoidAstMutation bool

	// TemplateQuasis exists because upstream's language-independent attacher special-cases
	// TemplateLiteral and reads its quasis, so comments inside one template expression do not move to
	// another. That is JavaScript knowledge in a generic place; here it is a hook the JavaScript printer
	// supplies and every other language leaves nil.
	TemplateQuasis func(node N) ([]N, bool)
}

// Options is upstream's options object as the core and printers read it.
type Options[N Node[N]] struct {
	Printer      *Printer[N]
	OriginalText string

	// Settings is the language's own options (formatoptions.Options for every printer today), carried as a
	// value the language type-asserts, so this package does not import the oracle engine's package.
	Settings any

	// EmbeddedLanguageFormatting is upstream's option: "auto" (the default) or "off".
	EmbeddedLanguageFormatting string

	// TextToDoc formats embedded text in another language. Nil means no embedding is possible.
	TextToDoc TextToDoc

	// Comments is every comment in the file, upstream's options[commentsPropertyInOptions].
	Comments []N

	// NodeCount is about how many nodes the tree holds, when the language knows it cheaply, or zero. The
	// print cache, one entry per node printed, starts with room for that many rather than growing by
	// insertion, which was 124 + 79 MB of a cold run on ahra (#wcgw0n4). It never changes the output.
	NodeCount int

	// printedComments is upstream's options[Symbol.for("printedComments")]: comments printed by a
	// printer directly or covered by a prettier-ignore, which ensureAllCommentsPrinted must not flag.
	printedComments map[N]bool
}

// MarkPrinted records a comment as printed outside printComments, upstream's printedComments.add.
func (options *Options[N]) MarkPrinted(comment N) {
	if options.printedComments == nil {
		options.printedComments = map[N]bool{}
	}
	options.printedComments[comment] = true
}

// isNil is upstream's `value === undefined || value === null` for a value the stack holds. A typed nil
// pointer stored in an interface is not nil to Go, and it is nothing to print to upstream.
func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func:
		return reflected.IsNil()
	}
	return false
}

// PrintAstToDoc is upstream's printAstToDoc, src/main/ast-to-doc.js: attach comments, preprocess,
// resolve embedded languages, then print from the root, wrapping each node's doc in its comments.
//
// Cursor handling is not ported. cohere formats whole files and never tracks a cursor.
func PrintAstToDoc[N Node[N]](ast N, comments []N, options *Options[N]) (doc.Doc, error) {
	options.Comments = comments
	if options.printedComments == nil {
		options.printedComments = map[N]bool{}
	}

	AttachComments(ast, comments, options)
	if options.Printer.Preprocess != nil {
		ast = options.Printer.Preprocess(ast, options)
	}

	cache := make(map[N]doc.Doc, options.NodeCount)
	path := NewAstPath(ast)
	path.Settings = options.Settings
	embeds := map[N]doc.Doc{}

	var mainPrint PrintFunc
	mainPrintInternal := func(args any) doc.Doc {
		ensurePrintingNode(path, options)
		value := path.Value()
		if isNil(value) {
			return doc.Text("")
		}
		node, isNode := asNode[N](value)
		shouldCache := isNode && args == nil
		if shouldCache {
			if cached, present := cache[node]; present {
				return cached
			}
		}
		printed := callPluginPrintFunction(path, options, mainPrint, args, embeds)
		if shouldCache {
			cache[node] = printed
		}
		return printed
	}
	mainPrint = func(selector any, args any) doc.Doc {
		switch typed := selector.(type) {
		case nil:
			return mainPrintInternal(args)
		case *AstPath[N]:
			if typed == path {
				return mainPrintInternal(args)
			}
			panic("printing: print was called with a path other than the one being printed")
		case []any:
			return Call(path, func(*AstPath[N]) doc.Doc { return mainPrintInternal(args) }, typed...)
		case []string:
			names := make([]any, len(typed))
			for index, name := range typed {
				names[index] = name
			}
			return Call(path, func(*AstPath[N]) doc.Doc { return mainPrintInternal(args) }, names...)
		default:
			return Call(path, func(*AstPath[N]) doc.Doc { return mainPrintInternal(args) }, selector)
		}
	}

	printEmbeddedLanguages(path, mainPrint, options, embeds)

	printed := callPluginPrintFunction(path, options, mainPrint, nil, embeds)
	if err := ensureAllCommentsPrinted(options); err != nil {
		return nil, err
	}
	return printed, nil
}

// callPluginPrintFunction is upstream's callPluginPrintFunction, without cursor labels.
func callPluginPrintFunction[N Node[N]](path *AstPath[N], options *Options[N], print PrintFunc, args any, embeds map[N]doc.Doc) doc.Doc {
	node, _ := path.Node()
	printer := options.Printer

	var printed doc.Doc
	switch {
	case printer.HasPrettierIgnore != nil && printer.HasPrettierIgnore(path):
		printed = printIgnored(path, options, print, args)
	default:
		if embedded, present := embeds[node]; present {
			printed = embedded
		} else {
			printed = printer.Print(path, options, print, args)
		}
	}

	// JSXElement prints its own comments, because it adds parentheses around them; that is what
	// WillPrintOwnComments is for.
	if printer.PrintComment != nil && len(node.CommentData().Comments) > 0 &&
		(printer.WillPrintOwnComments == nil || !printer.WillPrintOwnComments(path, options)) {
		printed = PrintComments(path, printed, options, nil)
	}
	return printed
}

// printIgnored is upstream's printIgnored, src/main/print-ignored.js: a prettier-ignore'd node prints as
// its original text, and every comment inside it counts as printed.
func printIgnored[N Node[N]](path *AstPath[N], options *Options[N], print PrintFunc, args any) doc.Doc {
	node, _ := path.Node()
	start := options.Printer.LocStart(node)
	end := options.Printer.LocEnd(node)
	for _, comment := range options.Comments {
		if options.Printer.LocStart(comment) >= start && options.Printer.LocEnd(comment) <= end {
			options.MarkPrinted(comment)
		}
	}
	if options.Printer.PrintPrettierIgnored != nil {
		return options.Printer.PrintPrettierIgnored(path, options, print, args)
	}
	return doc.Text(options.OriginalText[start:end])
}

// ensurePrintingNode is upstream's createPrintPreCheckFunction: print() may only be called on a
// property the parent's visitor keys name. Upstream checks this outside production builds; it is kept
// here because a printer port calling print on a non-node property is a bug worth a loud stop.
func ensurePrintingNode[N Node[N]](path *AstPath[N], options *Options[N]) {
	if path.IsRoot() {
		return
	}
	key, hasKey := path.Key()
	parent, hasParent := path.Parent()
	if !hasKey || !hasParent {
		return
	}
	if slices.Contains(options.Printer.VisitorKeys(parent), key) {
		return
	}
	panic(fmt.Sprintf("printing: print() called on %q of a %s, which is not one of its visitor keys %v",
		key, parent.Type(), options.Printer.VisitorKeys(parent)))
}

// printEmbeddedLanguages is upstream's printEmbeddedLanguages, src/main/multiparser.js, made
// synchronous.
//
// An embed that fails is dropped and the node prints normally, as upstream's try/catch drops it. That
// is observable: a fenced code block that does not parse prints as written rather than failing the file.
func printEmbeddedLanguages[N Node[N]](path *AstPath[N], print PrintFunc, options *Options[N], embeds map[N]doc.Doc) {
	if options.EmbeddedLanguageFormatting != "" && options.EmbeddedLanguageFormatting != "auto" {
		return
	}
	printer := options.Printer
	if printer.Embed == nil {
		return
	}
	if printer.MayHoldEmbed != nil {
		if root, isNode := asNode[N](path.Value()); isNode && !printer.MayHoldEmbed(root) {
			return
		}
	}
	visitorKeys := printer.EmbedVisitorKeys
	if visitorKeys == nil {
		visitorKeys = printer.VisitorKeys
	}

	type pending struct {
		print func(TextToDoc, PrintFunc, *AstPath[N], *Options[N]) (doc.Doc, error)
		node  N
		stack []any
	}
	var calls []pending

	// Each key is boxed once for the path's names, rather than on every node that has it, and the callback
	// Call takes is made once, rather than once per key of every node.
	boxedKeys := map[string]any{}
	var recurse func(*AstPath[N], int, any)
	recurseInto := func(inner *AstPath[N]) struct{} { recurse(inner, 0, nil); return struct{}{} }
	recurse = func(current *AstPath[N], _ int, _ any) {
		value := current.Value()
		node, isNode := asNode[N](value)
		if isNil(value) || !isNode || (printer.HasPrettierIgnore != nil && printer.HasPrettierIgnore(current)) {
			return
		}
		for _, key := range visitorKeys(node) {
			boxedKey, known := boxedKeys[key]
			if !known {
				boxedKey = key
				boxedKeys[key] = boxedKey
			}
			if isSlice(node.Field(key)) {
				current.Each(recurse, boxedKey)
			} else {
				Call(current, recurseInto, boxedKey)
			}
		}
		if result := printer.Embed(current, options); result != nil {
			calls = append(calls, pending{print: result, node: node, stack: append([]any(nil), current.stack...)})
		}
	}
	recurse(path, 0, nil)

	original := path.stack
	for _, call := range calls {
		path.stack = call.stack
		textToDoc := options.TextToDoc
		if textToDoc == nil {
			textToDoc = func(string, string) (doc.Doc, error) {
				return nil, fmt.Errorf("printing: no TextToDoc is configured for embedded languages")
			}
		}
		embedded, err := call.print(textToDoc, print, path, options)
		if err == nil && embedded != nil {
			embeds[call.node] = embedded
		}
	}
	path.stack = original
}
