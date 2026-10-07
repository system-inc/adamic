package lower

import (
	"math"
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Constant enums become the same object JavaScript constructs, including numeric
// reverse mappings. Computed initializers and merged declarations need separate
// lowering, and are rejected before any initializer effects could disappear.
func (l *lowering) enumDeclaration(node *ast.Node) ([]ir.Statement, error) {
	if l.function != nil || ast.IsEnumConst(node) {
		return nil, l.notYet(node, "a local or const enum")
	}
	symbol := l.symbol(node.Name())
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, l.notYet(node, "a merged enum")
	}
	object := ir.ObjectLiteral{NoReuse: true}
	positions := map[string]int{}
	field := func(name string, value ir.Expression) {
		if index, found := positions[name]; found {
			object.Fields[index].Value = value
			return
		}
		positions[name] = len(object.Fields)
		object.Fields = append(object.Fields, ir.Field{Name: name, Value: value})
	}
	for _, member := range node.AsEnumDeclaration().Members.Nodes {
		name := member.Name().Text()
		if name == "__proto__" {
			return nil, l.notYet(member, "an enum member that changes its prototype")
		}
		constant := l.checker.GetConstantValue(member)
		if text, stringValue := constant.(string); stringValue {
			field(name, ir.StringConstant{Index: l.constant(text)})
			continue
		}
		value := reflect.ValueOf(constant)
		if !value.IsValid() || value.Kind() != reflect.Float64 {
			return nil, l.notYet(member, "an enum member with a computed initializer")
		}
		number := value.Float()
		if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number || math.Abs(number) > 2147483647 {
			return nil, l.notYet(member, "an enum reverse mapping outside small integer keys")
		}
		key := strconv.FormatFloat(number, 'f', 0, 64)
		if number == 0 {
			key = "0"
		}
		field(name, ir.NumberConstant{Value: number})
		field(key, ir.StringConstant{Index: l.constant(name)})
	}
	local, known := l.local(node.Name())
	if !known {
		return nil, l.notYet(node, "an enum without module storage")
	}
	return []ir.Statement{ir.Declare{Local: local, Value: object}}, nil
}

func (l *lowering) enumElement(node *ast.Node, object ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	symbol := l.checker.GetTypeAtLocation(access.Expression).Symbol()
	if symbol == nil || symbol.Flags&ast.SymbolFlagsEnum == 0 {
		return nil, false, nil
	}
	key := ast.SkipParentheses(access.ArgumentExpression)
	if key.Kind != ast.KindNumericLiteral && key.Kind != ast.KindStringLiteral {
		return nil, true, l.notYet(node, "an enum index without a literal key")
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	return ir.Property{Object: object, Name: key.Text(), Of: of, Absent: true}, true, nil
}
