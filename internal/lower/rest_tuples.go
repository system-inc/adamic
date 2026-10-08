package lower

import (
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// fixedRestTuples requires distinct, fixed lengths, so arity selects one storage shape.
func (l *lowering) fixedRestTuples(proven *checker.Type) ([]*checker.Type, bool) {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	seen := map[int]bool{}
	for _, member := range members {
		if !checker.IsTupleType(member) {
			return nil, false
		}
		elements := l.checker.GetTypeArguments(member)
		length := l.checker.GetPropertyOfType(member, "length")
		if length == nil {
			return nil, false
		}
		literal := l.checker.GetTypeOfSymbol(length)
		if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 || reflect.ValueOf(literal.AsLiteralType().Value()).Float() != float64(len(elements)) || seen[len(elements)] {
			return nil, false
		}
		seen[len(elements)] = true
		for _, element := range elements {
			of, known := l.kept(element)
			if !known || slotless(of) {
				return nil, false
			}
		}
	}
	return members, true
}

// Tuple rest is represented by numbered fields and a length field. Keep the whole
// tuple local: only reads and fixed-call forwarding are admitted, so array identity,
// mutation and array methods cannot expose this storage convention.
func (l *lowering) proveRestTupleUses(declaration, parameter *ast.Node) error {
	symbol := l.symbol(parameter.Name())
	var stopped error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if stopped != nil {
			return true
		}
		if ast.IsIdentifier(node) && l.symbol(node) == symbol {
			parent := node.Parent
			allowed := false
			if parent != nil {
				switch parent.Kind {
				case ast.KindPropertyAccessExpression:
					access := parent.AsPropertyAccessExpression()
					allowed = access.Expression == node && access.Name().Text() == "length" && !ast.IsAssignmentTarget(parent)
				case ast.KindElementAccessExpression:
					access := parent.AsElementAccessExpression()
					allowed = access.Expression == node && access.ArgumentExpression.Kind == ast.KindNumericLiteral && !ast.IsAssignmentTarget(parent)
				case ast.KindSpreadElement:
					call := parent.Parent
					allowed = call != nil && call.Kind == ast.KindCallExpression && len(call.AsCallExpression().Arguments.Nodes) == 1
				}
			}
			if !allowed {
				stopped = l.notYet(node, "a tuple rest escaping read-only fixed-call storage")
				return true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	declaration.Body().ForEachChild(visit)
	return stopped
}

func (l *lowering) packRestTuple(call *ast.CallExpression, parameters []*ast.Symbol, proven *checker.Type) ([]ir.Expression, error) {
	members, known := l.fixedRestTuples(proven)
	if !known {
		return nil, l.notYet(call.AsNode(), "a tuple rest without distinct fixed storage shapes")
	}
	fixed := len(parameters) - 1
	count := len(call.Arguments.Nodes) - fixed
	var selected *checker.Type
	for _, member := range members {
		if len(l.checker.GetTypeArguments(member)) == count {
			selected = member
		}
	}
	if selected == nil {
		return nil, l.notYet(call.AsNode(), "a tuple rest call without a fixed arity shape")
	}
	elements := l.checker.GetTypeArguments(selected)
	// This object never escapes as an array; the source-use proof admits only reads.
	tuple := ir.ObjectLiteral{}
	arguments := []ir.Expression{}
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(argument, "a spread into tuple rest storage")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		if index < fixed {
			arguments = append(arguments, value)
			continue
		}
		of, _ := l.kept(elements[index-fixed])
		value = fit(value, of)
		if value.Type() != of {
			return nil, l.notYet(argument, "a tuple rest argument with another representation")
		}
		tuple.Fields = append(tuple.Fields, ir.Field{Name: strconv.Itoa(index - fixed), Value: value})
	}
	tuple.Fields = append(tuple.Fields, ir.Field{Name: "length", Value: ir.NumberConstant{Value: float64(count)}})
	return append(arguments, tuple), nil
}

// A sole spread of a plain fixed tuple has no intervening argument effects. Read
// its numbered fields once each; all other tuple spread evaluation stays NotYet.
func (l *lowering) fixedTupleCallArguments(call *ast.CallExpression) ([]ir.Expression, bool, error) {
	if len(call.Arguments.Nodes) != 1 || call.Arguments.Nodes[0].Kind != ast.KindSpreadElement {
		return nil, false, nil
	}
	spread := call.Arguments.Nodes[0]
	node := spread.AsSpreadElement().Expression
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if !checker.IsTupleType(proven) {
		return nil, false, nil
	}
	if !ast.IsIdentifier(node) {
		return nil, true, l.notYet(spread, "a fixed tuple spread other than a plain local")
	}
	if _, local := l.local(node); !local {
		return nil, true, l.notYet(spread, "a fixed tuple spread other than a plain local")
	}
	if _, known := l.fixedRestTuples(proven); !known {
		return nil, true, l.notYet(spread, "a tuple spread without a fixed storage shape")
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Object {
		return nil, true, l.notYet(spread, "a tuple spread with another representation")
	}
	arguments := []ir.Expression{}
	for index, element := range l.checker.GetTypeArguments(proven) {
		of, _ := l.kept(element)
		arguments = append(arguments, ir.Property{Object: value, Name: strconv.Itoa(index), Of: of})
	}
	return arguments, true, nil
}
