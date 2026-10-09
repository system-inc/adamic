package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// voidDomElementsFile is where the fixtures pretend to live.
//
// A `.tsx` name because most of the corpus is JSX and would not parse otherwise. Upstream gates
// itself on `ctx.source_type().is_jsx()`, so a non-JSX file is skipped there before the rule runs.
// That gate is not reproduced here, and the reason is that it is not a judgment about code: it is
// oxc declining to run a JSX rule over a file whose parser never produces a JSX node. The
// `createElement` half of this rule decides about ordinary call expressions and is just as correct
// in a `.ts` file, and our harness offers a rule only the files the config selects, so a gate here
// would remove findings upstream removes only because it had already parsed the file differently.
const voidDomElementsFile = "/repository/source/VoidDomElements.tsx"

// TestVoidDomElementsNoChildrenFires covers every failing input upstream ships.
//
// Ten inputs and ten diagnostics in the snapshot, so it is one finding per input and there was no
// per-input count to recover. Both halves of the rule are represented: four JSX elements and six
// `createElement` calls, and within the calls both the positional third argument and the two
// property names that carry children inside the props object.
func TestVoidDomElementsNoChildrenFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a br wrapping text", "<br>Foo</br>;"},
		{"a br with a children prop", "<br children='Foo' />;"},
		{"an img with a spread and a children prop", "<img {...props} children='Foo' />;"},
		{"a br with the danger prop", "<br dangerouslySetInnerHTML={{ __html: 'Foo' }} />;"},
		{"createElement for a br with a third argument", "React.createElement('br', {}, 'Foo');"},
		{"createElement for a br with a children prop", "React.createElement('br', { children: 'Foo' });"},
		{"createElement for a br with the danger prop", "React.createElement('br', { dangerouslySetInnerHTML: { __html: 'Foo' } });"},
		{"a bare createElement for an img with a third argument", "\n                import React, {createElement} from 'react';\n                createElement('img', {}, 'Foo');\n            "},
		{"a bare createElement for an img with a children prop", "\n                import React, {createElement} from 'react';\n                createElement('img', { children: 'Foo' });\n            "},
		{"a bare createElement for an img with the danger prop", "\n                import React, {createElement} from 'react';\n                createElement('img', { dangerouslySetInnerHTML: { __html: 'Foo' } });\n            "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
		})
	}
}

// TestVoidDomElementsNoChildrenStaysSilent covers every clean input upstream ships.
//
// The shape of this list is the rule's whole discrimination written out: the same six inputs as the
// fail list with `div` in place of `br`, which pins that the element name is what decides, plus the
// arity cases (`createElement('img')` and `createElement()`) and the props-as-a-variable case that
// upstream declines because it reads only an object literal written at the call site.
func TestVoidDomElementsNoChildrenStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a div wrapping text", "<div>Foo</div>;"},
		{"a div with a children prop", "<div children='Foo' />;"},
		{"a div with the danger prop", "<div dangerouslySetInnerHTML={{ __html: 'Foo' }} />;"},
		{"createElement for a div with a third argument", "React.createElement('div', {}, 'Foo');"},
		{"createElement for a div with a children prop", "React.createElement('div', { children: 'Foo' });"},
		{"createElement for a div with the danger prop", "React.createElement('div', { dangerouslySetInnerHTML: { __html: 'Foo' } });"},
		{"createElement for an img with no props", "React.createElement('img');"},
		{"createElement with no arguments at all", "React.createElement();"},
		{"createElement for an img whose props are a variable", "\n                const props = {};\n                React.createElement('img', props);\n            "},
		{"a bare createElement for a div", "\n                import React, {createElement} from 'react';\n                createElement('div');\n            "},
		{"a bare createElement for an img", "\n                import React, {createElement} from 'react';\n                createElement('img');\n            "},
		{"a class component rendering a div", "\n                import React, {createElement, PureComponent} from 'react';\n                class Button extends PureComponent {\n                    handleClick(ev) {\n                        ev.preventDefault();\n                    }\n                    render() {\n                        return <div onClick={this.handleClick}>Hello</div>;\n                    }\n                }\n            "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenPointsAtTheElementName asserts where each finding lands.
//
// Message ids cannot see this and the two halves of the rule point at different things, so both are
// pinned. A JSX element reports the tag identifier alone, not the element and not the opening tag:
// upstream passes `identifier.span`. A `createElement` call reports the string literal in the first
// argument, quotes included, because upstream passes `element_name.span` and a `StringLiteral`'s
// span in oxc covers the whole token. Both were read off the release binary's own columns rather
// than inferred: `<br>Foo</br>;` prints 1:2 over two columns, and `React.createElement('br', ...)`
// prints 1:21 over four, which is `'br'` with its quotes and not `br`.
func TestVoidDomElementsNoChildrenPointsAtTheElementName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   string
	}{
		{"the tag identifier of an element wrapping text", "<br>Foo</br>;", "br"},
		{"the tag identifier of a self-closing element", "<br children='Foo' />;", "br"},
		{"the string literal including its quotes", "React.createElement('br', {}, 'Foo');", "'br'"},
		{"a double-quoted literal including its quotes", "React.createElement(\"br\", {}, \"Foo\");", "\"br\""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding covers %q, wanted %q", reported, testCase.reported)
			}
		})
	}
}

// TestVoidDomElementsNoChildrenNamesTheElementInTheMessage asserts the rendered text exactly.
//
// Equality rather than `strings.Contains`, because a substring predicate over an interpolated value
// is weaker than the property it guards: a message naming the wrong element still contains the
// right prefix. The name is interpolated from the tag, so the two halves are checked separately and
// the `createElement` half is checked against the literal's *value* rather than its source text,
// which is the one place the two could differ.
func TestVoidDomElementsNoChildrenNamesTheElementInTheMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantMessage string
	}{
		{"a JSX element", "<br>Foo</br>;", "Void DOM element <br /> cannot receive children."},
		{"a longer element name", "<menuitem>Foo</menuitem>;", "Void DOM element <menuitem /> cannot receive children."},
		{"a createElement call names the literal's value, without quotes", "React.createElement('img', {}, 'Foo');", "Void DOM element <img /> cannot receive children."},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.wantMessage {
				t.Errorf("message is %q, wanted %q", got, testCase.wantMessage)
			}
		})
	}
}

// TestVoidDomElementsNoChildrenCoversEveryVoidElement walks the whole list.
//
// Sixteen names, and the list is the rule's entire subject, so a typo in any one of them would make
// the rule quietly narrower on exactly one element and every other fixture would stay green.
// Upstream's corpus exercises only `br` and `img`, so fourteen of these sixteen are untested there.
func TestVoidDomElementsNoChildrenCoversEveryVoidElement(t *testing.T) {
	t.Parallel()

	voidElements := []string{
		"area", "base", "br", "col", "embed", "hr", "img", "input",
		"keygen", "link", "menuitem", "meta", "param", "source", "track", "wbr",
	}

	for _, elementName := range voidElements {
		t.Run(elementName, func(t *testing.T) {
			t.Parallel()
			sourceText := "<" + elementName + " children='Foo' />;"
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
		})
	}
}

// TestVoidDomElementsNoChildrenIsCaseSensitive pins that the comparison is exact.
//
// Measured on the release binary rather than assumed, and all three spellings are silent there:
// `<BR>Foo</BR>`, `<Br>Foo</Br>`, and `React.createElement('BR', {}, 'Foo')`. Upstream compares
// `identifier.name` against a lowercase list with no folding, and the reason it is right rather
// than merely faithful is that JSX resolves a capitalized tag to a component in scope: `<BR>` and
// `<Img>` render whatever those names are bound to, which is not the void HTML element and may
// legitimately take children. A port that folded case would report on component references.
func TestVoidDomElementsNoChildrenIsCaseSensitive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an all-caps tag is a component reference", "<BR>Foo</BR>;"},
		{"a capitalized tag is a component reference", "<Br>Foo</Br>;"},
		{"a capitalized component named for a void element", "<Img children='Foo' />;"},
		{"an all-caps string in createElement", "React.createElement('BR', {}, 'Foo');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenDeclinesSpreads pins that a spread contributes nothing.
//
// This one is worth stating carefully, because the natural reading of the code is the wrong one. A
// JSX spread is not invisible to a listener: it is a `JsxSpreadAttribute` sitting in the same
// properties list, and a rule that looked for that kind would see `<br {...{children: 1}} />`
// perfectly well. Upstream sees it too and deliberately answers false, matching
// `JSXAttributeItem::SpreadAttribute(_) => false`, and the same arm exists for the object literal
// in a `createElement` call as `ObjectPropertyKind::SpreadProperty(_) => false`.
//
// So the decline is a judgment rather than a blind spot, and it is reproduced rather than improved
// on. Reading a spread would mean deciding what an arbitrary expression evaluates to; the sibling
// rule `no-danger-with-children` does resolve spreads through the checker, and this rule does not,
// which is why this one declares no checker at all. Both are silent on an inline object literal
// spread, and this rule is additionally silent on the variable form that the sibling would follow.
//
// All three measured on the release binary: `<br {...{children: 1}} />`,
// `React.createElement('br', {...{children: 1}})` and `React.createElement('br', {...props})` are
// each silent there. Upstream's own corpus contains the case that proves a spread does not *stop*
// the search either: `<img {...props} children='Foo' />` reports, because the written attribute
// beside the spread still answers.
func TestVoidDomElementsNoChildrenDeclinesSpreads(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a JSX spread of an object literal carrying children", "<br {...{children: 1}} />;"},
		{"a JSX spread of a variable", "<br {...props} />;"},
		{"a spread property carrying children in the props object", "React.createElement('br', {...{children: 1}});"},
		{"a spread property naming a variable", "React.createElement('br', {...props});"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenReadsOnlyStaticIdentifierKeys pins which property keys answer.
//
// Upstream matches `PropertyKey::StaticIdentifier` alone and lets every other key shape fall
// through to false, so a quoted key and a computed key are both silent even though both name
// `children` unambiguously. That is narrower than the sibling rules in this package, which reach
// for `ast.TryGetTextOfPropertyName` and therefore resolve a string key and a computed key holding
// a literal. The narrowness is deliberate here and it is upstream's: measured on the release
// binary, `React.createElement('br', {'children': 'Foo'})` and
// `React.createElement('br', {['children']: 'Foo'})` are both silent.
//
// Stating it matters more than usual, because `TryGetTextOfPropertyName` is the house answer to
// reading a key and reaching for it here would silently widen the rule past upstream on two inputs
// no imported fixture covers. The command that established it is in the doc comment on the rule.
func TestVoidDomElementsNoChildrenReadsOnlyStaticIdentifierKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a quoted children key", "React.createElement('br', {'children': 'Foo'});"},
		{"a computed children key holding a literal", "React.createElement('br', {['children']: 'Foo'});"},
		{"a quoted danger key", "React.createElement('br', {'dangerouslySetInnerHTML': {__html: 'Foo'}});"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenTreatsAnyChildAsContent pins that children are counted, not read.
//
// The JSX half asks `!jsx_el.children.is_empty()` and nothing else, so it never looks at what a
// child contains. Two consequences, and both were measured on the release binary because both look
// like defects from the outside and neither is.
//
// A child that renders nothing still counts. `<br>\n</br>` written across two lines reports, and so
// does `<br>{}</br>`, because a formatting newline is a `JsxText` node and an empty expression
// container is a node too. The sibling rule `no-danger-with-children` filters exactly this shape
// out through `ast.IsWhitespaceOnlyJsxText`, and that difference between the two rules is real:
// upstream applies the whitespace test in that rule and not in this one. Reaching for the same
// predicate here because the neighbouring rule has it would make this rule silent on an input
// upstream reports.
//
// And an element with no children at all is clean even written with a closing tag, so `<br></br>`
// is silent while `<br> </br>` reports on the single space.
func TestVoidDomElementsNoChildrenTreatsAnyChildAsContent(t *testing.T) {
	t.Parallel()

	reports := []struct {
		name       string
		sourceText string
	}{
		{"a single space between the tags", "<br> </br>;"},
		{"a formatting newline between the tags", "<br>\n</br>;"},
		{"an empty expression container", "<br>{}</br>;"},
		{"an expression container holding a string", "<br>{'x'}</br>;"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
		})
	}

	t.Run("a closing tag with nothing between it and the opening tag", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, "<br></br>;")
		rule_testing.ExpectClean(t, result)
	})
}

// TestVoidDomElementsNoChildrenDeclinesNonIdentifierTags pins which tag shapes participate.
//
// Upstream destructures `JSXElementName::Identifier` and returns on anything else, so a member
// expression and a namespaced name never reach the void-element list. `<Foo.br>x</Foo.br>` is
// silent on the release binary, which is the right answer rather than a gap: a dotted tag resolves
// to a component property and renders whatever that is.
//
// This is also the shape that panics if a rule reads a tag's text before checking its kind, because
// a JSX member-expression tag parses to a `PropertyAccessExpression` here and `Node.Text()` panics
// outright on that kind. The rule asks `jsx.ElementParts` and then checks for `KindIdentifier`, so
// the kind check is load-bearing against a crash as well as against a wrong verdict.
func TestVoidDomElementsNoChildrenDeclinesNonIdentifierTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a member expression tag", "<Foo.br>x</Foo.br>;"},
		{"a namespaced attribute name never matches children", "<br xlink:children='Foo' />;"},
		{"a dashed attribute name is not the children prop", "<br data-children='Foo' />;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenAcceptsEveryCalleeSpellingUpstreamDoes pins the callee test.
//
// This rule calls oxc's shared `is_create_element_call`, unlike `no-danger-with-children` beside it
// which writes a narrower test inline, so the two rules genuinely disagree about which calls they
// look at and the shelf helper is the right one here for that reason. All four spellings were
// measured on the release binary and all four report: a bare `createElement`, a dotted member on
// any object rather than only `React`, a computed member, and `document` excluded by name.
func TestVoidDomElementsNoChildrenAcceptsEveryCalleeSpellingUpstreamDoes(t *testing.T) {
	t.Parallel()

	reports := []struct {
		name       string
		sourceText string
	}{
		{"a bare createElement", "createElement('br', {}, 'Foo');"},
		{"a computed member", "React['createElement']('br', {}, 'Foo');"},
		{"an object that is not React", "Preact.createElement('br', {}, 'Foo');"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
		})
	}

	// `document.createElement` builds a DOM node rather than a React element, and it is the one
	// call sharing the property name. Upstream excludes it by name in both member arms.
	t.Run("document.createElement is excluded by name", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, "document.createElement('br', {}, 'Foo');")
		rule_testing.ExpectClean(t, result)
	})
}

// TestVoidDomElementsNoChildrenMatchesUpstreamOnParenthesizedCallees is the paren measurement.
//
// Measuring this was mandatory rather than optional and it found a real divergence, so the rule
// carries a deliberate narrowing that the shelf helper would otherwise undo.
//
// `internal/utilities/react.IsCreateElementCall` runs `ast.SkipParentheses` over the callee before
// matching it. Upstream does not: `is_create_element_call` matches `call_expr.callee` against
// `Expression::StaticMemberExpression`, `ComputedMemberExpression` and `Identifier` directly, and a
// parenthesized expression is none of those in oxc's AST, so it falls to the `_ => false` arm.
// Measured on the release binary rather than reasoned about: `(React.createElement)('br', {}, 'Foo')`
// and `(createElement)('br', {}, 'Foo')` are both silent there, while the unparenthesized forms of
// both report. Calling the shelf helper unguarded would therefore make this rule report on two
// inputs upstream is silent on, and no imported fixture could catch it because the corpus writes no
// parenthesized callee at all.
//
// The rule checks the callee's kind itself before handing the call to the helper, which reproduces
// upstream's fall-through without reimplementing the three accepted spellings. The helper's own
// paren-skipping stays correct for the rules that read it the way oxc's other callers do; this is a
// difference between callers, not a defect in the helper, and it is noted here rather than changed
// there because other rules in this package depend on the current behavior.
//
// The first argument is a separate question and falls the same way: `React.createElement(('br'), ...)`
// is silent, because upstream matches `Argument::StringLiteral` and a parenthesized literal is not
// one. That needs no special handling here since the rule checks the argument's kind directly.
func TestVoidDomElementsNoChildrenMatchesUpstreamOnParenthesizedCallees(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a parenthesized member callee", "(React.createElement)('br', {}, 'Foo');"},
		{"a parenthesized bare callee", "(createElement)('br', {}, 'Foo');"},
		{"a parenthesized element name argument", "React.createElement(('br'), {}, 'Foo');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenCountsArgumentsBeforeReadingThem pins the arity gates.
//
// Upstream applies four gates in order and each one is a separate early return, so each is worth a
// case. No arguments at all, a first argument that is not a string literal, fewer than two
// arguments, and a second argument that is not an object literal all decline before the property
// search happens. The `null` and non-object cases are the ones that look wrong from the outside:
// `React.createElement('br', null, 'Foo')` really does pass children positionally and really is
// silent upstream, because the second argument must destructure as an `ObjectExpression` before the
// third is ever consulted. Measured on the release binary, and reproduced rather than improved on.
func TestVoidDomElementsNoChildrenCountsArgumentsBeforeReadingThem(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a null props argument with a third argument", "React.createElement('br', null, 'Foo');"},
		{"a props argument that is not an object", "React.createElement('br', 'notanobject', 'Foo');"},
		{"a first argument that is not a string literal", "React.createElement(Component, {}, 'Foo');"},
		{"a template literal element name", "React.createElement(`br`, {}, 'Foo');"},
		{"a props variable rather than a literal", "const props = {children: 'Foo'}; React.createElement('br', props);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestVoidDomElementsNoChildrenReportsOncePerElement pins the count.
//
// An element carrying children two ways at once is still one finding, because the two conditions
// are a single disjunction feeding a single report rather than two independent checks. Upstream's
// corpus never writes an element that satisfies both, so nothing imported pins it. Measured on the
// release binary: `<br dangerouslySetInnerHTML={{__html: 'Foo'}}>Foo</br>` reports once, and
// `React.createElement('br', {children: 'Foo'}, 'Bar')` reports once.
func TestVoidDomElementsNoChildrenReportsOncePerElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the danger prop and children between the tags", "<br dangerouslySetInnerHTML={{ __html: 'Foo' }}>Foo</br>;"},
		{"a children prop and children between the tags", "<br children='Foo'>Foo</br>;"},
		{"a children prop and a positional third argument", "React.createElement('br', { children: 'Foo' }, 'Bar');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidDomElementsNoChildren, voidDomElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "voidDomElementsNoChildren")
		})
	}
}
