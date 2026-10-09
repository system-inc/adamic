package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageIncompatibleLibrary = rule.Message{
	Id: "incompatibleLibrary",
	Description: "This library's API returns functions that cannot be memoized without showing " +
		"stale data. The API hands back a new function on every render on purpose, because that " +
		"is how it tells you something changed. Memoizing the component that calls it, by hand " +
		"or by compiler, freezes that function at its first value and the screen stops following " +
		"the data. React Compiler skips optimizing this component rather than break it, so the " +
		"component silently loses memoization. Read the value out at the call site and pass the " +
		"plain data onward, rather than passing this API's own functions into memoized code.",
}

// IncompatibleLibrary flags a call to a library API the React Compiler refuses to memoize around.
//
//	valid:   function Component(props) { return <div>{props.text}</div>; }
//	valid:   import {useReactTable} from '@tanstack/react-table';   (imported, never called)
//	valid:   function helper() { return useReactTable(); }          (not a component or hook)
//	invalid: function Component() { const t = useReactTable(); return <div>{t}</div>; }
//	invalid: function Component() { const {watch} = useForm(); return <div>{watch()}</div>; }
//
// # What this rule is, and why the measurement that unblocked it understates it
//
// The research pass that cleared this rule described it as "a config mechanism, not three names",
// on the evidence that the three shipped entries live outside `globals.rs` and that all three
// upstream fixtures import a synthetic module a test harness registers. Both observations are
// correct and the conclusion drawn from them is not, which matters because it decides whether this
// rule can ever fire on this tree.
//
// The distinction is against `configuration.go`, which really is the shape that measurement describes.
// That rule's default table is *self-consistent at every entry*, so it is structurally unable to
// report without a user-supplied provider cohere has no channel for. This one is the opposite:
// every one of the three default entries carries a live `knownIncompatible` message, and the
// message is the entire firing condition. The table is not a schema waiting to be filled, it is
// three real libraries with three real diagnostics already in it.
//
// **Measured on the executable rather than argued.** React 7.1.1 driven through the ESLint Linter
// interface, with no configuration beyond enabling the rule:
//
//	import {useReactTable} from '@tanstack/react-table';    reports
//	import {useVirtualizer} from '@tanstack/react-virtual'; reports
//	import {useForm} from 'react-hook-form'; watch() called  reports
//	import {useSomething} from 'some-random-module';        silent
//
// So the honest port is the three names, and this rule fires on real source with no configuration
// at all. The seam below exists so the synthetic-module fixtures can also be run, not because the
// shipped table is unreachable.
//
// The second, independent confirmation is oxc's own lint corpus. oxc *does* ship a lint rule here,
// which is a correction to the standing belief that this family has none on that side: its single
// failing case imports the genuine `@tanstack/react-table` rather than the synthetic module, which
// is only a sensible corpus if the default table fires. React and oxc agree on the span and on the
// message text, so there is no divergence to record for this rule.
//
// # What decides a finding
//
// Three conditions, all measured against React 7.1.1 rather than read off the emit site:
//
//	the callee resolves to a named import from a module in the table
//	the entry for that imported name carries an incompatible message
//	the call happens inside a component or a hook
//
// The third is load-bearing and is the whole reason this rule is not a grep. Measured, holding the
// call fixed and moving only the enclosing function's name:
//
//	function Component() { useReactTable(); }   reports
//	function useThing()  { useReactTable(); }   reports
//	function helper()    { useReactTable(); }   silent
//	const t = useReactTable();                  silent   (top level)
//	class Widget { render() { useReactTable(); } }  silent
//
// It has to be *called*. An import that is never used, or one passed onward as a value, is silent
// at upstream, because the diagnostic is raised while inferring the effects of a call rather than
// while resolving the import.
//
// # One finding per function, and why that is not a bug to fix
//
// Upstream `throw`s at the emit site rather than pushing onto an error list, which aborts
// compilation of the enclosing function. Measured consequence: a component calling two different
// incompatible APIs reports **once**, on the first in source order, and the same call written twice
// also reports once. Reproduced here rather than corrected, since reporting both would be a
// different rule than the one being ported, and the count is a visible behavior a differential run
// would show.
//
// # Where this port is narrower than upstream, stated with the command that established it
//
// A local alias defeats it: `const local = useReactTable; local();` is **silent at React too**,
// measured, so the narrowness is upstream's and is reproduced rather than introduced. But this port
// is genuinely narrower in one place, and it is worth naming rather than leaving for a reader to
// discover. Upstream tracks the *value* through its intermediate representation, so it follows the
// import through a destructure of a hook's return object. This port matches syntactically. The
// consequence is confined to `react-hook-form`, the only entry whose message hangs off a nested
// property, and it is handled by matching the destructured property name against the entry's nested
// table. What is not handled is an arbitrary chain of re-assignments between the hook call and the
// property access, which needs the binding-resolution pass this rule deliberately does not build.
//
// Deliberately NOT reproduced: with a JavaScript-only parser, ESLint reads `import type {x}` as a
// default import named `type` and reports at the specifier with an empty span. That is an artifact
// of parsing TypeScript syntax without a TypeScript parser, not a judgment about the code, and our
// parser gets the shape right. Measured and declined rather than ported.
//
// # No intermediate representation, no type checker, and why
//
// Neither is needed and both were considered rather than skipped. The question this rule answers is
// "does this call site name an import from one of three modules, inside a component", which is
// answered by the import clause and the enclosing function's name. The checker would add nothing:
// the modules are third-party packages that a fixture cannot resolve and that carry no useful type
// information for this decision, and the module specifier is compared as a *string* at both
// upstreams. Declaring the checker would make every fixture need the typed harness to answer a
// question no type is consulted for.
var IncompatibleLibrary = rule.Rule{
	// Bare name, no namespace prefix. The inventory writes `react/incompatible-library` and the
	// parity guard tries an exact match before stripping the namespace, so the slash spelling would
	// also match while being the wrong thing to write.
	Name: "react-hooks/incompatible-library",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// A source-file pre-pass rather than a call-expression listener, because the
				// judgment needs the file's imports before it can read any call, and the walk is
				// pre-order with no exit hook. `KindSourceFile` fires before its children, so the
				// imports are gathered and the calls are judged in one pass here.
				incompatible := incompatibleLibraryBindings(node)
				if len(incompatible.direct) == 0 && len(incompatible.indirect) == 0 {
					return
				}
				incompatibleLibraryReport(ctx, node, incompatible)
			},
		}
	},
}

// incompatibleLibraryEntry is one library API the compiler refuses to memoize around.
//
// `message` is the per-entry text upstream prints beside the span, and it is the firing condition
// rather than decoration: an entry with no message is a configured API that memoizes fine.
//
// `properties` carries the nested case. `react-hook-form` declares nothing incompatible about
// `useForm` itself; the incompatibility is on the `watch` function its return object holds, which
// is why calling `useForm()` and never touching `watch` is silent.
type incompatibleLibraryEntry struct {
	message    string
	properties map[string]string
}

// incompatibleLibraryTable is the module table, and the seam a test can substitute at.
//
// A variable rather than a direct call for the same reason `configuration.go` has one, though the reason
// is weaker here and the difference is worth stating so a reader does not copy the wrong lesson.
// There, the seam is the *only* way to make the rule report at all. Here the shipped table fires on
// its own, and the seam exists so upstream's three fixtures — which import a module registered by a
// test harness rather than a real package — can be run as written rather than rewritten against a
// different library. Both proofs are then available: the shipped table reporting on real names, and
// upstream's own corpus reporting on its synthetic one.
var incompatibleLibraryTable = defaultIncompatibleLibraryTable

// defaultIncompatibleLibraryTable returns the incompatible APIs a module declares, if it declares any.
//
// Verbatim from React 7.1.1's `defaultModuleTypeProvider`, and oxc's
// `default_module_type_provider.rs` carries the same three entries with the same three message
// strings. The messages are copied rather than paraphrased because upstream prints them as the
// per-finding text, so a reworded copy is a visible divergence.
//
// The module name is compared **exactly**, and that is measured rather than assumed. Unlike
// `isKnownReactModule` in `configuration.go`, which lowercases before comparing, this comparison is
// case-sensitive and does not accept subpaths: `@TanStack/React-Table` and
// `@tanstack/react-table/core` are both silent at React 7.1.1. A lowercasing or prefix-matching
// version of this function would report where upstream does not.
func defaultIncompatibleLibraryTable(moduleName string) (map[string]incompatibleLibraryEntry, bool) {
	switch moduleName {
	case "react-hook-form":
		// The only nested entry. `useForm` is a hook that memoizes fine; its returned `watch`
		// function is what cannot be. Measured: `const f = useForm();` alone is silent, and
		// `const {watch} = useForm(); watch()` reports.
		return map[string]incompatibleLibraryEntry{
			"useForm": {properties: map[string]string{
				"watch": "React Hook Form's `useForm()` API returns a `watch()` function which cannot be memoized safely.",
			}},
		}, true
	case "@tanstack/react-table":
		return map[string]incompatibleLibraryEntry{
			"useReactTable": {message: "TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely"},
		}, true
	case "@tanstack/react-virtual":
		return map[string]incompatibleLibraryEntry{
			"useVirtualizer": {message: "TanStack Virtual's `useVirtualizer()` API returns functions that cannot be memoized safely"},
		}, true
	}
	return nil, false
}

// incompatibleLibraryLocals is what one file's imports bring into scope from the table.
//
// Two maps because the two shapes are judged at different nodes. `direct` is keyed by the local
// name a call must use as its callee. `indirect` is keyed by the local name of a hook whose *result*
// carries the incompatible property, and its value is that hook's nested property table.
type incompatibleLibraryLocals struct {
	direct   map[string]string
	indirect map[string]map[string]string
}

// incompatibleLibraryBindings reads a file's imports and returns the local names that reach the table.
//
// Local names rather than imported ones, because the callee at the use site spells the local. The
// *lookup* into the table uses the imported name, which is what makes an alias work: measured at
// React 7.1.1, `import {useReactTable as renamed}` still reports, with the span on `renamed`.
//
// A namespace import is deliberately not collected. Upstream does report on
// `table.useReactTable()`, with the span on `table` rather than on the property, and reproducing
// that is a member-access path this port does not build. Recorded at the declining site rather than
// silently omitted, and asserted as a known gap in the tests.
func incompatibleLibraryBindings(sourceFile *ast.Node) incompatibleLibraryLocals {
	locals := incompatibleLibraryLocals{
		direct:   map[string]string{},
		indirect: map[string]map[string]string{},
	}
	for _, statement := range sourceFile.AsSourceFile().Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		moduleName, named := incompatibleLibraryModuleName(statement)
		if !named {
			continue
		}
		entries, configured := incompatibleLibraryTable(moduleName)
		if !configured {
			continue
		}
		for _, element := range imports.BindingsOf(statement).Named {
			entry, present := entries[imports.ImportedNameOf(element)]
			if !present {
				continue
			}
			localName := incompatibleLibraryLocalName(element)
			if localName == "" {
				continue
			}
			if entry.message != "" {
				locals.direct[localName] = entry.message
			}
			if len(entry.properties) > 0 {
				locals.indirect[localName] = entry.properties
			}
		}
	}
	return locals
}

// incompatibleLibraryReport walks the file and reports the first offending call in each component.
//
// The per-function bookkeeping reproduces upstream's `throw`: the diagnostic aborts compilation of
// the enclosing function, so a second offending call in the same function never runs. Keyed by the
// enclosing function node rather than by a single flag, because two sibling components in one file
// each report their own first call.
func incompatibleLibraryReport(ctx rule.Context, sourceFile *ast.Node, locals incompatibleLibraryLocals) {
	reported := map[*ast.Node]bool{}

	// Property accesses reached through a destructure of an incompatible hook's result, keyed by
	// the local name bound to the property. Gathered as the walk descends rather than in a separate
	// pass, which is sound because a `const {watch} = useForm()` declaration precedes any use of
	// `watch` in the source order the walk follows.
	destructured := map[string]string{}

	// Locals bound to an incompatible hook's whole result, keyed by the local name, carrying that
	// hook's property table. This is the `const form = useForm()` half, and it is a different map
	// from `destructured` because the finding lands at a different node: here the call is
	// `form.watch()` and upstream's span is the OBJECT, while a destructured property is called
	// bare and the span is the property's own local.
	results := map[string]map[string]string{}

	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			incompatibleLibraryCollectDestructure(node, locals, destructured, results)
		}
		if node.Kind == ast.KindCallExpression {
			if span, message, offending := incompatibleLibraryCallMessage(node, locals, destructured, results); offending {
				enclosing := incompatibleLibraryEnclosingRoot(node)
				if enclosing != nil && !reported[enclosing] {
					reported[enclosing] = true
					ctx.ReportNode(span, rule.Message{
						Id:          messageIncompatibleLibrary.Id,
						Description: messageIncompatibleLibrary.Description + " " + message,
					})
				}
			}
		}
		node.ForEachChild(walk)
		return false
	}
	sourceFile.ForEachChild(walk)
}

// incompatibleLibraryCollectDestructure records `const {watch} = useForm()` shaped bindings.
//
// Only the object-pattern form, and only one level deep, which is the shape upstream's own fixture
// writes and the shape the single nested entry needs. A renamed destructure (`{watch: w}`) binds the
// property under a different local, so the *property* name decides the lookup and the *local* name
// is what the call site will spell. Measured at React 7.1.1: the renamed form reports, with the span
// on the renamed local.
func incompatibleLibraryCollectDestructure(declaration *ast.Node, locals incompatibleLibraryLocals, destructured map[string]string, results map[string]map[string]string) {
	node := declaration.AsVariableDeclaration()
	if node == nil || node.Initializer == nil || node.Name() == nil {
		return
	}
	if node.Initializer.Kind != ast.KindCallExpression {
		return
	}
	callee := node.Initializer.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return
	}
	properties, incompatible := locals.indirect[callee.Text()]
	if !incompatible {
		return
	}
	// `const form = useForm()` binds the whole result. Recorded so a later `form.watch()` can be
	// matched, which is a separate shape from destructuring the property out here.
	if node.Name().Kind == ast.KindIdentifier {
		results[node.Name().Text()] = properties
		return
	}
	if node.Name().Kind != ast.KindObjectBindingPattern {
		return
	}
	for _, element := range node.Name().AsBindingPattern().Elements.Nodes {
		binding := element.AsBindingElement()
		if binding == nil || binding.Name() == nil || binding.Name().Kind != ast.KindIdentifier {
			continue
		}
		// `PropertyName` is set only when the destructure renames, so it is the source-side name
		// when present and the local name doubles as it when absent.
		propertyName := binding.Name().Text()
		if binding.PropertyName != nil && binding.PropertyName.Kind == ast.KindIdentifier {
			propertyName = binding.PropertyName.Text()
		}
		if message, present := properties[propertyName]; present {
			destructured[binding.Name().Text()] = message
		}
	}
}

// incompatibleLibraryCallMessage answers whether a call is to an incompatible API, with what text,
// and at which node the finding points.
//
// The span is returned rather than assumed by the caller because the three shapes point at three
// different nodes, all measured against React 7.1.1:
//
//	useReactTable()          span is the callee identifier
//	watch()                  span is the destructured local
//	form.watch()             span is the OBJECT, not the property
//	useForm().watch()        span is the whole inner call expression
//
// The third is the one a reader would guess wrong, and a fixture asserting the property was written
// here and corrected by the measurement rather than the other way round.
func incompatibleLibraryCallMessage(
	call *ast.Node,
	locals incompatibleLibraryLocals,
	destructured map[string]string,
	results map[string]map[string]string,
) (*ast.Node, string, bool) {
	callee := call.AsCallExpression().Expression
	if callee == nil {
		return nil, "", false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		if message, present := locals.direct[callee.Text()]; present {
			return callee, message, true
		}
		if message, present := destructured[callee.Text()]; present {
			return callee, message, true
		}
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access.Expression == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return nil, "", false
		}
		propertyName := access.Name().Text()

		// `form.watch()`, where `form` holds a previously-bound hook result. The span is the
		// object, which is what upstream underlines.
		if access.Expression.Kind == ast.KindIdentifier {
			if properties, incompatible := results[access.Expression.Text()]; incompatible {
				if message, present := properties[propertyName]; present {
					return access.Expression, message, true
				}
			}
			return nil, "", false
		}

		// `useForm().watch()`, chained with no intervening binding. Upstream's span is the whole
		// inner call, which is the object of this access.
		if access.Expression.Kind == ast.KindCallExpression {
			inner := access.Expression.AsCallExpression().Expression
			if inner == nil || inner.Kind != ast.KindIdentifier {
				return nil, "", false
			}
			if properties, incompatible := locals.indirect[inner.Text()]; incompatible {
				if message, present := properties[propertyName]; present {
					return access.Expression, message, true
				}
			}
		}
	}
	return nil, "", false
}

// incompatibleLibraryEnclosingRoot returns the component or hook a call sits inside, if any.
//
// This is the component gate, and it reuses the react shelf's `IsComponentOrHookLike` and
// `IsReachableRootPosition` whole rather than restating them. Those two were derived over
// seventeen probe rounds against React's own rule and are already shared by three other rules here;
// a fourth private copy would be the divergence this tree's shelf rule exists to prevent.
//
// The node is returned rather than a boolean because the caller needs an identity to key the
// one-finding-per-function bookkeeping on.
func incompatibleLibraryEnclosingRoot(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if utilsreact.IsComponentOrHookLike(current) && utilsreact.IsReachableRootPosition(current) {
			return current
		}
	}
	return nil
}

// incompatibleLibraryLocalName reads the local name an import specifier binds.
//
// Separate from `imports.ImportedNameOf`, which reads through the alias to the module-side name.
// This rule needs both halves and they differ exactly when the import renames: the table is keyed by
// the imported name and the call site spells the local one.
func incompatibleLibraryLocalName(specifier *ast.Node) string {
	if specifier == nil || specifier.Kind != ast.KindImportSpecifier {
		return ""
	}
	if name := specifier.Name(); name != nil && name.Kind == ast.KindIdentifier {
		return name.Text()
	}
	return ""
}

// incompatibleLibraryModuleName reads the specifier text of an import declaration.
//
// The kind check is CRASH PROTECTION rather than a behavioral filter, and the distinction was
// established by a surviving mutant rather than by reading. Deleting it survives the whole suite,
// because no fixture asserting findings can see a panic. Probed over the shapes our parser produces
// for a malformed specifier, three of them take the linter down on `Text()`:
// `KindTemplateExpression`, `KindPropertyAccessExpression` and `KindParenthesizedExpression`. The
// last is the sharpest, since `from ('@tanstack/react-table')` carries a name that would match.
//
// The behavioral reading was tried first and is wrong: a recovered `KindIdentifier` does answer
// `Text()` cleanly, but no identifier can spell a scoped package name, so both versions reach
// silence there by different routes. `TestIncompatibleLibrarySurvivesAMalformedModuleSpecifier`
// pins the panics.
//
// `ast.IsStringLiteralLike` also accepts a no-substitution template, which is a genuine specifier
// rather than recovered input, so narrowing this to `ast.IsStringLiteral` would lose a real report.
func incompatibleLibraryModuleName(node *ast.Node) (string, bool) {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return "", false
	}
	if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return "", false
	}
	return declaration.ModuleSpecifier.Text(), true
}
