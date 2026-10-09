package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Integrity queries do not read fields. Mutations require a complete plain shape or a
// constructor proof distinguishing actual own descriptors from native internal slots.
func (l *lowering) objectIntegrityCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return l.objectIntegrityCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) objectIntegrityCallArguments(node *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {
	switch name {
	case "freeze", "seal", "preventExtensions", "isSealed", "isExtensible", "isFrozen":
	default:
		return nil, false, nil
	}
	if len(written) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	argument := written[0]
	proven := l.checker.GetTypeAtLocation(argument)
	mutation := name == "freeze" || name == "seal" || name == "preventExtensions"
	if mutation && proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
		value, err := l.expression(argument)
		return value, true, err
	}
	shape := l.objectIntegrityShape(argument, 0)
	if shape == 2 && name == "freeze" {
		return nil, true, l.notYet(argument, "Object.freeze on RegExp (catchable lastIndex write failures are not represented)")
	}
	if mutation && !l.exactObject(argument, 0) && shape == 0 {
		return nil, true, l.notYet(argument, "Object."+name+" on a shape not proven by a plain literal or its const binding")
	}
	// A tuple has an array length descriptor absent from its native object shape. A structural
	// object view may also hide a primitive or array, requiring tagged view dispatch first.
	if checker.IsTupleType(proven) {
		return nil, true, l.notYet(argument, "Object."+name+" on a tuple (its array descriptors are not represented)")
	}
	value, err := l.objectIntegrityValue(argument, node)
	if err != nil {
		return nil, true, err
	}
	if value.Type() == ir.Object && shape == 0 && !l.exactObject(argument, 0) {
		if hazard := l.prototypeHazard(argument, "valueOf"); hazard != "" {
			return nil, true, l.notYet(argument, "Object."+name+" through an object view ("+hazard+")")
		}
	}
	call := ir.ObjectCall{Method: name, IntegrityShape: shape, Returns: ir.Boolean}
	if mutation {
		if (value.Type() != ir.Object && value.Type() != ir.Map) || l.includesUndefined(proven) {
			return nil, true, l.notYet(argument, "Object."+name+" on other than a present plain object")
		}
		call.Returns = value.Type()
		call.Arguments = []ir.Expression{value}
	} else {
		call.Arguments = []ir.Expression{fit(value, ir.Union)}
	}
	return call, true, nil
}

func (l *lowering) objectTypeIncludesNull(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsNull != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if l.objectTypeIncludesNull(part) {
				return true
			}
		}
	}
	return false
}
