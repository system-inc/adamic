package core

import (
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnmodifiedLoopCondition = rule.Message{
	Id: "loopConditionNotModified",
	Description: "This variable is read by the loop condition and never changed inside the loop, " +
		"so the condition evaluates to the same thing every time and the loop either runs forever " +
		"or never runs at all. Whatever was meant to end the loop is not happening: either the " +
		"variable should be updated in the body, or the condition should be reading something that " +
		"does change.",
}

// NoUnmodifiedLoopCondition flags a variable a loop tests and never modifies.
//
//	valid:   var foo = 0; while (foo) { ++foo; }
//	valid:   var foo = 0; while (foo++) { }              the condition itself modifies it
//	valid:   var foo = 0; while (ok(foo)) { }            a call may do anything
//	valid:   var foo = 0; while (foo.ok) { }             a member access may be a getter
//	valid:   var foo = 0; while (foo) { update(); } function update() { ++foo; }
//	valid:   var foo = 0, bar = 9; while (foo < bar) { foo += 1; }   one of the group changes
//	invalid: var foo = 0; while (foo) { } foo = 1;
//	invalid: var foo = 0, bar = 9; while (foo < bar) { } foo = 1;    both names
//	invalid: for (var foo = 0; foo < 10; ) { } foo = 1;
//
// A loop whose condition can never change its answer is a loop that runs forever or not at all. The
// variable being written somewhere else in the file is what makes this reportable rather than
// obvious: the author clearly expected it to change, and the write is outside the loop.
//
// # Groups, and why `foo < bar` is one judgment rather than two
//
// A reference inside a BinaryExpression belongs to that expression as a group, and a group is only
// reported when EVERY member of it is unmodified. `while (foo < bar) { foo += 1; }` is clean in both
// names, because changing either one can change the comparison. When neither changes, both names are
// reported, which is why that case carries two findings rather than one.
//
// A ConditionalExpression is a group too, so `while (foo ? bar : baz) { foo += 1; }` is clean.
//
// # Dynamic expressions abandon the group entirely
//
// If the group holds a call, a member access, a `new`, a tagged template, or a `yield`, the whole
// search stops and nothing is reported, because any of those can have an effect the rule cannot
// see. That is why `while (foo === f(bar)) { }` and `while (foo === obj.bar) { }` are clean: a
// getter or a function call may be changing the world. The check does not descend into a nested
// function, since a function that is merely written is not called.
//
// # A modification counts when it can run in the loop, and that includes a called function
//
// A write inside the loop marks the condition modified. So does a write inside a named function
// DECLARATION, when that function is itself referenced from inside the loop. This is the pair the
// corpus states twice, once each way:
//
//	while (foo) { update(); } function update() { ++foo; }         clean, update is called here
//	while (foo) { update(); } function update(foo) { ++foo; }      reported, that is a parameter
//
// The second differs only in that `foo` is shadowed by a parameter, so the write is to a different
// binding entirely and never reaches the loop's variable.
//
// # The initializer is not a modification, except for `var`
//
// `isWriteReference` declines a reference marked as an initializer unless the binding is a `var`.
// `for (var foo = 0; foo < 10; ) { } foo = 1;` reports on that basis: the `foo = 0` in the head is
// an initializer on a `var`, so it counts as a write, yet it sits in the loop's own initializer,
// which is excluded from "in the loop" for a ForStatement. The write that would save it is outside.
//
// # Where the loop condition is, and the initializer that is not part of it
//
// Only `while`, `do`/`while` and `for` have a test. A `for...in` or `for...of` has no condition at
// all and is never judged. For a `for`, a reference inside the INITIALIZER is not "in the loop",
// which is what makes `var foo = 0, bar = 0; for (bar; foo;) { ++foo }` clean while the same shape
// with the write in the initializer is not.
//
// # This ports the INSTALLED build, and the clone has an option it does not
//
// The clone at 10.9.1 adds `checkConditionalExpressions`, which turns a ConditionalExpression from
// a group into three separate judgments. The installed eslint 10.8.1 build this repository runs has
// `schema: []` and no such option: configuring it throws at config validation with "should NOT have
// more than 0 items", measured by driving it through the linter interface.
//
// So this rule takes no options, matching what our gate compares against, and the three corpus cases
// carrying that option are recorded in the fixtures as unmeasurable against the oracle rather than
// asserted either way. When the installed build moves to 10.9, the option is the follow-up.
//
// # No fix
//
// The repair is either updating the variable in the body or rewriting the condition, and which one
// is wanted is the whole question the finding asks. Upstream ships no fixer either.
var NoUnmodifiedLoopCondition = rule.Rule{
	Name: "no-unmodified-loop-condition",

	// The discrimination is name resolution. `while (foo) { update(); } function update(foo) {
	// ++foo; }` and the same source without the parameter differ only in what `foo` binds to, and
	// upstream answers them differently.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The judgment is the whole file's, made once, at the first loop that has a condition. A file
		// with none cannot report, and most files have none, so they never build the reference index.
		judged := false
		judge := func(node *ast.Node) {
			if judged || ctx.TypeChecker == nil || unmodifiedLoopTestOf(node) == nil {
				return
			}
			judged = true
			reportUnmodifiedLoopConditions(ctx)
		}
		return rule.Listeners{
			ast.KindWhileStatement: judge,
			ast.KindDoStatement:    judge,
			ast.KindForStatement:   judge,
		}
	},
}

// unmodifiedLoopCondition is one reference read by a loop condition.
type unmodifiedLoopCondition struct {
	// identifier is the reference itself, and where the finding points.
	identifier *ast.Node

	// loop is the loop statement whose test holds this reference.
	loop *ast.Node

	// group is the BinaryExpression or ConditionalExpression this reference belongs to, or nil.
	// A group is reported only when every member of it is unmodified.
	group *ast.Node

	// modified records that something able to run in the loop writes this binding.
	modified bool
}

// reportUnmodifiedLoopConditions judges every variable in the file.
func reportUnmodifiedLoopConditions(ctx rule.Context) {
	// Same index the loop-function rule builds: every identifier grouped by the declaration it
	// resolves to. Upstream reads the equivalent off eslint-scope, walking every scope and every
	// variable in it; iterating the index's keys reaches the same set of bindings, since a binding
	// with no reference anywhere has nothing in a loop condition either.
	referencesByDeclaration := loopFuncReferenceIndex(ctx)

	// Declarations in a stable order, because map iteration is randomised and findings are sorted
	// by position at the end regardless. Sorting here as well keeps a group's members in source
	// order when they are reported together.
	declarations := make([]*ast.Node, 0, len(referencesByDeclaration))
	for declaration := range referencesByDeclaration {
		declarations = append(declarations, declaration)
	}
	sort.Slice(declarations, func(first, second int) bool {
		return declarations[first].Pos() < declarations[second].Pos()
	})

	reported := []*ast.Node{}
	groups := map[*ast.Node][]*unmodifiedLoopCondition{}
	groupOrder := []*ast.Node{}

	for _, declaration := range declarations {
		references := referencesByDeclaration[declaration]

		conditions := []*unmodifiedLoopCondition{}
		for _, identifier := range references {
			if condition := unmodifiedLoopConditionFor(ctx, identifier); condition != nil {
				conditions = append(conditions, condition)
			}
		}
		if len(conditions) == 0 {
			continue
		}

		for _, condition := range conditions {
			if condition.group == nil {
				continue
			}
			if _, seen := groups[condition.group]; !seen {
				groupOrder = append(groupOrder, condition.group)
			}
			groups[condition.group] = append(groups[condition.group], condition)
		}

		// Every write to this binding that could run inside the loop.
		modifiers := []*ast.Node{}
		for _, identifier := range references {
			if unmodifiedLoopIsWriteReference(identifier, declaration) {
				modifiers = append(modifiers, identifier)
			}
		}
		for _, condition := range conditions {
			for _, modifier := range modifiers {
				if condition.modified {
					break
				}
				condition.modified = unmodifiedLoopHasModifierInLoop(ctx, condition, modifier, referencesByDeclaration)
			}
		}

		// A condition in a group waits for every member of that group; the rest report now.
		for _, condition := range conditions {
			if !condition.modified && condition.group == nil {
				reported = append(reported, condition.identifier)
			}
		}
	}

	// A group reports only when nothing in it is modified, and then reports every member.
	for _, group := range groupOrder {
		members := groups[group]
		allUnmodified := true
		for _, member := range members {
			if member.modified {
				allUnmodified = false
				break
			}
		}
		if !allUnmodified {
			continue
		}
		for _, member := range members {
			reported = append(reported, member.identifier)
		}
	}

	// Source order. Upstream reports groups after every variable is done, so its own emission order
	// is neither traversal nor position, and ESLint sorts by position before handing over.
	sort.Slice(reported, func(first, second int) bool {
		return reported[first].Pos() < reported[second].Pos()
	})

	for _, identifier := range reported {
		ctx.ReportNode(identifier, rule.Message{
			Id: messageUnmodifiedLoopCondition.Id,
			Description: fmt.Sprintf("%s The variable is '%s'.",
				messageUnmodifiedLoopCondition.Description, identifier.Text()),
		})
	}
}

// unmodifiedLoopConditionFor builds the condition record for a reference, or nil when the reference
// is not read by a loop condition.
//
// Upstream's `toLoopCondition`. The climb stops at the first "sentinel" node, which is its way of
// saying "we have left the expression the condition is made of".
func unmodifiedLoopConditionFor(ctx rule.Context, identifier *ast.Node) *unmodifiedLoopCondition {
	// An initializer reference is never a condition read. `var foo = 0` is not the loop testing
	// `foo`, and upstream returns null for it before anything else.
	//
	// SUBSUMED by the sentinel walk below, and kept as upstream's own first check. Deleting it
	// leaves every fixture green: an initializer name's climb reaches a VariableStatement, which is
	// a sentinel and is not a loop, or in a `for` head reaches the ForStatement whose Condition is
	// a different slot than the Initializer the name sits in, so `test == child` fails either way.
	//
	// Probed over seven shapes including `for (var foo = 0; foo < 10; )`,
	// `var foo = 0, bar = foo; while (bar) { }` and `for (var a = 0, b = a; b; )`, with the guard
	// removed: initializer references landing on a loop test came back 0 while the control of
	// ordinary references landing on one came back 7.
	if unmodifiedLoopIsInitializerReference(identifier) {
		return nil
	}

	var group *ast.Node
	child := identifier
	for node := identifier.Parent; node != nil; node = node.Parent {
		if unmodifiedLoopIsSentinel(node) {
			if test := unmodifiedLoopTestOf(node); test != nil && test == child {
				return &unmodifiedLoopCondition{identifier: identifier, loop: node, group: group}
			}
			// A sentinel that is not a loop's test means the reference is somewhere else.
			return nil
		}

		if unmodifiedLoopIsGroupNode(node) {
			// A group holding anything dynamic is abandoned rather than recorded, because a call
			// or a getter in it may be changing exactly what the rule is about to call unchanging.
			if unmodifiedLoopHasDynamicExpressions(node) {
				return nil
			}
			group = node
		}

		child = node
	}
	return nil
}

// unmodifiedLoopIsGroupNode reports whether a node forms a group: an expression whose members are
// judged together, so that a change to any one of them can change the whole answer.
//
// # `&&` is NOT a group, and our AST hides that behind one kind
//
// Upstream tests `node.type === "BinaryExpression"`, and in the tree it walks, `&&`, `||` and `??`
// are a LogicalExpression rather than a BinaryExpression. Ours calls all of them
// KindBinaryExpression, so a direct translation of that test groups them and is wrong.
//
// Measured against the installed eslint 10.8.1 build, which is what settled it:
//
//	while (foo && bar) { ++bar; }        reports 'foo' ALONE, so no group
//	while (foo < bar) { }                reports BOTH, so a group
//	while (a < c && b < c) { ++a; }      reports 'b' and 'c', the two `<` groups, not the `&&`
//
// The third is the one that pins it both ways at once: the logical operator forms no group while
// the comparisons under it do, so the answer cannot be "never group" either. Three fixtures failed
// on this and none of them would have been written from reading the rule source, because the source
// reads correct.
//
// A comma is a SequenceExpression upstream, which SENTINEL_PATTERN does not match and this test
// does not either; ours spells it KindBinaryExpression too, so it is excluded here for the same
// reason as the logical operators.
func unmodifiedLoopIsGroupNode(node *ast.Node) bool {
	if node.Kind == ast.KindConditionalExpression {
		return true
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	switch node.AsBinaryExpression().OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindCommaToken:
		return false
	}
	return true
}

// unmodifiedLoopIsSentinel reports whether a node ends the climb out of a loop condition.
//
// Upstream's SENTINEL_PATTERN matches Call/Class/Function/Member/New/Yield expressions plus anything
// whose type name ends in Statement or Declaration. The kinds are enumerated here rather than
// pattern matched on a name, because our kind names do not spell the same words: a
// PropertyAccessExpression is upstream's MemberExpression, and an ElementAccessExpression is the
// computed form of the same thing.
//
// # Not `ast.IsDeclarationNode`, and the name is the trap
//
// The first version of this called `ast.IsStatement(node) || ast.IsDeclarationNode(node)` for the
// "ends in Statement or Declaration" half. `IsDeclarationNode` is `DeclarationData() != nil`, which
// is a question about how the node stores itself rather than about what it is, and measured on
// `while (foo < bar)` it answers TRUE for the BinaryExpression. So the climb stopped on the group
// node itself, no reference ever reached its loop, and the rule went silent on every grouped case:
// eight fixtures, which is what caught it.
//
// The shelf helper is accurate about what it computes and wrong for this question. Enumerated
// instead, so what ends the climb is written down where it can be read.
func unmodifiedLoopIsSentinel(node *ast.Node) bool {
	switch node.Kind {
	// SENTINEL_PATTERN's named expressions.
	case ast.KindCallExpression, ast.KindNewExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindYieldExpression,
		ast.KindClassExpression, ast.KindClassDeclaration,
		ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction:
		return true

	// The loops themselves, which is where the climb is trying to arrive.
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		return true

		// Everything else ending in Statement or Declaration. `ast.IsStatement` covers the statements
		// without the storage-shaped answer that `IsDeclarationNode` gives.
	}
	return ast.IsStatement(node) || ast.IsDeclarationStatement(node)
}

// unmodifiedLoopTestOf returns a loop's condition expression, or nil when the node is not a loop
// with one.
//
// `for...in` and `for...of` have no test at all and upstream's LOOP_PATTERN excludes them by name.
func unmodifiedLoopTestOf(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindWhileStatement:
		return node.AsWhileStatement().Expression
	case ast.KindDoStatement:
		return node.AsDoStatement().Expression
	case ast.KindForStatement:
		return node.AsForStatement().Condition
	}
	return nil
}

// unmodifiedLoopHasDynamicExpressions reports whether a group contains something that could change
// the world without the rule seeing it.
//
// Upstream's DYNAMIC_PATTERN is Call, Member, New, TaggedTemplate and Yield expressions, and its
// SKIP_PATTERN stops the walk at a nested function, because a function written inside the condition
// is not called by being written.
func unmodifiedLoopHasDynamicExpressions(root *ast.Node) bool {
	found := false
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		switch node.Kind {
		case ast.KindCallExpression, ast.KindNewExpression,
			ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindTaggedTemplateExpression, ast.KindYieldExpression:
			found = true
			return
		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression:
			// Written, not run.
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(root)
	return found
}

// unmodifiedLoopIsInitializerReference reports whether an identifier is the name being bound by a
// declaration that gives it a value.
//
// This is upstream's `reference.init`, which eslint-scope sets on exactly the initializer write of
// a declaration. Measured against it: `var foo = 0` marks position 4 with init true, while the
// later `foo = 1` and the read in the condition both have it false.
func unmodifiedLoopIsInitializerReference(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := parent.AsVariableDeclaration()
	return declaration.Name() == identifier && declaration.Initializer != nil
}

// unmodifiedLoopIsWriteReference reports whether an occurrence modifies the binding, in upstream's
// sense rather than in the plain one.
//
// `isWriteReference` declines an initializer reference UNLESS the binding is a `var`.
//
// The shape that depends on that exception is a `var` initializer inside a loop BODY, where the
// write counts as in the loop and makes the condition modified: `while (foo) { var foo = 0; }` is
// clean, and the same source with `let` reports, because the `let` shadows rather than writing the
// outer binding.
//
// It is NOT the corpus's own `for (var foo = 0; foo < 10; ) { } foo = 1;`, which reports whether or
// not this exception fires, since the head's write sits in the loop's initializer and is excluded
// from "in the loop" regardless. A mutant removing the exception survived that case, which is how
// the misreading in the first version of this comment was caught.
func unmodifiedLoopIsWriteReference(identifier *ast.Node, declaration *ast.Node) bool {
	if unmodifiedLoopIsInitializerReference(identifier) {
		list := loopFuncDeclarationListFor(declaration)
		if list == nil {
			return false
		}
		// A `var` list carries neither the let nor the constant flag.
		if list.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConstant) != 0 {
			return false
		}
	}
	return loopFuncWritesTo(identifier)
}

// unmodifiedLoopIsInLoop reports whether a reference sits in the part of a loop that runs each
// iteration.
//
// A ForStatement's initializer runs once, so a reference there is not in the loop. Upstream's
// `isInLoop` table says exactly this and says nothing else: the test and the update ARE in the
// loop, and so is the body.
func unmodifiedLoopIsInLoop(loop *ast.Node, identifier *ast.Node) bool {
	if identifier.Pos() < loop.Pos() || identifier.End() > loop.End() {
		return false
	}
	if loop.Kind == ast.KindForStatement {
		if initializer := loop.AsForStatement().Initializer; initializer != nil &&
			identifier.Pos() >= initializer.Pos() && identifier.End() <= initializer.End() {
			return false
		}
	}
	return true
}

// unmodifiedLoopHasModifierInLoop reports whether a write can run inside the loop.
//
// Directly, when the write itself is in the loop. Indirectly, when the write is inside a named
// function DECLARATION whose name is referenced from inside the loop, which is upstream's
// `hasModifierInLoop` reaching through a call.
func unmodifiedLoopHasModifierInLoop(
	ctx rule.Context,
	condition *unmodifiedLoopCondition,
	modifier *ast.Node,
	referencesByDeclaration map[*ast.Node][]*ast.Node,
) bool {
	if unmodifiedLoopIsInLoop(condition.loop, modifier) {
		return true
	}

	// Deliberately quieter than upstream: a write that can run while the loop is suspended. See
	// unmodifiedLoopWriterRunsWhileSuspended.
	if unmodifiedLoopWriterRunsWhileSuspended(condition.loop, modifier, referencesByDeclaration) {
		return true
	}

	// Only a function declaration, and only a named one. Upstream's
	// `getEncloseFunctionDeclaration` climbs to the first FunctionDeclaration and returns null when
	// it has no id, and it does not consider a function expression or an arrow at all: those have
	// no name to look up in the enclosing scope.
	function := unmodifiedLoopEnclosingFunctionDeclaration(modifier)
	if function == nil {
		return false
	}
	name := function.Name()
	if name == nil {
		return false
	}

	// Every reference to that function's own name, asking whether any of them is in the loop. This
	// is upstream looking the name up in the scope above the write and reading its references.
	for _, occurrence := range referencesByDeclaration[function] {
		if unmodifiedLoopIsInLoop(condition.loop, occurrence) {
			return true
		}
	}
	return false
}

// unmodifiedLoopWriterRunsWhileSuspended reports whether a write outside the loop can still run
// between two of its iterations. This is where cohere is deliberately quieter than upstream.
//
// Upstream credits a write only in the loop, or in a function declaration whose name the loop
// references. A loop that awaits hands the thread to whatever else is queued, so a flag set by a
// signal handler or an abort callback ends it, and upstream reports exactly that shape. The three real
// sites are `TasksWatchCommandLineInterface.ts:305`, `AhraOsMonitors.ts:548` and `RainbowMatrix.ts:689`
// in ahra; ESLint reports all three and all three are false.
//
// Two facts, both required:
//
//	the loop suspends        an `await` or `yield` in its test, body or update that belongs to the
//	                         loop's own function rather than to one nested in it
//	the write is elsewhere   in a different function activation from the loop's own, which can be
//	                         running while the loop is suspended
//
// "Elsewhere" is narrowed by when that activation can exist. A closure the loop's own function creates
// counts only if it is created before the loop could finish: a function or arrow expression written
// before the loop's end, or a hoisted declaration whose name is referenced before the loop's end (a
// handler registered in time). A hoisted declaration further out counts once anything references it.
// A write in an enclosing function's own body counts as it stands, since that activation resumes the
// moment the loop's function first suspends.
func unmodifiedLoopWriterRunsWhileSuspended(
	loop *ast.Node,
	modifier *ast.Node,
	referencesByDeclaration map[*ast.Node][]*ast.Node,
) bool {
	loopFunction := unmodifiedLoopNearestFunction(loop)
	if unmodifiedLoopNearestFunction(modifier) == loopFunction {
		// The loop's own activation, which cannot run while it is suspended.
		return false
	}

	// The outermost function holding the write that does not also hold the loop. Nil means the write
	// is in an enclosing function's own body, which is running whenever the loop is suspended.
	var closure *ast.Node
	for current := modifier.Parent; current != nil; current = current.Parent {
		if !ast.IsFunctionLikeDeclaration(current) {
			continue
		}
		if loop.Pos() >= current.Pos() && loop.End() <= current.End() {
			// Every ancestor of this function holds the loop too, so stopping here is an economy
			// rather than a decision: a mutant that kept climbing survived for exactly that reason.
			break
		}
		closure = current
	}

	if closure != nil {
		createdByTheLoopsFunction := unmodifiedLoopNearestFunction(closure) == loopFunction
		if closure.Kind == ast.KindFunctionDeclaration {
			if !unmodifiedLoopIsReferenced(closure, referencesByDeclaration, createdByTheLoopsFunction, loop.End()) {
				return false
			}
		} else if createdByTheLoopsFunction && closure.Pos() >= loop.End() {
			return false
		}
	}

	return unmodifiedLoopSuspends(loop)
}

// unmodifiedLoopIsReferenced reports whether anything names a hoisted function declaration, which is
// the only way it can run, and when `before` applies, whether it is named before that position.
func unmodifiedLoopIsReferenced(
	function *ast.Node,
	referencesByDeclaration map[*ast.Node][]*ast.Node,
	bounded bool,
	before int,
) bool {
	name := function.Name()
	for _, occurrence := range referencesByDeclaration[function] {
		if occurrence == name {
			continue
		}
		if !bounded || occurrence.Pos() < before {
			return true
		}
	}
	return false
}

// unmodifiedLoopNearestFunction returns the nearest function enclosing a node, or nil at the top level
// of the file, which is where a module's top-level `await` suspends.
func unmodifiedLoopNearestFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) {
			return current
		}
	}
	return nil
}

// unmodifiedLoopSuspends reports whether a loop yields the thread on some iteration: an `await` or a
// `yield` in its test, body or update, belonging to the loop's own function. A nested function's
// `await` suspends that function, not the loop, so the walk does not descend into one. A `for`
// initializer runs once before the first test, so a suspension there cannot end the loop.
func unmodifiedLoopSuspends(loop *ast.Node) bool {
	var initializer *ast.Node
	if loop.Kind == ast.KindForStatement {
		initializer = loop.AsForStatement().Initializer
	}
	found := false
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if found || node == nil || node == initializer {
			return found
		}
		if ast.IsFunctionLikeDeclaration(node) {
			return false
		}
		if node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindYieldExpression {
			found = true
			return true
		}
		node.ForEachChild(walk)
		return found
	}
	loop.ForEachChild(walk)
	return found
}

// unmodifiedLoopEnclosingFunctionDeclaration returns the nearest enclosing function declaration, or
// nil.
//
// The climb stops at the FIRST function declaration rather than skipping past an unnamed one, which
// is upstream's behaviour: it lands on whatever declaration encloses the write and does not keep
// searching outward for one that happens to have a name.
//
// That stop is the load-bearing part and it is what a mutant would have to attack. Upstream also
// returns null when the declaration it landed on has no id; here the caller asks `function.Name()`
// on the next line and returns false when it is nil, so restating the test inside this function is
// a duplicate rather than a guard. A mutant deleting it survived twice, including after a fixture
// written for it, and reading the caller is what explained why: both spellings return false on
// `export default function () { ++foo; }`, by different routes and with the same answer.
//
// The behaviour itself is pinned by TestNoUnmodifiedLoopConditionUnnamedFunctionDeclarationHasNoName
// ToReach, which asserts the verdict wherever it comes from.
//
// # The remaining mutant here is unreachable in well-formed source
//
// Adding `&& current.Name() != nil` to the stop, so the climb skips an unnamed declaration and
// keeps going outward, also survives. That is correct rather than a fixture gap: to distinguish the
// two spellings you need an UNNAMED function declaration NESTED inside a named one, and the only
// unnamed declaration the grammar has is `export default function`, which may not appear inside a
// function body.
//
// Measured, and this is the one place it is not simply absent: the TypeScript parser accepts
// `function d() { export default function () { ++foo; } }` through error recovery, and the
// installed eslint 10.8.1 build reports on it, which agrees with the stop written here. No fixture
// asserts it, because pinning a rule's behaviour on a syntax error records a fact about error
// recovery rather than about the rule. The distinguishing input is named here instead, so the next
// reader finding this mutant alive has it rather than having to re-derive it.
func unmodifiedLoopEnclosingFunctionDeclaration(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindFunctionDeclaration {
			return current
		}
	}
	return nil
}
