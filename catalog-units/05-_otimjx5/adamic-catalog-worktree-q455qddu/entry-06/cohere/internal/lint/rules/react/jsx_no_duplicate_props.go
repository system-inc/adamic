package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageJsxNoDuplicateProps = rule.Message{
	Id: "jsxNoDuplicateProps",
	Description: "This prop is written more than once on the same element. React builds the " +
		"props object by assigning each attribute in order, so every earlier copy is " +
		"overwritten and only the last one survives. The discarded values look like they are " +
		"being passed and are not. Remove one of them, or rename them so each prop is distinct.",
}

// JsxNoDuplicatePropsOptions configures the rule, as upstream's single options object.
type JsxNoDuplicatePropsOptions struct {
	// IgnoreCase compares names lowercased, so `<App A a />` reports. Off by default.
	IgnoreCase bool `json:"ignoreCase"`
}

// JsxNoDuplicateProps flags a JSX element that writes the same prop name twice.
//
//	valid:   <App a b c />
//	valid:   <App a b c A />
//	valid:   <App {...this.props} a b c />
//	invalid: <App a a />
//	invalid: <App a="a" {...this.props} b="b" a="a" />
//
// Ported from `react/jsx-no-duplicate-props`, read against oxc's `jsx_no_duplicate_props.rs` and
// against `eslint-plugin-react`'s own `jsx-no-duplicate-props.js`. The two sources disagree in two
// places and oxc is followed in both, because oxc is what the gate being replaced actually runs.
//
// # Where the two upstreams disagree, and which one this follows
//
// **The option surface.** ESLint carries an `ignoreCase` boolean that lowercases every name before
// comparing, so `<App foo Foo />` reports under it. oxc implements none of it, on the reasoning that
// props are case-sensitive in JSX. This rule takes it, because ESLint's schema accepts it and an
// accepted option is an implemented one (#d21war2): off, the default, `<App A a />` is clean, which
// four of upstream's twelve clean cases pin; on, it reports, which three of upstream's failing cases
// pin. A namespaced name is still declined under it, as upstream's `a:b` clean case says.
//
// **Where the finding points.** ESLint reports the whole `JSXAttribute`, initializer included. oxc
// reports the two *names*, passing `with_labels([span1, span2])` where `span1` is the earlier
// occurrence and `span2` the one just found. Our harness carries one range per finding, so the
// earlier occurrence is the one reported, which is what oxc renders as the diagnostic's primary
// position: the snapshot for `<App {...this.props} a="a" b="b" a="a" />` prints `1:22`, the first
// `a` after the spread rather than the second. Measured against the release oxlint binary as well
// as read from the snapshot, because a two-label diagnostic's primary span is not something the
// snapshot format states outright.
//
// # Once per extra occurrence, not once per pair
//
// The map is written on every attribute, so a third copy is compared against the second rather than
// the first. `<App a a a />` therefore reports twice and `<App a a a a />` three times. This is
// worth stating because the other reading, one finding per duplicated name, is the more natural
// design and produces a different count on exactly the input no upstream fixture covers: every fail
// case upstream ships has at most two copies of one name, so both readings agree on all nine and
// the corpus cannot tell them apart. Measured on the release oxlint binary rather than inferred,
// and pinned by a fixture below that upstream does not have.
//
// # Two element kinds where oxc has one
//
// oxc listens on `JSXOpeningElement` and gets both shapes, because its parser gives a self-closing
// element an opening element with a `self_closing` flag. Our parser gives it a distinct
// `JsxSelfClosingElement` with no `JsxOpeningElement` inside it, so listening on the literal
// translation of oxc's node would make this rule silent on every fail fixture upstream ships: all
// nine are self-closing. Both kinds are listened on here and a probe confirms they never both fire
// for one element, so there is no double-report to guard against.
//
// # What a spread does, and what it does not do
//
// A spread is a `JsxSpreadAttribute` in the properties list rather than an absent node, so it is
// skipped by kind rather than by being invisible. It neither reports nor resets what has been seen:
// `<App a="a" {...this.props} b="b" a="a" />` reports, because the map survives across it. What
// upstream cannot see is a name *inside* the spread, so `<App a {...{a: 1}} />` is silent in both
// tools even though it is the same collision. That is a real gap and it is reproduced rather than
// improved on: deciding what a spread contributes would mean evaluating an arbitrary expression,
// and a rule that guessed would report on code whose real props are unknown.
//
// # Namespaced names, which never participate
//
// `<svg xlink:href="x" xlink:href="y" />` is silent, in oxc and here. oxc destructures
// `JSXAttributeName::Identifier` and a namespaced name is the sibling variant, so it returns before
// reaching the map; `jsx.AttributeName` declines the same shape by requiring `KindIdentifier`, and a
// probe confirms it answers `named=false` there. The decline is correct for this rule rather than
// merely convenient, and the reason is that a namespaced name would need its prefix compared too:
// treating `xlink:href` as `href` would make it collide with a plain `href` that sets a different
// attribute entirely. Verified against the release oxlint binary, which is silent on both
// `<svg xlink:href="x" xlink:href="y" />` and `<svg href="x" xlink:href="y" />`.
//
// A dashed name is not the same case and does participate: `data-x` parses as a single identifier,
// so `<App data-x data-x />` reports in both tools.
var JsxNoDuplicateProps = rule.Rule{
	// No namespace prefix. The config writes `react/jsx-no-duplicate-props` and the parity guard
	// strips the namespace on a `/` boundary, so a rule named `react-jsx-no-duplicate-props` would
	// match no inventory entry, lint no files, and still pass every fixture in this package.
	Name: "react/jsx-no-duplicate-props",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule gets the zero value, which is upstream's case-sensitive default.
		settings, _ := rule.OptionsAs[JsxNoDuplicatePropsOptions](options)

		report := func(node *ast.Node) {
			_, attributes := jsx.ElementParts(node)
			if attributes == nil {
				return
			}
			properties := attributes.AsJsxAttributes().Properties
			if properties == nil {
				return
			}

			// Keyed by name, valued by the node of the earliest *unreported* occurrence, which is
			// what gets reported when the next copy arrives. The value is overwritten on every hit
			// rather than only on the first, which is the behavior described above: it makes a
			// third copy report against the second.
			seen := make(map[string]*ast.Node, len(properties.Nodes))

			for _, property := range properties.Nodes {
				// Declines a spread, which carries no name, and a namespaced name, which oxc also
				// declines by matching only the `Identifier` variant.
				name, named := jsx.AttributeName(property)
				if !named {
					continue
				}
				if settings.IgnoreCase {
					name = strings.ToLower(name)
				}
				nameNode := property.AsJsxAttribute().Name()
				if earlier, duplicated := seen[name]; duplicated {
					ctx.ReportNode(earlier, messageJsxNoDuplicateProps)
				}
				seen[name] = nameNode
			}
		}

		return rule.Listeners{
			// Both kinds, and neither is redundant. A self-closing element produces no
			// JsxOpeningElement in our tree, and every one of upstream's nine fail fixtures is
			// self-closing, so dropping the second arm would silence the entire imported corpus
			// while leaving the first arm's fixtures green.
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
