package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConstructorSuper requires a derived class constructor to call `super()` on every path, and
// forbids a `super()` call where there is no constructable superclass to call.
//
// The failure is a runtime one and it is total. A derived constructor that returns without calling
// `super()` throws a ReferenceError on every construction, and `super()` in a class that extends
// nothing is a syntax error the file will not even load with. Neither is a style question.
//
// Examples of incorrect code:
//
//	class A extends B { constructor() { } }
//	class A extends B { constructor() { if (a) super(); } }
//	class A extends B { constructor() { super(); super(); } }
//	class A extends null { constructor() { super(); } }
//
// Examples of correct code:
//
//	class A extends B { constructor() { super(); } }
//	class A extends B { constructor() { if (a) super(); else super(); } }
//	class A { constructor() { } }
//	class A extends null { constructor() { return a; } }
//
// # Four judgments, and which one an input lands in is the whole rule
//
// Upstream names them and this reproduces the names, because they are genuinely different findings
// rather than one finding with four wordings:
//
//	missingAll    a derived constructor with no reachable `super()` anywhere
//	missingSome   `super()` on some paths out of the constructor and not others
//	duplicate     a `super()` reachable on a path that has already called one
//	badSuper      a `super()` where the superclass cannot be constructed
//
// `missingAll` and `missingSome` anchor on the constructor, since the fix is to add a call and
// there is no single place the caret belongs. `duplicate` and `badSuper` anchor on the offending
// call, since there is one and it is the thing to delete.
//
// # There is no control flow graph here, and this is the third rule to establish that
//
// Upstream builds a real dataflow analysis on oxc's control flow graph: a five-state lattice
// (`Unreached`, `Never`, `Once`, `Multiple`, `Mixed`), a worklist over basic blocks, a join at
// every merge point, and special handling for back edges and explicit error edges. We have no such
// graph. TypeScript's flow graph is not a substitute and that has now been measured independently
// three times, by `getter-return`, by `no-this-before-super`, and here: `ast.FlowNode` carries
// `Antecedent` and `Antecedents` and no successor edges at all, because it exists for *narrowing*
// and is keyed on the expression positions where a type could change. There is no `super()`
// instruction in it to find, and no return or throw instruction either.
//
// So the analysis is structural, over the statement tree. The shape transferred from
// `getter-return` almost intact, which is worth saying precisely because it is the second time it
// has held:
//
//	an if merges its arms, and an arm that exits acceptably drops out of the merge
//	a try's finally overrides, and a catch means the try's own result is not guaranteed
//	a switch needs a default and every non-empty clause to agree
//	loops never establish anything, because the body may run zero times
//
// Where it did *not* transfer is the reason this rule needed a lattice rather than a boolean.
// `getter-return` asks a yes-or-no question ("does every path return") and a boolean answers it.
// This rule needs three values, because `super()` on *some* paths is a distinct finding from
// `super()` on *no* paths, and a boolean collapses exactly that distinction. `superSometimes` is
// the state that carries it, and it is why `combineBranches` is a real join rather than an `&&`.
//
// The second thing that did not transfer is the exit rule. `getter-return` treats every exit alike,
// because a `throw` and a `return x` both mean the getter did not fall off its end. Here they are
// the same as each other but different from a bare `return`, and the difference is the whole of two
// corpus pairs. A `throw` or a `return <expression>` is an *acceptable* exit: the constructor never
// completes normally, so the missing `super()` on that path cannot be observed, and upstream drops
// the path from the merge. A bare `return;` is not acceptable: it completes the constructor and
// yields an object whose `this` was never bound. Upstream's `is_acceptable_exit` reads exactly
// `Throw | Return(NotImplicitUndefined)`, and ESLint reaches the same place from the other
// direction with a `ReturnStatement` handler whose comment says a returned argument is a substitute
// for `super()`. Both are reproduced by `acceptableExit`.
//
// # Where the two upstreams disagree, and which one is reproduced
//
// The corpus was run through the real ESLint rule rather than read, and the two implementations
// agree on *whether* every one of the 87 inputs fires. They disagree on the message id for six of
// the 43 fail inputs, all involving a loop or a short circuit. The verdicts reproduced here are
// oxc's, because oxc's snapshot is the artifact that pins this port and choosing otherwise would
// make the imported corpus unable to check it.
//
// One of those six is an oxc defect rather than a judgment call, and it is reproduced knowingly.
// `class A extends B { constructor() { a && super(); } }` reports `missingAll` upstream and
// `missingSome` in ESLint, and ESLint is right: the `super()` is reachable whenever `a` is truthy,
// so it is on some paths rather than none. Oxc lands on `missingAll` because its control flow graph
// does not record the short-circuited right operand at all, so its `super_call_counts` map comes
// back empty and the first branch of its match fires. The finding is correct and only its wording
// is wrong, which is why it is reproduced rather than quietly improved on.
//
// # The classifier is ESLint's, and that is the one deliberate divergence
//
// Deciding whether `extends <expression>` names something constructable is a separate question from
// the path analysis, and the two upstreams solve it in opposite directions. Oxc writes a blacklist:
// `is_invalid_super_class` returns true for a numeric, string, boolean or bigint literal and for a
// binary expression, and false for everything else, so anything unrecognised is assumed
// constructable. ESLint writes a whitelist: `isPossibleConstructor` returns true for the ten node
// kinds that can hold a constructor and false for everything else.
//
// They agree on all 87 corpus inputs, which is why the corpus cannot separate them. They diverge on
// inputs the corpus does not contain, and every divergence is oxc missing a real defect:
//
//	class A extends undefined { constructor() { super(); } }
//	class A extends [] { constructor() { super(); } }
//	class A extends ({}) { constructor() { super(); } }
//	class A extends `x` { constructor() { super(); } }
//	class A extends /re/ { constructor() { super(); } }
//	class A extends (void 0) { constructor() { super(); } }
//
// Each of those is provably not a constructor and each is `badSuper` in ESLint and clean in oxc.
// The whitelist is used here. It is the stronger instrument, it costs nothing, and it agrees with
// oxc everywhere oxc has an opinion the corpus records. Six invented fixtures pin the divergence so
// a later reader can see it was a decision.
var ConstructorSuper = rule.Rule{
	Name: "constructor-super",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindConstructor: func(constructor *ast.Node) {
				checkConstructorSuper(ctx, constructor)
			},
		}
	},
}

// checkConstructorSuper is the whole decision for one constructor.
func checkConstructorSuper(ctx rule.Context, constructor *ast.Node) {
	class := constructor.Parent
	if class == nil ||
		(class.Kind != ast.KindClassDeclaration && class.Kind != ast.KindClassExpression) {
		return
	}

	body := constructor.AsConstructorDeclaration().Body
	if body == nil {
		// A constructor overload signature declares nothing executable, so there is no path
		// through it to have an opinion about.
		return
	}

	heritage := ast.GetClassExtendsHeritageElement(class)
	var superClass *ast.Node
	if heritage != nil {
		superClass = heritage.Expression()
	}

	// Collect first, judge second. Both the `badSuper` branch and the `duplicate` branch need the
	// list of calls, and the path analysis needs to know whether the list is empty, so gathering
	// once is cheaper than three walks and keeps the three answers reading the same tree.
	calls := collectSuperCalls(body.AsNode())

	switch classifySuperClass(superClass) {
	case superClassAbsent:
		// A base class. `super()` in one is a syntax error rather than a lint finding, and both
		// upstreams decline it for that reason: ESLint's parser refuses the file outright with
		// "super() call outside constructor of a subclass", and oxc's `SuperClassType::None`
		// branch reports only where a call was collected, which cannot happen in a file that did
		// not parse. Our parser is more permissive and does hand this rule a tree for
		// `class A { constructor() { super(); } }`, so the decline has to be explicit here where
		// upstream got it for free.
		//
		// A mutation sweep scores this `return` as a survivor and will keep doing so. With the
		// branch unreachable a base class matches no case and falls out of the switch, reporting
		// nothing by a different route, so no input can distinguish the two. That is subsumption
		// by the switch's absent default rather than a fixture gap, and adding a test for it would
		// assert nothing. It is kept rather than deleted because it is the only place a reader
		// learns that a base class was considered and deliberately passed over, and because a
		// later `default:` clause would make it load-bearing again without anything saying so.
		return

	case superClassNotConstructable:
		// `extends null`, `extends 100`, `extends (B += C)`. Every call is wrong and each is
		// reported, because each is separately deletable.
		if len(calls) > 0 {
			for _, call := range calls {
				ctx.ReportNode(call, rule.Message{
					Id: "badSuper",
					Description: "This `super()` call has no constructor to call, because the " +
						"`extends` clause names something that cannot be constructed.",
				})
			}
			return
		}
		// No call, and none is possible. The constructor is still required to produce a bound
		// `this` somehow, and the only way left is to return an object outright. Upstream checks
		// for a return carrying a value anywhere in the body and stays silent when it finds one,
		// which is `class A extends null { constructor() { return a; } }` in the pass list against
		// `class A extends null { constructor() { } }` in the fail list.
		if !bodyReturnsAValue(body.AsNode()) {
			reportMissingSuper(ctx, constructor, "missingAll")
		}
		return

	case superClassConstructable:
		checkConstructableSuperPaths(ctx, constructor, body.AsNode(), calls)
	}
}

// checkConstructableSuperPaths is the path analysis, run only once the superclass is known to be
// callable.
func checkConstructableSuperPaths(
	ctx rule.Context,
	constructor *ast.Node,
	body *ast.Node,
	calls []*ast.Node,
) {
	if len(calls) == 0 {
		// Nothing to be on some paths rather than others.
		reportMissingSuper(ctx, constructor, "missingAll")
		return
	}

	flow := superFlowOfStatement(body)

	// A `super()` written after an unconditional exit is in the source but on no path, so the
	// constructor has none and the finding is `missingAll` rather than `missingSome`.
	// `constructor() { return; super(); }` is upstream's case and it reports `missingAll` even
	// though a `super()` token is right there.
	//
	// The short-circuit cases land here too and that is upstream's defect, reproduced knowingly:
	// `a && super()` yields `superNever` because a short circuit establishes nothing, so it reads
	// as no path having a call. See the note on the rule.
	switch {
	case flow.calls == superNever:
		reportMissingSuper(ctx, constructor, "missingAll")
	case flow.calls == superSometimes || flow.escapedWithoutCalling:
		// `escapedWithoutCalling` is the second way to reach `missingSome`, and it is the one a
		// state lattice alone cannot express. `if (a) return; super();` comes out `superAlways`,
		// because the live path really does call it, and it is still `missingSome` because the
		// path that returned early is finished and never did.
		reportMissingSuper(ctx, constructor, "missingSome")
	}

	// Duplicate detection is a separate walk from the path analysis and has to be, because the two
	// ask opposite questions. The path analysis asks whether a call is guaranteed; this asks
	// whether a *second* call is possible. A loop is the case that separates them: its body may run
	// zero times, so it guarantees nothing, and it may also run twice, so it permits a duplicate.
	// The same loop is therefore `superNever` above and a duplicate source here, which is exactly
	// how upstream reads it and why its lattice needs both `Mixed` and `Multiple`.
	for _, duplicate := range findDuplicateSuperCalls(body) {
		ctx.ReportNode(duplicate, rule.Message{
			Id: "duplicate",
			Description: "This `super()` call can run on a path that already called `super()`, " +
				"which throws a ReferenceError because a constructor may only call it once.",
		})
	}
}

// reportMissingSuper anchors a missing-call finding on the constructor rather than on any one node
// inside it.
//
// The span runs from the `constructor` keyword to the closing brace, matching upstream's
// `constructor.span`, which is the method definition rather than the function value. Built through
// `rule.TokenRange` so a constructor preceded by a comment reports at the keyword and a
// `-next-line` suppression written above it can match.
func reportMissingSuper(ctx rule.Context, constructor *ast.Node, messageId string) {
	description := "This derived constructor never calls `super()`, so constructing the class " +
		"throws a ReferenceError before the body finishes."
	if messageId == "missingSome" {
		description = "This derived constructor calls `super()` on some paths and not others, " +
			"so the paths without it throw a ReferenceError."
	}
	ctx.ReportRange(rule.TokenRange(ctx.SourceFile, constructor), rule.Message{
		Id:          messageId,
		Description: description,
	})
}

// superClassKind is what an `extends` clause names, from this rule's point of view.
type superClassKind int

const (
	// superClassAbsent is a base class: no `extends` at all.
	superClassAbsent superClassKind = iota
	// superClassNotConstructable is an `extends` naming something that cannot be constructed,
	// including `null`. Every `super()` under one is wrong.
	superClassNotConstructable
	// superClassConstructable is an `extends` naming something that might hold a constructor.
	superClassConstructable
)

// classifySuperClass decides which of the three worlds a class lives in.
//
// `extends null` is its own case in the language and it is neither of the other two. Such a class
// is genuinely derived, so its constructor is still obliged to bind `this`, but it has no
// constructor to call, so `super()` throws. The only valid derived-from-null constructor is one
// that returns an object outright, which is why the `superClassNotConstructable` branch has a
// second check the other branches do not need. Upstream has three fixtures on it: `extends null` with
// no constructor is clean, with an empty constructor is `missingAll`, and with `return a;` is clean.
func classifySuperClass(superClass *ast.Node) superClassKind {
	if superClass == nil {
		return superClassAbsent
	}
	if isPossibleConstructor(superClass) {
		return superClassConstructable
	}
	return superClassNotConstructable
}

// isPossibleConstructor answers whether an expression could evaluate to something `new` works on.
//
// This is ESLint's whitelist rather than oxc's blacklist, and the choice is argued on the rule's
// doc comment. The direction is what matters: an expression nobody listed is *not* a constructor,
// so an unrecognised `extends` clause makes `super()` a finding. That is the reporting direction
// for this half of the rule, the opposite of the path analysis, and both are the safe direction for
// their own question.
func isPossibleConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindParenthesizedExpression:
		// Not in ESLint's list because its parser produces no such node. Ours does, so unwrapping
		// is required or every parenthesized corpus case reads as unrecognised. Twelve corpus
		// inputs are parenthesized and they split evenly across the verdict, so a missing unwrap
		// would be visible in both directions.
		return isPossibleConstructor(node.AsParenthesizedExpression().Expression)

	case ast.KindClassExpression, ast.KindFunctionExpression, ast.KindThisKeyword,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindYieldExpression,
		ast.KindTaggedTemplateExpression, ast.KindMetaProperty:
		// ESLint's `MemberExpression` covers both `a.b` and `a[b]`, which are separate kinds here.
		// `ChainExpression` has no counterpart: an optional chain is a PropertyAccessExpression
		// carrying a question-dot token, so `class A extends obj?.prop` lands on the line above
		// and is clean, as its pass case requires.
		return true

	case ast.KindNonNullExpression:
		// TypeScript only, so ESLint has no opinion. `B!` asserts `B` is not nullish and evaluates
		// to `B` itself, so it is constructable exactly when `B` is.
		return isPossibleConstructor(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		// Same reasoning. `B as C` is `B` at runtime.
		return isPossibleConstructor(node.AsAsExpression().Expression)

	case ast.KindIdentifier:
		// `undefined` is an identifier and is the one that is never a constructor. ESLint names it
		// explicitly for this reason and so does this.
		return node.Text() != "undefined"

	case ast.KindConditionalExpression:
		// Either branch being possible is enough, since only one runs and the rule reports only
		// what it can prove wrong. `extends (a ? B : C)` is a pass case and `extends (a ? 1 : 2)`
		// is `badSuper`.
		conditional := node.AsConditionalExpression()
		return isPossibleConstructor(conditional.WhenTrue) ||
			isPossibleConstructor(conditional.WhenFalse)

	case ast.KindBinaryExpression:
		return binaryIsPossibleConstructor(node.AsBinaryExpression())
	}

	// Literals, template literals, regular expressions, array and object literals, arrow
	// functions, unary and `void` expressions, and anything else all land here and answer false.
	// An arrow function is worth naming: it is a function but it has no `[[Construct]]` slot, so
	// `new (() => {})` throws and ESLint leaves it off the list deliberately.
	return false
}

// binaryIsPossibleConstructor handles every operator that produces a value worth asking about.
//
// The three groups are ESLint's and each is a different argument.
func binaryIsPossibleConstructor(binary *ast.BinaryExpression) bool {
	if binary.OperatorToken == nil {
		return false
	}

	switch binary.OperatorToken.Kind {
	case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindAmpersandAmpersandToken:
		// A plain assignment yields its right side. `&&` yields its right side when it does not
		// short circuit, and when it does short circuit the left side was falsy and so was not a
		// constructor either way, so only the right side needs asking about. `&&=` is the same
		// argument in assignment form. Upstream pins all three: `extends (B = C)`, `extends (5 && B)`
		// and `extends (B &&= C)` are clean while `extends (B = 5)`, `extends (B && 5)` and
		// `extends (B &&= 5)` are `badSuper`.
		return isPossibleConstructor(binary.Right)

	case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken,
		ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		// These can yield either operand, so either being possible is enough. `extends (B ||= 5)`
		// and `extends (B ?? 5)` are both clean, which is the pair that makes this distinct from
		// the `&&` line above: the literal `5` on the right does not condemn them.
		return isPossibleConstructor(binary.Left) || isPossibleConstructor(binary.Right)

	case ast.KindCommaToken:
		// A sequence yields its last operand. `extends (B, C)` is clean and `extends (B, 5)` is not.
		return isPossibleConstructor(binary.Right)
	}

	// Every remaining operator is arithmetic, bitwise, comparison, or their assignment forms. Each
	// yields a primitive or throws, and neither is a constructor. Six corpus inputs cover the
	// assignment forms and each is `badSuper`.
	return false
}

// collectSuperCalls gathers every `super(...)` belonging to this constructor, in source order.
//
// "Belonging to" is the load-bearing part. A `super()` inside a nested function, arrow, or class is
// a different constructor's business, or a syntax error, and either way it does not satisfy this
// one. Upstream has five fail cases resting on this and they are the shape that catches a walk that
// forgot to stop:
//
//	constructor() { var c = () => super(); }                     missingAll
//	constructor() { class C extends D { constructor() { super(); } } }  missingAll
//	constructor() { var c = class extends D { constructor() { super(); } } }  missingAll
//
// The arrow is the one worth pausing on, because an arrow *does* inherit the enclosing `super`
// binding, so `() => super()` is legal JavaScript that really would call this class's superclass
// constructor. It still does not count, because the arrow body does not run at the point it is
// written. The rule measures evaluation time rather than binding, which is the same distinction
// `no-this-before-super` draws for `this` and for the same reason.
func collectSuperCalls(node *ast.Node) []*ast.Node {
	var calls []*ast.Node
	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || isSeparateEvaluationContext(current) {
			return
		}
		if isSuperCall(current) {
			calls = append(calls, current)
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(node)
	return calls
}

// isSuperCall answers whether a node is `super(...)` rather than `super.x(...)`.
//
// The distinction is a fail case rather than a hypothetical: `constructor() { for (var a of b)
// super.foo(); }` is `missingAll` upstream, because `super.foo()` is a method call on the
// superclass prototype and calls no constructor at all. Asking whether the callee *is* the `super`
// keyword rather than whether it *contains* one is the whole difference.
func isSuperCall(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := node.AsCallExpression().Expression
	return callee != nil && callee.Kind == ast.KindSuperKeyword
}

// superState is how much a construct establishes about `super()` having been called.
//
// Three values rather than two, and the middle one is the reason. A boolean can say whether a call
// is guaranteed; it cannot separate "on some paths" from "on none", and those are two different
// findings with two different message ids. This is the lattice from upstream's `SuperCallState`
// with its duplicate-tracking states dropped, since duplicates are found by a separate walk here.
type superState int

const (
	// superNever is no path through this construct calling `super()`.
	superNever superState = iota
	// superSometimes is some but not all.
	superSometimes
	// superAlways is every path.
	superAlways
)

// superFlow is what one statement establishes: how much it calls, and whether control leaves.
type superFlow struct {
	calls superState
	// exits is true when control cannot continue past this statement.
	exits bool
	// acceptableExit distinguishes an exit that excuses a missing `super()` from one that does not.
	//
	// A `throw` or a `return <expression>` excuses it, because the constructor never completes
	// normally, so no caller ever holds the object whose `this` was never bound. A bare `return;`
	// does not, because it completes and yields exactly that object. Upstream's
	// `is_acceptable_exit` is `Throw | Return(NotImplicitUndefined)` and ESLint's `ReturnStatement`
	// handler marks a returned argument as a substitute for `super()`, which is the same rule
	// reached from the other side.
	//
	// Only meaningful when exits is true.
	acceptableExit bool
	// escapedWithoutCalling records that some path already left the constructor without calling
	// `super()` and without an excuse for it.
	//
	// It has to be carried separately from `calls` because it is a fact about a path that is
	// already over, and `combineSequential` deliberately lets a later `superAlways` absorb an
	// earlier state. That absorption is right for live paths and wrong for finished ones, and this
	// is the flag that keeps them apart. `if (a) return; super();` is the whole reason it exists.
	escapedWithoutCalling bool
}

// superFlowOfStatement is the must-analysis, and the direction of its uncertainty is what makes the
// port correct.
//
// Anything it cannot prove answers `superNever`, so an unrecognised construct is treated as not
// calling `super()` and the rule reports. For a rule about a guaranteed runtime crash that is the
// right bias, and it is upstream's too.
func superFlowOfStatement(node *ast.Node) superFlow {
	if node == nil {
		// An absent else arm, or an absent loop body. Establishes nothing and exits nothing, which
		// is what makes `if (a) super();` come out `superSometimes` without a special case.
		return superFlow{calls: superNever}
	}

	switch node.Kind {
	case ast.KindBlock:
		return superFlowOfStatements(node.AsBlock().Statements.Nodes)

	case ast.KindReturnStatement:
		if node.AsReturnStatement().Expression != nil {
			// A returned object replaces the one `super()` would have bound, so upstream treats
			// this path as excused. `class A extends B { constructor() { if (true) return a;
			// super(); } }` is a pass case and this line is why.
			return superFlow{calls: superNever, exits: true, acceptableExit: true}
		}
		// A bare `return;` yields the unbound `this`, which is the defect itself.
		// `constructor() { if (a) return; super(); }` is `missingSome`.
		return superFlow{calls: superNever, exits: true}

	case ast.KindThrowStatement:
		return superFlow{calls: superNever, exits: true, acceptableExit: true}

	case ast.KindIfStatement:
		statement := node.AsIfStatement()
		thenFlow := superFlowOfStatement(statement.ThenStatement)
		elseFlow := superFlowOfStatement(statement.ElseStatement)
		return combineBranches(thenFlow, elseFlow)

	case ast.KindTryStatement:
		return superFlowOfTry(node.AsTryStatement())

	case ast.KindSwitchStatement:
		return superFlowOfSwitch(node.AsSwitchStatement())

	case ast.KindLabeledStatement:
		// A label does not change what its statement establishes. A `break label` out of it would
		// reach the end of the labelled statement, which is already the fall-through path the
		// analysis models.
		return superFlowOfStatement(node.AsLabeledStatement().Statement)

	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		// A loop never establishes a call, because the body may run zero times. But it is not
		// `superNever` either when the body holds a call, and that distinction is worth eight of
		// the corpus's fail cases.
		//
		// `while (a) super();` reports `missingSome` rather than `missingAll`, because there are
		// executions in which `super()` ran and executions in which it did not, and the finding
		// names exactly that. Upstream reaches the same place from its control flow graph: when a
		// back edge encloses a `super()` it sets `loop_with_super`, and then pushes *both*
		// `NoSuper` and `CalledMultiple` into its path results, so its `some_missing` and its
		// `has_super_on_any_path` are true together.
		//
		// The matched pair is what pins it. `while (a) super();` is `missingSome` while
		// `for (a in b) for (c in d);` and `do { something(); } while (foo);` are `missingAll`,
		// and the only difference is whether the body holds a call.
		//
		// `do { ... } while (...)` runs at least once and could in principle establish a call.
		// Upstream declines to credit it, reporting `missingSome` for `do { super(); } while (foo);`,
		// and this reproduces that rather than silently improving on it. The improvement would be
		// sound and it would also make the imported corpus disagree with the port, which is the one
		// thing the corpus is for.
		if len(collectSuperCalls(node)) > 0 {
			return superFlow{calls: superSometimes}
		}
		return superFlow{calls: superNever}

	case ast.KindExpressionStatement:
		return superFlowOfExpression(node.AsExpressionStatement().Expression)
	}

	// Variable statements, empty statements, debugger, and everything else. A `super()` cannot
	// appear in one in a position that would establish anything: `var c = () => super()` puts it
	// inside an arrow, which does not count, and there is no other shape.
	return superFlow{calls: superNever}
}

// superFlowOfStatements walks a statement list in source order, stopping at the first exit.
//
// Stopping matters rather than being an optimisation. `constructor() { return; super(); }` is
// `missingAll` upstream even though the source plainly contains a `super()`, because the call is
// unreachable. Walking past the return would find it and report `missingSome`, which is the wrong
// finding for the right reason.
func superFlowOfStatements(statements []*ast.Node) superFlow {
	result := superFlow{calls: superNever}
	for _, statement := range statements {
		if result.exits {
			break
		}
		flow := superFlowOfStatement(statement)
		result.calls = combineSequential(result.calls, flow.calls)
		result.exits = flow.exits
		result.acceptableExit = flow.acceptableExit
		result.escapedWithoutCalling = result.escapedWithoutCalling || flow.escapedWithoutCalling
	}
	return result
}

// combineSequential joins what two statements establish when the second runs after the first.
//
// `superAlways` absorbs everything, because a second call does not un-call the first, and the
// second call being a duplicate is the other walk's business. `superSometimes` is sticky in the
// other direction: once some path has a call and some path does not, a later statement that
// establishes nothing leaves it that way.
func combineSequential(before superState, after superState) superState {
	if before == superAlways || after == superAlways {
		return superAlways
	}
	if before == superSometimes || after == superSometimes {
		return superSometimes
	}
	return superNever
}

// combineBranches joins two arms of a conditional, dropping an arm that exits acceptably.
//
// Dropping is the sharp part and it is what separates two upstream cases that differ by one word:
//
//	constructor() { if (a) throw Error(); super(); }   clean
//	constructor() { if (a) return; super(); }          missingSome
//
// The `throw` arm never completes the constructor, so there is no execution in which an object with
// an unbound `this` escapes, and upstream drops the path from its merge entirely. The bare `return`
// arm does complete it, so the path stays in the merge as one with no call, and merging that with
// the other arm's call yields `superSometimes`.
func combineBranches(first superFlow, second superFlow) superFlow {
	firstDropped := first.exits && first.acceptableExit
	secondDropped := second.exits && second.acceptableExit

	var calls superState
	switch {
	case firstDropped && secondDropped:
		// Both arms leave without completing, so the conditional as a whole does too and there is
		// no surviving path to have an opinion about.
		return superFlow{calls: superNever, exits: true, acceptableExit: true}
	case firstDropped:
		calls = second.calls
	case secondDropped:
		calls = first.calls
	default:
		calls = mergeStates(first.calls, second.calls)
	}

	return superFlow{
		calls: calls,
		// The conditional as a whole exits only when both arms do. An `if` with no else has an
		// absent arm whose flow is the zero value, so this comes out false without a special case.
		exits:          first.exits && second.exits,
		acceptableExit: first.exits && second.exits && first.acceptableExit && second.acceptableExit,
		// An arm that left the constructor without calling and without an excuse is a path that is
		// already finished and already wrong, and nothing written after the conditional can fix it.
		// `constructor() { if (a) return; super(); }` is `missingSome` for exactly this reason: the
		// `super()` after the `if` is real and does run on the other path, so the conditional's own
		// merge is `superAlways` by the time it reaches it, and without this flag the trailing call
		// would make the whole body come out clean.
		//
		// The matched pair is `if (a) throw Error(); super();`, which is clean. Both arms exit; the
		// difference is only whether the exit was acceptable, which is the same distinction
		// `acceptableExit` draws one line above and the reason it is recorded rather than folded
		// into `exits`.
		escapedWithoutCalling: first.escapedWithoutCalling || second.escapedWithoutCalling ||
			(first.exits && !first.acceptableExit && first.calls != superAlways) ||
			(second.exits && !second.acceptableExit && second.calls != superAlways),
	}
}

// mergeStates is the join: what two alternative paths establish together.
//
// Two paths that both call establish a call. Two that both do not establish none. One of each is
// exactly the `superSometimes` case, and it is the state a boolean analysis could not represent.
func mergeStates(first superState, second superState) superState {
	if first == second {
		return first
	}
	return superSometimes
}

// superFlowOfTry is the subtlest of the four and the corpus separates every shape it draws.
//
//	try {} finally { super(); }                clean:       the finally runs on every path
//	try { super(); } catch (err) {}            missingSome: the catch can swallow a throwing super()
//	try { a; } catch (err) { super(); }        missingSome: the catch may not run at all
//
// The second is the one that looks wrong and is not. `super()` can itself throw, partway through
// the superclass constructor, and then the catch runs and the constructor completes with `this`
// still unbound. Upstream reaches the same conclusion from its control flow graph by routing
// explicit error edges from the *pre-transfer* state of the block, which is the same statement in
// different words: nothing the try body achieved can be assumed on the exception path.
func superFlowOfTry(statement *ast.TryStatement) superFlow {
	// A finally that calls `super()` wins outright, because it runs on the normal path and the
	// exception path both.
	if statement.FinallyBlock != nil {
		finallyFlow := superFlowOfStatement(statement.FinallyBlock.AsNode())
		if finallyFlow.calls == superAlways {
			return finallyFlow
		}
	}

	tryFlow := superFlowOfStatement(statement.TryBlock.AsNode())
	if statement.CatchClause == nil {
		// No catch, so the only way out of the try block is through it.
		return tryFlow
	}

	catchFlow := superFlowOfStatement(statement.CatchClause.AsCatchClause().Block.AsNode())
	// The catch runs from the state at the try's *entry*, never from its exit, because an exception
	// can arrive from anywhere in the body including from inside `super()` itself. So the two
	// alternatives being merged are "the try completed" and "the catch ran instead", and the second
	// is credited only with what the catch itself establishes.
	return combineBranches(tryFlow, catchFlow)
}

// superFlowOfSwitch answers whether a switch calls `super()` on every path.
//
// Two conditions and both are pinned by the corpus. Without a default clause the scrutinee may
// match nothing and fall straight out, which is `switch (a) { case 0: super(); }` reporting
// `missingSome`. And every clause has to call, which is
// `switch (a) { case 0: break; default: super(); }` reporting `missingSome` against
// `switch (a) { case 0: super(); break; default: super(); }` being clean.
//
// Fallthrough is what makes this more than a loop over clauses. An empty clause is not a path out:
// `case 0: case 1: super();` runs the same statements for both, so the empty `case 0` calls by way
// of its successor. Only a trailing empty clause genuinely falls out of the switch.
func superFlowOfSwitch(statement *ast.SwitchStatement) superFlow {
	clauses := statement.CaseBlock.AsCaseBlock().Clauses.Nodes

	hasDefault := false
	for _, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		// The scrutinee may match nothing and fall straight out, so no path is guaranteed. But a
		// clause holding a call still means *some* execution called it, and that is a different
		// finding: `switch (a) { case 0: super(); }` is `missingSome` rather than `missingAll`.
		// The same asymmetry as a loop, for the same reason, and pinned by the same kind of
		// matched pair: this against `switch (a) { case 0: break; }`, which holds no call at all.
		if len(collectSuperCalls(statement.AsNode())) > 0 {
			return superFlow{calls: superSometimes}
		}
		return superFlow{calls: superNever}
	}

	everyClauseCalls := true
	anyClauseCalls := false
	for index, clause := range clauses {
		statements := clause.AsCaseOrDefaultClause().Statements.Nodes
		if len(statements) == 0 {
			if index == len(clauses)-1 {
				// Nothing after it to fall through to.
				everyClauseCalls = false
			}
			continue
		}
		flow := superFlowOfStatements(statements)
		if flow.calls == superAlways {
			anyClauseCalls = true
		} else {
			everyClauseCalls = false
			if flow.calls == superSometimes {
				anyClauseCalls = true
			}
		}
	}

	switch {
	case everyClauseCalls:
		return superFlow{calls: superAlways}
	case anyClauseCalls:
		return superFlow{calls: superSometimes}
	}
	return superFlow{calls: superNever}
}

// superFlowOfExpression handles the expression statements that can establish a call.
//
// Three shapes and they are three different answers, which is why this is not one recursive walk.
func superFlowOfExpression(node *ast.Node) superFlow {
	if node == nil {
		return superFlow{calls: superNever}
	}

	switch node.Kind {
	case ast.KindParenthesizedExpression:
		return superFlowOfExpression(node.AsParenthesizedExpression().Expression)

	case ast.KindCallExpression:
		if isSuperCall(node) {
			return superFlow{calls: superAlways}
		}

	case ast.KindConditionalExpression:
		// `a ? super() : super()` is a pass case and `a ? super() : b` is not, so the two branches
		// merge exactly as an `if` does. Nothing here exits, so `combineBranches` degenerates to
		// its merge.
		conditional := node.AsConditionalExpression()
		return combineBranches(
			superFlowOfExpression(conditional.WhenTrue),
			superFlowOfExpression(conditional.WhenFalse),
		)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			// A sequence runs both, so either establishing a call establishes one.
			return superFlow{calls: combineSequential(
				superFlowOfExpression(binary.Left).calls,
				superFlowOfExpression(binary.Right).calls,
			)}
		}
		if shortCircuitsRightOperand(binary.OperatorToken) {
			// Only the *right* operand short circuits. The left one always evaluates, which is
			// what separates upstream's two logical cases: `super() || super();` reports one
			// `duplicate` and nothing else, because the left call is guaranteed, while
			// `a && super();` reports `missingAll`, because the guaranteed operand holds no call
			// and the conditional one establishes nothing.
			//
			// That pair is also where oxc and ESLint part company. ESLint models the short circuit
			// as a real branch and reports `missingSome` for `a && super();`, which is the better
			// finding: the call is reachable whenever `a` is truthy, so it is on some paths rather
			// than none. Oxc's control flow graph does not record the short-circuited operand at
			// all, so its call map comes back empty and it reports `missingAll`. Oxc's verdict is
			// the one reproduced, because its snapshot is what pins this port.
			return superFlow{calls: superFlowOfExpression(binary.Left).calls}
		}
	}

	return superFlow{calls: superNever}
}

// shortCircuitsRightOperand answers whether an operator may skip its right operand.
//
// The three short-circuiting operators and their assignment forms. What they share is the only
// property this rule cares about: the right side is not guaranteed to evaluate, while the left side
// always is.
//
// Named for the property rather than for the operator family on purpose, because `isLogicalOperator`
// already exists in this package asking a different question: whether a node is a binary expression
// joined by one *specific* operator. Two helpers with the same name and different questions is how a
// package grows a subtle bug, so this one says what it decides.
func shortCircuitsRightOperand(operator *ast.Node) bool {
	if operator == nil {
		return false
	}
	switch operator.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// findDuplicateSuperCalls returns every `super()` that can run on a path where one already has.
//
// This is a second walk rather than part of the path analysis because the two questions point in
// opposite directions. The path analysis is a *must* analysis and resolves its uncertainty toward
// "no call"; this is a *may* analysis and resolves its uncertainty toward "a call already
// happened". Both directions are the reporting direction for their own question, which is the
// thing that makes them two walks instead of one traversal carrying two bits.
//
// The four duplicate cases upstream records are the shapes that pin it:
//
//	super(); super();                                    the second
//	super() || super();                                  the right operand
//	if (a) super(); super();                             the second
//	switch (a) { case 0: super(); default: super(); }     the second
//
// The third and fourth are why "already called" has to mean *may* rather than *must*. After
// `if (a) super();` a call has only possibly happened, and the following `super()` is still
// reported. A *must* reading would report neither, and a naive "everything after the first" reading
// would report `if (a) super(); else super();` too, which is a pass case: those two calls are
// alternatives rather than a sequence, and no execution reaches both.
//
// Loops are the case where this diverges from ESLint knowingly. A loop body holding a `super()` can
// obviously reach it twice, and ESLint reports the duplicate for `while (a) super();` on top of the
// `missingSome`. Oxc reports only the `missingSome`, because its `loop_with_super` path pushes
// `CalledMultiple` but its duplicate branch is gated on `super_call_spans.len() > 1` and a loop
// holds one span. Four corpus inputs would gain a second finding if the loop were credited here,
// which is why it is not: the snapshot records one diagnostic for each of them.
func findDuplicateSuperCalls(body *ast.Node) []*ast.Node {
	finder := &duplicateSuperFinder{}
	finder.walkStatement(body, false)
	return finder.duplicates
}

// duplicateSuperFinder walks a constructor body in source order tracking whether a `super()` may
// already have run.
//
// A struct rather than threaded returns because the walk is genuinely stateful in one direction and
// branching in the other: a statement sets the bit for its successors, while two arms of a
// conditional each read the entry bit and neither sees the other's.
type duplicateSuperFinder struct {
	// mayHaveCalled records that at least one execution reaching this point has called `super()`.
	mayHaveCalled bool
	// duplicates is every call found while mayHaveCalled was already true, in source order.
	duplicates []*ast.Node
}

// walkStatement walks one statement, returning nothing and mutating mayHaveCalled in place.
//
// `insideLoop` is carried down so a `super()` in a loop body can be recognised without being
// credited as a duplicate of itself. It is the one place where the same call is both the first and
// the second, and upstream declines to report it; see the note on findDuplicateSuperCalls.
func (finder *duplicateSuperFinder) walkStatement(node *ast.Node, insideLoop bool) {
	if node == nil || isSeparateEvaluationContext(node) {
		return
	}

	switch node.Kind {
	case ast.KindBlock:
		finder.walkStatementList(node.AsBlock().Statements.Nodes, insideLoop)
		return

	case ast.KindIfStatement:
		statement := node.AsIfStatement()
		finder.walkStatement(statement.Expression, insideLoop)
		// Each arm reads the entry state and neither sees the other's result, which is what keeps
		// `if (a) super(); else super();` clean. The merge is an OR because either arm having
		// called means a call may have happened.
		entry := finder.mayHaveCalled

		finder.mayHaveCalled = entry
		finder.walkStatement(statement.ThenStatement, insideLoop)
		thenCalled := finder.mayHaveCalled && !statementExits(statement.ThenStatement)

		finder.mayHaveCalled = entry
		finder.walkStatement(statement.ElseStatement, insideLoop)
		elseCalled := finder.mayHaveCalled && !statementExits(statement.ElseStatement)

		// An arm that leaves the constructor contributes nothing to what follows, because no
		// execution continues from it into the code after the `if`. That is what keeps
		// `for (const a of list) { if (a.foo) { super(a); return; } } super();` clean: the loop
		// body's call is on a path that returns, so it can never be the first of two.
		finder.mayHaveCalled = thenCalled || elseCalled
		return

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		finder.walkStatement(conditional.Condition, insideLoop)
		entry := finder.mayHaveCalled

		finder.mayHaveCalled = entry
		finder.walkStatement(conditional.WhenTrue, insideLoop)
		whenTrueCalled := finder.mayHaveCalled

		finder.mayHaveCalled = entry
		finder.walkStatement(conditional.WhenFalse, insideLoop)
		whenFalseCalled := finder.mayHaveCalled

		finder.mayHaveCalled = whenTrueCalled || whenFalseCalled
		return

	case ast.KindSwitchStatement:
		statement := node.AsSwitchStatement()
		finder.walkStatement(statement.Expression, insideLoop)
		// Clauses are walked in order with the state carried through, because fallthrough is real
		// and `case 0: super(); default: super();` reaches both. That is upstream's fourth
		// duplicate case and it is the reason the clauses are not treated as alternatives the way
		// an if's arms are.
		// Clauses are walked in order, but only a clause that falls through hands its state to
		// the next one. A `break`, `return` or `throw` ends the clause, so the following one
		// starts from the switch's entry state instead.
		//
		// That is the whole difference between two corpus cases one word apart:
		// `switch (a) { case 0: super(); break; default: super(); }` is clean, because the `break`
		// means no execution reaches both calls, while
		// `switch (a) { case 0: super(); default: super(); }` reports a `duplicate`, because the
		// first clause falls into the second and one execution runs both.
		entry := finder.mayHaveCalled
		anyClauseCalled := entry
		for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
			statements := clause.AsCaseOrDefaultClause().Statements.Nodes
			finder.walkStatementList(statements, insideLoop)
			anyClauseCalled = anyClauseCalled || finder.mayHaveCalled
			if clauseBreaksOut(statements) {
				finder.mayHaveCalled = entry
			}
		}
		finder.mayHaveCalled = anyClauseCalled
		return

	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		finder.walkLoop(node)
		return

	case ast.KindTryStatement:
		statement := node.AsTryStatement()
		entry := finder.mayHaveCalled

		finder.walkStatement(statement.TryBlock.AsNode(), insideLoop)
		tryCalled := finder.mayHaveCalled

		catchCalled := false
		if statement.CatchClause != nil {
			// The catch runs from the try's entry state, since an exception can arrive before
			// anything in the try body completed.
			finder.mayHaveCalled = entry
			finder.walkStatement(statement.CatchClause.AsCatchClause().Block.AsNode(), insideLoop)
			catchCalled = finder.mayHaveCalled
		}

		finder.mayHaveCalled = tryCalled || catchCalled
		if statement.FinallyBlock != nil {
			// The finally runs after either, so it sees whatever they left.
			finder.walkStatement(statement.FinallyBlock.AsNode(), insideLoop)
		}
		return

	case ast.KindCallExpression:
		if isSuperCall(node) {
			// Arguments evaluate before the call, so a `super()` nested in them is the earlier one.
			// `super(super())` is not legal, but `super(f(() => 0))` walks the same way and the
			// ordering costs nothing to get right.
			if arguments := node.AsCallExpression().Arguments; arguments != nil {
				for _, argument := range arguments.Nodes {
					finder.walkStatement(argument, insideLoop)
				}
			}
			if finder.mayHaveCalled && !insideLoop {
				finder.duplicates = append(finder.duplicates, node)
			}
			finder.mayHaveCalled = true
			return
		}
	}

	// Everything else contributes its children in source order with the state carried straight
	// through. That covers expression statements, returns, variable statements, and the logical
	// operators: `super() || super()` walks its left operand then its right, so the right is
	// reported, which is upstream's second duplicate case. Short circuiting does not save it,
	// because this is a *may* analysis and the non-short-circuiting execution exists.
	node.ForEachChild(func(child *ast.Node) bool {
		finder.walkStatement(child, insideLoop)
		return false
	})
}

// clauseBreaksOut answers whether a switch clause ends without falling into the next one.
//
// A `break` is the ordinary spelling and the only one the corpus uses, but a `return` or a `throw`
// ends the clause just as completely, so all three are asked about through the same helper the path
// analysis uses.
func clauseBreaksOut(statements []*ast.Node) bool {
	for _, statement := range statements {
		if statement.Kind == ast.KindBreakStatement {
			return true
		}
		if superFlowOfStatement(statement).exits {
			return true
		}
	}
	return false
}

// statementExits answers whether control cannot continue past a statement.
//
// A thin wrapper over the path analysis so the duplicate walk asks the same question one way
// instead of growing its own second opinion about what an exit is.
func statementExits(node *ast.Node) bool {
	return superFlowOfStatement(node).exits
}

// walkStatementList walks statements in order, stopping at the first unconditional exit.
//
// Stopping is what keeps an unreachable call from counting. `constructor() { return; super(); }`
// holds one `super()` and reports `missingAll`; a body holding two after a `return` reports the
// same rather than gaining a duplicate, because neither call is reachable and a duplicate is a
// second *reachable* call.
func (finder *duplicateSuperFinder) walkStatementList(statements []*ast.Node, insideLoop bool) {
	for _, statement := range statements {
		finder.walkStatement(statement, insideLoop)
		if superFlowOfStatement(statement).exits {
			return
		}
	}
}

// walkLoop walks a loop without letting its body count as a duplicate of itself.
//
// The head runs before the body and its state carries out, so it is walked normally. The body is
// walked with `insideLoop` set, which suppresses the duplicate report while still letting a call in
// it set the state for whatever follows the loop. That last part is what makes
// `while (a) super(); super();` report its trailing call, and it is also why the body cannot simply
// be skipped.
func (finder *duplicateSuperFinder) walkLoop(node *ast.Node) {
	// The body is not exempt from duplicate reporting; see the note below on what `insideLoopBody`
	// actually suppresses.
	const insideLoopBody = false

	var head []*ast.Node
	var body *ast.Node

	switch node.Kind {
	case ast.KindWhileStatement:
		statement := node.AsWhileStatement()
		head = []*ast.Node{statement.Expression}
		body = statement.Statement
	case ast.KindDoStatement:
		statement := node.AsDoStatement()
		head = []*ast.Node{statement.Expression}
		body = statement.Statement
	case ast.KindForStatement:
		statement := node.AsForStatement()
		head = []*ast.Node{statement.Initializer, statement.Condition, statement.Incrementor}
		body = statement.Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		statement := node.AsForInOrOfStatement()
		head = []*ast.Node{statement.Initializer, statement.Expression}
		body = statement.Statement
	}

	for _, part := range head {
		finder.walkStatement(part, false)
	}

	// The body is walked exactly once, from the state the head left, and a call inside it is
	// reported normally. `while (a) { super(); super(); }` runs both on one iteration, so the
	// second is a duplicate by the same reading as anywhere else, and oxc agrees: its duplicate
	// branch is gated on `super_call_spans.len() > 1` and a body with two calls holds two spans.
	//
	// Walking once is the whole of the divergence from ESLint. ESLint models the back edge, so a
	// *single* call in a loop body is a duplicate of itself on the next iteration and it reports
	// `while (a) super();` as `missingSome` and `duplicate` both. Oxc's span gate does not, because
	// one call is one span, and its four loop cases each record exactly one diagnostic. Not walking
	// the body a second time is what reproduces that, and ESLint's is the more useful finding.
	//
	// There is deliberately no merge back of the entry state here. It read as the load-bearing line
	// and it was subsumed: `mayHaveCalled` is monotonic, set to true and never back to false, so
	// after the body walk it already carries whatever the entry state held. A mutation sweep scored
	// `mayHaveCalled = entry` against `mayHaveCalled = mayHaveCalled || entry` as a survivor and no
	// fixture could have caught it, because the two expressions cannot differ. Deleted rather than
	// tested, so the next reader does not spend the same half hour proving the same thing.
	finder.walkStatement(body, insideLoopBody)
}

// bodyReturnsAValue answers whether any `return <expression>` appears in a constructor body.
//
// Only used by the `extends null` branch, where returning an object is the only way to produce a
// bound `this` at all. Upstream's `has_return_with_value` walks the same set of nesting statements
// and this reproduces it, including that it does *not* ask whether the return is on every path: one
// anywhere is enough to stay silent. That is looser than the rest of the rule and it is upstream's
// looseness, reproduced rather than tightened, since no corpus case separates the two readings.
func bodyReturnsAValue(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if isSeparateEvaluationContext(node) {
		// A `return x` inside a nested function returns from that function.
		return false
	}
	if node.Kind == ast.KindReturnStatement {
		return node.AsReturnStatement().Expression != nil
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if bodyReturnsAValue(child) {
			found = true
			return true
		}
		return false
	})
	return found
}
