package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageForbiddenTypeDefaultParam builds the finding for one offending default.
//
// Upstream interpolates two placeholders and uses the forbidden type's name twice, including the
// ungrammatical "a/an" that avoids choosing an article. Both are reproduced verbatim rather than
// tidied, because the text is what a reader greps for and a paraphrase makes this rule's findings
// stop matching every other tool's.
//
// A rendered-text assertion covers this in the test file. A message-id assertion could not see any
// of it: an interpolation defect leaves the id and the count untouched.
func messageForbiddenTypeDefaultParam(propertyName string, forbiddenType string) rule.Message {
	return rule.Message{
		Id: "forbiddenTypeDefaultParam",
		Description: propertyName + " has a/an " + forbiddenType +
			" as default prop. This could lead to potential infinite render loop in React. " +
			"Use a variable reference instead of " + forbiddenType + ".",
	}
}

// NoObjectTypeAsDefaultProp reports a referential value used as a destructured prop default.
//
//	valid:   function C({a = 1}) { return <div/>; }
//	valid:   function C({a = someReference}) { return <div/>; }
//	valid:   function helper({a = {}}) { return 1; }, not a component
//	invalid: function C({a = {}}) { return <div/>; }
//	invalid: const C = ({a = []}) => <div/>;
//
// A fresh object, array, function, or element is a new reference on every render, so a child
// memoized on that prop re-renders every time and an effect depending on it never settles. A
// primitive default is fine because it compares equal.
//
// # What counts as forbidden
//
// Nine shapes, measured one per run against the installed build on 2026-08-27: an object literal,
// an array literal, an arrow function, a function expression, a class expression, a construction
// expression, a JSX element, a regular expression literal, and a call to `Symbol`.
//
// Two of those are special cases rather than node kinds and both are narrow. A regular expression
// is a literal upstream, separated from a string by the presence of a `regex` field. And the Symbol
// case keys on the CALLEE being the bare identifier `Symbol`, so `Symbol.for(1)` is silent,
// measured. Everything else is silent, including a plain call, a template literal, a tagged
// template, a member access, and every primitive.
//
// # Only the first parameter, and only its top level
//
// Upstream reads `component.node.params[0]` and requires it to be an object pattern, then loops its
// own properties without recursing. So a destructured SECOND parameter is silent, and a nested
// pattern like `{x: {a = {}}}` is silent too. Both measured. That looks like an oversight and it is
// reproduced rather than corrected, because widening it changes which files report.
//
// # No fix
//
// The repair is to hoist the value to a module constant and name it, which means choosing that name
// and a place to put it. Upstream ships no fixer either.
var NoObjectTypeAsDefaultProp = rule.Rule{
	Name:             "react/no-object-type-as-default-prop",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The checker is reached through the component detector, which resolves a bare
				// `createElement` back to a React binding. A nil checker would make this rule
				// silently narrower rather than crash, which is the more dangerous failure, so it
				// declines outright and a test asserts the typed harness is required.
				if ctx.TypeChecker == nil {
					return
				}

				for _, component := range collectDetectedComponents(ctx, node) {
					reportForbiddenDefaults(ctx, component.node)
				}
			},
		}
	},
}

// reportForbiddenDefaults checks one component's first parameter for offending defaults.
func reportForbiddenDefaults(ctx rule.Context, component *ast.Node) {
	parameters := componentParameters(component)
	if parameters == nil || len(parameters.Nodes) == 0 {
		return
	}

	// Upstream requires `params[0].type === 'ObjectPattern'`, so a component destructuring only its
	// second parameter is silent.
	pattern := parameters.Nodes[0].AsParameterDeclaration().Name()
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return
	}

	// Every element of an object binding pattern is a KindBindingElement, so there is no kind
	// guard here. An earlier draft carried one and a mutation sweep could not kill it; probing the
	// parser over eight shapes settled why. Shorthand, renamed, rest, string key, computed key, an
	// elision, and two malformed sources that error recovery had to repair all produce that one
	// kind and nothing else. If a future parser change introduces a second kind, `AsBindingElement`
	// is where it would surface.
	for _, element := range pattern.AsBindingPattern().Elements.Nodes {
		binding := element.AsBindingElement()

		// No default means no fresh reference, and upstream's filter requires an assignment
		// pattern before it looks at anything else. That decline happens inside
		// `forbiddenDefaultTypeOf`, which takes a nil initializer through
		// `skipParenthesesOptional` unchanged and rejects it on its own nil check, so an explicit
		// guard here would be the same decision written twice. An earlier draft carried one and a
		// mutation sweep showed no input could distinguish the two versions.
		initializer := binding.Initializer

		// Upstream reads `prop.key.name`, which is the property name in the object pattern. A
		// shorthand `{a = {}}` has no explicit key, so the bound name is the key. A computed key
		// yields no name upstream and is skipped here for the same reason.
		// `destructuredPropertyName` answers for every shape a binding element with an initializer
		// can take, so its second return is not tested here. Enumerated on 2026-08-27 across
		// fourteen shapes with the predicate run directly: shorthand, renamed, string key, numeric
		// key, computed identifier, computed call, computed string, computed arithmetic, computed
		// member, an empty computed key, a template key, nested object and array patterns, and a
		// rest element. None answered false.
		//
		// An earlier draft declined on that second return and was silently narrower than upstream
		// on four computed-key shapes, which a surviving mutant on this branch exposed. The verdict
		// above is what the branch would now be testing, and it names the callers it rests on: if
		// `destructuredPropertyName` grows an arm that can answer false, this needs the guard back.
		propertyName, _ := destructuredPropertyName(binding)

		forbiddenType, isForbidden := forbiddenDefaultTypeOf(initializer)
		if !isForbidden {
			continue
		}

		// Upstream reports on `prop.value`, the whole assignment pattern, so the span covers the
		// name, the equals sign and the value rather than the value alone. Measured by slicing the
		// installed build's reported range: `a = {}` for `function C({a = {}, b = []})`.
		ctx.ReportNode(element, messageForbiddenTypeDefaultParam(propertyName, forbiddenType))
	}
}

// componentParameters returns the parameter list of a component node, or nil.
//
// The detector hands back a pragma wrapper CALL for a memo or forwardRef component and a class or
// an object literal for the older forms, none of which carry parameters. Upstream reads
// `component.node.params` and gets undefined for those, so `hasUsedObjectDestructuringSyntax`
// declines them; returning nil here is the same decline.
func componentParameters(component *ast.Node) *ast.NodeList {
	switch component.Kind {
	case ast.KindFunctionDeclaration:
		return component.AsFunctionDeclaration().Parameters
	case ast.KindFunctionExpression:
		return component.AsFunctionExpression().Parameters
	case ast.KindArrowFunction:
		return component.AsArrowFunction().Parameters
	case ast.KindMethodDeclaration:
		return component.AsMethodDeclaration().Parameters
	}
	return nil
}

// destructuredPropertyName returns the name upstream would read from `prop.key.name`.
//
// A shorthand element carries no property name node and the bound identifier is the key. An
// explicit `{key: value = default}` carries both, and the KEY is what upstream names in the
// message, not the local binding.
func destructuredPropertyName(binding *ast.BindingElement) (string, bool) {
	if key := binding.PropertyName; key != nil {
		switch key.Kind {
		case ast.KindIdentifier:
			return key.Text(), true

		case ast.KindComputedPropertyName:
			// Upstream reads `prop.key.name`, and for a computed key the key node IS the inner
			// expression, so `{[k]: a = {}}` renders as `k`. Measured: it reports and names `k`.
			inner := key.AsComputedPropertyName().Expression
			if inner != nil && inner.Kind == ast.KindIdentifier {
				return inner.Text(), true
			}
			// A computed key that is not a plain identifier still REPORTS upstream, rendering the
			// word "undefined". Measured on 2026-08-27 over `[k()]`, `['lit']`, `[1+1]` and
			// `[o.p]`: all four report. An earlier draft of this port declined them and was
			// silently narrower on four real shapes; a surviving mutant on this branch is what
			// exposed it, after a first hypothesis about why it survived turned out to be wrong.
			//
			// The name rendered is the key's own source text, for the reason given in the string
			// and numeric arm below.
			return computedKeySourceText(key), true

		case ast.KindStringLiteral, ast.KindNumericLiteral:
			// # A stated divergence, in the message text only
			//
			// Upstream reports these and renders the literal word "undefined" as the property
			// name, because a string or numeric key node has no `name` field for `prop.key.name`
			// to read. Measured on 2026-08-27: `{'str': a = {}}` and `{1: a = {}}` both report,
			// both saying "undefined has a/an object literal as default prop".
			//
			// Which inputs report is reproduced exactly. What the message SAYS is not: rendering
			// the word "undefined" into user-facing text is a defect this port can see and does
			// not have to carry, and our node holds the real key. So these render `str` and `1`.
			//
			// Recorded here rather than left silent because the next reader comparing our output
			// against upstream's on such a file will see a difference and needs to know it was
			// authored. `TestNoObjectTypeAsDefaultPropRendersRealKeyNamesWhereUpstreamSaysUndefined`
			// pins both halves.
			return key.Text(), true
		}
		return "", false
	}

	name := binding.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.Text(), true
}

// forbiddenDefaultTypeOf names the offending kind of a default value, or reports that it is fine.
//
// The returned string is interpolated into the message twice, so these are upstream's own labels
// rather than descriptions of our node kinds.
func forbiddenDefaultTypeOf(initializer *ast.Node) (string, bool) {
	// Our parser keeps parentheses where upstream's folds them away, so `{a = ({})}` would reach
	// upstream's switch as the object literal itself. Unwrapped in a loop because `(({}))` nests.
	initializer = skipParenthesesOptional(initializer)
	if initializer == nil {
		return "", false
	}

	switch initializer.Kind {
	case ast.KindObjectLiteralExpression:
		return "object literal", true
	case ast.KindArrayLiteralExpression:
		return "array literal", true
	case ast.KindArrowFunction:
		return "arrow function", true
	case ast.KindFunctionExpression:
		return "function expression", true
	case ast.KindClassExpression:
		return "class expression", true
	case ast.KindNewExpression:
		return "construction expression", true
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		// A fragment is deliberately absent from this arm. Upstream's forbidden set names
		// `JSXElement`, and its parser spells a fragment as `JSXFragment`, a different type, so
		// `{a = <></>}` is silent there even though it allocates a fresh element on every render
		// exactly like `{a = <div/>}` does.
		//
		// Measured against the installed build on 2026-08-27: the element form reports and the
		// fragment form does not. This reads as an upstream oversight and it is reproduced rather
		// than corrected. An earlier draft of this port included the fragment as an obvious
		// improvement and over-reported on it, and the differential table is what caught it, since
		// upstream's corpus writes no fragment default at all.
		return "JSX element", true

	case ast.KindRegularExpressionLiteral:
		// Upstream reaches this through `Literal` plus a non-null `regex` field, because its parser
		// spells every literal as one node kind. Ours gives a regular expression its own kind, so
		// the same decision is one case arm rather than a field test.
		return "regex literal", true

	case ast.KindCallExpression:
		// Only a call to the bare identifier `Symbol`. Upstream tests
		// `callee.type === 'Identifier' && callee.name === 'Symbol'`, so `Symbol.for(1)` is a
		// member callee and is silent. Measured against the installed build.
		callee := skipParenthesesOptional(initializer.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "Symbol" {
			return "Symbol literal", true
		}
	}
	return "", false
}

// computedKeySourceText renders a computed property key that has no plain name.
//
// Upstream renders the word "undefined" for these, which is a defect in text a person reads. The
// key's own source is used instead, so `{[k()]: a = {}}` names `k()`. Which inputs report is
// unaffected; only the rendered name differs. See destructuredPropertyName for the full note.
func computedKeySourceText(key *ast.Node) string {
	inner := key.AsComputedPropertyName().Expression
	if inner == nil {
		return "[]"
	}
	sourceFile := ast.GetSourceFileOfNode(key)
	if sourceFile == nil {
		return "[]"
	}
	// The range is taken from the node rather than read through `Text()`, which panics on several
	// expression kinds and would take every rule in this package down for the whole file.
	start, end := inner.Pos(), inner.End()
	if start < 0 || end > len(sourceFile.Text()) || start >= end {
		return "[]"
	}
	return text.TrimWhitespace(sourceFile.Text()[start:end])
}
