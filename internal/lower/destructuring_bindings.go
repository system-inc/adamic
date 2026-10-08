package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// destructureArray uses the built-in array iterator's live length, with sticky
// exhaustion. A default can mutate the source between steps, but cannot revive
// an iterator which has already returned done. Rest copies what remains.
func (l *lowering) destructureArray(pattern *ast.Node, source *checker.Type, held int) ([]ir.Statement, error) {
	elementType := l.checker.GetElementTypeOfArrayType(source)
	if elementType == nil && l.isLibraryType(source, "RegExpExecArray", "RegExpMatchArray") {
		elementType = l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(source, "0"))
	}
	if elementType == nil {
		return nil, l.notYet(pattern, "destructuring an array without an element type")
	}
	element, known := l.representation(elementType)
	if !known || slotless(element) {
		return nil, l.notYet(pattern, "destructuring an array with an unrepresented element")
	}
	array := ir.Read{Local: held, Of: ir.Array}
	done := l.iterationLocal("destructuring_done", ir.Boolean, l.functionIndex)
	statements := []ir.Statement{ir.Declare{Local: done, Value: ir.BooleanConstant{Value: false}}}
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		position := ir.NumberConstant{Value: float64(index)}
		if binding.Kind != ast.KindOmittedExpression && binding.Name() != nil && binding.AsBindingElement().DotDotDotToken != nil {
			if !ast.IsIdentifier(binding.Name()) {
				return nil, l.notYet(binding, "a nested array rest binding")
			}
			restElement, err := l.elementType(binding.Name())
			if err != nil {
				return nil, err
			}
			if restElement != element {
				return nil, l.notYet(binding, "an array rest binding held otherwise than its source")
			}
			rest := ir.Conditional{Condition: ir.Read{Local: done, Of: ir.Boolean}, WhenTrue: ir.ArrayLiteral{Element: element}, WhenNot: ir.ArraySlice{Array: array, Arguments: []ir.Expression{position}}, Of: ir.Array}
			declared, err := l.declareDestructured(binding, l.checker.GetTypeAtLocation(binding.Name()), rest)
			if err != nil {
				return nil, err
			}
			statements = append(statements, declared...)
			continue
		}
		statements = append(statements, ir.Assign{Local: done, Value: ir.Binary{Operator: ir.Or, Left: ir.Read{Local: done, Of: ir.Boolean}, Right: ir.Binary{Operator: ir.GreaterOrEqual, Left: position, Right: ir.Length{Array: array}}}})
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		value := ir.Conditional{Condition: ir.Read{Local: done, Of: ir.Boolean}, WhenTrue: fit(ir.Undefined{}, ir.Maybe(element)), WhenNot: ir.ArrayIndex{Array: array, Index: position, Element: element}, Of: ir.Maybe(element)}
		declared, err := l.declareDestructured(binding, elementType, value)
		if err != nil {
			return nil, err
		}
		statements = append(statements, declared...)
	}
	return statements, nil
}

// declareDestructured evaluates a field once, then a lazy default only for
// undefined, then binds the name or recursively reads a nested pattern.
func (l *lowering) declareDestructured(binding *ast.Node, fieldType *checker.Type, value ir.Expression) ([]ir.Statement, error) {
	declared := binding.AsBindingElement()
	statements := []ir.Statement{}
	if declared.Initializer != nil {
		// Reference slots collapse null and undefined. Do not guess which was read.
		if l.includesNull(fieldType) && value.Type() != ir.Union {
			return nil, l.notYet(binding, "a destructuring default whose slot cannot distinguish null from undefined")
		}
		raw := l.iterationLocal("destructuring_field", value.Type(), l.functionIndex)
		statements = append(statements, ir.Declare{Local: raw, Value: value})
		read := ir.Read{Local: raw, Of: value.Type()}
		fallback, err := l.expression(declared.Initializer)
		if err != nil {
			return nil, err
		}
		of, known := l.representation(l.checker.GetTypeAtLocation(binding.Name()))
		if !known {
			return nil, l.notYet(binding, "a destructuring default without a representation")
		}
		present, fallback := fit(read, of), fit(fallback, of)
		if present.Type() != of || fallback.Type() != of {
			return nil, l.notYet(binding, "a destructuring default held otherwise than its name")
		}
		test := ir.Expression(ir.IsUndefined{Value: read})
		if read.Type() == ir.Number || read.Type() == ir.Boolean {
			test = ir.BooleanConstant{Value: false}
		}
		if read.Type() == ir.Union {
			test = ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: read}, Right: ir.StringConstant{Index: l.constant("undefined")}}
		}
		value = ir.Conditional{Condition: test, WhenTrue: fallback, WhenNot: present, Of: of}
		fieldType = l.checker.GetTypeAtLocation(binding.Name())
	}
	if binding.Name().Kind == ast.KindObjectBindingPattern || binding.Name().Kind == ast.KindArrayBindingPattern {
		held := l.iterationLocal("nested_destructured", value.Type(), l.functionIndex)
		statements = append(statements, ir.Declare{Local: held, Value: value})
		nested, err := l.destructureFrom(binding.Name(), fieldType, value.Type(), held)
		return append(statements, nested...), err
	}
	if !ast.IsIdentifier(binding.Name()) {
		return nil, l.notYet(binding, "a destructured name that isn't plain")
	}
	local, err := l.declareLocal(binding.Name())
	if err != nil {
		return nil, err
	}
	of := l.result.Locals[local].Type
	value = fit(value, of)
	if value.Type() != of {
		return nil, l.notYet(binding, "a destructured name held otherwise than its field")
	}
	return append(statements, ir.Declare{Local: local, Value: value}), nil
}
