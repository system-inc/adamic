package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageJsxIdentifierNotDefined = rule.Message{
	Id: "jsxIdentifierNotDefined",
	Description: "This JSX tag names a component that has no declaration in scope, so React " +
		"receives `undefined` where a component was meant and throws at render rather than at " +
		"build. Almost always a misspelling or a missing import. Import the component, or fix " +
		"the spelling to match the binding you meant.",
}

// JsxNoUndef reports a JSX tag whose component name resolves to no declaration.
//
//	valid:   var React, App; React.render(<App />);
//	valid:   var React; React.render(<img />);          (intrinsic element)
//	valid:   var React; React.render(<x-gif />);        (custom element)
//	valid:   var React; React.render(<Apppp:Foo />);    (namespaced name)
//	valid:   <this.props.tag />                         (rooted at this)
//	invalid: var React; React.render(<App />);
//	invalid: var React; React.render(<appp.Foo />);     (the object is the reference)
//
// Ported from `react/jsx-no-undef`, read against oxc's `jsx_no_undef.rs`. Both of the questions
// this rule's shape turns on were measured rather than reasoned about, because both are places
// where our tree's syntax model differs from the one upstream is written against.
//
// # Where the intrinsic line actually sits, and why the shim cannot tell us
//
// Upstream never asks "is this tag lowercase". It asks the parser, which has already split the tag
// into `JSXElementName::Identifier` for an intrinsic and `JSXElementName::IdentifierReference` for
// a component, and then declines the first variant outright. The whole judgment lives in oxc's
// parser at `crates/oxc_parser/src/jsx/mod.rs:178`.
//
// typescript-go draws no such distinction. Both forms arrive as a plain `ast.KindIdentifier`, so a
// port that anchors on the kind alone has to reproduce that split by hand, and reproducing only the
// obvious half of it means every `<div>` in the tree reports. The predicate is `isComponentName`
// below and it is upstream's, byte for byte in its effect, measured against the release oxlint
// binary on the four shapes a reading of "lowercase means intrinsic" would get wrong: `<_foo />`
// and `<$foo />` report, `<테스트 />` reports because a non-ASCII first byte is always a reference,
// and `<Foo-bar />` stays silent because a dash makes it a custom element whatever its case.
//
// # A member expression's object is a reference at any case
//
// The case test applies only to a bare tag name. `<app.Foo />` and `<appp.Foo />` differ in the
// corpus by whether `app` is declared, not by case, because oxc's parser builds every member
// object as an `IdentifierReference` unconditionally. Applying the case predicate to the leftmost
// object would have made the corpus's `<appp.Foo />` and `<appp.foo.Bar />` fail cases silent, and
// they are two of the eight findings upstream snapshots.
//
// # Why the checker rather than a scope walk
//
// This is name resolution rather than a scope flag, so it is the kind of `ctx.scoping()` question
// that costs the checker. Established by probe rather than by reading the call site: the four
// corpus cases that separate a real port from a plausible one are all resolution questions no
// syntactic walk answers. `enum A { App }` declares `App` as a member and not as a value binding,
// so it reports, while `var App` beside the same enum does not; `import App = require('./app')`
// and `import App = Foo.App` both bind; and `{ const App = null; }` binds only inside its block.
// `GetSymbolAtLocation` returned the right answer on all four.
//
// Resolution alone is not sufficient, which is the other half of why the case predicate is not
// optional: `<img />` and `<x-gif />` resolve to no symbol either, exactly like the undeclared
// components. Nothing in the type graph separates an intrinsic element from a typo.
//
// # Globals, and allowGlobals
//
// Upstream walks the scope chain from the tag outward and, in a module, stops at the module scope
// unless `allowGlobals` is set. So by default a tag naming something only the global scope declares
// reports in a module: a configured global, or a lib declaration such as `<Map />`. ESLint's flat
// config reads every file as a module except a `.cjs` one, which is commonjs and reaches the global
// scope with or without the option. All three measured against the installed build (#d21war2).
//
// The checker answers a wider question than the walk does, since `GetSymbolAtLocation` finds a
// global as readily as a local. So a resolved symbol counts as in scope only when one of its
// declarations is in this file; an import counts, because its symbol is the alias the import
// declares here. Under `allowGlobals`, or in a `.cjs` file, any resolved symbol counts. Before
// #d21war2 every resolved symbol counted, which was `allowGlobals` on in every file.
var JsxNoUndef = rule.Rule{
	Name:             "react/jsx-no-undef",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An unconfigured rule gets the zero value, which is upstream's `allowGlobals: false`.
		settings, _ := rule.OptionsAs[JsxNoUndefOptions](options)
		stopsAtModuleScope := !settings.AllowGlobals && !strings.HasSuffix(ctx.SourceFile.FileName().AsString(), ".cjs")

		check := func(node *ast.Node) {
			tagName, _ := jsx.ElementParts(node)
			reference := resolvableJsxReference(tagName)
			if reference == nil {
				return
			}
			if ctx.TypeChecker == nil {
				return
			}
			if symbol := ctx.TypeChecker.GetSymbolAtLocation(reference); symbol != nil {
				if !stopsAtModuleScope || jsxNoUndefDeclaredInFile(symbol, ctx.SourceFile) {
					return
				}
			}
			ctx.ReportNode(reference, messageJsxIdentifierNotDefined)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// JsxNoUndefOptions configures the rule, as upstream's single options object.
type JsxNoUndefOptions struct {
	// AllowGlobals lets a tag in a module resolve to a global-scope declaration. Off by default.
	AllowGlobals bool `json:"allowGlobals"`
}

// jsxNoUndefDeclaredInFile reports whether a symbol has a declaration in this file, which is what
// upstream's walk reaches when it stops at the module scope.
//
// A `declare global` member declared here counts too. That reads as a global, and it was measured
// rather than assumed: typescript-eslint's scope manager puts it in the file's own scopes, so
// `declare global { var G: any } export {}; <G />;` is clean under the default at 10.8.1.
func jsxNoUndefDeclaredInFile(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	for _, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) == sourceFile {
			return true
		}
	}
	return false
}

// resolvableJsxReference returns the identifier a JSX tag name asks scope to resolve, or nil when
// the tag names something that is not a binding lookup at all.
//
// Three shapes decline, matching the three variants upstream's `get_resolvable_ident` returns
// `None` for:
//
//   - an intrinsic or custom element, which is `KindIdentifier` failing `isComponentName`
//   - a namespaced name, which typescript-go gives its own kind, so no text inspection is needed
//   - anything rooted at `this`, which resolves to the enclosing class rather than to a binding
//
// The `this` case is declined structurally rather than by resolution, and that is deliberate:
// probed, `<this.props.tag />` inside a class resolves to the class's own symbol and would have
// passed anyway. Leaning on that would have made the rule's agreement with upstream accidental,
// dependent on the element sitting inside a class, and `<this.foo />` at the top level of a module
// resolves to nothing and would then have reported where upstream is silent.
// A nil tagName needs no arm of its own. `jsx.ElementParts` returns nil only for a node that is
// neither JSX opening form, which the two listeners above never hand it, and a JSX element the
// parser produced always carries a tag name: typescript-go's own parser dereferences `TagName()`
// unguarded on both forms. Written first as an explicit `tagName == nil` guard, which a mutation
// sweep then showed no fixture could reach, because nothing can construct the input. It is subsumed
// by the switch rather than deleted, since a nil node matches no arm and falls to the final return.
func resolvableJsxReference(tagName *ast.Node) *ast.Node {
	if tagName == nil {
		return nil // Unreachable from the listeners above; see the comment on this function.
	}

	switch tagName.Kind {
	case ast.KindIdentifier:
		if !isComponentName(tagName.Text()) {
			return nil
		}
		return tagName

	case ast.KindPropertyAccessExpression:
		// The leftmost object carries the reference, at any case. `<A.B.C.D />` reports on `A`.
		leftmost := tagName
		for leftmost.Kind == ast.KindPropertyAccessExpression {
			leftmost = leftmost.AsPropertyAccessExpression().Expression
		}
		if leftmost.Kind != ast.KindIdentifier {
			return nil
		}
		return leftmost
	}

	return nil
}

// isComponentName reports whether a bare JSX tag name is a reference to a binding rather than an
// intrinsic or custom element.
//
// This is oxc's parser predicate rather than a rule of our own, transcribed from
// `crates/oxc_parser/src/jsx/mod.rs:178` and then measured against the release oxlint binary. Read
// in full because two thirds of it is counterintuitive:
//
//	dash anywhere            custom element, never a reference. `<Foo-bar />` is silent.
//	non-ASCII first byte     always a reference. `<테스트 />` reports.
//	ASCII lowercase first    intrinsic. `<div />` is silent.
//	anything else ASCII      a reference, so `_` and `$` report as well as `A-Z`.
//
// `this` is not special-cased here because typescript-go parses a bare `<this />` as
// `KindThisKeyword`, which never reaches this function. Upstream needs the name check because its
// parser routes `this` through the same identifier path.
func isComponentName(name string) bool {
	if name == "" {
		return false
	}

	for index := 0; index < len(name); index++ {
		if name[index] == '-' {
			return false
		}
	}

	first := name[0]
	if first >= 0x80 {
		return true
	}
	return first < 'a' || first > 'z'
}
