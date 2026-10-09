package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// A forward reference no ordering can remove, and that cannot run early.
//
// The rule's repair is "move the declaration above its first use", and for most forward references
// that is the whole answer. It has no answer for declarations that reference each other: two
// functions that call each other, or two ORM entity classes whose relation decorators each name the
// other. Whichever is written first refers forward to the second. Reporting both directions asks for
// an ordering that does not exist, so the only way to a clean file was a suppression (ruled by
// @system_cohere, 2026-10-03, from 15 such sites in api).
//
// So a forward reference is exempt only when both of these hold:
//
//   - It is part of a cycle among the statements of one block: the declaration it reaches reaches back,
//     directly or through other statements in the block, to the statement holding the reference. A
//     forward reference outside a cycle still reports, because moving the declaration fixes it.
//   - The reference cannot run before its target exists. Either the target is a function declaration,
//     which is initialized with its body before anything in the block runs, or the reference sits in a
//     nullary arrow or function passed straight to a decorator, the deferred-thunk idiom
//     (`@OneToMany(() => Post)`) that exists because the class it names is not initialized yet.
//
// Everything else in a cycle still reports, because it can run in the temporal dead zone: a static or
// field initializer, which runs as the class is defined; a thunk passed to an ordinary call, which may
// call it on the spot (`items.map(() => B)`); an immediately invoked function; a callback taking
// parameters. The one trust left is that a decorator does not call its thunk while the class is being
// defined. Such a decorator throws a ReferenceError on the file's first import, so it cannot hide.

// useBeforeDefineCycles answers the cycle question for one file, building each block's reference
// graph only when a candidate reference first needs it.
type useBeforeDefineCycles struct {
	ctx rule.Context
	// graphs maps a block to its statements' references to one another: statement to the statements
	// it reaches.
	graphs map[*ast.Node]map[*ast.Node]map[*ast.Node]bool
}

func newUseBeforeDefineCycles(ctx rule.Context) *useBeforeDefineCycles {
	return &useBeforeDefineCycles{ctx: ctx, graphs: map[*ast.Node]map[*ast.Node]map[*ast.Node]bool{}}
}

// exempts reports whether a forward reference from identifier to declaration is one no ordering can
// remove and that cannot run before declaration exists.
func (cycles *useBeforeDefineCycles) exempts(identifier *ast.Node, declaration *ast.Node) bool {
	if !runsAfterItsTargetExists(identifier, declaration) {
		return false
	}
	target := useBeforeDefineStatementOf(declaration)
	if target == nil {
		return false
	}
	source := useBeforeDefineStatementUnder(identifier, target.Parent)
	if source == nil || source == target {
		return false
	}
	return cycles.reaches(target.Parent, target, source)
}

// runsAfterItsTargetExists reports whether a reference cannot run before the declaration it names
// holds its value: the declaration is a function declaration, or the reference is in a nullary thunk
// handed straight to a decorator.
func runsAfterItsTargetExists(identifier *ast.Node, declaration *ast.Node) bool {
	if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
		return true
	}
	for ancestor := identifier.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind != ast.KindArrowFunction && ancestor.Kind != ast.KindFunctionExpression {
			if ast.IsFunctionLikeDeclaration(ancestor) || ast.IsClassStaticBlockDeclaration(ancestor) {
				return false
			}
			continue
		}
		// The innermost function holding the reference decides, so a thunk inside some other callback
		// is judged as that callback.
		if len(ancestor.Parameters()) != 0 {
			return false
		}
		call := ancestor.Parent
		if call == nil || call.Kind != ast.KindCallExpression || call.Parent == nil || call.Parent.Kind != ast.KindDecorator {
			return false
		}
		for _, argument := range call.AsCallExpression().Arguments.Nodes {
			if argument == ancestor {
				return true
			}
		}
		return false
	}
	return false
}

// useBeforeDefineStatementOf returns the statement a declaration belongs to in its block: the
// function or class declaration itself, or the variable statement around a variable.
func useBeforeDefineStatementOf(declaration *ast.Node) *ast.Node {
	for node := declaration; node != nil; node = node.Parent {
		if node.Parent != nil && isUseBeforeDefineBlock(node.Parent) {
			return node
		}
	}
	return nil
}

// useBeforeDefineStatementUnder returns the statement of block that holds node, or nil when node is
// not inside block.
func useBeforeDefineStatementUnder(node *ast.Node, block *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if current.Parent == block {
			return current
		}
	}
	return nil
}

func isUseBeforeDefineBlock(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

// reaches reports whether from reaches to through references among block's statements.
func (cycles *useBeforeDefineCycles) reaches(block *ast.Node, from *ast.Node, to *ast.Node) bool {
	graph := cycles.graph(block)
	seen := map[*ast.Node]bool{from: true}
	queue := []*ast.Node{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for next := range graph[current] {
			if next == to {
				return true
			}
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

// graph builds, once per block, which of its statements each statement references.
func (cycles *useBeforeDefineCycles) graph(block *ast.Node) map[*ast.Node]map[*ast.Node]bool {
	if graph, built := cycles.graphs[block]; built {
		return graph
	}
	graph := map[*ast.Node]map[*ast.Node]bool{}
	block.ForEachChild(func(statement *ast.Node) bool {
		var visit func(*ast.Node)
		visit = func(current *ast.Node) {
			if current.Kind == ast.KindIdentifier && !isDeclaringName(current) {
				declarations := rule.DeclarationsIn(cycles.ctx.SourceFile, useBeforeDefineResolveBinding(cycles.ctx, current))
				if declaration := useBeforeDefineBindingDeclaration(declarations); declaration != nil {
					if target := useBeforeDefineStatementOf(declaration); target != nil && target.Parent == block && target != statement {
						if graph[statement] == nil {
							graph[statement] = map[*ast.Node]bool{}
						}
						graph[statement][target] = true
					}
				}
			}
			current.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return false
			})
		}
		visit(statement)
		return false
	})
	cycles.graphs[block] = graph
	return graph
}
