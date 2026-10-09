package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnnecessaryBind = rule.Message{
	Id: "unexpected",
	Description: "This calls `.bind()` on a function that never mentions `this`, so the binding " +
		"changes nothing the function can observe. What it does change is cost and identity: every " +
		"`.bind()` allocates a fresh function object, so the value is no longer equal to the one " +
		"before it, which quietly breaks `removeEventListener` and any memoization keyed on the " +
		"callback. An arrow function is never fixable this way either, since its `this` is already " +
		"decided by where it was written. Drop the call.",
}

// NoExtraBind flags a `.bind()` whose receiver cannot use the binding.
//
//	valid:   var a = function() { return this.b }.bind(c)
//	valid:   var a = function(b) { return b }.bind(c, d)
//	valid:   var a = f.bind(a)
//	valid:   var a = function() { return () => this; }.bind(b)
//	invalid: var a = function() { return 1; }.bind(b)
//	invalid: var a = (() => { return 1; }).bind(b)
//	invalid: var a = function() { return 1; }?.["bind"](b)
//
// Binding a function that never reads `this` is a no-op with two costs. It allocates a new function
// every time, and the result is a different object from the one that went in, so a listener
// registered with `handler.bind(this)` can never be removed with `handler`, and anything keyed on
// callback identity sees a new key. An arrow function is the sharper case: its `this` is fixed at
// the point it is written and `.bind()` cannot move it, so binding one is always pointless whether
// or not it mentions `this`.
//
// # What decides it
//
// The receiver must be a function expression or an arrow function; the property must be `bind`
// spelled statically, so a dot, a string subscript, or a template with no substitutions; and the
// call must pass exactly one argument that is not a spread. Two arguments means the extra ones are
// partial application, which does something regardless of `this`, and a spread means the count is
// not knowable.
//
// Then: a function expression is reported only if no `this` appears in its own scope. `this` inside
// a nested function belongs to that function, so it does not save the outer bind, and `this` inside
// a nested ARROW does, because an arrow takes `this` from where it is written. An arrow receiver
// needs no such scan and is always reported, since binding cannot change its `this` either way.
//
// # Why this gathers over the file rather than tracking a stack
//
// Upstream keeps a stack, pushing on function entry and popping on exit, and reports at the pop,
// because it has to know whether a `this` appeared anywhere in the body before it can judge the
// function. There is no exit listener here and the walk is pre-order, so the same information is
// gathered in one pass over the file: for each candidate receiver, walk its body and look for a
// `this` that belongs to it. That answers the same question the stack answers, and it does it
// without depending on visit order.
//
// # The fixer, and the two things it refuses
//
// The repair removes the `.bind` and the argument list, which is two disjoint spans rather than one,
// because parentheses may sit between them: `(function(){}.bind ) (obj)` has to lose `.bind` and
// ` (obj)` while keeping the parenthesis in the middle. Upstream builds exactly these two spans and
// so does this.
//
// It declines twice, and both declines are decisions rather than gaps. An argument that is not
// obviously side-effect free is left alone, because `f.bind(b())` still calls `b()` and removing the
// call would stop it: upstream's allowed set is literal, identifier, `this`, and function
// expression, and it says in a comment that the set is stricter than it needs to be. And a comment
// anywhere inside the removal span is left alone, because deleting the span would delete the comment
// with it. Twelve of upstream's thirty-two failing cases carry `output: null` for one of these two
// reasons, and reproducing the decline is the point: a fixer that repaired them would delete a
// comment or a side effect unattended.
//
// A trailing comment AFTER the call survives, because it is outside both spans:
// `function(){}.bind(b)/**/` becomes `function(){}/**/`. That falls out of the spans rather than
// being handled, and it is stated because it looks like an oversight.
var NoExtraBind = rule.Rule{
	Name: "no-extra-bind",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				var visit func(*ast.Node)
				visit = func(current *ast.Node) {
					if current == nil {
						return
					}
					switch current.Kind {
					case ast.KindFunctionExpression:
						// A function expression can only be reported when nothing in its own scope
						// reads `this`, which is what upstream's stack is for.
						if !mentionsOwnThis(current) {
							reportIfBoundOnce(ctx, current)
						}
					case ast.KindArrowFunction:
						// No scan. An arrow's `this` comes from where it was written, so `.bind()`
						// cannot change it and the binding is pointless either way. Upstream has a
						// separate listener for exactly this reason, and its corpus pins it with
						// `(() => { return this; }).bind(b)`, which reports.
						reportIfBoundOnce(ctx, current)
					}
					current.ForEachChild(func(child *ast.Node) bool {
						visit(child)
						return false
					})
				}
				visit(node)
			},
		}
	},
}

// mentionsOwnThis reports whether a `this` anywhere inside a function belongs to that function.
//
// The whole function rather than its body, and that is measured rather than tidy. Parameter defaults
// are evaluated with the function's own `this`, so `function(x = this) {}` reads its own `this` and
// upstream is silent on binding it. Upstream gets this without deciding anything, because its
// `ThisExpression` listener fires for every `this` visited between the function's entry and its
// exit, and the parameter list is inside that window. Scanning only the body reports it, which is a
// false positive nothing in the corpus can see: it writes no `this` in a parameter position.
// Measured against the installed rule.
//
// The recursion stops at a nested function expression, a function declaration, and any accessor or
// method, because each of those rebinds `this` and a `this` inside one says nothing about the
// function being judged. It does NOT stop at an arrow function, which is the whole distinction: an
// arrow takes `this` from where it is written, so `function() { return () => this; }` does read its
// own `this` and upstream leaves it alone. Both halves are in the corpus, as a failing case with a
// nested function and a passing case with a nested arrow.
//
// Class declarations are not crossed either. A method body's `this` is the instance, and a class
// written inside a bound function is not evidence about that function.
func mentionsOwnThis(function *ast.Node) bool {
	if function == nil {
		return false
	}
	found := false
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		// The function under test is where the walk starts, so its own kind must not end it.
		if current != function {
			switch current.Kind {
			case ast.KindFunctionExpression,
				ast.KindFunctionDeclaration,
				ast.KindMethodDeclaration,
				ast.KindGetAccessor,
				ast.KindSetAccessor,
				ast.KindConstructor,
				ast.KindClassDeclaration,
				ast.KindClassExpression:
				return
			}
		}
		if current.Kind == ast.KindThisKeyword {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(function)
	return found
}

// boundCall is the shape a reportable `.bind()` has, once it has been recognised.
type boundCall struct {
	// access is the `*.bind` member access, either a property access or an element access.
	access *ast.Node

	// property is the node the finding points at: the `bind` identifier for a dot access, or the
	// string or template literal INCLUDING its quotes for a subscript. Measured against the
	// installed rule, which reports `['bind']` over six columns rather than four.
	property *ast.Node

	// call is the call expression the whole thing sits in, which may be separated from `access` by
	// parentheses.
	call *ast.Node

	// calleeEnd is where the callee of that call ends, which is the access's own end when nothing
	// intervenes and the outermost parenthesis's end when something does. The fixer's second span
	// begins after it, so that a parenthesis written between the member access and the argument
	// list is outside both spans and survives.
	calleeEnd int

	// argument is the single argument, already known not to be a spread.
	argument *ast.Node
}

// reportIfBoundOnce reports a function that is the receiver of a single-argument `.bind()`.
func reportIfBoundOnce(ctx rule.Context, function *ast.Node) {
	bound := boundCallFor(function)
	if bound == nil {
		return
	}
	ctx.ReportRangeWithFixes(rule.TokenRange(ctx.SourceFile, bound.property),
		messageUnnecessaryBind, bindRemovalFixes(ctx, *bound)...)
}

// boundCallFor recognises `<function>.bind(<one argument>)` around a function node.
//
// Optional chaining needs no unwrapping here and that is a real difference from upstream rather than
// an omission. An estree tree wraps an optional chain in a `ChainExpression` node, so upstream has
// to look through it in three places; typescript-go records the optionality as a token on the access
// and the call themselves, so `f?.bind(b)`, `f.bind?.(b)` and `f?.['bind'](b)` are the same two
// nodes as the non-optional forms with a flag set. All six optional shapes are in the corpus and all
// six report.
func boundCallFor(function *ast.Node) *boundCall {
	// The receiver may be wrapped in parentheses, and three of upstream's failing cases write it
	// that way: `(function(){}).bind(this)` and both arrow forms, since an arrow needs parentheses
	// to be a member-access receiver at all. Upstream never sees these, because an estree tree has
	// no node for a parenthesis and its `node.parent` is already the member access.
	receiver := function
	for receiver.Parent != nil && receiver.Parent.Kind == ast.KindParenthesizedExpression {
		receiver = receiver.Parent
	}

	access := receiver.Parent
	if access == nil {
		return nil
	}

	// Both arms below check that the function is the RECEIVER of the access rather than some other
	// child of it, and a mutant dropping either one survives every fixture. Both survivals were
	// resolved by probe rather than by adding a case, and they survive for two different reasons.
	//
	// The property arm is UNREACHABLE. A property access has exactly two children, the receiver and
	// the name, and the name is always an identifier or a private name, so a function expression
	// reaching here can only be the receiver. Probed over 11 shapes and nothing produced the other
	// arrangement.
	//
	// The element arm IS reachable -- `o[function(){}](b)` puts the function in the subscript -- and
	// is SUBSUMED by `isStaticallyBind` four lines below, which asks for a string or a bare template
	// and answers false for a function expression. So no input can distinguish the two versions.
	//
	// Both are kept because they say what the rule means, and because the subsumption is one edit
	// away from being false: widening `isStaticallyBind` to accept anything whose text is `bind`
	// would make the element arm load-bearing again with nothing to notice. The verdicts were taken
	// over `boundCallFor`'s single caller.
	var property *ast.Node
	switch access.Kind {
	case ast.KindPropertyAccessExpression:
		propertyAccess := access.AsPropertyAccessExpression()
		if propertyAccess.Expression != receiver {
			return nil
		}
		name := propertyAccess.Name()
		// A private name cannot be `bind` and reads as one through `Text()`, which strips the hash.
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "bind" {
			return nil
		}
		property = name
	case ast.KindElementAccessExpression:
		elementAccess := access.AsElementAccessExpression()
		if elementAccess.Expression != receiver {
			return nil
		}
		argument := elementAccess.ArgumentExpression
		if !isStaticallyBind(argument) {
			return nil
		}
		property = argument
	default:
		return nil
	}

	// Parentheses may sit between the access and the call: `(function(){}.bind)(this)` is in the
	// corpus, as is its optional twin. Upstream reaches the same shape by skipping closing-paren
	// tokens; here the parentheses are nodes, so the climb is explicit.
	current := access
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}

	call := current.Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return nil
	}
	callExpression := call.AsCallExpression()
	if callExpression.Expression != current {
		return nil
	}
	if callExpression.Arguments == nil || len(callExpression.Arguments.Nodes) != 1 {
		return nil
	}
	argument := callExpression.Arguments.Nodes[0]
	// A spread makes the count unknowable, so `f.bind(...c)` is clean. In the corpus as a passing
	// case, and it parses here as one argument node of a different kind rather than as none.
	if argument.Kind == ast.KindSpreadElement {
		return nil
	}

	return &boundCall{
		access:    access,
		property:  property,
		call:      call,
		calleeEnd: current.End(),
		argument:  argument,
	}
}

// isStaticallyBind reports whether a subscript names `bind` without evaluating anything.
//
// `f['bind']` and `f[`bind`]` are the same access as `f.bind` and report; `f[bind]` and
// `f[`bi${n}d`]` read a variable and are clean, both in the corpus. A template with any substitution
// parses as a different node kind here, which is what separates the two template cases.
func isStaticallyBind(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.Text() == "bind"
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.Text() == "bind"
	}
	return false
}

// bindRemovalFixes builds the repair, or nothing where upstream declines to repair.
//
// Two spans rather than one, because parentheses may sit between the member access and the argument
// list and must survive: `(function(){}.bind ) (obj)` loses `.bind` and ` (obj)` and keeps the
// parenthesis between them. The first span runs from the end of the receiver to the end of the
// property; the second from the end of the member access to the end of the call.
//
// Both spans start at a node END rather than at a token start, which is deliberate: the `.`, `?.`,
// `[` and `(` tokens are not nodes here, and anchoring on the neighbouring node's end sweeps them up
// without having to find them. It also means the spans are contiguous with the text that stays, so
// the rewrite cannot leave a stray separator behind.
func bindRemovalFixes(ctx rule.Context, bound boundCall) []rule.Fix {
	// An argument that might do something is left alone, because removing the call would stop it
	// happening. Upstream's set is literal, identifier, `this`, function expression, and its comment
	// says the set is stricter than it needs to be, so an arrow argument is declined here too even
	// though evaluating one is inert.
	if !isSideEffectFreeArgument(bound.argument) {
		return nil
	}

	sourceText := ctx.SourceFile.Text()

	// Both spans begin at a TOKEN start rather than at the preceding node's end, and that is the
	// whole reason this reaches for the scanner. `function(){}/**/.bind(b)` has a comment between
	// the receiver and the dot, and upstream repairs it to `function(){}/**/` -- the comment is
	// BEFORE the first token it removes, so it survives and does not trip the comment check either.
	// Anchoring the span on the receiver's end instead swallows that comment, which turns a case
	// upstream repairs into one this declines, and turns three more into silent comment deletion.
	// `scanner.SkipTrivia` advances past whitespace and comments to the next real token, which is
	// upstream's `getTokenAfter(..., isNotClosingParenToken)` for the shapes that can occur here.
	//
	// The first span ends at the ACCESS end rather than at the property end, and the difference is
	// the closing bracket of a computed access. Upstream takes "the property name or the `]` token",
	// which is one token in the dotted case and a different one in the subscript case; here the
	// access node's own end is both. Ending at the property leaves a stray `]` behind, and the
	// rewritten file still parses in some positions, which is exactly the damage the fix engine's
	// parse guard cannot catch.
	firstRemoval := core.NewTextRange(
		scanner.SkipTrivia(sourceText, receiverOf(bound.access).End()),
		bound.access.End())

	// The second span starts after any parentheses between the access and the argument list, which
	// is what keeps `(function(){}.bind ) (obj)` from losing the parenthesis that has to survive.
	// `bound.calleeEnd` is where the parenthesis climb finished, so the skip from there lands on the
	// `(` or `?.` that opens the arguments.
	secondRemoval := core.NewTextRange(
		scanner.SkipTrivia(sourceText, bound.calleeEnd),
		bound.call.End())

	// A comment anywhere between the start of the first span and the end of the second is enough to
	// decline, INCLUDING one in the gap the two spans leave alone. Upstream asks the same question
	// over the same outer bounds rather than per span, and `function(){}.bind/**/(b)` is the case
	// that separates them: the comment sits between the spans and upstream declines.
	if commentsExistBetween(ctx, firstRemoval.Pos(), secondRemoval.End()) {
		return nil
	}

	return []rule.Fix{rule.RemoveRange(firstRemoval), rule.RemoveRange(secondRemoval)}
}

// receiverOf reads the object side of a member access.
func receiverOf(access *ast.Node) *ast.Node {
	if access.Kind == ast.KindPropertyAccessExpression {
		return access.AsPropertyAccessExpression().Expression
	}
	return access.AsElementAccessExpression().Expression
}

// isSideEffectFreeArgument reproduces upstream's deliberately narrow allowed set.
//
// Literal, identifier, `this`, and function expression. Upstream's own comment calls this stricter
// than it needs to be and keeps it for simplicity, so widening it would be an improvement rather
// than a port: `f.bind(b.c)` reads a property, which can run a getter, and is declined; `f.bind(b++)`
// and `f.bind(b())` plainly are not free. All three are in the corpus with `output: null`.
//
// The literal arm is the one that needs measuring rather than reading, because "literal" names a
// different set of shapes in each tree. An estree `Literal` is a string, a number, a regular
// expression, `null`, a boolean, or a bigint -- and NOT a template, which is its own node kind, nor
// a negative number, which is a unary expression over a literal. Both of those are inside the
// intuitive reading of "literal" and outside upstream's, and both are reported-but-not-fixed there.
// A no-substitution template was in this list first, on that intuitive reading; measured against the
// installed rule, a bare-template argument reports and is NOT repaired. The corpus writes no literal
// argument at all, so nothing imported could have separated the two readings.
func isSideEffectFreeArgument(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindThisKeyword,
		ast.KindFunctionExpression,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	}
	return false
}

// commentsExistBetween reports whether any comment starts inside a span.
//
// Cached per file through the shelf's comment reader rather than rescanning, since this runs once
// per finding and a file can hold many.
func commentsExistBetween(ctx rule.Context, from int, to int) bool {
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= from && comment.Range.Pos() < to {
			return true
		}
	}
	return false
}
