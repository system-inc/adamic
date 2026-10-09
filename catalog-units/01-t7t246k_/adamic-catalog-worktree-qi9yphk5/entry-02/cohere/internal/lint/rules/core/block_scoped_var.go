package core

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// BlockScopedVar reports a `var` used outside the block it was declared in.
//
//	valid:   function f() { var a = 1; a = 2; }
//	valid:   function f() { { let a = 1; } }
//	invalid: function f() { { var a = 1; } a = 2; }
//	invalid: if (true) { var a; } a;
//	invalid: for (var i = 0;;) {} i;
//
// # What the rule actually asks
//
// `var` is function-scoped, so a declaration inside a block is visible after that block ends, and
// the language will not stop you reading it there. That is legal and almost never intended: the
// braces say one thing and the binding does another, and code moved out of the block later breaks
// in a way the original author could not have predicted.
//
// So the rule pretends `var` were block-scoped and reports every use that would then be out of
// scope. It is not a style preference: every finding is a place where the written structure and the
// real binding disagree.
//
// # The block, not the scope
//
// The container is the nearest enclosing BLOCK-like node, which is a block statement, a loop, a
// switch, a catch clause, a class static block, or the file itself. Notably NOT a function: a
// function body is a block, so a `var` at the top of one is contained by that block and every use
// inside the function is fine.
//
// Upstream maintains this as a stack pushed and popped by a dozen listeners. We have parents, so
// the same container is found by walking up from the declaration, which is the same answer without
// the bookkeeping.
//
// # Every declaration of a merged name is checked against every reference
//
// `if (a) { var x = 1; } else { var x = 2; }` declares one variable twice, and both spellings are
// out of the other's block, so it reports FOUR times: each declaration name is itself a reference,
// and each is outside the other declaration's block. That reads like over-reporting and it is
// upstream's behaviour, measured; a port checking only non-declaration references reports zero here,
// on a case upstream reports six findings for at three branches.
var BlockScopedVar = rule.Rule{
	Name:             "block-scoped-var",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(file *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				checkBlockScopedVar(ctx, file)
			},
		}
	},
}

// blockScopedVarFinding is one out-of-scope use paired with the declaration it escapes.
//
// Collected rather than reported as they are found, because the gathering loop is nested
// declaration-outer and reference-inner, which groups findings by declaration while upstream's come
// out ordered by where the USE is. `blockScopedVarReport` restores that order.
type blockScopedVarFinding struct {
	reference       *ast.Node
	declarationName *ast.Node
}

// checkBlockScopedVar walks a file once and reports every out-of-scope `var` use.
//
// One pass over the tree gathers both halves: the `var` declaration names, and every identifier that
// could be a reference. Asking the checker per identifier is what connects them, and it is asked
// once per identifier rather than once per (declaration, identifier) pair, so the cost is linear
// rather than quadratic in a file full of `var`.
func checkBlockScopedVar(ctx rule.Context, file *ast.Node) {
	var declarations []*ast.Node
	var identifiers []*ast.Node

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			identifiers = append(identifiers, node)
		}
		if node.Kind == ast.KindVariableDeclaration && blockScopedVarIsVarDeclaration(node) {
			blockScopedVarCollectBindingNames(node.Name(), &declarations)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(file)
	if len(declarations) == 0 {
		return
	}

	// One symbol lookup per identifier, reused across every declaration below.
	symbols := make(map[*ast.Node]any, len(identifiers))
	for _, identifier := range identifiers {
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier); symbol != nil {
			symbols[identifier] = symbol
		}
	}

	var findings []blockScopedVarFinding
	for _, declarationName := range declarations {
		declarationSymbol, resolved := symbols[declarationName]
		if !resolved {
			// A name the checker cannot resolve has no references we can find, which is the
			// conservative direction: reporting on it would be a guess about what it binds to.
			continue
		}
		container := blockScopedVarContainerOf(declarationName, file)
		if container == nil {
			continue
		}
		for _, identifier := range identifiers {
			if symbols[identifier] != declarationSymbol {
				continue
			}
			// A reference inside the declaration's own block is exactly what `var` should have
			// meant, so only the ones outside it are findings.
			if identifier.Pos() >= container.Pos() && identifier.End() <= container.End() {
				continue
			}
			findings = append(findings, blockScopedVarFinding{
				reference:       identifier,
				declarationName: declarationName,
			})
		}
	}

	blockScopedVarReport(ctx, findings)
}

// blockScopedVarReport emits findings ordered by where the USE is, then by the declaration.
//
// The gathering loop is nested declaration-outer, reference-inner, which groups findings by
// declaration. Upstream's emission order is the other way round: it reports at the reference, and
// the engine orders findings by position, so all the findings AT one use come out together and are
// tie-broken by the declaration each names.
//
// `if (foo) { var a = 1; } else if (bar) { var a = 2; } else { var a = 3; }` is what makes this
// visible: six findings across three uses, and the two orders differ on every one of them while
// producing the identical multiset. No count-based fixture can see it, and the pairing of a span
// with its message is exactly what a reader checks first.
func blockScopedVarReport(ctx rule.Context, findings []blockScopedVarFinding) {
	slices.SortStableFunc(findings, func(left, right blockScopedVarFinding) int {
		if left.reference.Pos() != right.reference.Pos() {
			return left.reference.Pos() - right.reference.Pos()
		}
		return left.declarationName.Pos() - right.declarationName.Pos()
	})
	for _, finding := range findings {
		line, column := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
			rule.TokenRange(ctx.SourceFile, finding.declarationName).Pos())
		ctx.ReportNode(finding.reference, rule.Message{
			Id: "outOfScope",
			Description: fmt.Sprintf(
				"`%s` is declared with `var` inside a block on line %d column %d, and used here, "+
					"outside it. That works because `var` is function-scoped rather than "+
					"block-scoped, so the braces around the declaration say one thing and the "+
					"binding does another. Declare it with `let` in the scope that actually "+
					"needs it, or move the declaration out to where it is used.",
				finding.reference.Text(), line+1, column+1),
		})
	}
}

// blockScopedVarIsVarDeclaration reports whether a declaration was written with `var`.
//
// `let` and `const` are already block-scoped, so they are not this rule's business, and neither is
// a `using` declaration. The flag lives on the declaration LIST rather than on the declaration, so
// this reaches through the parent rather than asking the node itself.
//
// # This guard is SUBSUMED, and it is kept anyway
//
// Forcing it to accept every declaration kind changes no verdict, and it was measured rather than
// argued: the checker already enforces block scoping, so a reference to a `let` outside its block
// resolves to no symbol at all. Probed on `function f(){ { let a = 1; } a; }` -- zero references
// resolve to the declaration, against one for the `var` spelling of the same shape -- so the
// container test below can never see a block-scoped declaration to report.
//
// Kept because it says what the rule is about, and because the subsumption is a property of the
// CHECKER rather than of this rule: a resolution that ever answered more broadly would turn every
// `let` in the tree into a finding, and this line is the only thing that would still refuse. The
// paired fixtures in `TestBlockScopedVarIgnoresBlockScopedDeclarations` pin the behaviour whichever
// half is doing the work.
func blockScopedVarIsVarDeclaration(declaration *ast.Node) bool {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return false
	}
	// A `var` list carries none of the block-scoped flags. Testing for the absence rather than for
	// a `var` flag is what upstream's `node.kind !== "var"` does, and it keeps any future
	// block-scoped spelling out of this rule by default rather than in it.
	return list.Flags&ast.NodeFlagsBlockScoped == 0
}

// blockScopedVarCollectBindingNames appends every identifier a binding name introduces.
//
// A declaration is usually one name, and can be a destructuring pattern introducing several:
// `{ var { foo, bar } = baz; } bar;` is one of upstream's own reporting cases, so a port reading
// only `declaration.Name()` as an identifier goes silent on it.
func blockScopedVarCollectBindingNames(name *ast.Node, into *[]*ast.Node) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		*into = append(*into, name)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		name.ForEachChild(func(element *ast.Node) bool {
			if element.Kind == ast.KindBindingElement {
				blockScopedVarCollectBindingNames(element.Name(), into)
			}
			return false
		})
	}
}

// blockScopedVarContainerOf returns the block-like node a declaration name sits directly inside.
//
// Upstream pushes a range onto a stack for each of these seven kinds and reads the top of the stack
// when it meets a declaration. Walking up from the declaration reaches the same node, and the list
// of kinds is upstream's exactly:
//
//	a block statement, which covers a function body and a bare block alike
//	the three for statements, whose head is inside the container along with the body
//	a switch statement, whose cases share one block
//	a catch clause, whose parameter and body share one
//	a class static block
//
// The file is the outermost container, so a top-level `var` is contained by everything.
//
// A function is deliberately absent from that list even though `var` is function-scoped. It does not
// need to be there: a function body IS a block statement, so the block arm already stops the walk at
// the right place, and adding a function arm would change no answer while suggesting the rule is
// about function scope, which is the thing it is arguing against.
func blockScopedVarContainerOf(node *ast.Node, file *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBlock,
			ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindSwitchStatement,
			ast.KindCatchClause,
			ast.KindClassStaticBlockDeclaration:
			return parent
		case ast.KindSourceFile:
			return parent
		}
	}
	return file
}
