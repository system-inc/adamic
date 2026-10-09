package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageCondAssign = rule.Message{
	Id: "condAssign",
	Description: "This condition assigns rather than compares, so it does not test what it reads " +
		"as testing: it writes the value and then branches on whatever was written, which makes " +
		"the branch depend on the assigned value rather than on any comparison. It is almost " +
		"always a `==` or `===` typed as `=`. If the assignment is deliberate, wrap it in its own " +
		"parentheses to say so.",
}

// NoCondAssignMode is the option value, spelled as the config spells it.
//
// A named string type rather than a struct because this rule's option is a bare string in the
// config tuple: `["error", "always"]` puts `"always"` where other rules put an object. The decoder
// is a plain json.Unmarshal, which decodes a JSON string into this directly.
type NoCondAssignMode = string

const (
	// NoCondAssignExceptParens permits an assignment wrapped in its own parentheses, which is the
	// established way to write "the assignment here is deliberate". This is the default.
	NoCondAssignExceptParens NoCondAssignMode = "except-parens"

	// NoCondAssignAlways reports every assignment reachable from a conditional test, parenthesized
	// or not.
	NoCondAssignAlways NoCondAssignMode = "always"
)

// NoCondAssignOptions is the decoded option.
type NoCondAssignOptions = NoCondAssignMode

// NoCondAssign flags an assignment in a conditional test.
//
//	valid:   if (x === 0) { }
//	valid:   if ((someNode = someNode.parentNode) !== null) { }
//	valid:   if ((a = b));                       (except-parens, the default)
//	valid:   while (obj.key) { obj.key = false }
//	valid:   switch (foo) { case a = b: bar(); }
//	invalid: if (x = 0) { }
//	invalid: var b = (x = 0) ? 1 : 0;
//	invalid: if ((a = b));                       (always)
//
// Mistyping `==` as `=` inside a condition is one of the oldest defects in the language and one of
// the quietest: the code parses, runs, assigns, and then branches on the assigned value, so the
// condition is not merely wrong but has a side effect the reader does not expect.
//
// # The option is about parentheses, and the default is the permissive one
//
// There is a real idiom here that the rule must not break. `while ((match = regex.exec(s)) !== null)`
// is the standard way to iterate matches, and the extra parentheses are how an author says the
// assignment was meant. Under `except-parens`, the default, an assignment only reports when it is
// *directly* the test, so wrapping it silences the rule. Under `always`, parentheses stop meaning
// anything and every assignment reachable from the test reports.
//
// Getting that backwards fires on exactly the code people write to declare their intent, so the
// fixtures pin both directions and the unconfigured default separately.
//
// # Five test positions, and the ones that are not tests
//
// `if`, `while`, `do`, the middle clause of `for`, and the test of a ternary. Not the `for`
// initializer or updater, which are statement positions where an assignment is the point, and not a
// `case` label, which is compared rather than tested. Upstream's pass list covers all four of those
// near misses.
//
// # An upstream asymmetry, reproduced
//
// A ternary test has its parentheses stripped in *both* modes, while the four statement forms strip
// them only under `always`. So `if ((a = b));` passes under the default and
// `((a = b)) ? x : y` reports under it. That is inconsistent, and it looks like an oversight rather
// than a decision: nothing in the rule or its documentation distinguishes a ternary's parentheses
// from an if's. It is reproduced here because upstream's own fail corpus pins it, with
// `(((3496.29)).bkufyydt = 2e308) ? foo : bar;` listed as a failure under no options at all. A port
// that "fixed" it would disagree with the reference on a case the reference explicitly tests.
//
// # One finding per assignment, where upstream emits two
//
// Upstream runs two branches that overlap. Its statement branch reports an assignment sitting
// directly in a test, and under `always` its assignment branch walks ancestors and reports any
// assignment inside a test's span, which includes that same direct one. Nine inputs in its own fail
// corpus therefore snapshot the identical diagnostic twice, at byte-identical spans, which is how
// 21 failing inputs produce 30 diagnostics.
//
// That is a defect rather than a gap upstream chose: a duplicated finding at one span is not a
// second thing to fix, and it double-counts in any total. The walk under `always` already covers
// the direct case on its own, so this reports through one path per mode and emits one finding. The
// divergence is stated rather than silent, and the fixtures assert the single count.
var NoCondAssign = rule.Rule{
	Name: "no-cond-assign",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Unrecognized and absent both land on the permissive mode. A rule enabled with a bare
		// severity must not silently run the strict half, and a typo must not silently escalate.
		always := false
		if mode, configured := rule.OptionsAs[NoCondAssignOptions](options); configured && mode == NoCondAssignAlways {
			always = true
		}

		// Under `always`, every finding comes from the assignment itself walking up to see whether
		// it sits in a test. That single path covers a bare assignment in the test position too,
		// which is the overlap upstream reports twice.
		if always {
			return rule.Listeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					if !isAssignmentExpression(node) {
						return
					}
					if assignmentSitsInAConditionalTest(node) {
						reportCondAssign(ctx, node)
					}
				},
			}
		}

		// Under `except-parens`, only an assignment that IS the test reports, so parentheses around
		// it make the test a ParenthesizedExpression and the rule stays silent.
		//
		// `test` is nil for `for(;;)`, and `isAssignmentExpression` answers false for nil on its
		// first line, so there is no separate nil guard. A mutation proved that: removing the
		// `test == nil` check changed no behavior on any input, because nothing past it
		// dereferences a nil.
		reportIfTestIsAssignment := func(test *ast.Node) {
			if !isAssignmentExpression(test) {
				return
			}
			reportCondAssign(ctx, test)
		}

		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				reportIfTestIsAssignment(node.AsIfStatement().Expression)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				reportIfTestIsAssignment(node.AsWhileStatement().Expression)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				reportIfTestIsAssignment(node.AsDoStatement().Expression)
			},
			ast.KindForStatement: func(node *ast.Node) {
				// A `for` with no test is `for(;;)`, which has no node to look at.
				reportIfTestIsAssignment(node.AsForStatement().Condition)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				// Parentheses stripped even in this mode, which is the upstream asymmetry above.
				reportIfTestIsAssignment(ast.SkipParentheses(node.AsConditionalExpression().Condition))
			},
		}
	},
}

// isAssignmentExpression reports whether a node is an assignment of any operator.
//
// `IsAssignmentOperator` rather than a comparison against `=`, because the compound forms assign
// too. Upstream's fail corpus includes `for(; x+=1 ;)`, and the logical forms `&&=`, `||=`, and
// `??=` are assignments the corpus predates.
func isAssignmentExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	return binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
}

// assignmentSitsInAConditionalTest walks up from an assignment looking for a test position holding
// it.
//
// The walk stops at a function, an arrow, or a block, which is what separates `if (a = b)` from
// `if (cond) { a = b }`. Without that stop, every assignment in a conditional's *body* reports,
// since the body is inside the statement. Upstream's pass list covers both the braced bodies and
// the unbraced ones, where there is no block to stop at and the containment check does the work.
//
// Containment rather than identity is the second half: the assignment may be nested arbitrarily
// deep inside the test, as in `if (someNode || (someNode = parentNode))`, and comparing nodes for
// equality would find only the direct case.
func assignmentSitsInAConditionalTest(assignment *ast.Node) bool {
	for ancestor := assignment.Parent; ancestor != nil; ancestor = ancestor.Parent {
		var test *ast.Node
		switch ancestor.Kind {
		case ast.KindIfStatement:
			test = ancestor.AsIfStatement().Expression
		case ast.KindWhileStatement:
			test = ancestor.AsWhileStatement().Expression
		case ast.KindDoStatement:
			test = ancestor.AsDoStatement().Expression
		case ast.KindForStatement:
			test = ancestor.AsForStatement().Condition
		case ast.KindConditionalExpression:
			test = ancestor.AsConditionalExpression().Condition

		case ast.KindArrowFunction, ast.KindBlock:
			// A body, not a test. The assignment runs when the body runs, which is the case this
			// rule is not about.
			//
			// Two kinds rather than upstream's six, because the other four are unreachable here and
			// a mutation sweep proved it: removing FunctionDeclaration, FunctionExpression,
			// MethodDeclaration, or SourceFile from this set changed no behavior on any input,
			// including object methods, getters, class methods, and nested function declarations in
			// a test position. Every one of those bodies is a Block, so the walk stops at the Block
			// before it ever reaches the function node above it, and SourceFile is unreachable
			// because a walk that gets that far has already returned false at the Block or run out
			// of ancestors.
			//
			// ArrowFunction earns its place on one shape only: `() => a = 1` has an expression body
			// with no Block, so it is the single function form whose body the Block case cannot
			// catch. Removing it is caught by a fixture; removing the other four is not, which is
			// what says they are redundant rather than untested.
			//
			// Listing them anyway would read as four handled cases, and the next reader would have
			// no way to tell which of the six are load-bearing.
			return false
		}

		if test == nil {
			continue
		}
		// Inclusive on both ends, so an assignment that is exactly the test counts. That is what
		// lets this one path cover the direct case as well as the nested one.
		if test.Pos() <= assignment.Pos() && assignment.End() <= test.End() {
			return true
		}
	}
	return false
}

// reportCondAssign points the finding at the assignment operator rather than the whole expression.
//
// `rule.TokenRange` rather than the operator token's own position: a node's `Pos()` runs from before
// its leading trivia, so on `while (a /* = */ = b) {}` a span built from `OperatorToken.Pos()`
// covers ` /* = */ =` and anchors the finding inside the comment. Upstream carries that exact input
// in its fail corpus for this reason, and `TokenRange` scans past the trivia to the token itself.
func reportCondAssign(ctx rule.Context, assignment *ast.Node) {
	operator := assignment.AsBinaryExpression().OperatorToken
	ctx.ReportRange(rule.TokenRange(ctx.SourceFile, operator), messageCondAssign)
}
