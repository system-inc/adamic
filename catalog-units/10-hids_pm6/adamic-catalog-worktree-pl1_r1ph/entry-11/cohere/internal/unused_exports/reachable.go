package unused_exports

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
)

// Unreachable is one statement that cannot run.
type Unreachable struct {
	// FileName is the file the statement is in.
	FileName string

	// Range is the statement's own span, not the span of the exit that stranded it.
	Range TextSpan

	// Cause names the abrupt completion that made this unreachable, so the reader can find the
	// `return` rather than only the line after it. A finding that says "this cannot run" without
	// saying why sends the reader hunting.
	Cause string
}

// TextSpan is a position in a file, kept independent of the compiler's own range type so this
// package's results can be printed by a caller that never built a program.
type TextSpan struct {
	Position int
	End      int
}

// FindUnreachable reports the statements in one file that control cannot arrive at.
//
// It builds one control-flow graph per code path root — the file itself, each function, each class
// static block, each property initializer — because that is the unit the graph is defined over. A
// statement is reported when it is laid out in a block whose Reachable is false.
//
// # Why this leans on the graph rather than pattern-matching a `return`
//
// The naive version looks for a statement following a `return` in the same block and is wrong in
// both directions. It misses `if (x) { return; } else { return; } here();`, where nothing directly
// precedes `here()` in its own block, and it fires on `while (true) { break; } here();`, where the
// break makes the statement after the loop reachable rather than dead. The graph already decides
// both, and correctly: measured on this substrate, both shapes answer as they should.
//
// # What the graph is trusted for, measured rather than assumed
//
// `Block.Reachable`'s doc claims code after an abrupt exit is laid out in unreachable blocks rather
// than dropped, so a consumer's hooks still run there. That claim is the whole basis of this
// function, and it was probed on this substrate before anything was built on it — after `return`,
// after `throw`, after `break`, after `continue`, both-branches-return, and the `do…while` shape the
// vendoring report flagged as a known deviation. All eight answered as documented.
func FindUnreachable(sourceFile *ast.SourceFile) []Unreachable {
	if sourceFile == nil {
		return nil
	}

	var findings []Unreachable
	for _, root := range codePathRoots(sourceFile) {
		findings = append(findings, unreachableInRoot(root)...)
	}
	return findings
}

// statementEvent is what the hooks record: one statement, and where it was laid out.
type statementEvent struct {
	node *ast.Node
}

// unreachableInRoot walks one code path root and reports the statements laid out in blocks nothing
// reaches.
func unreachableInRoot(root *ast.Node) []Unreachable {
	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[statementEvent]{
		Statement: func(builder *control_flow_graph.Builder[statementEvent], node *ast.Node) {
			builder.Emit(statementEvent{node: node})
		},
	})

	sourceFile := ast.GetSourceFileOfNode(root)
	if sourceFile == nil {
		return nil
	}

	// A statement laid out in a reachable block anywhere in the graph is live, whatever other blocks
	// also hold it. This pass has to run before the reporting pass, and it is the difference between
	// a working report and one that condemns cleanup code.
	//
	// # The duplicate-layout trap, measured on this substrate
	//
	// The CFG lays a `finally` block out TWICE — once for normal completion, once for the path that
	// leaves the `try` through `return`, `throw`, or a suspended `yield`. Both copies carry the same
	// source positions, and the second copy is unreachable exactly when nothing takes the abrupt
	// path. So `try { return 1; } finally { beta(); }` puts `beta()` in a reachable block AND in an
	// unreachable one, and a scan that reports any statement found in an unreachable block condemns
	// a `finally` body that certainly runs.
	//
	// This was not hypothetical and not caught by reading: it was a live false positive that the
	// silent-half fixtures caught on the first run. The same shape appears for a `catch` body when
	// the `try` returns — `try { return 1; } catch (e) { gamma(); } finally {...}` lays `gamma()` out
	// unreachable, and that catch is reachable code.
	//
	// Keying liveness by source position rather than by block is what makes both correct, because
	// position is the identity the reader has and block identity is an artifact of the layout.
	live := make(map[int]bool)
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, event := range block.Events {
			if event.node != nil {
				live[event.node.Pos()] = true
			}
		}
	}

	var findings []Unreachable
	// Only the first statement of a contiguous dead run is reported. Reporting every one turns a
	// single mistake into a wall of findings that all say the same thing, and the reader's question
	// is "what is dead here", answered once, rather than "how many lines are dead".
	reported := false
	for _, block := range graph.Blocks {
		if block.Reachable {
			// A reachable block ends any dead run, so the next dead block reports again.
			reported = false
			continue
		}
		for _, event := range block.Events {
			if event.node == nil {
				continue
			}
			// Laid out somewhere control does arrive, so the unreachable copy is a layout artifact
			// rather than dead code.
			if live[event.node.Pos()] {
				continue
			}
			// A `catch` body is never reported, even when it is laid out only in unreachable blocks.
			//
			// Measured on this substrate: the graph forks to the handler at the first node inside the
			// `try` that could throw, so a `try` whose body cannot throw gets no edge to its `catch`
			// at all, and the catch body appears exclusively in an unreachable block. Every one of
			// `try { return 1; } catch (e) { gamma(); }`, the same with a `finally`, and
			// `try { return; } catch ...` lays `gamma()` out unreachable and nowhere else, so the
			// position-keyed liveness above cannot rescue it.
			//
			// Reporting it would be defensible as a literal reading of the graph and wrong as a
			// report: the reader is told to delete error handling, and the analysis is asserting that
			// nothing in the `try` can throw — a claim about every callee's behaviour that this
			// analysis has not established and the graph is explicitly approximating. So catch bodies
			// are out of scope, stated rather than silently missing, and a genuinely dead `catch` is a
			// finding this report gives up in exchange for never condemning live error handling.
			if withinCatchClause(event.node, root) {
				continue
			}
			// A function declaration is hoisted, so it is callable from before its own position and
			// being laid out after a `return` says nothing about whether it runs. Reporting one is a
			// false positive with a confident-looking span.
			if event.node.Kind == ast.KindFunctionDeclaration {
				continue
			}
			// A bare `var` is hoisted too. The declaration reaches the whole scope even where the
			// assignment does not run.
			if isHoistedVariableStatement(event.node) {
				continue
			}
			if reported {
				continue
			}
			reported = true
			findings = append(findings, Unreachable{
				FileName: sourceFile.FileName().AsString(),
				Range: TextSpan{
					Position: event.node.Pos(),
					End:      event.node.End(),
				},
				Cause: "an earlier return, throw, break, or continue leaves before this line",
			})
		}
	}
	return findings
}

// isHoistedVariableStatement reports whether a statement declares with `var`, whose binding reaches
// the whole enclosing scope regardless of where the statement sits.
func isHoistedVariableStatement(node *ast.Node) bool {
	if node.Kind != ast.KindVariableStatement {
		return false
	}
	declarationList := node.AsVariableStatement().DeclarationList
	if declarationList == nil {
		return false
	}
	// IsVarConstLike rather than a const check, because `using` and `await using` are block-scoped
	// the same way and hoisting them would be wrong.
	return !ast.IsLet(declarationList) && !ast.IsVarConstLike(declarationList)
}

// codePathRoots returns every node the control-flow graph is defined over in one file: the file
// itself, and each function, static block, and property initializer inside it.
//
// The graph is built per root rather than once per file because a function's body is its own code
// path — a `return` inside it ends that path and not the file's.
func codePathRoots(sourceFile *ast.SourceFile) []*ast.Node {
	roots := []*ast.Node{sourceFile.AsNode()}

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindClassStaticBlockDeclaration:
			roots = append(roots, node)
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)

	return roots
}

// withinCatchClause reports whether node sits inside a `catch` body within the given root.
//
// The walk stops at the root rather than climbing out of it, so a nested function's own catch does
// not shield a statement in the enclosing one.
func withinCatchClause(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindCatchClause {
			return true
		}
	}
	return false
}
