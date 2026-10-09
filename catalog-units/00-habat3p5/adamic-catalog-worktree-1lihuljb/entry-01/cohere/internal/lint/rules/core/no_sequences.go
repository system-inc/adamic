package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedCommaExpression = rule.Message{
	Id: "unexpectedCommaExpression",
	Description: "This uses the comma operator, which evaluates the expression on its left, throws " +
		"the result away, and yields the expression on its right. Two statements are hiding inside " +
		"one, and the discarded half is the part a reader skips: `a = b(), c()` calls both and " +
		"assigns only the second, which reads as a typo for `a = b()`. Split it into separate " +
		"statements, or wrap it in its own parentheses to say the sequencing was deliberate.",
}

// NoSequencesOptions configures whether an explicitly parenthesized sequence is excused.
type NoSequencesOptions struct {
	// AllowInParentheses is upstream's only option and it defaults to TRUE, which is why it is a
	// pointer rather than a bool.
	//
	// A zero-value struct would read as `false` and silently invert the rule: every `var foo = (1,
	// 2)` in the tree would report, and no fixture built by handing `RunWithOptions` a struct could
	// see it, because such a fixture states the value it is testing. The nil case is the one the
	// live config actually produces, since a rule configured as bare "error" is handed no options at
	// all, and it is the case the default has to survive.
	AllowInParentheses *bool `json:"allowInParentheses"`
}

// allowInParentheses reads the option, supplying upstream's default for an absent or nil value.
func (options NoSequencesOptions) allowInParentheses() bool {
	if options.AllowInParentheses == nil {
		return true
	}
	return *options.AllowInParentheses
}

// NoSequences flags the comma operator outside the two places it is idiomatic.
//
//	valid:   var arr = [1, 2];
//	valid:   var a = 1, b = 2;
//	valid:   var foo = (1, 2);
//	valid:   for (i = 1, j = 2;; i++, j++);
//	valid:   if ((doSomething(), !!test));
//	invalid: 1, 2;
//	invalid: a = 1, 2
//	invalid: if (doSomething(), !!test);
//	invalid: while((1) , 2);
//
// Most commas in JavaScript are separators rather than operators: array elements, object properties,
// arguments, and the declarators of one `var` all use the token and none of them is this rule's
// subject. The comma *operator* is the one in expression position, and it is worth a rule because
// the two readings look identical and the wrong one is the common one. `a = b(), c()` binds tighter
// than it reads: the assignment takes `b()`, the comma sequences it with `c()`, and the value the
// author probably wanted is discarded.
//
// # Two exemptions, and both are about intent rather than safety
//
// The `for` statement's initializer and update are the one position where sequencing is the only way
// to say the thing, so upstream exempts them outright. The `for` *condition* is not exempt, since a
// condition discarding its left half is the defect the rule is about: `for (; a, b ;)` reports while
// `for (a, b;;)` and `for (;; a, b)` do not.
//
// Everywhere else, wrapping the sequence in its own parentheses is read as the author saying the
// sequencing was deliberate, and the rule stands down. That is the `allowInParentheses` option and
// it defaults to on.
//
// # Where our tree and upstream's disagree about the shape, and why the port is not structural
//
// Two differences, and each would ship a defect on its own.
//
// **Chained commas are one node upstream and several here.** ESTree flattens `a, b, c` into a single
// `SequenceExpression` holding three expressions, and reports once, at the first comma. typescript-go
// parses the same source as nested `BinaryExpression` nodes with `KindCommaToken`, so a listener
// firing per node would report twice on `a, b, c` and three times on `a, b, c, d`. Measured against
// the installed rule: both report exactly once, at column 2. So this walks to the outermost node of a
// chain, reports there, and points at the chain's leftmost comma. A parenthesized inner sequence is
// genuinely a separate `SequenceExpression` upstream, and the parenthesis is what stops the chain
// here too, which is why `a, (b, c)` reports once at column 2 and `(a, b), c` once at column 7.
//
// **The parentheses upstream counts are not all nodes here.** Upstream tests parenthesization
// positionally over tokens, reading the token before and after the node at depth 1, and it demands
// *two* levels for a node the grammar already parenthesizes: the tests of `if`, `while`, `do...while`
// and `switch`, and the object of `with`. In our tree those grammar parentheses are not nodes at all
// -- `if ((a, b));` produces exactly one `KindParenthesizedExpression` -- so counting node levels
// gives the same answer at one level that upstream gets at two. The arrow body is the exception and
// the reason `requiresExtraParens` survives here in any form: `a => (b, c)` does produce a
// `ParenthesizedExpression`, because those parentheses belong to the expression rather than to a
// statement's grammar, so the arrow body needs two nodes where every other position needs one.
// Measured against the installed rule, which reports `a => (b, c)` and is silent on `a => ((b, c))`.
//
// The consequence is that this rule counts paren *nodes* and upstream counts paren *tokens*, and the
// counts differ by one in five of the six positions. That is fidelity to the decision rather than to
// the mechanism: the inputs that report are the same inputs.
//
// # One measured divergence, in emission order rather than in content
//
// Under `allowInParentheses: false` a nested pair like `(a, b), c` reports twice, and upstream emits
// the inner finding before the outer one while this emits the outer first. Both walks are pre-order;
// what differs is which node is outer. Upstream's outer node is the flattened sequence whose reported
// comma is the LATER one, and its inner node is the parenthesized child whose comma is earlier, so
// pre-order and source order happen to agree there and disagree here.
//
// Five of 264 differential rows differ this way and no row differs in content: same count, same
// spans, same message. It is invisible downstream because `report.Write` sorts every finding by file
// and position before printing, so the printed order is source order either way. Stated rather than
// corrected, because reordering emission would mean holding findings back to sort them and the only
// thing it would buy is agreement with an artifact nobody reads.
//
// Measured by running all 66 probe inputs under all four option shapes through both implementations.
//
// No fix. Upstream ships none, and the repair is a judgment the rule cannot make: splitting a
// sequence into statements is only correct where the expression is already a statement, and the
// alternative repair of adding parentheses silences the finding without addressing it.
var NoSequences = rule.Rule{
	Name: "no-sequences",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The live config passes bare "error", which reaches a rule as nil options rather than as a
		// zero-value struct, so the type assertion fails and the default has to come from the zero
		// value of NoSequencesOptions. Its AllowInParentheses is a nil pointer, which the accessor
		// reads as upstream's `true`.
		var settings NoSequencesOptions
		if parsed, ok := rule.OptionsAs[NoSequencesOptions](options); ok {
			settings = parsed
		}
		allowInParentheses := settings.allowInParentheses()

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if node.AsBinaryExpression().OperatorToken.Kind != ast.KindCommaToken {
					return
				}

				// One report per chain, taken at the OUTERMOST node. `a, b, c` is one node upstream
				// and two here, and it is the outermost of ours that occupies the position
				// upstream's flattened node occupies -- which is what every check below reads. A
				// parenthesized inner sequence is a separate node upstream as well, so a
				// ParenthesizedExpression parent ends the chain and starts a new one.
				//
				// The direction is the whole of it, and the inner-first spelling passes 62 fixtures.
				// Standing down on the outer node instead reports the inner one, whose parent is the
				// outer comma expression rather than the statement, so the `for` exemption and the
				// paren allowance below both look at the wrong node and decline to fire. Measured
				// against the installed rule, that spelling reports five inputs upstream is silent
				// on, including `for (a, b, c;;)` and `(a, b, c);`. Found by a mutation sweep and by
				// nothing in the imported corpus, which never writes a three-element chain.
				// No check for WHICH side of the parent this is, and that omission is measured. A
				// mutant weakening it to a tautology survived every fixture, so the question was
				// whether a comma expression can be the direct RIGHT operand of another. It cannot:
				// comma is left-associative, so `a, b, c` parses as `(a, b), c` and reaching the
				// right side needs parentheses, which interpose a node and end the chain. Probed
				// over 17 shapes including assignments, ternaries, `yield`, and five-element chains,
				// with zero right-side occurrences. Taken over this one caller.
				if isCommaExpression(node.Parent) {
					return
				}

				// The initializer and the update of a `for`, and not its condition. Upstream's
				// comment says commas in a call or a `new` are argument separators and never reach
				// here, which is true of our parse as well.
				if standsInForHeaderSlot(node) {
					return
				}

				if allowInParentheses && parenthesizedEnough(node) {
					return
				}

				comma := leftmostCommaOf(ctx, node)
				if comma == nil {
					return
				}
				ctx.ReportRange(*comma, messageUnexpectedCommaExpression)
			},
		}
	},
}

// standsInForHeaderSlot reports whether a node is the initializer or the update of a `for`.
//
// The climb through parentheses is not a convenience. Upstream compares `node === node.parent.init`
// on a tree where parentheses are not nodes, so its comparison sees through any number of them, and
// it makes that comparison BEFORE the paren allowance rather than after. Two of upstream's own clean
// cases turn on exactly that ordering, and both are written with `allowInParentheses: false` so the
// allowance cannot be what excuses them:
//
//	for ((i = 0, j = 0); test; );     clean
//	for (; test; (i++, j++));         clean
//
// Without the climb our tree sees a ParenthesizedExpression parent, the slot comparison misses, the
// allowance is switched off, and both report. Measured against the installed rule, which is also
// silent on `for (((i = 0, j = 0)); test; );` at two levels.
//
// The climb stops at the first parent that is not a parenthesis, so it never crosses into a
// different expression: `for ((a, b), c; ; )` exempts the outer sequence and reports the inner one,
// whose parent is that outer sequence rather than the statement. Measured, at column 8.
func standsInForHeaderSlot(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindForStatement {
		return false
	}
	forStatement := parent.AsForStatement()
	return forStatement.Initializer == current || forStatement.Incrementor == current
}

// isCommaExpression reports whether a node is itself a comma operator rather than something else.
func isCommaExpression(node *ast.Node) bool {
	return node != nil &&
		node.Kind == ast.KindBinaryExpression &&
		node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken
}

// parenthesizedEnough reports whether the author wrote enough parentheses to claim the sequencing.
//
// One `KindParenthesizedExpression` everywhere except an arrow body, which needs two. See the rule's
// doc comment for why that asymmetry is fidelity rather than a special case: it is the one position
// where the parentheses upstream's grammar table names are a node in our tree instead of raw tokens.
func parenthesizedEnough(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindParenthesizedExpression {
		return false
	}
	if !isArrowBody(parent) {
		return true
	}
	grandparent := parent.Parent
	return grandparent != nil && grandparent.Kind == ast.KindParenthesizedExpression
}

// isArrowBody reports whether a node stands where a concise arrow body stands.
//
// Upstream's `parenthesized` table also names the tests of `if`, `while`, `do...while` and `switch`
// and the object of `with`, and every one of those is absent here on purpose: their parentheses are
// grammar rather than expression, so typescript-go builds no node for them and the level upstream
// has to discount does not exist to be discounted.
//
// No `Body == node` comparison, and that omission is measured rather than assumed. A mutant that
// weakened the comparison to a tautology survived every fixture, so the question was whether a
// PARENTHESIZED EXPRESSION whose parent is an arrow can ever be something other than that arrow's
// body. Probed over 21 shapes -- parenthesized and destructured and defaulted parameters, generics,
// `async`, a parenthesized return type, an arrow inside a default value, an arrow as a class
// property -- and it cannot. An arrow's parameters parse as `KindParameter` and their parentheses
// are grammar, so `((a, b)) => c` produces no ParenthesizedExpression at all, and parenthesizing the
// whole arrow makes the arrow the paren's child rather than its parent. The comparison was written
// first and removed once the probe showed nothing could reach the false branch; it is recorded here
// because the next reader will want to add it back.
//
// The verdict names the callers it was taken over: `parenthesizedEnough` is the only one. A second
// caller reaching this with a node that is not a ParenthesizedExpression voids it.
func isArrowBody(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && parent.Kind == ast.KindArrowFunction
}

// leftmostCommaOf returns the range of the first comma token in a chain of comma operators.
//
// Upstream reports at `getTokenAfter(node.expressions[0], isCommaToken)`, which on its flattened node
// is the comma following the first expression of the whole chain. Our nested parse reaches the same
// token by descending the left spine, because comma is left-associative and the deepest left operand
// is upstream's `expressions[0]`.
//
// The operator token carries leading trivia in its `Pos`, so `((1)) , (2)` would report a span
// starting at the space rather than at the comma. `rule.TokenRange` scans past it, which is the same
// correction `ctx.ReportNode` applies and the reason this does not build the range from `Pos()`
// directly.
func leftmostCommaOf(ctx rule.Context, node *ast.Node) *core.TextRange {
	current := node
	for isCommaExpression(current.AsBinaryExpression().Left) {
		current = current.AsBinaryExpression().Left
	}
	operator := current.AsBinaryExpression().OperatorToken
	if operator == nil {
		return nil
	}
	commaRange := rule.TokenRange(ctx.SourceFile, operator)
	return &commaRange
}

// DecodeNoSequencesOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, because that helper reports an error on empty
// input and the config layer turns the error into nil, so a rule reached through it can never see
// the difference between "no options were written" and "an option was written as false". For an
// option defaulting to false that distinction is invisible; for this one it inverts the rule.
func DecodeNoSequencesOptions(raw []byte) (any, error) {
	var options NoSequencesOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}
