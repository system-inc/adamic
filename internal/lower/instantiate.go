package lower

import (
	"reflect"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The cycle finder judges every instantiation of a generic class by its own types (cycles.go). Inside
// Maker<Node>, a new Box<T>() is a Box<Node>, though the source only ever writes Box<T>; the checker
// knows how to make that type, but its shim doesn't export the way. These two pull it from the
// checker's own package, as the shim's generated code does for what it exports. They belong in
// cohere's shim (extra-shim.json); until they're there, this is where they are.

// typeMapper belongs to the checker. Its identity must survive lowering snapshots;
// an opaque empty struct would look like lowering-owned state to a typed copier.
type typeMapper = checker.TypeMapper

//go:linkname newTypeMapper github.com/microsoft/TypeScript/tsc/internal/checker.newTypeMapper
func newTypeMapper(sources []*checker.Type, targets []*checker.Type) *typeMapper

//go:linkname instantiateType github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).instantiateType
func instantiateType(receiver *checker.Checker, proven *checker.Type, mapper *typeMapper) *checker.Type

// typeMapperOf is what a class's type parameters stand for in a type of it, as a mapper the checker
// instantiates with, or the one in force when the class has none.
func (l *lowering) typeMapperOf(declaration *ast.Node, classType *checker.Type) *typeMapper {
	parameters := declaration.TypeParameters()
	arguments := l.checker.GetTypeArguments(classType)
	if len(parameters) == 0 || len(parameters) != len(arguments) {
		return l.typeMapper
	}
	sources := []*checker.Type{}
	for _, parameter := range parameters {
		sources = append(sources, l.checker.GetTypeAtLocation(parameter.Name()))
	}
	return newTypeMapper(sources, arguments)
}

// closureNodeOf is where a function value is written.
func (l *lowering) closureNodeOf(function int) *ast.Node {
	for _, closure := range l.closureRecords {
		if closure.function == function {
			return closure.node
		}
	}
	return nil
}

// concrete is a type as the instantiation being lowered makes it: its type parameters, and those of
// every instantiation it's inside, replaced by what they stand for.
func (l *lowering) concrete(proven *checker.Type) *checker.Type {
	if l.typeMapper == nil || proven == nil {
		return proven
	}
	return instantiateType(l.checker, proven, l.typeMapper)
}

// classNodeFor is where a class type's class is declared, to point at.
func (l *lowering) classNodeFor(proven *checker.Type) *ast.Node {
	if symbol := proven.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		return symbol.Declarations[0]
	}
	return nil
}

// resolvedTypeMapper reads the checker-owned mapper by field name, without
// duplicating the checker's private Signature layout. The shim exposes the type
// but not this field; a changed shim shape must refuse instead of guessing.
func resolvedTypeMapper(signature *checker.Signature) *checker.TypeMapper {
	if signature == nil {
		return nil
	}
	field := reflect.ValueOf(signature).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.Type() != reflect.TypeOf((*checker.TypeMapper)(nil)) || field.IsNil() {
		return nil
	}
	return (*checker.TypeMapper)(unsafe.Pointer(field.UnsafePointer()))
}
