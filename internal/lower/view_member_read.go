package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// readViewMember is the checked own-member boundary. Syntax supplies its field
// and receiver certificates before a read can expose the stored value.
func (l *lowering) readViewMember(node *ast.Node, property ir.Property, field *ast.Symbol, receiver *checker.Type) ir.Expression {
	property.Readiness = sourceExpression(node)
	if property.View == "" {
		property.View = sourceExpression(node)
	}
	property.ViewWhere = l.program.Where(node)
	if receiver != nil {
		property.ViewReceiverTypeID = l.viewReceiverTypeID(receiver)
	}
	if field != nil {
		declared := l.checker.GetTypeOfSymbol(field)
		property.ViewTypeID = int(declared.Id())
		property.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]
		property.ViewType = l.checker.TypeToString(declared)
		property.ViewAllowed = l.viewLiterals(declared)
	}
	if field != nil {
		for _, declaration := range field.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration {
				initializer := declaration.AsPropertyDeclaration().Initializer
				if l.lazyAssertionInitializer(initializer) && !l.uninitializedInitializer(initializer) {
					property.Readiness = sourceExpression(initializer)
				}
			}
		}
	}
	if field == nil || field.Flags&ast.SymbolFlagsOptional == 0 {
		return property
	}
	property.Absent = true
	if declared, _ := l.representation(l.checker.GetTypeOfSymbol(field)); declared.IsMaybe() && property.Of == declared.Present() {
		property.Of = declared
		return fit(property, declared.Present())
	}
	return property
}

// Nullable receivers read the present object's contract. A generic receiver
// uses its instantiated type, or its constraint when the binder remains rigid.
// Neither operation relates unrelated objects merely because their names match.
func (l *lowering) viewReceiverTypeID(receiver *checker.Type) int {
	receiver = l.concrete(receiver)
	if receiver.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(receiver); constraint != nil {
			receiver = l.concrete(constraint)
		}
	}
	return int(l.checker.GetNonNullableType(receiver).Id())
}

// A union receiver can carry a checked member's allocation. Preserve that
// boundary for its common field without marking unrelated receiver types.
func (l *lowering) checkedViewReceiverField(receiver *checker.Type, name string) bool {
	id := l.viewReceiverTypeID(receiver)
	if l.result.CheckedFields[checkedViewFieldKey(id, name)] {
		return true
	}
	receiver = l.checker.GetNonNullableType(l.concrete(receiver))
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range receiver.Types() {
			if l.checkedViewReceiverField(member, name) {
				l.result.CheckedFields[checkedViewFieldKey(id, name)] = true
				return true
			}
		}
	}
	return false
}

// checkedViewMembers holds the source once and checks declared own slots before
// exposing their presence or copying values. Copies retain hidden fields and tags.
func (l *lowering) checkedViewMembers(node, source *ast.Node, value ir.Expression, member *ast.Node) (ir.Expression, error) {
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source))
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	checks := []ir.Statement{}
	for _, field := range l.checker.GetPropertiesOfType(receiver) {
		if member != nil && field.Name != member.Text() {
			continue
		}
		if accessorSymbol(field) {
			continue
		}
		prototype := false
		staticStorage := false
		for _, declaration := range field.Declarations {
			staticStorage = staticStorage || declaration.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(declaration, ast.ModifierFlagsStatic)
			if declaration.Kind == ast.KindMethodDeclaration || declaration.Kind == ast.KindMethodSignature || declaration.Name() != nil && declaration.Name().Kind == ast.KindPrivateIdentifier {
				prototype = true
			}
		}
		// Constructor built-ins such as prototype are checker properties, not
		// represented own data slots. Declared static fields keep parent lookup.
		if prototype || l.isStaticType(receiver) && !staticStorage {
			continue
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || of == ir.Weak || censusFieldSlotless(of) {
			return nil, l.notYet(node, "a view member operation with an unsupported field representation")
		}
		read := l.readViewMember(node, ir.Property{Object: held, Name: field.Name, Of: of, Absent: field.Flags&ast.SymbolFlagsOptional != 0, View: sourceExpression(node) + " (field " + field.Name + ")"}, field, receiver)
		checks = append(checks, ir.Evaluate{Value: read})
	}
	if len(checks) == 0 {
		return value, nil
	}
	if l.includesUndefined(l.checker.GetTypeAtLocation(source)) {
		b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: held}}, Then: checks})
	} else {
		b.body = append(b.body, checks...)
	}
	return b.finish("checked_view_members", held), nil
}

// Wait until readiness has resolved every receiver certificate, including views
// created after a function declaration. A helper containing only ordinary slot
// reads and an identity return no longer enforces a boundary. Removing its call
// exposes the original operand to ownership and spread reuse without evaluating
// that operand twice. Helpers with a view or readiness check remain intact.
func removeUncheckedMemberHelpers(program *ir.Program) {
	removable := map[int]bool{}
	for index, function := range program.Functions {
		if function.Name != "checked_view_members" || len(function.Parameters) != 1 || len(function.Body) == 0 {
			continue
		}
		parameter := function.Parameters[0]
		result, ok := function.Body[len(function.Body)-1].(ir.Return)
		if !ok {
			continue
		}
		read, ok := result.Value.(ir.Read)
		if !ok || read.Local != parameter || read.Readiness != "" {
			continue
		}
		var plain func([]ir.Statement) bool
		plain = func(body []ir.Statement) bool {
			for _, statement := range body {
				switch statement := statement.(type) {
				case ir.Evaluate:
					value := statement.Value
					for {
						unwrap, ok := value.(ir.Unwrap)
						if !ok {
							break
						}
						value = unwrap.Value
					}
					property, ok := value.(ir.Property)
					if !ok || property.View != "" || property.Readiness != "" || property.Method {
						return false
					}
					object, ok := property.Object.(ir.Read)
					if !ok || object.Local != parameter || object.Readiness != "" {
						return false
					}
				case ir.If:
					condition, ok := statement.Condition.(ir.Unary)
					if !ok || condition.Operator != ir.Not || len(statement.Else) != 0 {
						return false
					}
					undefined, ok := condition.Operand.(ir.IsUndefined)
					if !ok {
						return false
					}
					object, ok := undefined.Value.(ir.Read)
					if !ok || object.Local != parameter || object.Readiness != "" || !plain(statement.Then) {
						return false
					}
				default:
					return false
				}
			}
			return true
		}
		removable[index] = plain(function.Body[:len(function.Body)-1])
	}
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			node := mapped.Interface()
			if call, ok := node.(ir.Call); ok && call.Virtual == 0 && len(call.Arguments) == 1 && !program.CallExpandsArguments(call) && call.Returns == call.Arguments[0].Type() {
				targets := program.CallTargets(call)
				if len(targets) == 1 && removable[targets[0]] {
					node = call.Arguments[0]
				}
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(node))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		}
		return value
	}
	program.Main = transform(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = transform(reflect.ValueOf(program.Functions[index].Body)).Interface().([]ir.Statement)
	}
}
