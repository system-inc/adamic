package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A finite string key selects an ordinary data field. All candidate fields
// must have the same checker type and representation; no synthetic context
// or dynamic index signature is needed. Capture receiver and key once.
func (l *lowering) finitePropertyIndex(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	receiver := l.concrete(l.checker.GetTypeAtLocation(access.Expression))
	of, known := l.representation(receiver)
	if !known || of != ir.Object || checker.IsTupleType(receiver) {
		return nil, false, nil
	}
	keyType := l.concrete(l.checker.GetTypeAtLocation(access.ArgumentExpression))
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	names := []string{}
	var common *checker.Type
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsStringLiteral == 0 {
			return nil, false, nil
		}
		name, literal := member.AsLiteralType().Value().(string)
		if !literal {
			return nil, false, nil
		}
		field := l.checker.GetPropertyOfType(receiver, name)
		if l.inheritedLibrarySymbol(field) {
			return nil, true, l.prototypeRead(node, name)
		}
		if field == nil || field.Flags&(ast.SymbolFlagsMethod|ast.SymbolFlagsOptional) != 0 || accessorSymbol(field) || isClassInstance(receiver) {
			return nil, true, l.notYet(node, "a finite computed read outside ordinary required data fields")
		}
		if err := l.erasedMethodField(node, receiver, name); err != nil {
			return nil, true, err
		}
		declared := l.concrete(l.checker.GetTypeOfSymbol(field))
		if common == nil {
			common = declared
		} else if !checker.Checker_isTypeIdenticalTo(l.checker, common, declared) {
			return nil, true, l.notYet(node, "a finite computed read whose fields have different stored types")
		}
		names = append(names, name)
	}
	if access.QuestionDotToken != nil || l.includesUndefined(receiver) || l.includesNull(receiver) || len(names) == 0 {
		return nil, true, l.notYet(node, "a finite computed read on a possibly absent receiver")
	}
	result, known := l.representation(common)
	if !known || slotless(result) {
		return nil, true, l.notYet(node, "a finite computed field without a supported stored representation")
	}
	observed, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	if observed != result {
		return nil, true, l.notYet(node, "a finite computed read narrowed to another representation")
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	function, reads := l.stringHelper("finite_property_index", []ir.Expression{object, key})
	l.result.Functions[function].Returns = result
	readObject, readKey := reads[0], reads[1]
	for _, name := range names {
		l.result.Functions[function].Body = append(l.result.Functions[function].Body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant(name)}}, Then: []ir.Statement{ir.Return{Value: ir.Property{Object: readObject, Name: name, Of: result}}}})
	}
	l.result.Functions[function].Body = append(l.result.Functions[function].Body, ir.Panic{Message: ir.StringConstant{Index: l.constant("computed field key outside its proven union")}})
	return ir.Call{Function: function, Arguments: []ir.Expression{object, key}, Returns: result}, true, nil
}
