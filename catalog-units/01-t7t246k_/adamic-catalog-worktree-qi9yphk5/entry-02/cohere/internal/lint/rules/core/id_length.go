package core

import (
	"fmt"
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// IdLengthSettings is the decoded option surface.
type IdLengthSettings struct {
	// Minimum is the shortest name allowed. Upstream's default is 2.
	Minimum int
	// Maximum is the longest name allowed. Zero means unbounded, which is upstream's `Infinity`
	// default: the option has no maximum unless one is configured.
	Maximum int
	// HasMaximum distinguishes "no maximum configured" from a configured maximum of zero, which
	// would reject every name.
	HasMaximum bool
	// CheckProperties is upstream's `properties !== "never"`. Default true.
	CheckProperties bool
	// Exceptions are names exempted verbatim.
	Exceptions map[string]bool
	// ExceptionPatterns are compiled regular expressions; a name matching any of them is exempt.
	ExceptionPatterns []*regexp.Regexp
}

// DefaultIdLengthSettings is upstream's `defaultOptions`, spelled out because the zero value of the
// struct is not it: Minimum would be 0 and CheckProperties false, both the opposite of the default.
func DefaultIdLengthSettings() IdLengthSettings {
	return IdLengthSettings{Minimum: 2, CheckProperties: true}
}

// idLengthRawOptions is the wire shape.
//
// Minimum and Properties are pointers so an absent key is distinguishable from a written zero or
// empty string, which mean different things for a default-2 and a default-"always" option.
type idLengthRawOptions struct {
	Min               *int     `json:"min"`
	Max               *int     `json:"max"`
	Exceptions        []string `json:"exceptions"`
	ExceptionPatterns []string `json:"exceptionPatterns"`
	Properties        *string  `json:"properties"`
}

// DecodeIdLengthOptions reads the option object off the config.
//
// Hand rolled rather than routed through `rule.DecodeOptionsInto` because two of the defaults are
// non-zero -- a minimum of 2 and properties "always" -- so a zero-value struct is a different rule
// rather than a weaker one: it would exempt every name from the minimum and stop checking
// properties at the same time.
//
// Empty input decodes to those defaults, which is what a rule configured as a bare "error" is
// handed and what upstream's `defaultOptions` says it means.
//
// An exception pattern that does not compile is returned as an error rather than dropped. Upstream
// would throw at `new RegExp(pattern, "u")`, and silently dropping one would exempt fewer names
// than the author asked for while looking configured.
func DecodeIdLengthOptions(raw []byte) (any, error) {
	settings := DefaultIdLengthSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var wire idLengthRawOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}

	if wire.Min != nil {
		settings.Minimum = *wire.Min
	}
	if wire.Max != nil {
		settings.Maximum = *wire.Max
		settings.HasMaximum = true
	}
	// Upstream's `options.properties !== "never"`, so anything other than the literal "never"
	// leaves property checking on, including an absent key.
	if wire.Properties != nil && *wire.Properties == "never" {
		settings.CheckProperties = false
	}
	if len(wire.Exceptions) > 0 {
		settings.Exceptions = make(map[string]bool, len(wire.Exceptions))
		for _, name := range wire.Exceptions {
			settings.Exceptions[name] = true
		}
	}
	for _, pattern := range wire.ExceptionPatterns {
		// Upstream compiles with the `u` flag, which makes the pattern operate on code points.
		// Go's regexp is already code-point oriented, so there is no flag to add.
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return settings, fmt.Errorf("id-length exceptionPattern %q does not compile: %w", pattern, err)
		}
		settings.ExceptionPatterns = append(settings.ExceptionPatterns, compiled)
	}
	return settings, nil
}

// idLengthSettingsFrom recovers the settings from whatever the config layer handed over.
func idLengthSettingsFrom(options any) IdLengthSettings {
	if settings, ok := rule.OptionsAs[IdLengthSettings](options); ok {
		return settings
	}
	return DefaultIdLengthSettings()
}

var messageIdLengthTooShort = rule.Message{
	Id: "tooShort",
	Description: "A name this short carries no information about what it holds, so every reader " +
		"has to reconstruct that from the surrounding code. Name it for what it is.",
}

var messageIdLengthTooShortPrivate = rule.Message{
	Id: "tooShortPrivate",
	Description: "A private field name this short carries no information about what it holds, so " +
		"every reader has to reconstruct that from the surrounding code. Name it for what it is.",
}

var messageIdLengthTooLong = rule.Message{
	Id: "tooLong",
	Description: "A name this long is usually carrying context that belongs in its type or its " +
		"scope rather than in its spelling.",
}

var messageIdLengthTooLongPrivate = rule.Message{
	Id: "tooLongPrivate",
	Description: "A private field name this long is usually carrying context that belongs in its " +
		"type or its scope rather than in its spelling.",
}

// IdLength reports an identifier whose name is shorter or longer than the configured bounds.
//
//	valid:   var foo = 1;                      (default minimum of 2)
//	valid:   foo.a;                            (reading a property is not naming one)
//	valid:   var { a: b } = obj;               (the source object's key is not ours)
//	valid:   import { a } from 'm';            (the exporting module chose that name)
//	valid:   var 葛󠄀 = 2;                       (one GRAPHEME, at min 1 max 1)
//	invalid: var a = 1;                        (default minimum of 2)
//	invalid: obj.a = 1;                        (a WRITE creates a property with that name)
//	invalid: class C { #a = 1; }               (reported as #a)
//
// # The rule is a dispatch table on the parent, and what it does NOT check is the interesting half
//
// Upstream's `SUPPORTED_EXPRESSIONS` maps a parent's ESTree type to either `true` or a predicate.
// A name whose parent is absent from that table is never reported at all, which is how the rule
// avoids flagging every reference to a short name -- only the DECLARATION of `a` is reported, and
// the twenty uses of it are not. That inversion is the rule's whole shape: the table is a list of
// naming sites rather than a list of exclusions.
//
// Member access is in the table but gated: reading `foo.a` is fine because the object may be one
// we do not own, while `foo.a = 1` names a property we are creating. Same split as `id-denylist`,
// and upstream's predicate also reaches the destructuring-into-a-member form `({p: o.q} = {})`.
//
// # Grapheme counting, and why a rune count is wrong in BOTH directions
//
// Upstream measures with `getGraphemeCount`, which uses `Intl.Segmenter`. Go has no grapheme
// segmenter in its standard library and this module has no dependency that provides one, so the
// question was whether a rune count would do. It would not. Measured by replaying the whole corpus
// against a rune-counting mutant of the installed rule:
//
//	rune counting differs from grapheme counting on 2 of 181 corpus cases
//	  graphemes=0 runes=1   "var 葛󠄀 = 2"  {"min":1,"max":1}
//	  graphemes=1 runes=0   "var 葛󠄀 = 2"  default options
//
// One identifier, wrong in both directions, so it is a divergence rather than a conservative
// approximation. `葛󠄀` is a base character plus a variation selector: two runes, one grapheme.
//
// The count is `text.GraphemeCount`, which follows Unicode's grapheme cluster rules from Go's own
// character tables and is scored against `Intl.Segmenter` there. It began here as a count that only
// skipped combining marks and Hangul jamo, which is all an identifier can hold: the parser rejects
// regional indicators and emoji joiner sequences as identifier characters outright, so no name this
// rule sees contains one. It moved to the shelf when ban-ts-comment needed the same count over
// comment text, which can hold both.
//
// # No fix
//
// Upstream offers none and none is possible: the repair is to rename, which is a refactor reaching
// every reference rather than a lint edit.
var IdLength = rule.Rule{
	Name: "id-length",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := idLengthSettingsFrom(options)

		check := func(node *ast.Node) {
			checkIdLengthName(ctx, node, settings)
		}
		return rule.Listeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
		}
	},
}

// checkIdLengthName judges one identifier.
func checkIdLengthName(ctx rule.Context, node *ast.Node, settings IdLengthSettings) {
	parent := node.Parent
	if parent == nil {
		return
	}

	// A private name's `Text()` carries the leading `#`, and upstream matches against the name
	// without it: its `node.name` for a PrivateIdentifier excludes the hash, and the message
	// renders the hash back in. So both the length and the exception lookup use the bare name.
	name := node.Text()
	if node.Kind == ast.KindPrivateIdentifier && len(name) > 0 && name[0] == '#' {
		name = name[1:]
	}

	length := text.GraphemeCount(name)
	tooShort := length < settings.Minimum
	tooLong := settings.HasMaximum && length > settings.Maximum
	if !tooShort && !tooLong {
		return
	}
	if settings.Exceptions[name] {
		return
	}
	for _, pattern := range settings.ExceptionPatterns {
		if pattern.MatchString(name) {
			return
		}
	}

	if !idLengthIsNamingSite(node, parent, settings) {
		return
	}

	message := messageIdLengthTooShort
	if tooLong {
		message = messageIdLengthTooLong
	}
	if node.Kind == ast.KindPrivateIdentifier {
		message = messageIdLengthTooShortPrivate
		if tooLong {
			message = messageIdLengthTooLongPrivate
		}
	}

	rendered := fmt.Sprintf("Identifier name '%s' is too short (< %d). %s",
		name, settings.Minimum, message.Description)
	if tooLong {
		rendered = fmt.Sprintf("Identifier name '%s' is too long (> %d). %s",
			name, settings.Maximum, message.Description)
	}
	if node.Kind == ast.KindPrivateIdentifier {
		if tooLong {
			// Upstream's `tooLongPrivate` message spells the hash OUTSIDE the quotes --
			// `Identifier name #'{{name}}' is too long` -- while every other message puts it
			// inside. That is upstream's own inconsistency and it is reproduced rather than
			// tidied, because the message text is a verdict a fixture asserts.
			rendered = fmt.Sprintf("Identifier name #'%s' is too long (> %d). %s",
				name, settings.Maximum, message.Description)
		} else {
			rendered = fmt.Sprintf("Identifier name '#%s' is too short (< %d). %s",
				name, settings.Minimum, message.Description)
		}
	}

	ctx.ReportNode(node, rule.Message{Id: message.Id, Description: rendered})
}

// idLengthIsNamingSite is upstream's `SUPPORTED_EXPRESSIONS` lookup: does this identifier NAME
// something, as opposed to referring to something already named.
//
// A parent kind absent from this switch means the name is never reported, which is what keeps the
// rule from flagging every use of a short variable.
func idLengthIsNamingSite(node *ast.Node, parent *ast.Node, settings IdLengthSettings) bool {
	switch parent.Kind {
	// Upstream's unconditional `true` entries. `VariableDeclarator` is conditional there, on
	// `parent.id === node`, and that is the `Name()` test below.
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		return declaration != nil && declaration.Name() == node

	// A catch binding reaches the same kind as an ordinary variable here, because our parser gives
	// `catch (e)` a VariableDeclaration whose parent is the CatchClause. Upstream lists
	// `CatchClause: true` separately, and both land on the arm above, so nothing extra is needed --
	// but it is worth stating, because the two look like different cases in upstream's table.

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassDeclaration, ast.KindClassExpression:
		// The function's or class's own NAME, not its parameters.
		return idLengthNodeNameOf(parent) == node

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyDeclaration:
		// Upstream's MethodDefinition and PropertyDefinition, both unconditional.
		return idLengthNodeNameOf(parent) == node

	// Upstream's RestElement and ArrayPattern, and its parameter naming generally. A parameter's
	// name and a binding-pattern element are both naming sites; a binding element's PROPERTY name
	// is the source object's key and is handled below.
	case ast.KindParameter:
		declaration := parent.AsParameterDeclaration()
		return declaration != nil && declaration.Name() == node

	case ast.KindBindingElement:
		return idLengthBindingElementIsNamingSite(node, parent, settings)

	// Upstream's Property entry, whose two halves are an object literal key and an object pattern.
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return idLengthPropertyIsNamingSite(node, parent, settings)

	// Upstream's MemberExpression entry, gated on `properties` and on being a write.
	case ast.KindPropertyAccessExpression:
		if !settings.CheckProperties {
			return false
		}
		access := parent.AsPropertyAccessExpression()
		if access == nil || access.Name() != node {
			return false
		}
		return idLengthMemberIsWritten(parent)

	// Upstream's three import entries. Only the LOCAL name is ours, and a named import is ours only
	// when it RENAMES: `import { x } from 'y'` binds the exporting module's own name, which we did
	// not choose, so upstream exempts it.
	//
	// Upstream spells that as `getModuleExportName(parent.imported) !== getModuleExportName(...)`,
	// and the test is easy to get backwards because ESTree fills `imported` even when no `as` is
	// written -- with the same name -- so the inequality is false and the specifier is exempt. Our
	// parser leaves PropertyName NIL in that case, so a nil PropertyName means "no rename" and must
	// return false rather than true. An earlier draft returned true and reported every short named
	// import, which two corpus cases caught.
	//
	// `import { 'x' as x }` writes the exported name as a string literal, and upstream compares
	// through `getModuleExportName` so `'x'` and `x` are the same name. Measured, `Text()` on the
	// literal already yields `x`, so the plain comparison below handles it.
	case ast.KindImportSpecifier:
		specifier := parent.AsImportSpecifier()
		if specifier == nil || specifier.Name() != node {
			return false
		}
		if specifier.PropertyName == nil {
			return false
		}
		return specifier.PropertyName.Text() != node.Text()

	case ast.KindImportClause:
		clause := parent.AsImportClause()
		return clause != nil && clause.Name() == node

	case ast.KindNamespaceImport:
		namespace := parent.AsNamespaceImport()
		return namespace != nil && namespace.Name() == node
	}
	return false
}

// idLengthNodeNameOf reads a declaration's own name node, or nil when it has none.
func idLengthNodeNameOf(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return node.Name()
}

// idLengthBindingElementIsNamingSite handles a DECLARATION destructuring.
//
// One `KindBindingElement` carries what ESTree splits over a Property's key and value:
//
//	var { a }        PropertyName nil,  Name() = a     upstream's shorthand
//	var { p: q }     PropertyName = p,  Name() = q     key and value
//	var [ a ]        PropertyName nil,  Name() = a     an array pattern element
//
// Upstream's Property predicate for an ObjectPattern reports the VALUE when the two differ, and
// reports the shared name only when `properties` is on. The binding form draws the same line: the
// new binding is always ours, and the key is only ours when it IS the binding.
func idLengthBindingElementIsNamingSite(node *ast.Node, parent *ast.Node,
	settings IdLengthSettings) bool {

	element := parent.AsBindingElement()
	if element == nil {
		return false
	}
	// Only the BINDING is ours. In `var { a: b } = x` the key `a` is the source object's own
	// property name, so it is never reported even when it is too short -- measured against the
	// installed rule, `var { a: longEnough } = x` at min 2 is clean while `var { a: b } = x`
	// reports on `b` alone.
	//
	// One test rather than two. An earlier draft also returned early on `PropertyName == node`,
	// which reads as the explicit statement of that exclusion and is entirely redundant with this
	// line: a key is by definition not the name, so it exits here anyway. The mutation sweep is
	// what established the redundancy -- deleting the PropertyName guard alone survived every
	// fixture, while deleting both together is caught by six.
	if element.Name() != node {
		return false
	}
	// An array pattern's elements are positional, so no property name is involved and `properties`
	// does not gate them. Upstream lists `ArrayPattern: true`, unconditional.
	if parent.Parent != nil && parent.Parent.Kind == ast.KindArrayBindingPattern {
		return true
	}
	// A rest element likewise names something new rather than echoing a key.
	if element.DotDotDotToken != nil {
		return true
	}
	// A shorthand `var { a } = x` re-uses the object's key as the binding, so upstream governs it
	// with `properties`. A rename `var { p: q } = x` coins `q`, which is ours regardless.
	if element.PropertyName != nil {
		return true
	}
	return settings.CheckProperties
}

// idLengthPropertyIsNamingSite handles an object literal key and an ASSIGNMENT destructuring, which
// our parser spells with the same nodes.
func idLengthPropertyIsNamingSite(node *ast.Node, parent *ast.Node, settings IdLengthSettings) bool {
	objectLiteral := parent.Parent
	if objectLiteral == nil || objectLiteral.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	// An assignment destructuring: the object literal is standing in for a pattern.
	if idDenylistIsDestructuringTarget(objectLiteral) {
		switch parent.Kind {
		case ast.KindShorthandPropertyAssignment:
			// `({ a } = x)`: key and value are the same name, so upstream reports it only when
			// `properties` is on.
			assignment := parent.AsShorthandPropertyAssignment()
			return assignment != nil && assignment.Name() == node && settings.CheckProperties
		case ast.KindPropertyAssignment:
			// `({ p: q } = x)`: only the new binding is ours.
			assignment := parent.AsPropertyAssignment()
			return assignment != nil && assignment.Initializer == node
		}
		return false
	}

	// An ordinary object literal key, which upstream gates on `properties` and excludes for import
	// attributes and computed keys. A computed key's identifier has a ComputedPropertyName parent
	// and never reaches here.
	if !settings.CheckProperties {
		return false
	}
	if idDenylistIsImportAttributeKey(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		return assignment != nil && assignment.Name() == node
	case ast.KindShorthandPropertyAssignment:
		assignment := parent.AsShorthandPropertyAssignment()
		return assignment != nil && assignment.Name() == node
	}
	return false
}

// idLengthMemberIsWritten is upstream's MemberExpression predicate: the property is being written
// rather than read.
//
// Two shapes qualify, and the second is easy to miss: a plain assignment `o.a = 1`, and a
// destructuring whose target is a member access, `({ p: o.q } = {})`. Upstream spells the second as
// a Property inside an ObjectPattern whose parent is the left of an assignment.
func idLengthMemberIsWritten(access *ast.Node) bool {
	parent := access.Parent
	if parent == nil {
		return false
	}

	// `o.a = 1`
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		return binary != nil && binary.Left == access &&
			binary.OperatorToken.Kind == ast.KindEqualsToken
	}

	// `({ p: o.q } = {})`
	if parent.Kind == ast.KindPropertyAssignment {
		assignment := parent.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer != access {
			return false
		}
		objectLiteral := parent.Parent
		return objectLiteral != nil && objectLiteral.Kind == ast.KindObjectLiteralExpression &&
			idDenylistIsDestructuringTarget(objectLiteral)
	}
	return false
}
