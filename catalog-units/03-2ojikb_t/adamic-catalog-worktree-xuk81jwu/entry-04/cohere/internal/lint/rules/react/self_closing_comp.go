package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// SelfClosingCompOptions configures which element kinds the rule judges.
//
// Both keys default to TRUE upstream, which is why this does not use `rule.DecodeOptionsInto`: the
// generic helper would hand back a zero-value struct on absent input, both fields false, and the
// rule would silently report nothing while every fixture built from a struct still passed.
type SelfClosingCompOptions struct {
	// Component judges elements whose tag does not begin with a lowercase letter, which is the
	// custom-component half. Absent means true.
	Component bool

	// Html judges elements whose tag begins with a lowercase letter, which upstream calls a DOM
	// component. Absent means true.
	Html bool
}

// selfClosingCompWire is the on-the-wire shape, with pointers so an absent key stays distinguishable
// from an explicit false.
//
// This distinction is the whole reason the decoder is hand-rolled. With plain bools, `{}` and
// `{"component": false, "html": false}` decode identically and the rule turns itself off.
type selfClosingCompWire struct {
	Component *bool `json:"component"`
	Html      *bool `json:"html"`
}

// DefaultSelfClosingCompOptions is the unconfigured answer: both halves on.
//
// Upstream spells it `Object.assign({}, {component: true, html: true}, context.options[0])`, so an
// absent option object and an empty one are the same rule. The corpus states that directly: it
// ships the same sources under no options and under `[]`, both reporting.
func DefaultSelfClosingCompOptions() SelfClosingCompOptions {
	return SelfClosingCompOptions{Component: true, Html: true}
}

// DecodeSelfClosingCompOptions reads this rule's configuration from the config layer.
//
// Our config layer unwraps the severity tuple before dispatch, so this receives upstream's option
// OBJECT rather than upstream's one-element array. Empty input is a rule configured as a bare
// `"error"` and resolves to the default.
func DecodeSelfClosingCompOptions(raw []byte) (any, error) {
	options := DefaultSelfClosingCompOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire selfClosingCompWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	if wire.Component != nil {
		options.Component = *wire.Component
	}
	if wire.Html != nil {
		options.Html = *wire.Html
	}
	return options, nil
}

var messageNotSelfClosing = rule.Message{
	Id:          "notSelfClosing",
	Description: "Empty components are self-closing",
}

// SelfClosingComp reports an element written with a separate closing tag when it has no children.
//
//	valid:   <div>text</div>
//	valid:   <div> </div>, a single space on one line
//	valid:   <div />
//	invalid: <div></div>            fixed to <div />
//	invalid: <div>\n  \n</div>      fixed to <div />
//
// # Which elements each option covers
//
// Upstream splits on `isDOMComponent`, which is the regular expression `/^[a-z]/` over the tag
// name. So the split is purely "does the tag start with a lowercase letter", and it lands in places
// that read oddly: `<foo.bar>` is html, `<my-element>` is html, `<_foo>` is component, and
// `<svg:rect>` is html. All measured against the installed build on 2026-08-27.
//
// Both options default true, so under the default configuration the split is invisible and every
// empty element reports. It only matters once one half is turned off.
//
// # Which children count as empty
//
// Either no children at all, or exactly one text child that CONTAINS A NEWLINE and is otherwise
// whitespace. The newline requirement is not decoration: `<div> </div>` on one line is silent while
// `<div>\n</div>` reports, measured both ways.
//
// A non-breaking space is deliberately not whitespace here. Upstream strips with
// `/(?!\xA0)\s/g`, whose negative lookahead spares U+00A0, so a text child holding only newlines
// and a non-breaking space keeps a non-empty remainder and is silent. Measured against the
// installed build. That character is written on purpose when it is written at all, which is why
// upstream preserves it, and a port using a plain whitespace trim would delete somebody's
// deliberate spacing along with the closing tag.
//
// # The fix
//
// Upstream replaces from the opening tag's own `>` through the closing tag's `>` with ` />`, which
// collapses the whole `></div>` tail in one edit and normalizes the spacing before the slash. The
// corpus asserts the resulting text on all twelve reporting cases and every one is reproduced.
var SelfClosingComp = rule.Rule{
	Name: "react/self-closing-comp",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultSelfClosingCompOptions()
		if configured, isConfigured := rule.OptionsAs[SelfClosingCompOptions](options); isConfigured {
			settings = configured
		}

		return rule.Listeners{
			ast.KindJsxElement: func(node *ast.Node) {
				element := node.AsJsxElement()
				opening, closing := element.OpeningElement, element.ClosingElement
				if opening == nil || closing == nil {
					return
				}

				// A self-closing element is a different node kind here, so it never reaches this
				// listener at all. Upstream needs an explicit `!node.selfClosing` because both
				// forms arrive at its single JSXOpeningElement handler.
				tagName := opening.AsJsxOpeningElement().TagName
				if !selfClosingCompJudges(settings, tagName) {
					return
				}
				if !selfClosingCompChildrenAreEmpty(element.Children) {
					return
				}

				// The replaced range runs from the opening tag's final `>` through the closing
				// tag's final `>`, which is upstream's `[node.range[1] - 1, closingElement.range[1]]`.
				// Taking it from the node ends rather than by scanning for the character keeps it
				// correct when the tag spans lines.
				replaced := core.NewTextRange(opening.End()-1, closing.End())
				ctx.ReportNodeWithFixes(opening, messageNotSelfClosing,
					rule.ReplaceRange(replaced, " />"))
			},
		}
	},
}

// selfClosingCompJudges reports whether the configuration covers this element's tag.
//
// Upstream's `isComponent` additionally requires the name node to be a JSX identifier or a member
// expression, which excludes a namespaced name from the component half. But a namespaced name is
// also not a DOM component unless it starts lowercase, and `elementType` renders `svg:rect` with
// its namespace intact, so `<svg:rect>` lands in the html half and `<SVG:rect>` in neither.
// Measured: `<svg:rect></svg:rect>` reports under the default and under `{component: false}`.
func selfClosingCompJudges(settings SelfClosingCompOptions, tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}
	if selfClosingCompIsDomTag(tagName) {
		return settings.Html
	}
	// A namespaced name that is not lowercase-first belongs to neither half, matching upstream's
	// `isComponent` declining a JSXNamespacedName.
	if tagName.Kind == ast.KindJsxNamespacedName {
		return false
	}
	return settings.Component
}

// selfClosingCompIsDomTag reports whether a tag name begins with a lowercase letter, which is
// upstream's `COMPAT_TAG_REGEX` of `/^[a-z]/` applied to the rendered element type.
//
// The test is on the FIRST character of the whole rendered name, so a member expression is judged
// by its leftmost segment: `<foo.bar>` is a DOM component and `<Foo.Bar>` is not. Both measured.
//
// The range is ASCII a through z specifically, matching the regular expression rather than Unicode
// lowercase, so a tag beginning with a lowercase non-ASCII letter is a component rather than html.
// No corpus case writes one and no rule behaviour turns on it under the default configuration,
// where both halves are on.
func selfClosingCompIsDomTag(tagName *ast.Node) bool {
	rendered := selfClosingCompElementType(tagName)
	if rendered == "" {
		return false
	}
	first := rendered[0]
	return first >= 'a' && first <= 'z'
}

// selfClosingCompElementType renders a tag name the way upstream's `elementType` does.
//
// Only the leading character is ever read, so the interesting part is which node contributes it: a
// member expression contributes its leftmost object, and a namespaced name contributes its
// namespace.
func selfClosingCompElementType(tagName *ast.Node) string {
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.Text()

	case ast.KindPropertyAccessExpression:
		// `<Foo.Bar.Baz />` nests to the left, so the loop walks to the leftmost object.
		expression := tagName.AsPropertyAccessExpression().Expression
		for expression != nil && expression.Kind == ast.KindPropertyAccessExpression {
			expression = expression.AsPropertyAccessExpression().Expression
		}
		if expression != nil && expression.Kind == ast.KindIdentifier {
			return expression.Text()
		}

	case ast.KindJsxNamespacedName:
		if namespace := tagName.AsJsxNamespacedName().Namespace; namespace != nil {
			return namespace.Text()
		}

	case ast.KindThisKeyword:
		// `<this.Foo />` renders as `this`, which is lowercase and therefore html upstream.
		return "this"
	}
	return ""
}

// selfClosingCompChildrenAreEmpty reports whether the element has nothing between its tags.
//
// Two accepting shapes, reproducing upstream's `childrenIsEmpty` and `childrenIsMultilineSpaces`.
// No children at all, or exactly one text child that contains a newline and is whitespace apart
// from any non-breaking space.
//
// # Why this is not the parser's own ContainsOnlyTriviaWhiteSpaces
//
// That field answers what looks like the identical question, and its own documentation states the
// JSX rule exactly as upstream does: whitespace containing a newline is not rendered, whitespace on
// one line is. Reaching for it is the obvious move and it is wrong here by one character.
//
// Probed on 2026-08-27 across seven text shapes with the two predicates side by side. They agree on
// every row except a text child of newline, non-breaking space, newline, where the parser answers
// TRUE and upstream is silent. Taking the parser's answer would report that element and then apply
// a fix that DELETES the non-breaking space along with the closing tag, which is the one character
// upstream's negative lookahead exists to protect and the one somebody typed on purpose.
//
// So the predicate is hand-rolled, and the divergence is recorded rather than the helper adopted.
func selfClosingCompChildrenAreEmpty(children *ast.NodeList) bool {
	if children == nil || len(children.Nodes) == 0 {
		return true
	}
	if len(children.Nodes) != 1 {
		return false
	}

	child := children.Nodes[0]
	if child.Kind != ast.KindJsxText {
		return false
	}
	// `ast.Node.Text()` PANICS on a JsxText node, and the walk recovers per file rather than per
	// rule, so one such call costs every rule in this package every finding in that file while the
	// run still prints a plausible summary. The field is read directly instead. Pinned by
	// TestSelfClosingCompDoesNotPanicOnJsxText, because no findings assertion can see a panic.
	text := child.AsJsxText().Text
	if !strings.Contains(text, "\n") {
		return false
	}

	// Upstream's `/(?!\xA0)\s/g` removes every whitespace character EXCEPT the non-breaking space,
	// then requires the remainder to be empty. So a non-breaking space survives the strip and makes
	// the child non-empty, which is how a deliberate one is preserved rather than deleted by the
	// fix. Reproduced by treating it as a non-whitespace character here.
	// The non-breaking space needs no test of its own here: it is deliberately absent from
	// `isJavaScriptWhitespace`, so it already falls out as content. An earlier draft carried an
	// explicit `character == '\u00a0'` guard above this loop and a mutation sweep showed it could
	// not change any answer, which is the same decision written in two places. The single site is
	// the whitespace set, and its comment is where the reasoning lives.
	for _, character := range text {
		if !isJavaScriptWhitespace(character) {
			return false
		}
	}
	return true
}

// isJavaScriptWhitespace reports whether a rune is matched by the regular expression class `\s` in
// JavaScript, which is what upstream strips with.
//
// Spelled out as code points rather than deferred to `unicode.IsSpace`, for two reasons. The two
// sets disagree: JavaScript's class includes the byte order mark and Go's does not. And a literal
// whitespace character written into Go source is invisible to review and, in the byte order mark's
// case, does not compile at all.
//
// The non-breaking space, U+00A0, is deliberately ABSENT from this set even though JavaScript's
// class matches it, because upstream's negative lookahead spares it. Its exclusion is what makes a
// deliberate non-breaking space count as content; see the caller.
func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ',
		'\u0085',                     // next line
		'\u1680',                     // ogham space mark
		'\u2000', '\u2001', '\u2002', // en quad, em quad, en space
		'\u2003', '\u2004', '\u2005', // em space, three-per-em, four-per-em
		'\u2006', '\u2007', '\u2008', // six-per-em, figure space, punctuation space
		'\u2009', '\u200a', // thin space, hair space
		'\u2028', '\u2029', // line separator, paragraph separator
		'\u202f', // narrow no-break space
		'\u205f', // medium mathematical space
		'\u3000', // ideographic space
		'\ufeff': // byte order mark
		return true
	}
	return false
}
