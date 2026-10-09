package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Builtin Error contributes the same owned prefix as a declared base class.
// Its initializer writes the existing object, preserving the derived identity.
func (l *lowering) errorBase(name string) *instance {
	key := "$builtin:" + name
	if existing := l.instances[key]; existing != nil {
		return existing
	}
	if l.instances == nil {
		l.instances = map[string]*instance{}
	}
	base := &instance{class: len(l.result.Classes) + 1, constructor: len(l.result.Functions), initializer: len(l.result.Functions) + 1, methods: map[string]int{}, slots: map[string]int{}, staticMethods: map[string]bool{}, hasDescendants: true}
	empty := ir.StringConstant{Index: l.constant("")}
	fields := []ir.Field{{Name: "name", Value: ir.StringConstant{Index: l.constant(name)}}, {Name: "message", Value: empty}}
	l.result.Classes = append(l.result.Classes, ir.Class{Name: "builtin_" + name, BuiltinError: name, Constructor: base.constructor, Fields: fields, OwnStart: 2})
	l.instances[key] = base
	object := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: base.constructor})
	argument := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "message", Type: ir.String, Function: base.constructor})
	receiver := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: base.initializer})
	message := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "message", Type: ir.String, Function: base.initializer})
	read := func(local int, of ir.Type) ir.Expression { return ir.Read{Local: local, Of: of} }
	l.result.Functions = append(l.result.Functions,
		ir.Function{Name: "builtin_" + name + "_new", Parameters: []int{argument}, Returns: ir.Object, Body: []ir.Statement{
			ir.Declare{Local: object, Value: ir.ObjectLiteral{Fields: fields, Class: base.class}},
			ir.Evaluate{Value: ir.Call{Function: base.initializer, Arguments: []ir.Expression{read(object, ir.Object), read(argument, ir.String)}}},
			ir.Return{Value: read(object, ir.Object)},
		}},
		ir.Function{Name: "builtin_" + name + "_initialize", Parameters: []int{receiver, message}, Receiver: true, Body: []ir.Statement{
			ir.SetProperty{Object: read(receiver, ir.Object), Name: "name", Value: ir.StringConstant{Index: l.constant(name)}},
			ir.SetProperty{Object: read(receiver, ir.Object), Name: "message", Value: ir.Coalesce{Value: read(message, ir.String), Fallback: empty, Of: ir.String}},
		}})
	return base
}

func (l *lowering) errorAncestry(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	if l.isLibraryType(proven, "Error", "RangeError", "TypeError", "ReferenceError") {
		return true
	}
	if !isClassInstance(proven) {
		return false
	}
	for _, base := range l.classBases(proven) {
		if l.errorAncestry(base) {
			return true
		}
	}
	return false
}
