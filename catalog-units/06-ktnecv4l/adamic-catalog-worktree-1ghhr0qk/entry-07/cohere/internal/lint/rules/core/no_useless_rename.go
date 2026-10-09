package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnnecessarilyRenamed = rule.Message{
	Id: "unnecessarilyRenamed",
	Description: "This renames a binding to the name it already had, so the rename does nothing " +
		"and the longer spelling only invites a reader to look for a difference that is not there.",
}

// NoUselessRenameSettings is the decoded option surface.
//
// All three options default to FALSE, so the zero value is correct here and the default-inversion
// hazard does not apply. They are decoded through pointers anyway so an explicit false stays
// distinguishable from an absent key, which keeps the shape right if a default ever moves.
type NoUselessRenameSettings struct {
	IgnoreDestructuring bool
	IgnoreImport        bool
	IgnoreExport        bool
}

// DefaultNoUselessRenameSettings is upstream's `{ignoreDestructuring: false, ignoreImport: false,
// ignoreExport: false}`.
func DefaultNoUselessRenameSettings() NoUselessRenameSettings {
	return NoUselessRenameSettings{}
}

type noUselessRenameWireShape struct {
	IgnoreDestructuring *bool `json:"ignoreDestructuring"`
	IgnoreImport        *bool `json:"ignoreImport"`
	IgnoreExport        *bool `json:"ignoreExport"`
}

// DecodeNoUselessRenameOptions reads the option object off the config.
func DecodeNoUselessRenameOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultNoUselessRenameSettings(), nil
	}

	var wire noUselessRenameWireShape
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return DefaultNoUselessRenameSettings(), err
	}

	settings := DefaultNoUselessRenameSettings()
	if wire.IgnoreDestructuring != nil {
		settings.IgnoreDestructuring = *wire.IgnoreDestructuring
	}
	if wire.IgnoreImport != nil {
		settings.IgnoreImport = *wire.IgnoreImport
	}
	if wire.IgnoreExport != nil {
		settings.IgnoreExport = *wire.IgnoreExport
	}
	return settings, nil
}

// NoUselessRename flags a destructuring, import or export that renames a binding to its own name.
//
//	valid:   let {foo} = obj;
//	valid:   let {foo: bar} = obj;
//	valid:   import {foo as bar} from "foo";
//	invalid: let {foo: foo} = obj;
//	invalid: import {foo as foo} from "foo";
//	invalid: export {foo as foo};
//
// # Three sites, one judgment, three options
//
// A destructured property, an import specifier and an export specifier each carry an original name
// and a local one, and the rule reports when they match. Each site has its own option to turn it
// off, all three defaulting to on.
//
// # Shorthand is not a rename, and neither is a computed key
//
// `let {foo} = obj` writes one name and renames nothing, so there is nothing to report; upstream
// tests the shorthand flag before reading a name. A computed key is worse than uninteresting: the
// key is not known until run time, so `let {[foo]: foo} = obj` might or might not be a rename and
// upstream declines rather than guessing. Both are clean cases in its corpus and a port reading the
// two names without these tests reports every shorthand destructuring in the tree.
//
// # A string key counts, and an escape does not change the name
//
// `let {'foo': foo} = obj` is a rename to the same name and reports, because a string key and an
// identifier name compare by VALUE. That is also why `let {a: a} = obj` reports: the escape is
// spelling rather than identity, and both sides are the name `a`. Our parser hands back the cooked
// text for both, so the comparison needs no unescaping here, and the corpus carries five escape
// cases that would fail loudly if it did.
//
// # The fix rewrites the whole specifier, and refuses two shapes
//
// The repair replaces the property or specifier with just its local part, turning `{foo: foo}` into
// `{foo}`. It refuses when a comment sits in the part being discarded, since applying it unattended
// would delete the comment, and it refuses a parenthesized left side of a default assignment,
// because `({foo: (foo) = a} = obj)` cannot become `({(foo) = a} = obj)`: parentheses are not legal
// in a shorthand property. Twenty one of upstream's 107 failing cases carry `output: null` for one
// of those two reasons, and both are ported.
//
// The comment test is a COUNT rather than a presence test, which matters: a comment inside the part
// being KEPT survives the rewrite, so `({foo: foo /**/} = {})` still fixes while
// `({foo /**/: foo} = {})` does not. Upstream compares the comments inside the whole node against
// those inside the replacement, and that comparison is what draws the line.
var NoUselessRename = rule.Rule{
	Name: "no-useless-rename",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoUselessRenameSettings](options)
		if !ok {
			settings = DefaultNoUselessRenameSettings()
		}

		return rule.Listeners{
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if settings.IgnoreDestructuring {
					return
				}
				checkUselessRenameInBindingPattern(ctx, node)
			},
			// A destructuring ASSIGNMENT rather than a declaration reaches the parser as an object
			// literal on the left of an `=`, so its properties are property assignments rather than
			// binding elements. Upstream sees one ObjectPattern for both because its parser
			// reinterprets; ours does not, so both shapes are visited.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if settings.IgnoreDestructuring || !isADestructuringTarget(node) {
					return
				}
				checkUselessRenameInObjectLiteral(ctx, node)
			},
			ast.KindImportSpecifier: func(node *ast.Node) {
				if settings.IgnoreImport {
					return
				}
				specifier := node.AsImportSpecifier()
				if specifier == nil {
					return
				}
				// An import keeps its local name, which is the one after `as`.
				checkUselessRenameSpecifier(ctx, node, specifier.PropertyName, specifier.Name(),
					specifier.Name())
			},
			ast.KindExportSpecifier: func(node *ast.Node) {
				if settings.IgnoreExport {
					return
				}
				specifier := node.AsExportSpecifier()
				if specifier == nil {
					return
				}
				// An export keeps its local name too, and in an export that is the name BEFORE
				// `as`, which the parser stores as the property name.
				checkUselessRenameSpecifier(ctx, node, specifier.PropertyName, specifier.Name(),
					specifier.PropertyName)
			},
		}
	},
}

// isADestructuringTarget says whether an object literal is really a destructuring pattern.
//
// The parser produces an object LITERAL for `({foo: foo} = obj)` and only the surrounding assignment
// says it is a pattern. Without this test the rule would report an ordinary object literal
// `({foo: foo})`, which upstream leaves alone: its parser never calls that an ObjectPattern, so no
// listener of its fires. That input is a clean case in its corpus.
func isADestructuringTarget(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary != nil && binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == current
		case ast.KindForInStatement, ast.KindForOfStatement:
			return parent.AsForInOrOfStatement().Initializer == current
		case ast.KindParenthesizedExpression, ast.KindPropertyAssignment,
			ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
			// Keep walking: a nested pattern is a target when the outermost one is.
			continue
		default:
			return false
		}
	}
	return false
}

// checkUselessRenameInBindingPattern judges the declaration form, `let {foo: foo} = obj`.
func checkUselessRenameInBindingPattern(ctx rule.Context, node *ast.Node) {
	for _, element := range node.AsBindingPattern().Elements.Nodes {
		if element.Kind != ast.KindBindingElement {
			continue
		}
		binding := element.AsBindingElement()
		// No property name is shorthand, which renames nothing.
		if binding.PropertyName == nil {
			continue
		}
		// A computed key is not known until run time, so upstream declines rather than guessing.
		//
		// SUBSUMED rather than load bearing, and kept because it states the reason where a reader
		// looks for it. `nameTextOf` answers the empty string for any kind that is not an
		// identifier or a string, and a computed property name is neither, so the comparison below
		// already declines every one of them. Measured over five shapes including a computed key
		// whose expression is a string literal: byte identical findings with the guard present and
		// absent, and all five agree with the installed rule.
		if binding.PropertyName.Kind == ast.KindComputedPropertyName {
			continue
		}
		name := binding.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			continue
		}
		if nameTextOf(binding.PropertyName) != name.Text() {
			continue
		}
		reportUselessRename(ctx, element, keptTextRangeOfBindingElement(ctx, binding))
	}
}

// checkUselessRenameInObjectLiteral judges the assignment form, `({foo: foo} = obj)`.
func checkUselessRenameInObjectLiteral(ctx rule.Context, node *ast.Node) {
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			continue
		}
		assignment := property.AsPropertyAssignment()
		if assignment.Name() == nil ||
			assignment.Name().Kind == ast.KindComputedPropertyName {
			continue
		}
		value := assignment.Initializer
		if value == nil {
			continue
		}
		// The local name, looking through parentheses and through a default assignment, which is
		// what makes `({foo: (foo) = a} = obj)` a finding at all.
		local := localNameOfDestructuringValue(value)
		if local == nil || nameTextOf(assignment.Name()) != local.Text() {
			continue
		}
		reportUselessRename(ctx, property, valueRangeOfPropertyAssignment(ctx, assignment))
	}
}

// localNameOfDestructuringValue reads the binding a destructured value introduces.
//
// It looks through parentheses, since `({foo: (foo)} = obj)` is a rename to the same name and one of
// upstream's failing cases, and through a default, since `({foo: foo = 1} = obj)` is too.
func localNameOfDestructuringValue(value *ast.Node) *ast.Node {
	for value != nil && value.Kind == ast.KindParenthesizedExpression {
		value = value.AsParenthesizedExpression().Expression
	}
	if value == nil {
		return nil
	}
	if value.Kind == ast.KindBinaryExpression {
		binary := value.AsBinaryExpression()
		if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			return nil
		}
		left := binary.Left
		for left != nil && left.Kind == ast.KindParenthesizedExpression {
			left = left.AsParenthesizedExpression().Expression
		}
		if left != nil && left.Kind == ast.KindIdentifier {
			return left
		}
		return nil
	}
	if value.Kind == ast.KindIdentifier {
		return value
	}
	return nil
}

// checkUselessRenameSpecifier judges an import or export specifier.
//
// A nil property name is the un-renamed form, `import {foo}`, which renames nothing. Upstream words
// the same test as a range comparison, because its parser gives both fields the same node when there
// is no rename; ours leaves the property name nil, which asks the question more directly.
//
// The `keep` argument is which of the two names the repair writes, and the two sites differ. An
// import keeps its LOCAL name, so `import {'foo' as foo}` becomes `import {foo}`. An export keeps
// its local name too, which is the FIRST one written, so `export {foo as 'foo'}` becomes
// `export {foo}` while `export {'foo' as foo}` becomes `export {'foo'}`.
//
// That asymmetry looks like an inconsistency and is not: in both cases upstream writes the text of
// `node.local`, and `local` is the second name in an import and the first in an export. Both
// spellings are in the corpus with opposite outputs, so nothing but an output comparison could tell
// them apart, and a port keeping the same side at both sites passes every message-id assertion.
func checkUselessRenameSpecifier(ctx rule.Context, node *ast.Node, propertyName *ast.Node,
	local *ast.Node, keep *ast.Node) {
	if propertyName == nil || local == nil {
		return
	}
	if nameTextOf(propertyName) != nameTextOf(local) {
		return
	}
	reportUselessRename(ctx, node, rangeOfNode(ctx, keep))
}

// nameTextOf reads a name's value, whether it is written as an identifier or as a string.
//
// A string key and an identifier compare by VALUE rather than by spelling, which is why
// `let {'foo': foo} = obj` reports. The same applies to an escape: our parser hands back the cooked
// text, so `a` and `a` are one name here without any unescaping, and upstream's five escape
// cases are what pin it.
func nameTextOf(name *ast.Node) string {
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return name.Text()
	}
	return ""
}

// reportUselessRename reports a finding, with the repair when it can be offered safely.
func reportUselessRename(ctx rule.Context, node *ast.Node, kept textSpan) {
	if !kept.ok {
		ctx.ReportNode(node, messageUnnecessarilyRenamed)
		return
	}

	// A comment in the discarded part would be deleted by the rewrite, so the repair is withheld and
	// the finding still fires. Counted rather than merely detected: a comment inside the part being
	// KEPT survives, so `({foo: foo /**/} = {})` still fixes.
	// The node's span starts at its first TOKEN rather than at Pos(), because Pos() reaches back
	// over leading trivia and would count a comment written BEFORE the specifier as being inside
	// it. Upstream's getCommentsInside asks about the node's own range, where a leading comment is
	// not inside, so `import {/* comment */foo as foo}` still fixes. Seven of upstream's fix
	// vectors are that shape and every one of them was a withheld repair here first.
	if commentsInsideCount(ctx, rule.TokenRange(ctx.SourceFile, node).Pos(), node.End()) >
		commentsInsideCount(ctx, kept.start, kept.end) {
		ctx.ReportNode(node, messageUnnecessarilyRenamed)
		return
	}

	source := ctx.SourceFile.Text()
	if kept.start < 0 || kept.end > len(source) || kept.start > kept.end {
		ctx.ReportNode(node, messageUnnecessarilyRenamed)
		return
	}

	ctx.ReportNodeWithFixes(node, messageUnnecessarilyRenamed,
		ctx.ReplaceNode(node, source[kept.start:kept.end]))
}

// textSpan is the part of a node the repair keeps, or a statement that no repair is available.
type textSpan struct {
	start int
	end   int
	ok    bool
}

// keptTextRangeOfBindingElement gives the text a declaration-form repair keeps.
//
// Everything from the local name through the end of the element, which carries any default with it:
// `{foo: foo = 1}` keeps `foo = 1`. The start is the name's token start rather than its Pos(), since
// Pos() sits before leading trivia and a repair built from it would swallow the space after the
// colon.
func keptTextRangeOfBindingElement(ctx rule.Context, binding *ast.BindingElement) textSpan {
	name := binding.Name()
	if name == nil {
		return textSpan{}
	}
	// TokenRange rather than Pos(), because Pos() sits before leading trivia: the space after the
	// colon in `{foo: foo}` belongs to nobody, and a repair built from Pos() writes `{ foo}`. That
	// was 86 failing fix vectors here before this line, and it reads as a broken fixer rather than
	// as a range bug.
	return textSpan{start: rule.TokenRange(ctx.SourceFile, name).Pos(),
		end: binding.AsNode().End(), ok: true}
}

// valueRangeOfPropertyAssignment gives the text an assignment-form repair keeps.
//
// Upstream refuses a parenthesized left side of a default here, because parentheses are not legal in
// a shorthand property: `({foo: (foo) = a} = obj)` cannot become `({(foo) = a} = obj)`. That refusal
// is the reason this returns a span rather than a range, so the caller can report without a fix.
func valueRangeOfPropertyAssignment(ctx rule.Context, assignment *ast.PropertyAssignment) textSpan {
	value := assignment.Initializer
	if value == nil {
		return textSpan{}
	}
	if value.Kind == ast.KindBinaryExpression {
		binary := value.AsBinaryExpression()
		if binary.Left != nil && binary.Left.Kind == ast.KindParenthesizedExpression {
			return textSpan{}
		}
	}
	// A bare parenthesized local is unwrapped, so `({foo: (foo)} = obj)` repairs to `({foo} = obj)`
	// rather than to `({(foo)} = obj)`, which is not legal shorthand. Upstream reaches the same
	// place from the other side: its parser hands the rule the identifier and it writes that node's
	// text. A parenthesized DEFAULT is refused above rather than unwrapped, because the parentheses
	// there may be load bearing for precedence.
	kept := value
	for kept.Kind == ast.KindParenthesizedExpression {
		inner := kept.AsParenthesizedExpression().Expression
		if inner == nil || inner.Kind != ast.KindIdentifier {
			break
		}
		kept = inner
	}
	return textSpan{start: rule.TokenRange(ctx.SourceFile, kept).Pos(), end: kept.End(), ok: true}
}

// rangeOfNode gives a whole node's span as the kept text.
//
// TokenRange rather than Pos() for the same reason as the binding element above: Pos() sits before
// leading trivia, so the space after `as` in `{foo as foo}` would be carried into the repair and
// produce `{ foo}`.
func rangeOfNode(ctx rule.Context, node *ast.Node) textSpan {
	if node == nil {
		return textSpan{}
	}
	return textSpan{start: rule.TokenRange(ctx.SourceFile, node).Pos(), end: node.End(), ok: true}
}

// commentsInsideCount counts the comments lying wholly inside a span.
//
// Upstream's `getCommentsInside` returns comments contained by the node, which is why this asks for
// containment rather than for overlap. The comment cache is per file, so this costs one scan across
// every finding in a file.
func commentsInsideCount(ctx rule.Context, start int, end int) int {
	count := 0
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= start && comment.Range.End() <= end {
			count++
		}
	}
	return count
}
