package lower

import (
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) isNever(node *ast.Node) bool {
	return reducedOptionalSource(l.checker, l.concrete(l.checker.GetTypeAtLocation(node))).Flags()&checker.TypeFlagsNever != 0
}

// A flow type is not an unreachability proof: a call can invalidate a narrowing.
// Use ordinary IR so both backends stop identically, before loading an impossible
// field representation. Receiver and index effects still run in source order.
func (l *lowering) neverExpression(node *ast.Node) (ir.Expression, error) {
	operands := []ir.Expression{}
	var effect ir.Expression
	of := ir.Number
	switch node.Kind {
	case ast.KindIdentifier:
	case ast.KindPropertyAccessExpression:
		object, err := l.expression(node.AsPropertyAccessExpression().Expression)
		if err != nil {
			return nil, err
		}
		operands = append(operands, object)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		object, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		operands = append(operands, object, index)
	default:
		var value ir.Expression
		var err error
		if node.Kind == ast.KindCallExpression {
			value, err = l.callOrMethod(node)
		} else {
			value, err = l.uncheckedValue(node)
		}
		if err != nil {
			return nil, err
		}
		if value.Type() != 0 {
			of = value.Type()
			operands = append(operands, value)
		} else {
			effect = value
			switch call := value.(type) {
			case ir.Call:
				operands = append(operands, call.Arguments...)
			case ir.CallClosure:
				operands = append(operands, ir.ClosureOperands(call)...)
			default:
				return nil, l.notYet(node, "a never expression without a callable representation")
			}
		}
	}
	if node.Kind == ast.KindIdentifier && (node.Parent == nil || node.Parent.Kind != ast.KindVariableDeclaration) {
		if local, known := l.local(node); known {
			of = l.result.Locals[local].Type
		}
	}

	condition := node
	for condition.Parent != nil && condition.Parent.Kind == ast.KindParenthesizedExpression {
		condition = condition.Parent
	}
	if parent := condition.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindIfStatement:
			if parent.AsIfStatement().Expression == condition {
				of = ir.Boolean
			}
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == condition {
				of = ir.Boolean
			}
		case ast.KindWhileStatement:
			if parent.AsWhileStatement().Expression == condition {
				of = ir.Boolean
			}
		case ast.KindDoStatement:
			if parent.AsDoStatement().Expression == condition {
				of = ir.Boolean
			}
		case ast.KindForStatement:
			if parent.AsForStatement().Condition == condition {
				of = ir.Boolean
			}
		}
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		contextual = l.impliedTarget(node)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindAsExpression {
		contextual = l.checker.GetTypeAtLocation(node.Parent)
	}
	if contextual != nil {
		if reducedOptionalSource(l.checker, l.concrete(contextual)).Flags()&checker.TypeFlagsNever == 0 {
			if representation, known := l.representation(contextual); known {
				of = representation
			}
		}
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindReturnStatement && l.function != nil && l.function.Returns != 0 {
		of = l.function.Returns
	}
	function := len(l.result.Functions)
	parameters := []int{}
	reads := []ir.Expression{}
	for _, operand := range operands {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "never_operand", Type: operand.Type(), Function: function})
		parameters = append(parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: operand.Type()})
	}
	body := []ir.Statement{}
	if effect != nil {
		switch call := effect.(type) {
		case ir.Call:
			call.Arguments = reads
			effect = call
		case ir.CallClosure:
			call.Closure, call.Arguments = reads[0], reads[1:]
			effect = call
		}
		body = append(body, ir.Evaluate{Value: effect})
	}
	file := ast.GetSourceFileOfNode(node)
	expression := strings.TrimSpace(file.Text()[node.Pos():node.End()])
	where := filepath.Base(l.program.Where(node))
	message := "unreachable expression " + expression + " at " + where
	body = append(body, ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "never_check", Parameters: parameters, Returns: of, Body: body})
	return ir.Call{Function: function, Arguments: operands, Returns: of}, nil
}
