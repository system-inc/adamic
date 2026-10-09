// Package jsx answers the questions a rule asks about a JSX element: what tag it names, what
// attributes it writes, and what a given attribute's text is.
//
// Lifted out of `internal/rules/next/` on 2026-08-23, where it had been shared across five
// `@next/next` rules and was invisible to every other namespace. The eighteen ordinary
// `eslint-plugin-react` rules ask the same questions, and left where it was, `react` would have
// re-decided attribute matching and the two would have drifted permanently. That drift is the
// thing this shelf exists to prevent.
//
// It lives under `internal/utilities/` because `TestRulePackagesStayLeaves` refuses an import that
// pushes a rule package off the leaf of the graph, and that prefix is on the allowlist: this
// package moves when JSX handling changes, not when a rule is written.
package jsx

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// Attribute names are matched two ways because upstream matches them two ways, and which one a
// rule uses is part of what that rule decides rather than a detail of how it looks things up.
//
// `no-css-tags`, `no-sync-scripts` and `no-img-element` compare exactly. `google-font-preconnect`
// and `google-font-display` go through oxc's `has_jsx_prop_ignore_case`. Both spellings are
// preserved per rule rather than unified, because unifying would mean choosing for upstream: made
// case-insensitive everywhere, a rule would start reporting `<link REL="stylesheet">` that its
// original ignores; made exact everywhere, one would stop reporting what its original catches.
//
// JSX itself is case-sensitive, so the insensitive matchers describe upstream's leniency rather
// than anything the language requires.
type NameMatch func(candidate string, wanted string) bool

func MatchExactly(candidate string, wanted string) bool {
	return candidate == wanted
}

func MatchIgnoringCase(candidate string, wanted string) bool {
	return strings.EqualFold(candidate, wanted)
}

// AttributeName returns the plain identifier name a JSX attribute is written with.
//
// The second return separates "this is not a named attribute" from an attribute whose name happens
// to be empty, which a caller reporting on the name node has to be able to tell apart.
//
// Three shapes decline, and each is a real thing a rule will meet:
//
//   - a spread (`<div {...properties} />`), which is not a JsxAttribute and carries no name
//   - a namespaced name (`<svg xlink:href=... />`), matching oxc, which destructures
//     `JSXAttributeName::Identifier` everywhere and declines anything else
//   - a nil name, which the parser can produce on malformed input
//
// This exists because the same three-part guard was written twice inside this file, at the two
// loops below, and `@system_cohere_format` found a third rule about to hand-roll it: a rule that
// reports *on the attribute* rather than on the element needs the node, and neither element-level
// helper gives it one. Two copies of a guard inside the package whose doc comment says the shelf
// prevents drift is the drift, so it is one function now and both loops call it.
func AttributeName(property *ast.Node) (string, bool) {
	if property == nil || property.Kind != ast.KindJsxAttribute {
		return "", false
	}
	name := property.AsJsxAttribute().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	// The flag is unconditionally true once the kind check has passed, and a sweep confirms it: an
	// identifier attribute name cannot be empty in JSX the parser accepted, so making the flag
	// depend on the text changes no answer. It stays a plain true because the question the flag
	// answers is "is this a named attribute", which the kind check has already decided.
	return name.Text(), true
}

// StringAttributeValue returns the text of a JSX attribute whose value is a plain string literal.
//
// The second return reports whether such an attribute was found at all, which callers need in order
// to tell "absent" apart from "present and empty".
//
// Only a string literal answers. `href={someExpression}` and a spread (`<link {...props} />`) both
// decline, and that is upstream's judgment rather than a simplification: a rule that decides from
// the text of an attribute cannot decide anything about a value it would have to evaluate, so
// guessing would report on code whose real value is unknown. oxc expresses the same rule as
// `get_prop_value(...).as_string_literal()`.
//
// The value is decoded as ESLint's parser decodes it: typescript-estree runs a JSX attribute's
// string through unescapeStringLiteralText, so `href="&#47;about"` reads `/about` to an ESLint rule.
// Returning the raw text made every rule reading an attribute this way disagree with ESLint on an
// encoded value (#51y9jh2).
//
// Shared rather than copied because several rules read an attribute exactly this way, and a second
// implementation would be free to drift on the spread case, which is the one a porter is most
// likely to miss.
func StringAttributeValue(attributes *ast.Node, name string, matches NameMatch) (string, bool) {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return "", false
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return "", false
	}

	for _, property := range properties.Nodes {
		// A spread carries no name to match and no literal to read. It is skipped rather than
		// treated as absent, so an element that spreads its attributes simply has nothing to say
		// about this one.
		if property.Kind != ast.KindJsxAttribute {
			continue
		}

		attributeName, named := AttributeName(property)
		if !named || !matches(attributeName, name) {
			continue
		}
		attribute := property.AsJsxAttribute()
		if attribute.Initializer == nil || attribute.Initializer.Kind != ast.KindStringLiteral {
			return "", false
		}
		return text.UnescapeStringLiteralText(attribute.Initializer.Text()), true
	}

	return "", false
}

// ElementParts returns the tag name and attributes of either JSX opening form.
//
// A self-closing `<link />` parses as a JsxSelfClosingElement and never produces a
// JsxOpeningElement, and it is the shape most of these elements are written in, so a rule that
// reads only one form is silent on the common case.
func ElementParts(node *ast.Node) (tagName *ast.Node, attributes *ast.Node) {
	switch node.Kind {
	case ast.KindJsxOpeningElement:
		element := node.AsJsxOpeningElement()
		return element.TagName, element.Attributes
	case ast.KindJsxSelfClosingElement:
		element := node.AsJsxSelfClosingElement()
		return element.TagName, element.Attributes
	}
	return nil, nil
}

// IsIntrinsicElementNamed reports whether a tag name is the given lowercase HTML element.
//
// A member-expression name (`<Foo.link />`) or a namespaced one is a component reference that never
// emits the HTML element, so comparing text alone would flag code that renders nothing of the kind.
func IsIntrinsicElementNamed(tagName *ast.Node, name string) bool {
	return tagName != nil && tagName.Kind == ast.KindIdentifier && tagName.Text() == name
}

// HasAttributeNamed reports whether a JSX element writes an attribute with this name.
//
// Presence rather than value, which is a different question from StringAttributeValue and is the
// one some rules actually ask: `<script src=... async>` is exempt because `async` is written at
// all, whatever it is set to. Reading the value here would make a bare `async` (no initializer, the
// idiomatic spelling) look absent.
//
// A spread declines for the same reason as above: it carries no name to match, so an element that
// spreads its attributes reports nothing rather than falsely reporting the attribute missing.
func HasAttributeNamed(attributes *ast.Node, name string, matches NameMatch) bool {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return false
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return false
	}

	for _, property := range properties.Nodes {
		attributeName, named := AttributeName(property)
		if named && matches(attributeName, name) {
			return true
		}
	}

	return false
}
