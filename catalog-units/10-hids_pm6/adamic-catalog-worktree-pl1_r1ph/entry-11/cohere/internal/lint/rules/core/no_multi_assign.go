package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoMultiAssign = rule.Message{
	Id: "unexpectedChain",
	Description: "This chains one assignment inside another, so a single statement writes to two " +
		"names and only one of them is where a reader looks. In `let a = b = c` the `b` is not " +
		"declared at all, so a chain inside a declaration silently creates or writes an outer " +
		"binding. Write each assignment on its own line.",
}

// NoMultiAssignOptions carries upstream's one option.
//
// `IgnoreNonDeclaration` drops the bare-chain arm, leaving only the two arms that sit under a
// declaration. It defaults to FALSE, which is the safe direction for the zero value: an absent key
// and an explicit false mean the same thing, so unlike a default-true option this one does not
// invert when the config layer hands the rule nil. The pointer is kept anyway so the two stay
// distinguishable if the default ever moves, and so the decoder's own behaviour is testable.
type NoMultiAssignOptions struct {
	IgnoreNonDeclaration *bool `json:"ignoreNonDeclaration"`
}

// DecodeNoMultiAssignOptions turns the configured object into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so that empty input returns the defaults instead
// of an error the config layer turns into a nil the rule then reads as a zero value. That path is
// harmless for this rule, whose one option defaults to false, and it is written out anyway because
// the harmlessness is a property of today's default rather than of the code.
//
// The configured shape is the bare object rather than upstream's `[{...}]`: cohere's config layer
// strips the severity-and-options tuple before dispatch, so a fixture copying ESLint's spelling
// fails on every row.
func DecodeNoMultiAssignOptions(raw []byte) (any, error) {
	var options NoMultiAssignOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// NoMultiAssign flags an assignment nested inside another assignment or an initializer.
//
//	valid:   var a = 1, b = 2;
//	valid:   var a = 1 + (b === 10 ? 5 : 4);
//	valid:   class C { [foo = 0] = 0 }
//	invalid: var a = b = c;
//	invalid: a = b = c
//	invalid: class C { field = foo = 0 }
//
// Upstream is written as three selectors and this is the same three, anchored the other way round.
// A selector language matches a child by naming its parent and the field it sits in; a kind listener
// receives the node and asks what its parent is. Same question, and the parent test is what makes
// `class C { [foo = 0] = 0 }` clean, because that assignment's parent is the computed key rather
// than the field.
//
//	VariableDeclarator > AssignmentExpression.init      an assignment as a declarator's initializer
//	PropertyDefinition > AssignmentExpression.value     an assignment as a class field's initializer
//	AssignmentExpression > AssignmentExpression.right   an assignment as another's right-hand side
//
// The third is dropped when `ignoreNonDeclaration` is set, which is exactly upstream's own
// conditional push onto the selector list.
//
// # What counts as an assignment, measured rather than assumed
//
// ESTree's `AssignmentExpression` covers all sixteen assignment operators, not just `=`, and
// upstream's selector therefore does too. Measured against the installed build at 10.8.1: every one
// of `= += -= *= /= %= **= <<= >>= >>>= &= |= ^= &&= ||= ??=` reports in `var q = w OP e`. The
// corpus writes only `=`, so a port reading the corpus alone would ship a rule silent on fifteen
// of them.
//
// `ast.IsAssignmentOperator` answers for exactly those sixteen and false for `==`, `+`, `&&`, `,`,
// `instanceof` and `??`, probed against this list rather than taken from its name.
//
// # Parentheses, which are load-bearing here in both positions
//
// Espree produces no node for a parenthesized expression, so upstream's `.init` and `.right` fields
// already see through them. Measured: `var a = (b = c);` reports and `x = (y = z)` reports.
// typescript-go does produce a node, so the initializer and the right operand are both skipped
// through before the kind is read, or those two inputs go silent.
//
// The reported SPAN keeps the parentheses that belong to the assignment's own operands.
// `var a = (b) = (((c)))` reports at column 9, which is the `(` of `(b)`, because that parenthesis
// is inside the assignment rather than around it. That falls out of reporting the assignment node
// itself and is asserted rather than assumed.
var NoMultiAssign = rule.Rule{
	Name: "no-multi-assign",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare "error" is handed nil, so the type assertion yields the zero
		// value. That is correct here only because the one option defaults to false; the fallback is
		// written out rather than relied on so a later default change cannot invert the rule
		// silently.
		settings, _ := rule.OptionsAs[NoMultiAssignOptions](options)
		ignoreNonDeclaration := settings.IgnoreNonDeclaration != nil && *settings.IgnoreNonDeclaration

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				if !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				if !nestedAssignmentIsChained(node, ignoreNonDeclaration) {
					return
				}
				ctx.ReportNode(node, messageNoMultiAssign)
			},
		}
	},
}

// nestedAssignmentIsChained answers whether this assignment sits in one of upstream's three slots.
//
// The slot is decided by the parent AND by which field of the parent holds this node, which is what
// the selector's `.init`, `.value` and `.right` say. Testing the parent kind alone would report
// `class C { [foo = 0] = 0 }`, whose assignment is the computed key rather than the initializer, and
// `a, b = c`, whose parent is a binary expression that is not an assignment.
//
// # Parentheses move the parent, which is where the first draft was wrong
//
// Espree gives a parenthesized expression no node, so upstream's `.init` and `.right` are already
// the assignment. typescript-go inserts one, and it goes on BOTH sides of the link: in
// `var a = (b = c);` the assignment's own `Parent` is the parenthesized expression rather than the
// declaration, so a switch reading `node.Parent.Kind` never sees the slot at all.
//
// So the climb is upward, and the downward skip stays too. The upward climb finds the real parent;
// the downward skip confirms this node is the one that field arrives at rather than some deeper
// assignment inside it. Both directions were measured: `var a = (b = c);` and `x = (y = z)` report
// upstream and went silent on the draft that only skipped downward.
func nestedAssignmentIsChained(node *ast.Node, ignoreNonDeclaration bool) bool {
	// Climb past every parenthesis, carrying the outermost one so the downward test below compares
	// against the node the parent's field actually holds.
	outermost := node
	for outermost.Parent != nil && outermost.Parent.Kind == ast.KindParenthesizedExpression {
		outermost = outermost.Parent
	}
	parent := outermost.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	// The two declaration arms test which FIELD holds the node, matching upstream's `.init` and
	// `.value`, and both tests are SUBSUMED by the kind switch above them rather than load-bearing.
	// A mutation rewriting either to a bare `true` survived every fixture in this file, so the
	// question was settled by probe rather than by argument: across thirty shapes including source
	// the parser had to recover from, no assignment expression ever reached either arm from a
	// position other than the initializer. A declarator's name and a class field's name are not
	// expressions, and a computed key reparents to `KindComputedPropertyName`, which this switch
	// does not name.
	//
	// Kept rather than deleted, which is the opposite of what a surviving mutant usually earns. The
	// equivalence is a property of what typescript-go's parser can produce rather than of this code,
	// so it is exactly the kind of verdict that expires: a parser change, or a new caller reaching
	// this helper with a node from somewhere else, voids it silently. The tests also say out loud
	// which slot each arm is, which is upstream's own selector and the thing a reader is checking
	// this against.
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == outermost
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == outermost
	case ast.KindBinaryExpression:
		if ignoreNonDeclaration {
			return false
		}
		outer := parent.AsBinaryExpression()
		// The outer expression has to be an assignment too. `a, b = c` and `a || (b = c)` are
		// binary expressions holding an assignment on the right and neither is a chain.
		if outer.OperatorToken == nil || !ast.IsAssignmentOperator(outer.OperatorToken.Kind) {
			return false
		}
		// The RIGHT side specifically, which is upstream's `.right`. An assignment on the left, as
		// in `(a = b) = c`, is not a chain and the parser can produce it from recovered source.
		return outer.Right == outermost
	}
	return false
}
