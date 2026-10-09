package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUselessAssignment = rule.Message{
	Id: "noUselessAssignment",
	Description: "This assigns a value that nothing ever reads. Every path leaving this line either " +
		"ends or writes the variable again before any code looks at it, so the value computed here is " +
		"discarded. That is usually a symptom rather than the bug: the write was meant to be read and " +
		"an early return, a wrong branch, or a second assignment swallowed it, or the wrong variable " +
		"was named on the left. Remove the assignment, or use the value it produces.",
}

// NoUselessAssignment flags a write whose value no later read can observe, a dead store.
//
//	valid:   let v = 'a'; f(v); v = 'b'; f(v);
//	valid:   let v = 1; for (let i = 0; i < 10; i++) { f(v); v = 2; }
//	valid:   let v = 1; f(v); setTimeout(() => f(v)); v = 2;
//	valid:   let v; try { v = 1; h(); v = 2; } catch {} return v;
//	invalid: let v = 'a'; f(v); v = 'dead';
//	invalid: function f(c) { let v = 1; if (c) { v = 2; return; } g(v); }
//
// The value assigned here is computed and then thrown away. On its own that is only wasted work, but
// the shapes that produce it are nearly always a mistake somewhere else: a branch that returns before
// the value is used, a second write that lands first, or a left-hand side naming the wrong variable.
// Upstream's own corpus is mostly those three.
//
// # Why this is flow-sensitive and cannot be answered from the symbol table
//
// Whether a write is dead is not a property of the write. `v = 2` is dead in one file and live in
// the next, and nothing about the assignment node or the symbol distinguishes them. The question is
// whether any read of the same binding is reachable from this write without an intervening write,
// which is a path property of the program and needs the control flow graph. The symbol table can say
// that a binding is read somewhere; it cannot say that the read happens after this particular write,
// which is the entire discrimination. `let v = 1; f(v); v = 2;` and `let v = 1; v = 2; f(v);` have
// identical symbol tables and opposite verdicts.
//
// # Why the graph decides the algorithm, and which graph
//
// Upstream runs a backward liveness dataflow over oxc's control-flow graph: it walks blocks in
// reverse, unions the live sets of each block's successors, and reports a write whose bit is not
// live. That shape needs successor edges.
//
// TypeScript's own flow graph has none. `ast.FlowNode` carries `Antecedent` and `Antecedents` and
// nothing else, so every edge points from a node to what preceded it, because it exists for type
// narrowing rather than for this question. An earlier version of this rule inverted the question
// instead of the graph, asking each read whether its own antecedent chain passed back through the
// candidate write. That is the same predicate computed from the edges that existed, and it worked
// well enough to reproduce 33 of upstream's 54 diagnostics.
//
// `internal/lint/ecmascript/control_flow_graph` now supplies a real basic-block graph with forward `Successors`, so
// the question is asked in the direction it is posed and this is upstream's shape rather than a
// reconstruction of it. The analysis lives in `no_useless_assignment_analysis.go`.
//
// # What that recovered, measured against the same corpus
//
// The inversion reproduced 33 of 54 diagnostics and matched 27 of 43 failing inputs exactly. This
// reproduces all 54 and matches all 43. Five shapes it declined are recovered, and each one was
// declined for a reason the successor graph removes rather than for a reason that stopped being
// true: an update expression and a destructuring target now carry a write the graph places where it
// happens, a self-referential right-hand side is ordered before its own store by the graph's own
// evaluation order, a write inside a `try` block can kill an earlier write while still never being
// reported itself, and a destructured declarator is reached by walking up through its binding
// pattern.
//
// # The direction of error, and the thing that actually changed
//
// The inversion also shipped a false positive, and finding it is the more important half of this
// rewrite. A write inside an `if` with no `else` does not run on every path, so it cannot kill an
// earlier write; the inversion expressed the kill as a barrier in the flow graph, and a barrier
// stops the backward walk regardless of which path it sits on. Against the real tree that reported
// 34 writes at 34 locations and every one of them was wrong, all of the same shape: an accumulator
// seeded before a search, `let bestDistance = Infinity`, `let mostRecentTime = 0`,
// `let scope: string | null = null`. The release binary reports nothing on any of the twenty three
// files involved, with a control case in the same invocation firing to prove the rule was running.
//
// No imported case could see it, because upstream's corpus writes that shape only with an `else`
// arm present, where the kill is real and both implementations agree.
// `TestNoUselessAssignmentConditionalWritesDoNotKill` pins it now.
//
// # Scope of this port, stated rather than implied
//
// Measured against upstream's imported corpus of 71 clean and 43 failing inputs: this reproduces
// all 54 of upstream's diagnostics, matches all 43 failing inputs exactly, and reports nothing on
// any of the 71 clean cases.
//
// Two declines remain, each a silent miss rather than a wrong report:
//
//	unreachable writes            a write that exists only in a block control cannot arrive at is
//	                              not judged. `cohere --unused` reports the statement instead, and
//	                              names the exit that stranded it, which is the actionable finding.
//	captured bindings             a read from another code path root silences every write to that
//	                              binding. Upstream carries the same guard as `has_captured_read`
//	                              and it is exactly as coarse; see below.
//
// The captured-read guard was re-examined rather than inherited, because a successor graph is
// exactly what a narrowing would need. It cannot be narrowed here, and the reason is upstream's
// rather than ours: a closure's body is its own code path root with no edge to or from the
// enclosing one, so no ordering between a write and a captured read is available on either side.
// Upstream sets `has_captured_read` from any cross-scope read with no ordering condition and
// suppresses every write to that binding, and three of its clean cases exist to pin it. What DID
// separate cleanly is the other direction: a cross-root WRITE silences only itself rather than the
// whole binding, which is upstream's `has_same_parent_variable_scope` tested per write. Collapsing
// the two into one flag costs upstream's failing case 19, where a closure writes the variable and
// never reads it.
//
// No fix and no suggestion. Upstream offers neither, and its snapshot contains no fix output; the
// reason is that the right-hand side can have effects. Deleting `v = f()` removes the call, so the
// repair is not meaning-preserving and cannot be applied unattended. Even as a suggestion it would
// have to choose between deleting the statement and keeping the expression, and which one is right
// depends on whether the call matters.
var NoUselessAssignment = rule.Rule{
	Name: "no-useless-assignment",

	// Two separate needs. Symbol identity decides which occurrences name the same binding, so a
	// shadow in an inner block is not confused with the outer variable. And the flow nodes this
	// walks are populated by the binder as a side effect of building the program, so a run without
	// the checker sees `FlowNode == nil` everywhere and the rule goes silent rather than wrong.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				analyzeDeadStoresByLiveness(ctx, sourceFile)
			},
		}
	},
}

// isLocalVariableDeclaration reports whether a declaration is a `let` or `var` this rule may judge.
//
// A `const` cannot be reassigned, so its initializer is the only write and reporting it would
// duplicate the unused-variable judgment rather than the dead-store one. A parameter, a function, a
// class, and an import are all declined for the same reason: the write this rule reasons about is a
// store into a mutable local.
func isLocalVariableDeclaration(declaration *ast.Node) bool {
	// A destructured binding declares through a chain of binding elements and patterns rather than
	// directly on the declarator, so the declarator is walked up to rather than required to be the
	// node itself. Measured: in `let [a, b] = arr` the symbol's declaration is a
	// `KindBindingElement` whose parent is a `KindArrayBindingPattern` and whose grandparent is the
	// `KindVariableDeclaration`. Requiring the node's own kind declined every destructured binding,
	// which cost the two diagnostics upstream reports on exactly that shape.
	for declaration != nil && (declaration.Kind == ast.KindBindingElement ||
		declaration.Kind == ast.KindObjectBindingPattern ||
		declaration.Kind == ast.KindArrayBindingPattern) {
		declaration = declaration.Parent
	}
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	// The declarator sits inside a declaration list that carries the let/const/var flag.
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	if list.Flags&ast.NodeFlagsConst != 0 {
		return false
	}

	// An exported binding can be read by another module, so no write to it is provably dead. The
	// statement wrapping the list carries the modifier.
	statement := list.Parent
	if statement != nil && statement.Kind == ast.KindVariableStatement {
		if statement.ModifierFlags()&ast.ModifierFlagsExport != 0 {
			return false
		}
	}
	return true
}

// exportedNames collects every binding name the file hands out through any export form.
//
// The modifier check above sees `export let foo` and nothing else, which is the smaller half of the
// surface. A binding is equally exported by a later `export { foo }`, by `export { foo as bar }`,
// and by `export default foo`, and in every one of those the declaration carries no modifier at all.
// Upstream reads its module record's `exported_bindings` map, which is populated by all of them; we
// have no module record here, so the export clauses are walked directly.
//
// Collected in one pass and consulted by name, rather than walked per candidate. The per-candidate
// form was half of a 3.3 second cost on the real tree.
//
// Measured rather than reasoned: before this existed the rule reported
// `let foo = 'used'; export { foo }; console.log(foo); foo = 'unused like but exported';`, which is
// upstream clean case 15 and exists precisely to pin this. That was the last false positive on the
// imported corpus.
func exportedNames(sourceFile *ast.Node) map[string]bool {
	exported := map[string]bool{}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindExportSpecifier:
			specifier := node.AsExportSpecifier()
			// `export { foo }` names the local in `Name`; `export { foo as bar }` puts the local in
			// `PropertyName` and the public name in `Name`. The local is the one that matters,
			// because that is the binding writes here target.
			local := specifier.Name()
			if specifier.PropertyName != nil {
				local = specifier.PropertyName
			}
			if local != nil {
				exported[local.Text()] = true
			}
		case ast.KindExportAssignment:
			// `export default foo` and `export = foo` both hand the binding out whole.
			if expression := node.AsExportAssignment().Expression; expression != nil &&
				expression.Kind == ast.KindIdentifier {
				exported[expression.Text()] = true
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)
	return exported
}

// isInsideTryBlock reports whether a node sits in the `try` block of a try statement.
//
// Upstream refuses to report any write inside a try, under the name `is_in_try_block`, and the
// reason is that a try block has an edge out of every operation rather than only its end. A write
// there may be abandoned by a throw from the next line, so the value it overwrote is still
// observable on the catch and finally paths, and calling it dead is wrong. Three of upstream's clean
// cases are exactly this shape, including one where the try nests two deep.
//
// The catch and finally clauses are not covered: code there runs after the throw has been handled,
// so it has ordinary flow and an ordinary dead store is a real one.
func isInsideTryBlock(node *ast.Node) bool {
	for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
		if parent.Kind == ast.KindTryStatement && parent.AsTryStatement().TryBlock == child {
			return true
		}
	}
	return false
}

// resolveOccurrence reports the binding an identifier occurrence refers to.
//
// Plain `GetSymbolAtLocation` is wrong for one shape and only one: a shorthand property. In
// `return { v }` the identifier resolves to the *property's* symbol rather than to the variable it
// reads, so the read is filed under a symbol nothing else touches and the variable looks unread.
// The effect is not a missed finding but a false positive, because a write whose only read is a
// shorthand property then looks dead.
//
// Found on real code rather than on the corpus, which contains no shorthand at all. The dry run
// reported 50 findings in one file, and every one of them traced to
// `let newStartTime: Date; switch (...) { ... newStartTime = ...; } return { newStartTime, ... }`.
// Reading those findings is what surfaced it; the imported corpus was green throughout.
func resolveOccurrence(ctx rule.Context, identifier *ast.Node) *ast.Symbol {
	if parent := identifier.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
		if symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
			return symbol
		}
	}
	return ctx.TypeChecker.GetSymbolAtLocation(identifier)
}

// isNonReferenceIdentifier reports whether an identifier occurrence cannot be a reference to a local
// variable, so it need never be resolved or scanned.
//
// A property name in `o.v` or `{ v: 1 }` shares its spelling with a variable and resolves to a
// different symbol, so keeping it was never wrong, only wasteful. A shorthand property is
// deliberately absent from this list: it looks like a property name and is a real read, which is the
// distinction `resolveOccurrence` exists to draw.
func isNonReferenceIdentifier(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// `o.v` reads `o` and names `v`; only the name half is skipped.
		return parent.AsPropertyAccessExpression().Name() == identifier
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() == identifier
	case ast.KindQualifiedName, ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
		ast.KindPropertySignature, ast.KindMethodSignature, ast.KindGetAccessor,
		ast.KindSetAccessor, ast.KindEnumMember, ast.KindImportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindTypeParameter,
		ast.KindTypeReference, ast.KindJsxAttribute:
		return parent.Name() == identifier
	}
	return false
}
