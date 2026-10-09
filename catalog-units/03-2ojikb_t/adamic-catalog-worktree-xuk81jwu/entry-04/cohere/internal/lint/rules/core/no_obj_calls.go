package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// nonCallableGlobals are the namespace objects that exist to hold members and are not functions:
// ESLint's list, Temporal included.
var nonCallableGlobals = []string{"Atomics", "JSON", "Math", "Reflect", "Intl", "Temporal"}

// nonCallableTraceMap asks the tracker for every call and construction of each of them.
var nonCallableTraceMap = func() map[string]*reference.TraceMap {
	traceMap := map[string]*reference.TraceMap{}
	for _, name := range nonCallableGlobals {
		traceMap[name] = &reference.TraceMap{Call: true, Construct: true}
	}
	return traceMap
}()

// NoObjCalls flags calling or constructing a global namespace object, following ESLint's rule of the
// same name.
//
//	valid:   let area = r => 2 * Math.PI * r * r;
//	valid:   let object = JSON.parse("{}");
//	valid:   var Math; Math();
//	valid:   function foo(JSON) { new JSON(); }
//	invalid: let math = Math();                       // unexpectedCall
//	invalid: var x = globalThis.JSON();               // unexpectedCall
//	invalid: let j = JSON; j();                        // unexpectedRefCall
//	invalid: var foo = bar ? baz : JSON; foo();        // unexpectedRefCall
//
// `Math`, `JSON`, `Reflect`, `Atomics`, `Intl` and `Temporal` are ordinary objects holding methods,
// not functions and not constructors. Calling one throws a TypeError on the line it appears on, so
// this is a crash the code has never survived rather than a style preference.
//
// # How a call is found
//
// Through the shelf's reference tracker, as ESLint finds it through eslint-utils' ReferenceTracker:
// from every reference to one of these globals, or to the global object (`globalThis.JSON`), through
// every name it is copied into and every expression that hands it on unchanged, to a call or a `new`.
// So `var foo = bar ? baz : JSON; foo();`, `(globalThis?.Reflect)()` and `const { JSON: j } =
// globalThis; j();` all report. A local binding that shares a global's spelling is not the global,
// which is half of ESLint's clean corpus (`var Math; Math();`), and a file that writes to a global
// never has it followed.
//
// The port this replaced followed oxc: a backward walk from the callee that read only a plain alias or
// a dotted `globalThis` member, left Temporal out, and took a local named `globalThis` for the global.
// Each was recorded at its line as a deliberate divergence, and #jjfa7qb's authority rule settled them
// all ESLint's way.
//
// # What a finding names
//
// The whole call or new expression. `unexpectedCall` when the callee spells the global it calls
// (`Math()`, `globalThis.Math()`, `const { JSON } = globalThis; JSON()`), and `unexpectedRefCall`
// naming both when it does not (`j()` calling JSON). A callee with no name of its own,
// `(a ? JSON : b)()`, is the second kind. One call reaching two of them, `(a ? JSON : Math)()`,
// reports twice, once for each.
var NoObjCalls = rule.Rule{
	Name: "no-obj-calls",

	// A local that shares a global's spelling is told apart from the global by asking where its
	// symbol is declared.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// Every trace starts at a reference to one of these names, so a file that never
				// spells any of them has nothing to follow, and the index is never built.
				if !spellsANonCallableGlobal(ctx.SourceFile.Text()) {
					return
				}
				tracker := reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)
				for _, tracked := range tracker.GlobalReferences(nonCallableTraceMap) {
					callee := tracked.Node.Expression()
					ctx.ReportNode(tracked.Node, objCallsMessage(calleeName(callee), tracked.Path[0]))
				}
			},
		}
	},
}

// spellsANonCallableGlobal is the cheap test that gates the trace.
func spellsANonCallableGlobal(text string) bool {
	for _, name := range nonCallableGlobals {
		if strings.Contains(text, name) {
			return true
		}
	}
	for _, name := range reference.DefaultGlobalObjectNames {
		if strings.Contains(text, name) {
			return true
		}
	}
	return false
}

// calleeName returns the name a callee is written with, which is ESLint's getReportNodeName: an
// identifier's own name, or the static name of the member it reads, through parentheses. A callee of
// any other shape has none.
func calleeName(callee *ast.Node) string {
	callee = ast.SkipParentheses(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		name, _ := property.Name(callee.AsPropertyAccessExpression().Name(), property.Named|property.Private)
		return name
	case ast.KindElementAccessExpression:
		name, _ := property.Name(callee.AsElementAccessExpression().ArgumentExpression, property.Quoted|property.Templated|property.Numeric)
		return name
	}
	return ""
}

// objCallsMessage builds the finding for one call: unexpectedCall when the callee spells the global,
// unexpectedRefCall when it reaches the global through another name or none.
func objCallsMessage(name string, global string) rule.Message {
	if name == global {
		return rule.Message{
			Id: "unexpectedCall",
			Description: fmt.Sprintf(
				"This calls `%s`, which is an object holding members rather than a function. "+
					"Calling or constructing it throws a TypeError at runtime every time this line "+
					"is reached. Call a method on it instead.", global),
		}
	}
	callee := "a value"
	if name != "" {
		callee = "`" + name + "`"
	}
	return rule.Message{
		Id: "unexpectedRefCall",
		Description: fmt.Sprintf(
			"This calls %s, which holds `%s`, an object holding members rather than a function. "+
				"Calling or constructing it throws a TypeError at runtime every time this line is "+
				"reached, and the other name hides that from a reader. Call a method on `%s` instead.",
			callee, global, global),
	}
}
