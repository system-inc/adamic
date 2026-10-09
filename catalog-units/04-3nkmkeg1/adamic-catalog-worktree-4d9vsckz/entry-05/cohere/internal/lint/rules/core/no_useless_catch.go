package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnnecessaryCatch = rule.Message{
	Id: "unnecessaryCatch",
	Description: "This `catch` does nothing but rethrow what it caught, so the whole `try` has no " +
		"effect: the error propagates exactly as it would with no `try` at all. It reads as if the " +
		"error is being handled, which is the harm, since the next reader has to check. Delete the " +
		"`try` and `catch` and keep the body, or handle the error here.",
}

var messageUnnecessaryCatchClause = rule.Message{
	Id: "unnecessaryCatchClause",
	Description: "This `catch` does nothing but rethrow what it caught, so it changes nothing: the " +
		"error propagates as it would without it, and the `finally` still runs either way. Delete the " +
		"`catch` clause and keep `try`/`finally`, or handle the error here.",
}

// NoUselessCatch flags a `catch` whose first statement rethrows the error it caught.
//
//	valid:   try { f(); } catch (error) { log(error); }
//	valid:   try { f(); } catch (error) { cleanup(); throw error; }
//	valid:   try { f(); } catch (error) { throw new WrappedError(error); }
//	valid:   try { f(); } finally { cleanup(); }
//	invalid: try { f(); } catch (error) { throw error; }
//
// Catching an error and immediately rethrowing it leaves runtime behavior exactly as it was without
// the `try`. The cost is not the wasted frames, it is the reading: the shape says an error is being
// handled here, so anyone tracing where a failure goes has to stop and confirm that it isn't.
//
// Which node is reported depends on `finally`, and the distinction is the repair rather than the
// diagnosis. Without a `finally` the entire `try` is doing nothing, so the whole statement is
// reported and the repair is to unwrap it. With a `finally` the `try` still has work to do, so only
// the `catch` clause is reported and the repair keeps `try`/`finally` intact. Two messages, because
// one message that named the wrong repair would be worse than no message.
//
// Only the first statement is examined, which is ESLint's rule and not a shortcut. A rethrow in
// first position makes every statement after it unreachable, so the clause is useless whatever else
// it contains; a rethrow after other statements means the clause did something first and is doing
// real work.
//
// A destructured binding is exempt: `catch ({ message }) { throw message; }` rethrows a property
// rather than the error, which is a different value and a different behavior. So is a bare `catch`
// with no binding, which has nothing to name and so cannot be rethrowing what it caught.
//
// Where TypeScript adds shapes ESLint has no view on, we follow what survives compilation. `throw
// (error)` is reported, since parentheses are grouping and ESTree does not even keep them. `throw
// error!` and `throw error as Error` are reported too: both are erased before anything runs, so the
// thrown value is the caught binding unchanged and the clause is exactly as useless. rslint treats
// those last two as valid, on the reading that the wrapper makes it not a plain rethrow. We diverge
// deliberately, because ESLint's rule is about what the code does at runtime and at runtime there is
// no wrapper.
//
// No fix. Removing a `try` is a structural edit whose correct form depends on what the author meant
// the clause to become.
var NoUselessCatch = rule.Rule{
	Name: "no-useless-catch",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindCatchClause: func(node *ast.Node) {
				catchClause := node.AsCatchClause()
				if catchClause == nil {
					return
				}

				// A bare `catch` names nothing, so nothing it throws can be the thing it caught.
				if catchClause.VariableDeclaration == nil {
					return
				}
				binding := catchClause.VariableDeclaration.AsVariableDeclaration()
				if binding == nil || binding.Name() == nil {
					return
				}
				// A destructured binding rethrows a property, which is a different value.
				if binding.Name().Kind != ast.KindIdentifier {
					return
				}
				caughtName := binding.Name().AsIdentifier().Text

				if catchClause.Block == nil {
					return
				}
				block := catchClause.Block.AsBlock()
				if block == nil || block.Statements == nil || len(block.Statements.Nodes) == 0 {
					return
				}

				firstStatement := block.Statements.Nodes[0]
				if firstStatement.Kind != ast.KindThrowStatement {
					return
				}
				throwStatement := firstStatement.AsThrowStatement()
				if throwStatement == nil || throwStatement.Expression == nil {
					return
				}

				thrown := unwrapErasedExpressions(throwStatement.Expression)
				if thrown == nil || thrown.Kind != ast.KindIdentifier {
					return
				}
				if thrown.AsIdentifier().Text != caughtName {
					return
				}

				tryStatement := node.Parent
				if tryStatement == nil || tryStatement.Kind != ast.KindTryStatement {
					return
				}
				tryData := tryStatement.AsTryStatement()
				if tryData == nil {
					return
				}

				if tryData.FinallyBlock != nil {
					ctx.ReportNode(node, messageUnnecessaryCatchClause)
					return
				}
				ctx.ReportNode(tryStatement, messageUnnecessaryCatch)
			},
		}
	},
}

// unwrapErasedExpressions strips the wrappers that are gone by the time the code runs, so what is
// left is the value actually thrown.
//
// Parentheses are grouping and never reach ESTree at all. A non-null assertion and an `as` are
// TypeScript's alone and erase to nothing. All three leave the thrown value identical to the caught
// binding, which is the condition the rule is about.
func unwrapErasedExpressions(expression *ast.Node) *ast.Node {
	for expression != nil {
		switch expression.Kind {
		case ast.KindParenthesizedExpression:
			expression = expression.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			expression = expression.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			expression = expression.AsAsExpression().Expression
		default:
			return expression
		}
	}
	return nil
}
