package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedExceptionAssignment = rule.Message{
	Id: "unexpectedExceptionAssignment",
	Description: "This assigns to the caught exception, which discards the only reference to what " +
		"actually went wrong. Anything after this point in the handler reports the new value " +
		"rather than the error, so a rethrow rethrows the wrong thing and a log line describes " +
		"something that never happened. The usual intent is a second binding: declare one instead " +
		"of overwriting the parameter.",
}

// NoExAssign flags an assignment to the parameter of a catch clause.
//
//	valid:   try { work(); } catch(error) { report(error); }
//	valid:   try { work(); } catch(error) { const wrapped = wrap(error); throw wrapped; }
//	invalid: try { work(); } catch(error) { error = wrap(error); }
//	invalid: try { work(); } catch(error) { error ??= fallback; }
//
// The damage is that it is silent. Overwriting the binding does not fail, and the handler keeps
// running with a value that is no longer the error, so the failure surfaces later as a log line or a
// rethrown exception that describes the wrong thing. That is the shape of defect that survives
// review, because the code reads as though it is enriching the error.
//
// Every assignment operator counts, not just `=`. A compound assignment writes to the same binding,
// and `error ??= fallback` is the form that looks most innocent while doing exactly this.
//
// A destructured catch parameter is not reported. `catch({ message })` binds new names rather than
// the exception itself, so assigning to one of them does not lose the error.
//
// No fix. The repair is to introduce a new binding, which means choosing a name and rewriting every
// later use, and the rule cannot know which later uses meant the original error.
var NoExAssign = rule.Rule{
	Name: "no-ex-assign",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCatchClause: func(node *ast.Node) {
				clause := node.AsCatchClause()
				if clause.VariableDeclaration == nil || clause.Block == nil {
					return
				}

				// Only a plain identifier binds the exception itself. A destructuring pattern binds
				// new names taken out of it, so writing to one of those loses nothing.
				name := clause.VariableDeclaration.AsVariableDeclaration().Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				reportAssignmentsTo(ctx, clause.Block, name.Text(), messageUnexpectedExceptionAssignment)
			},
		}
	},
}

// reportAssignmentsTo walks a subtree reporting every assignment whose target is the named binding.
//
// Compound assignments are included, since `name += x` and `name ??= x` write to the binding exactly
// as `name = x` does, and the compound forms are the ones that read as harmless.
func reportAssignmentsTo(ctx rule.Context, root *ast.Node, targetName string, message rule.Message) {
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil {
			return
		}

		if current.Kind == ast.KindBinaryExpression {
			binary := current.AsBinaryExpression()
			if binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				if left := ast.SkipParentheses(binary.Left); left != nil &&
					left.Kind == ast.KindIdentifier && left.Text() == targetName {
					ctx.ReportNode(left, message)
				}
			}
		}

		// An update expression is a write too, and upstream's `is_write()` counts it. Its corpus
		// never exercises one, which is how the gap survived the port: every fixture on both sides
		// used `=` or a compound operator, and `e++` reported nothing while `e += 1` reported.
		//
		// Prefix and postfix both, since `++e` and `e++` differ only in what they evaluate to and
		// neither leaves the binding alone.
		if current.Kind == ast.KindPrefixUnaryExpression || current.Kind == ast.KindPostfixUnaryExpression {
			if operand, operator := updateOperandAndOperator(current); reference.IsUpdateOperator(operator) {
				if operand != nil && operand.Kind == ast.KindIdentifier && operand.Text() == targetName {
					ctx.ReportNode(operand, message)
				}
			}
		}

		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(root)
}

// updateOperandAndOperator reads the operand and operator of a unary expression, either fixity.
func updateOperandAndOperator(node *ast.Node) (*ast.Node, ast.Kind) {
	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		return ast.SkipParentheses(unary.Operand), unary.Operator
	case ast.KindPostfixUnaryExpression:
		unary := node.AsPostfixUnaryExpression()
		return ast.SkipParentheses(unary.Operand), unary.Operator
	}
	return nil, ast.KindUnknown
}
