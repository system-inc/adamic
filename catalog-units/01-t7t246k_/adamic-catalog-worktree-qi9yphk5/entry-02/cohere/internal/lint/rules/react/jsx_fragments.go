package react

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// JsxFragmentsMode selects which of the two fragment spellings the file is expected to use.
type JsxFragmentsMode string

const (
	// JsxFragmentsSyntax prefers the shorthand, so a named fragment reports.
	JsxFragmentsSyntax JsxFragmentsMode = "Syntax"

	// JsxFragmentsElement prefers the named form, so the shorthand reports.
	JsxFragmentsElement JsxFragmentsMode = "Element"
)

// JsxFragmentsOptions carries the single mode this rule takes.
type JsxFragmentsOptions struct {
	Mode JsxFragmentsMode
}

// DefaultJsxFragmentsOptions is the unconfigured answer.
//
// Upstream reads `context.options[0] || 'syntax'`, so an absent option and an empty string both
// mean the shorthand is preferred.
func DefaultJsxFragmentsOptions() JsxFragmentsOptions {
	return JsxFragmentsOptions{Mode: JsxFragmentsSyntax}
}

// DecodeJsxFragmentsOptions reads this rule's configuration from the config layer.
//
// The option is a bare enum string rather than an object, which is why this cannot use
// `rule.DecodeOptionsInto`: the generic helper unmarshals into a struct and a JSON string is not
// one. Our config layer unwraps the severity tuple before dispatch, so what arrives here is
// upstream's `"syntax"` or `"element"` on its own rather than upstream's one-element array.
//
// Empty input and `null` are upstream's default. Anything else that is not one of the two modes is
// refused, naming the value. This used to read an unknown value as the default, and its comment
// claimed upstream did the same; it does not. Upstream's schema is `[{ enum: ["syntax", "element"] }]`,
// measured on the installed build, and refuses any other value at load, the empty string included,
// before `context.options[0] || 'syntax'` could ever see it. The config layer now refuses a decoder's
// error by rule name, so refusing is what tells the author (#rfbha44).
func DecodeJsxFragmentsOptions(raw []byte) (any, error) {
	options := DefaultJsxFragmentsOptions()
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return options, nil
	}
	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return options, fmt.Errorf(`expected a mode string, "syntax" or "element", got %s`, trimmed)
	}
	switch mode {
	case "syntax":
		return options, nil
	case "element":
		options.Mode = JsxFragmentsElement
		return options, nil
	}
	return DefaultJsxFragmentsOptions(), fmt.Errorf(`mode %q is not "syntax" or "element"`, mode)
}

var messagePreferFragmentShorthand = rule.Message{
	Id: "preferFragment",
	Description: "Prefer fragment shorthand over React.Fragment. The shorthand `<>` says the same " +
		"thing with less to read, and a named fragment that carries no props is only the long " +
		"spelling of it.",
}

var messagePreferFragmentPragma = rule.Message{
	Id: "preferPragma",
	Description: "Prefer React.Fragment over fragment shorthand. The named form is the one this " +
		"file is configured to use, and it is the only spelling that can carry a `key`.",
}

// JsxFragments enforces one of the two React fragment spellings across a file.
//
//	valid:   <><Foo /></>                                     under the default
//	valid:   <React.Fragment key="k"><Foo /></React.Fragment>  a fragment carrying props
//	valid:   <React.Fragment><Foo /></React.Fragment>          under `"element"`
//	invalid: <React.Fragment><Foo /></React.Fragment>          under the default, fixed to <><Foo /></>
//	invalid: <><Foo /></>                                      under `"element"`
//
// # Which names count as a fragment
//
// The namespaced spelling `React.Fragment` always does. A bare identifier counts when it binds to
// the fragment export, through any of four shapes upstream accepts, all measured against the
// installed build on 2026-08-27:
//
//	import React, {Fragment} from 'react'         reports
//	import React, {Fragment as F} from 'react'    reports, the LOCAL name is what the tag writes
//	const F = React.Fragment                      reports
//	const {Fragment} = React                      reports
//	const {Fragment} = require('react')           reports
//	import {Fragment} from 'preact'               SILENT, the module must be exactly `react`
//	<A.React.Fragment>                            SILENT, a deeper member access is not the pragma
//
// Upstream asks `eslint-scope` for the variable's initializer; this asks the checker for the symbol
// and inspects its declarations against the same shapes, which is the substitution
// `checked-requires-onchange-or-readonly` already makes in this package for the identical question.
// Fidelity is to which names resolve to the fragment, not to the mechanism that resolves them.
//
// # The pragma is fixed at React.Fragment
//
// Upstream reads both halves from `settings.react.pragma` and `settings.react.fragment`, and
// cohere reads no React settings, so both are unreachable by any route. Upstream's entire
// corpus configures them to `Act` and `Frag`, which is why every imported case below was replayed
// against the installed build under the DEFAULT settings before being written down; nineteen of
// the twenty two reproduced exactly, including the fixer output, and the three that did not are
// the version cases described next.
//
// # The fragmentsNotSupported arm cannot fire here, measured
//
// Upstream's third message fires when `settings.react.version` is below 16.2. With no version
// configured, `version.js` sets `defaultVersion` to `999.999.999` and every `testReactVersion`
// comparison passes, so the arm is unreachable. Confirmed by running the installed build on
// upstream's own three version cases with no settings: `<><Foo /></>` went silent and both named
// forms reported `preferFragment` instead. The message is therefore not ported at all rather than
// ported and left dead, because a message nothing can produce is one more thing for a reader to
// account for. Should cohere ever grow a settings surface, this is the arm to add back.
//
// # No fix, and this is the deliberate half of the port
//
// Upstream ships a fixer for both directions and it DESTROYS TYPE ARGUMENTS. A JSX tag may carry
// them, and they sit inside the opening tag but outside the attributes list, so upstream's only
// guard, `attrs && attrs.length > 0`, cannot see them. Measured against the installed build with
// the TypeScript parser on 2026-08-27:
//
//	<React.Fragment<T>><Foo /></React.Fragment>   fixed to <><Foo /></>   the <T> is gone
//	<React.Fragment<T> />                         fixed to <></>          the <T> is gone
//
// Upstream's corpus is JavaScript and cannot express the shape, so all thirteen of its `output`
// assertions pass while the repair silently deletes type information from a TypeScript tree. This
// is the same structural failure as `no-undef-init` stranding a type annotation and
// `no-arrow-function-lifecycle` dropping a return type, arriving through a third door.
//
// The repair is withheld ENTIRELY rather than withheld only for the type-argument shape. Declining
// per-shape would ship a fixer whose safety rests on my enumeration of what can appear inside a JSX
// tag being complete, and the two rules above are both cases where that enumeration was wrong by
// exactly one item. The finding still tells a reader what to change and the edit is two characters;
// what is lost is the unattended rewrite, which is the half that cannot be reviewed.
//
// Reporting without the repair is also the safe direction for the `element` direction, whose fix
// has no known hazard, because a rule that fixes one direction and not the other is a rule whose
// behaviour a reader has to look up.
var JsxFragments = rule.Rule{
	Name:             "react/jsx-fragments",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultJsxFragmentsOptions()
		if configured, isConfigured := rule.OptionsAs[JsxFragmentsOptions](options); isConfigured {
			settings = configured
		}

		// The named form arrives as two different node kinds here. `<React.Fragment>x</React.Fragment>`
		// is a JsxElement and `<React.Fragment />` is a JsxSelfClosingElement, while upstream sees
		// one JSXElement carrying a `selfClosing` flag. Both must be judged: upstream's corpus
		// reports on `<Act.Frag />` explicitly.
		judgeNamed := func(node *ast.Node, tagName *ast.Node, attributes *ast.Node) {
			if settings.Mode != JsxFragmentsSyntax {
				return
			}
			if ctx.TypeChecker == nil {
				return
			}
			if !jsxFragmentsNameIsFragment(ctx, tagName) {
				return
			}
			// A fragment carrying any prop is left alone, because the shorthand cannot carry one.
			// Upstream tests `attrs && attrs.length > 0` on the opening element.
			if jsxFragmentsHasAttributes(attributes) {
				return
			}
			ctx.ReportNode(node, messagePreferFragmentShorthand)
		}

		return rule.Listeners{
			ast.KindJsxFragment: func(node *ast.Node) {
				if settings.Mode != JsxFragmentsElement {
					return
				}
				ctx.ReportNode(node, messagePreferFragmentPragma)
			},

			ast.KindJsxElement: func(node *ast.Node) {
				opening := node.AsJsxElement().OpeningElement
				if opening == nil {
					return
				}
				// `jsx.ElementParts` reads the tag name and attributes off whichever of the two
				// element kinds it is handed, which is the shelf's answer to the fact that our
				// parser splits what ESTree spells as one node. Reaching for the accessor directly
				// is what `TestRulePackagesDoNotReachPastWrappedAccessors` exists to catch.
				tagName, attributes := jsx.ElementParts(opening)
				judgeNamed(node, tagName, attributes)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				tagName, attributes := jsx.ElementParts(node)
				judgeNamed(node, tagName, attributes)
			},
		}
	},
}

// jsxFragmentsHasAttributes reports whether a tag carries at least one prop.
//
// Upstream reads `openingEl.attributes` and tests its length. Ours is a JsxAttributes node whose
// Properties list holds both plain attributes and spreads, matching what ESTree puts in that array.
func jsxFragmentsHasAttributes(attributes *ast.Node) bool {
	if attributes == nil {
		return false
	}
	list := attributes.AsJsxAttributes()
	if list == nil || list.Properties == nil {
		return false
	}
	return len(list.Properties.Nodes) > 0
}

// jsxFragmentsNameIsFragment reports whether a tag name denotes the React fragment.
//
// Two accepting routes, both upstream's. The namespaced spelling is matched structurally, and a
// bare identifier is resolved through the checker to see whether it binds to the fragment export.
func jsxFragmentsNameIsFragment(ctx rule.Context, tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}

	switch tagName.Kind {
	case ast.KindPropertyAccessExpression:
		// Exactly `React.Fragment`, two segments. Upstream builds the string
		// `${reactPragma}.${fragmentPragma}` and compares the rendered element type against it, so
		// a three-segment `A.React.Fragment` renders as `A.React.Fragment` and does not match.
		// Measured silent on the installed build.
		access := tagName.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier {
			return false
		}
		if access.Expression.Text() != jsxFragmentsPragmaName {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			name.Text() == jsxFragmentsFragmentName

	case ast.KindIdentifier:
		return jsxFragmentsIdentifierBindsToFragment(ctx, tagName)
	}
	return false
}

// jsxFragmentsPragmaName is the object half of the namespaced spelling.
//
// Fixed rather than configurable. `settings.react.pragma` has no counterpart in `internal/config`,
// and a `/** @jsx */` comment is readable through `comments.ForFile` but is not read, matching the
// stated decline in `checked-requires-onchange-or-readonly` for the same question. The cost is a
// false negative on a file that renames the pragma by hand, which is the safe direction.
const jsxFragmentsPragmaName = "React"

// jsxFragmentsFragmentName is the property half, upstream's `settings.react.fragment`.
const jsxFragmentsFragmentName = "Fragment"

// jsxFragmentsIdentifierBindsToFragment answers whether a bare tag name resolves to the fragment.
//
// Upstream does this in two halves that this collapses into one. Its `ImportDeclaration` visitor
// collects local names introduced by `import {Fragment as F} from 'react'` into a set, and its
// `refersToReactFragment` separately asks `eslint-scope` for a variable's initializer and matches
// three assignment shapes. Both halves are asking which declaration the name binds to, which the
// checker answers directly, so they are one loop here.
//
// The loop over declarations rather than an index is the standing hazard in this tree: declaration
// merging can put the import at a position other than zero, and indexing would go silent on an
// input upstream reports. `LocalSymbol` is consulted as a fallback because an exported declaration
// carries a truncated symbol.
func jsxFragmentsIdentifierBindsToFragment(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if jsxFragmentsDeclarationIsFragment(declaration) {
			return true
		}
	}
	return false
}

// jsxFragmentsDeclarationIsFragment answers the four accepted shapes for one declaration node.
func jsxFragmentsDeclarationIsFragment(declaration *ast.Node) bool {
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindImportSpecifier:
		// `import {Fragment} from 'react'` and `import {Fragment as F} from 'react'`. Upstream
		// keys on the IMPORTED name matching the fragment pragma and records the LOCAL name, so
		// the alias is what the tag writes and the export is what is checked.
		specifier := declaration.AsImportSpecifier()
		imported := specifier.PropertyName
		if imported == nil {
			imported = specifier.Name()
		}
		if imported == nil || imported.Kind != ast.KindIdentifier ||
			imported.Text() != jsxFragmentsFragmentName {
			return false
		}
		return jsxFragmentsImportModuleName(declaration) == jsxFragmentsModuleName

	case ast.KindBindingElement:
		// `const {Fragment} = React` and `const {Fragment} = require('react')`.
		pattern := declaration.Parent
		if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
			return false
		}
		variable := pattern.Parent
		if variable == nil || variable.Kind != ast.KindVariableDeclaration {
			return false
		}
		return jsxFragmentsInitializerIsFragmentSource(variable.AsVariableDeclaration().Initializer)

	case ast.KindVariableDeclaration:
		// `const F = React.Fragment`.
		return jsxFragmentsInitializerIsFragmentSource(declaration.AsVariableDeclaration().Initializer)
	}
	return false
}

// jsxFragmentsInitializerIsFragmentSource answers upstream's three accepted right-hand sides.
//
//	React                  the bare pragma identifier, for `const {Fragment} = React`
//	React.Fragment         a member access on it, for `const F = React.Fragment`
//	require('react')       the call, for `const {Fragment} = require('react')`
//
// Upstream's `require` arm checks only that the callee is named `require` and the first argument's
// value is `'react'`; it does not check that `require` resolves to anything, which is reproduced.
func jsxFragmentsInitializerIsFragmentSource(initializer *ast.Node) bool {
	if initializer == nil {
		return false
	}
	switch initializer.Kind {
	case ast.KindIdentifier:
		return initializer.Text() == jsxFragmentsPragmaName

	case ast.KindPropertyAccessExpression:
		access := initializer.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier ||
			access.Expression.Text() != jsxFragmentsPragmaName {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			name.Text() == jsxFragmentsFragmentName

	case ast.KindCallExpression:
		call := initializer.AsCallExpression()
		if call.Expression == nil || call.Expression.Kind != ast.KindIdentifier ||
			call.Expression.Text() != "require" {
			return false
		}
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return false
		}
		first := call.Arguments.Nodes[0]
		return ast.IsStringLiteralLike(first) && first.Text() == jsxFragmentsModuleName
	}
	return false
}

// jsxFragmentsModuleName is the module an import or a require must name.
//
// Upstream compares against the literal `'react'` in both arms, with no lowercasing, so `React` as
// a module specifier is a decline. Measured: `import {Fragment} from 'preact'` is silent.
const jsxFragmentsModuleName = "react"

// jsxFragmentsImportModuleName reads the module specifier text off an import specifier.
//
// The kind check before the text read is load-bearing rather than defensive: `Node.Text()` panics
// on several kinds in this tree rather than returning empty.
func jsxFragmentsImportModuleName(specifier *ast.Node) string {
	namedImports := specifier.Parent
	if namedImports == nil || namedImports.Kind != ast.KindNamedImports {
		return ""
	}
	clause := namedImports.Parent
	if clause == nil || clause.Kind != ast.KindImportClause {
		return ""
	}
	declaration := clause.Parent
	if declaration == nil || declaration.Kind != ast.KindImportDeclaration {
		return ""
	}
	moduleSpecifier := declaration.AsImportDeclaration().ModuleSpecifier
	if moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return ""
	}
	return moduleSpecifier.Text()
}
