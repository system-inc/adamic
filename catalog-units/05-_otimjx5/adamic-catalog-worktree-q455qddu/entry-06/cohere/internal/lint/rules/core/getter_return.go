package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/descriptor"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// GetterReturnOptions configures whether a bare `return;` satisfies the rule.
//
// This is the whole option surface. ESLint's `meta.schema` is one object with a single
// `allowImplicit` boolean and `additionalProperties: false`, so there is nothing else to read and
// a case configuring anything else configures nothing.
type GetterReturnOptions struct {
	// AllowImplicit lets a getter satisfy the rule with a `return` carrying no expression, which
	// yields `undefined`. Off by default, because a getter returning nothing is usually the bug
	// this rule exists to find rather than a deliberate choice.
	AllowImplicit bool `json:"allowImplicit"`
}

// GetterReturn requires a getter to return a value on every path out of it.
//
// A getter looks like a property read at the call site, so a path through it that returns nothing
// silently yields `undefined` where the reader expected a value. That is the defect: not a missing
// statement, but a property that lies about being a property.
//
// Three shapes count as a getter here, matching upstream: a `get` accessor in a class, a `get`
// accessor in an object literal, and a function sitting under the `get` key of a property
// descriptor passed to `Object.defineProperty`, `Object.defineProperties`, `Object.create`, or
// `Reflect.defineProperty`.
//
// Examples of incorrect code:
//
//	var foo = { get bar() {} };
//	class foo { get bar() { if (baz) { return true; } } }
//	Object.defineProperty(foo, "bar", { get: function () {} });
//
// Examples of correct code:
//
//	var foo = { get bar() { return true; } };
//	class foo { get bar() { if (baz) { return true; } else { return false; } } }
//	var foo = { get bar() { throw new Error("nope"); } };
//
// # Why there is no control flow graph here, and why there does not need to be
//
// Upstream reaches for oxc's CFG (`ctx.cfg()`) and walks it depth-first looking for a basic block
// that neither returns nor throws and has no outgoing edge. We have no CFG and are not building
// one, so the question is whether TypeScript's flow nodes are a substitute. Measured rather than
// assumed: they are not, and the reason is that they answer a different question.
//
// `Node.FlowNodeData()` is real and populated, but TypeScript's flow graph exists for *narrowing*
// and is keyed on the expression positions where a type could change. Probed on three getter bodies
// drawn from upstream's own corpus, counting statements that carry a flow node at all:
//
//	if (baz) { return true; }              4 statements, 2 with a flow node
//	try { return a(); } catch {}           4 statements, 1 with a flow node
//	for (let i=0;i<10;i++) { return i; }   3 statements, 1 with a flow node
//
// The middle line is the one that decides it. `try { return a(); } catch {}` is upstream's sharpest
// fail case, separated from the passing `try { return a(); } finally {}` by exactly the question
// this rule asks, and it carries one flow node in the whole body. There is no return instruction
// and no throw instruction to find, because a narrowing graph does not record them. Walking it
// would not be an approximation of the CFG answer; it would be reading an unrelated structure.
//
// So the analysis here is structural, over the statement tree, and it reproduces upstream's verdict
// on every discrimination its corpus draws. That is worth stating precisely, because "structural"
// invites the assumption that it is the weaker instrument. For this question it is not. "Does every
// path out of this body return" is a property of how statements nest, and the four constructs that
// make it interesting are all visible in the tree:
//
//	if        exits only when it has an else and both arms exit
//	try       exits when the finally exits, or when the try exits and the catch exits
//	switch    exits when it has a default and every non-empty clause exits
//	loops     never count, because the body may run zero times
//
// The loop rule is the one that looks like a limitation and is not. Upstream passes a loop whose
// body returns *followed by* a trailing return, and fails the identical loop without the trailing
// return, which is the same judgment reached here for the same reason: no static analysis can know
// the loop runs. A CFG gets this right by having a bypass edge around the loop, and the tree gets
// it right by declining to credit the body.
//
// The checker is asked one thing only: whether the `Object` or `Reflect` of a descriptor call is the
// global, so `let Object; Object.defineProperty(...)` is not a descriptor. Reachability never asks it.
//
// # TypeScript files are checked too
//
// The port this rule followed skipped TypeScript outright, as oxc does, because the compiler reports
// a `get` accessor with no return (TS2378), and typescript-eslint's eslint-recommended turns the rule
// off for .ts files for the same reason. The house does not: Kirk ruled on 2026-08-25 (#mnmx9s4)
// that the core correctness rules stay on in TypeScript in both engines, so that the engines agree on
// what runs, and Nexus turns getter-return on there. Skipping .ts files left cohere silent where
// ESLint reports, which every consumer's zero hid, and all 35 of ESLint's corpus rows read as missing
// (#jjfa7qb). It also missed what the compiler never checks: a descriptor's `get` is a plain function
// to TypeScript, so `Object.defineProperty(o, 'k', { get() {} })` was caught nowhere.
//
// # What this deliberately does not catch, matching upstream
//
// A getter whose only exit is inside a nested function does not count, because the nested function
// returns from itself. Upstream fails `get bar() { ~function () { return true; }() }` and so does
// this. Conversely `return` inside a nested function must not credit the outer getter, which is why
// the walk stops at every function boundary.
var GetterReturn = rule.Rule{
	Name: "getter-return",

	// Whether the `Object` or `Reflect` of a descriptor call is the global
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowImplicit := false
		if parsed, isParsed := rule.OptionsAs[GetterReturnOptions](options); isParsed {
			allowImplicit = parsed.AllowImplicit
		}

		isGlobal := func(identifier *ast.Node) bool {
			return ctx.TypeChecker != nil && rule.IsDeclaredOnlyInDeclarationFiles(ctx.TypeChecker.GetSymbolAtLocation(identifier))
		}

		check := func(node *ast.Node, body *ast.Node) {
			if body == nil || body.Kind != ast.KindBlock {
				return
			}
			if !isGetterFunction(node, isGlobal) {
				return
			}
			if bodyDefinitelyExits(body) {
				return
			}
			// ESLint's two ids: `expected` for a getter that never returns, `expectedAlways` for one
			// that returns on some paths and falls off the end on another
			if returnsAnywhere(body) {
				ctx.ReportRange(getterHeadRange(ctx.SourceFile, node), rule.Message{
					Id: "expectedAlways",
					Description: "This getter returns a value on some paths and falls off the end on " +
						"another, where reading the property yields undefined. Return a value from every path.",
				})
				return
			}
			ctx.ReportRange(getterHeadRange(ctx.SourceFile, node), rule.Message{
				Id: "expected",
				Description: "This getter never returns a value, so reading the property yields " +
					"undefined. Return the value it stands for.",
			})
		}

		return rule.Listeners{
			// A bare `return;` in a getter yields undefined, and ESLint reports the statement
			ast.KindReturnStatement: func(node *ast.Node) {
				if allowImplicit || node.AsReturnStatement().Expression != nil {
					return
				}
				function := ast.FindAncestor(node.Parent, ast.IsFunctionLike)
				if function == nil || function.Body() == nil || function.Body().Kind != ast.KindBlock || !isGetterFunction(function, isGlobal) {
					return
				}
				ctx.ReportNode(node, rule.Message{
					Id: "expected",
					Description: "This bare return leaves the getter with undefined, so reading the " +
						"property yields nothing. Return the value it stands for.",
				})
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				check(node, bodyOf(node))
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				check(node, bodyOf(node))
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				check(node, bodyOf(node))
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				check(node, bodyOf(node))
			},
		}
	},
}

// bodyOf returns the body a function-like node executes, or nil when it has none.
//
// An arrow function with an expression body returns that expression unconditionally, so it can
// never violate this rule. Upstream reaches the same conclusion in its `arrow_expr.is_expression()`
// early break. That filtering happens at the call site rather than here; see the note below.
func bodyOf(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindGetAccessor:
		if body := node.AsGetAccessorDeclaration().Body; body != nil {
			return body.AsNode()
		}
	case ast.KindFunctionExpression:
		if body := node.AsFunctionExpression().Body; body != nil {
			return body.AsNode()
		}
	case ast.KindArrowFunction:
		// No block-kind filter here, deliberately. An arrow with an expression body returns that
		// expression unconditionally and must not be reported, but the `check` closure already
		// declines any body that is not a block, so filtering in both places is one check doing
		// the work of two. A mutation sweep scored each gate as a survivor on its own and caught
		// them only when both were removed together, which is the signature of a redundant pair.
		// The gate is kept at `check` rather than here because that one also guards the three
		// non-arrow shapes.
		if body := node.AsArrowFunction().Body; body != nil {
			return body.AsNode()
		}
	case ast.KindMethodDeclaration:
		if body := node.AsMethodDeclaration().Body; body != nil {
			return body.AsNode()
		}
	}
	return nil
}

// getterHeadRange is where the finding points: the getter's head, not its whole body.
//
// ESLint's `getFunctionHeadLoc`, which keeps the caret on the declaration a reader has to change and
// keeps a multi-hundred-line getter from underlining itself. It runs from the property to the opening
// paren of the parameters, so `get bar() {}` reports `get bar` and `{ get: function () {} }` reports
// `get: function `. Where the function is a property's value, the property is where it starts. An arrow
// whose one parameter has no parentheses ends at that parameter. The old span ran to the body and
// read as a different finding on all 35 of ESLint's rows (#jjfa7qb).
func getterHeadRange(file *ast.SourceFile, node *ast.Node) core.TextRange {
	// TokenRange is what ReportNode uses, and it is what skips the leading trivia. Building the
	// range from node.Pos() directly is the defect that helper exists to prevent: a getter preceded
	// by a comment would report at the comment, and a `-next-line` suppression written above it
	// could never match.
	owner := node
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAssignment {
		owner = node.Parent
	}
	start := rule.TokenRange(file, owner).Pos()
	text := file.Text()
	end := node.ParameterList().Pos()
	if end > 0 && text[end-1] == '(' {
		end--
	} else {
		end = scanner.SkipTrivia(text, end)
	}
	return core.NewTextRange(start, end)
}

// returnsAnywhere answers whether a getter body holds a return of its own, outside any nested function
func returnsAnywhere(body *ast.Node) bool {
	found := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if found || ast.IsFunctionLike(node) {
			return found
		}
		if node.Kind == ast.KindReturnStatement {
			found = true
			return true
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return found
}

// isGetterFunction answers whether this function-like node is in getter position: a `get` accessor in
// a class or object literal, or the `get` of a property descriptor, which the `descriptor` shelf
// recognizes and `no-setter-return` shares for `set`.
//
// isGlobal answers whether the descriptor call's `Object` or `Reflect` is the global. ESLint's
// getter-return matches the spelling alone, and the global is asked here as no-setter-return does,
// since a shadowing local's `defineProperty` is not the platform method.
func isGetterFunction(node *ast.Node, isGlobal func(identifier *ast.Node) bool) bool {
	if node.Kind == ast.KindGetAccessor {
		// A `get` accessor is a getter in both a class and an object literal, and there is no
		// other thing it can be.
		return true
	}
	return descriptor.IsFunctionUnder(node, "get", isGlobal)
}

// bodyDefinitelyExits answers whether every path through a statement leaves the enclosing function.
//
// "Exits" rather than "returns" because a throw counts: upstream passes
// `get willThrowSoValid() { throw MyException() }`, and it is right to, since that path never
// yields a value at all rather than yielding the wrong one.
//
// Everything not named below answers false, which is the safe direction: an unrecognized statement
// is assumed not to exit, so the rule reports rather than staying silent. That is the correct bias
// for a correctness rule but it is also the one that produces false positives, which is why the
// recognized set covers every construct upstream's corpus exercises.
func bodyDefinitelyExits(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindReturnStatement:
		// A bare `return;` ends the path too. It yields undefined, which is reported at the
		// statement itself unless allowImplicit says it was meant, as ESLint reports it, so the
		// head is reported only for a path that falls off the end.
		return true

	case ast.KindThrowStatement:
		return true

	case ast.KindBlock:
		// A block exits when any statement in it does. Statements after an exiting one are
		// unreachable, so scanning past the first is harmless and stopping early is not required
		// for correctness.
		for _, statement := range node.AsBlock().Statements.Nodes {
			if bodyDefinitelyExits(statement) {
				return true
			}
		}
		return false

	case ast.KindIfStatement:
		// Both arms have to exit, and an absent else is one that does not. Without an else there is
		// a path that falls through, which is upstream's `get bar(){if(baz) {return true;}}` fail
		// case and the most common real defect this rule catches.
		//
		// There is deliberately no `ElseStatement != nil` conjunct here. It read as the load-bearing
		// part of this branch and it was subsumed: an absent else is a nil node, and the recursion
		// rejects nil on its first line, so the explicit check and the recursion agreed on every
		// input. A mutation sweep scored it as a survivor and no fixture could have caught it,
		// because none can exist. Deleted rather than tested, so the next reader does not spend the
		// same hour proving the same thing.
		statement := node.AsIfStatement()
		return bodyDefinitelyExits(statement.ThenStatement) &&
			bodyDefinitelyExits(statement.ElseStatement)

	case ast.KindTryStatement:
		return tryStatementExits(node.AsTryStatement())

	case ast.KindSwitchStatement:
		return switchStatementExits(node.AsSwitchStatement())

	case ast.KindLabeledStatement:
		// A label wraps a statement without changing whether it exits. `break label` would, but a
		// break out of a getter body reaches the end of the body, which is already not an exit.
		return bodyDefinitelyExits(node.AsLabeledStatement().Statement)

	case ast.KindWithStatement:
		return bodyDefinitelyExits(node.AsWithStatement().Statement)
	}

	// Loops land here and answer false on purpose. A `for`, `while`, or `for-in` body may run zero
	// times, so a return inside one is not a return on every path. Upstream agrees: it fails a
	// getter whose only return is inside a loop, and passes the identical loop followed by a
	// trailing return. `do-while` runs at least once and could in principle be credited, but
	// upstream does not credit it and this reproduces that rather than silently improving on it.
	return false
}

// tryStatementExits answers whether a try statement exits on every path.
//
// This is the subtlest of the four and the pair upstream uses to pin it is
// `try { return a(); } finally {}` passing against `try { return a(); } catch {}` failing. The
// difference is that a catch introduces a path: the try may throw partway through, and then only
// the catch runs. A finally introduces no path of its own.
func tryStatementExits(statement *ast.TryStatement) bool {
	// A finally that exits wins outright, since it runs on every path out of the try and the
	// catch both, and its own exit overrides theirs.
	if statement.FinallyBlock != nil &&
		bodyDefinitelyExits(statement.FinallyBlock.AsNode()) {
		return true
	}
	if !bodyDefinitelyExits(statement.TryBlock.AsNode()) {
		return false
	}
	if statement.CatchClause == nil {
		return true
	}
	return bodyDefinitelyExits(statement.CatchClause.AsCatchClause().Block.AsNode())
}

// switchStatementExits answers whether a switch exits on every path.
//
// Two conditions, and both are load-bearing. Without a default clause the scrutinee may match
// nothing and fall out of the switch entirely. And every clause has to exit, since one that does
// not is a path through.
//
// An empty clause is not a failure: `case A: case B: return x;` is a deliberate fallthrough and
// the empty `case A` exits by way of `case B`. Only a trailing empty clause genuinely falls out.
func switchStatementExits(statement *ast.SwitchStatement) bool {
	clauses := statement.CaseBlock.AsCaseBlock().Clauses.Nodes

	hasDefault := false
	for _, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		return false
	}

	for index, clause := range clauses {
		statements := clause.AsCaseOrDefaultClause().Statements.Nodes
		if len(statements) == 0 {
			// Falls through to the next clause, unless there is no next clause.
			if index == len(clauses)-1 {
				return false
			}
			continue
		}
		exits := false
		for _, inner := range statements {
			if bodyDefinitelyExits(inner) {
				exits = true
				break
			}
		}
		if !exits {
			return false
		}
	}
	return true
}
