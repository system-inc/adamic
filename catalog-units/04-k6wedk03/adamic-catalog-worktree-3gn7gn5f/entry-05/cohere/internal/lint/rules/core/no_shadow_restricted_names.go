package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoShadowRestrictedNamesOptions configures which names count as restricted.
//
// Spelled as an allowance rather than as upstream's `reportGlobalThis`, and deliberately. Upstream
// defaults that flag to true, so a Go field named `ReportGlobalThis` would default to false and
// quietly disable a third of the rule for every caller who configured nothing. Inverting it puts
// upstream's default on the zero value, where a struct built by `NoShadowRestrictedNamesOptions{}`
// behaves the way the rule is documented to behave.
type NoShadowRestrictedNamesOptions struct {
	// AllowGlobalThis stops reporting bindings named `globalThis`.
	//
	// It is separable from the rest because it is the newest of the six and the least dangerous:
	// `globalThis` arrived in ES2020, so a codebase predating it can hold a `globalThis` binding
	// that was legal when written, and shadowing it costs less than shadowing `undefined` does.
	// The other five are not configurable, upstream or here.
	AllowGlobalThis bool `json:"-"`
}

// UnmarshalJSON reads upstream's key, `reportGlobalThis`, and stores its inverse.
//
// The field is spelled as an allowance for the reason above, and the config is spelled as upstream
// spells it, so a config written for ESLint means the same thing here. Before this the field had no
// tag, so only the non-upstream `allowGlobalThis` reached it and `{"reportGlobalThis": false}` was
// refused as an unknown key (#4a4yse4).
func (options *NoShadowRestrictedNamesOptions) UnmarshalJSON(raw []byte) error {
	var wire struct {
		ReportGlobalThis *bool `json:"reportGlobalThis"`
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return err
	}
	options.AllowGlobalThis = wire.ReportGlobalThis != nil && !*wire.ReportGlobalThis
	return nil
}

// restrictedNames are the five names a binding may never take, whatever the options say.
//
// `NaN`, `Infinity` and `undefined` are value properties of the global object. `eval` and
// `arguments` are the strict-mode restricted identifiers. `globalThis` is the sixth and is handled
// separately because it is the one the options can turn off.
var restrictedNames = map[string]bool{
	"undefined": true,
	"NaN":       true,
	"Infinity":  true,
	"eval":      true,
	"arguments": true,
}

var messageShadowingRestrictedName = rule.Message{
	Id: "shadowingRestrictedName",
	Description: "This declares a binding named after a global the rest of the file assumes is the " +
		"global. Every read of that name inside the new scope gets this binding instead, so a " +
		"comparison against `undefined` compares against whatever was assigned here, `NaN` stops " +
		"being the value no comparison equals, and `arguments` stops being the call's arguments. " +
		"The damage is silent and it is not local to this line: the code that breaks is the code " +
		"that never mentioned the shadow and reads the name expecting the global. Rename the " +
		"binding.",
}

// NoShadowRestrictedNames flags a declaration whose name is one of the restricted globals.
//
//	valid:   function foo(bar) { var baz; }
//	valid:   try {} catch(e) {}
//	valid:   var undefined;
//	valid:   var undefined; doSomething(undefined);
//	valid:   import { undefined as undef } from 'foo';
//	valid:   enum Globals { undefined = 'undefined' }
//	invalid: function NaN(NaN) { var NaN; }
//	invalid: try {} catch(eval) {}
//	invalid: var undefined; undefined = 5;
//	invalid: class globalThis {}
//	invalid: import * as undefined from 'foo';
//
// The six names are `undefined`, `NaN`, `Infinity`, `eval`, `arguments` and `globalThis`. None is a
// reserved word, so every one of these declarations parses and most of them run, which is the whole
// problem: the mistake produces no error at the line that makes it and surfaces somewhere else
// entirely, in code that reads the name expecting the global and gets the shadow.
//
// # This rule is the inverse of its siblings and it changes the shape
//
// `no-class-assign` and `no-ex-assign` anchor on a declaration and hunt writes to it. This one
// anchors on a declaration and asks a question about the declaration itself, so there is no search:
// every finding sits at the binding identifier the listener already has. That makes the bulk of the
// rule syntactic, and the coverage question moves from "did I find every write" to "did I find
// every way to declare a name", which is a longer list than it looks. Upstream's densest input,
// `function NaN(NaN) { var NaN; !function NaN(NaN) { try {} catch(NaN) {} }; }`, is six bindings in
// five scopes and reports six times.
//
// # Why it still reads the checker
//
// One carve-out, and it is not syntactic. A bare `var undefined;` is legal and harmless, because
// the binding is initialized to the global's own value and reading it gets exactly what reading the
// global would. It stops being harmless the moment anything assigns to it, and upstream carves out
// only the harmless form:
//
//	var undefined;                          clean
//	var undefined; doSomething(undefined);  clean, the reference is a read
//	var undefined; undefined = 5;           reports, at the declaration
//
// Deciding between the second and the third means knowing whether a given `undefined = 5` writes to
// *this* binding or to some other one. Upstream answers with `get_resolved_references`, a find-all
// references index we do not have; the equivalent question our checker answers is which declaration
// an identifier binds to, so this walks the file once and asks that of each write.
//
// Node identity rather than declaration kind, and here the trap is live rather than theoretical.
// Measured on our own checker:
//
//	function f(){ var undefined; } var undefined; undefined = 1;
//
// holds two `KindVariableDeclaration` nodes with distinct pointers, and the write binds to the
// outer one. A kind comparison calls the inner declaration a match, carves nothing out, and reports
// a binding upstream leaves alone. `var undefined; var undefined;` is the control: those two merge
// to one symbol with one declaration pointer, which is why upstream's clean case for it stays clean.
//
// The shorthand accessor is live too. `({undefined} = obj)` resolves through `GetSymbolAtLocation`
// to the shorthand property's own symbol, which reads exactly like a write to some other binding,
// so the carve-out would hold and the case would go silent. `GetShorthandAssignmentValueSymbol`
// answers the value side. Measured, not assumed: the plain accessor returns a
// `KindShorthandPropertyAssignment` declaration here and the shorthand accessor returns the
// variable.
//
// # What the carve-out deliberately does not cover
//
// Only `undefined`, and only a bare variable declaration. `let globalThis; globalThis = 5;` reports
// once and would report even without the write, because `globalThis` has no carve-out at all. A
// destructured `var [undefined] = [1]` reports too: the binding takes a value out of the array, so
// it does not hold the global's value and the reasoning behind the carve-out does not apply.
//
// # Enum members
//
// `enum Globals { undefined = 'undefined' }` is clean. An enum member is a property of the enum
// object rather than a binding in any scope, so nothing named `undefined` is shadowed and every
// later read of the bare name still finds the global. Upstream carves this out explicitly and it is
// the only TypeScript-specific judgment in the rule.
//
// No fix. The repair is a rename, which means choosing a name and deciding which later uses of the
// old one meant this binding rather than the global, and those are the same uses the rule exists to
// warn about.
var NoShadowRestrictedNames = rule.Rule{
	Name: "no-shadow-restricted-names",

	// See the doc above: the `undefined` carve-out has to know whether a later write binds to the
	// declaration being judged, and no bounded walk answers that.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := NoShadowRestrictedNamesOptions{}
		if configured, isConfigured := rule.OptionsAs[NoShadowRestrictedNamesOptions](options); isConfigured {
			settings = configured
		}

		isRestricted := func(name string) bool {
			if restrictedNames[name] {
				return true
			}
			return name == "globalThis" && !settings.AllowGlobalThis
		}

		// report judges one binding identifier: the name position of some declaration.
		//
		// Every finding this rule produces comes through here, and the span is the identifier
		// rather than the declaration that contains it. That matches upstream, whose snapshot
		// underlines exactly the name in all 69 diagnostics, and it is the right choice
		// independently: the declaration is often correct code apart from what it is called, and
		// the name is the only part the reader has to change.
		report := func(name *ast.Node) {
			if name == nil || name.Kind != ast.KindIdentifier || !isRestricted(name.Text()) {
				return
			}
			if name.Text() == "undefined" && safelyShadowsUndefined(ctx, name) {
				return
			}
			ctx.ReportNode(name, messageShadowingRestrictedName)
		}

		// The declaration forms, and the list is the rule's coverage.
		//
		// Missing one does not fail loudly. It silently stops reporting a whole class of shadow
		// while every other fixture stays green, which is why upstream packs six binding forms into
		// one input rather than testing them apart.
		reportName := func(node *ast.Node) { report(node.Name()) }

		return rule.Listeners{
			// `var undefined`, `let NaN`, `const eval`. Also the binding of a `for (var eval of x)`
			// head, which is the same node kind.
			ast.KindVariableDeclaration: reportName,

			// `var {undefined} = obj`, `var [NaN] = arr`, `var {a, ...eval} = obj`. A binding
			// element inside a pattern is its own declaration and the pattern above it is not, so
			// this arm rather than the one above is what covers all four destructuring shapes.
			ast.KindBindingElement: reportName,

			// A function or arrow parameter, and a catch parameter, which typescript-go models as
			// a `KindVariableDeclaration` under the catch clause and so is already covered above.
			ast.KindParameter: reportName,

			// `function NaN() {}` and `!function NaN() {}`. Both bind the name, the declaration to
			// the enclosing scope and the expression to its own body, and both shadow.
			ast.KindFunctionDeclaration: reportName,
			ast.KindFunctionExpression:  reportName,

			// `class globalThis {}` and `(class globalThis {})`.
			ast.KindClassDeclaration: reportName,
			ast.KindClassExpression:  reportName,

			// `import undefined from 'foo'`. The clause's own name is the default binding; the
			// named and namespace forms are separate nodes below.
			ast.KindImportClause: reportName,

			// `import { undefined } from 'foo'` and `import { baz as undefined } from 'foo'`.
			// `Name()` is the local binding in both, which is what makes
			// `import { undefined as undef }` clean: the local name is `undef` and nothing is
			// shadowed.
			ast.KindImportSpecifier: reportName,

			// `import * as undefined from 'foo'`.
			ast.KindNamespaceImport: reportName,

			// Deliberately absent: KindEnumMember. An enum member is a property of the enum object
			// rather than a binding, so `enum Globals { undefined = 'u' }` shadows nothing.
			// Upstream skips it by an explicit test on the declaration kind; here the skip is that
			// no listener asks about it, which has the same effect and states itself.
		}
	},
}

// safelyShadowsUndefined reports whether a binding named `undefined` keeps the global's value.
//
// True only for the shape upstream carves out: a bare variable declaration with no initializer,
// none of whose references writes to it. Anything else either starts out holding something other
// than the global or is later made to.
//
// The write search is the whole file rather than the declaration's scope. A `var` binding hoists,
// so a write can sit textually before its declaration, and a nested block or function can hold one
// that still binds to the outer declaration. Anchoring the search on the file and the match on
// declaration identity is what makes the position of the write irrelevant.
func safelyShadowsUndefined(ctx rule.Context, name *ast.Node) bool {
	// A binding in any other declaration form is not the carved-out shape. A parameter named
	// `undefined` receives whatever the caller passed, a destructured one takes a value out of the
	// object, and an import binds the module's export, so none of them holds the global's value.
	if name.Parent == nil || name.Parent.Kind != ast.KindVariableDeclaration {
		return false
	}

	// `var undefined = 5` starts out wrong regardless of what happens later. The catch clause's
	// parameter is also a `KindVariableDeclaration` in this AST and has no initializer, so the
	// check below is what keeps `catch(undefined)` reporting: its parent is a catch clause rather
	// than a declaration list.
	declaration := name.Parent.AsVariableDeclaration()
	if declaration.Initializer != nil {
		return false
	}
	if name.Parent.Parent == nil || name.Parent.Parent.Kind != ast.KindVariableDeclarationList {
		return false
	}

	// The engine hands every rule a nil checker when the program could not be built. Without one
	// this cannot tell a write to this binding from a write to another, and the safe direction is
	// to decline the carve-out and report, which is what a missing checker gets here.
	if ctx.TypeChecker == nil {
		return false
	}

	anchor := declarationAnchoredAt(ctx, name)
	if anchor == nil {
		return false
	}

	sourceFile := ast.GetSourceFileOfNode(name)
	if sourceFile == nil {
		return false
	}

	written := false
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || written {
			return
		}
		// The text comparison is a pre-filter rather than a discrimination: symbol identity already
		// implies it, since an identifier spelled differently cannot resolve to this declaration.
		// It is here because it is far cheaper than a checker call and this walk visits every
		// identifier in the file.
		if current.Kind == ast.KindIdentifier &&
			current.Text() == "undefined" &&
			reference.WritesToBinding(current) &&
			resolvesToDeclaration(ctx, current, anchor) {
			written = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	return !written
}
