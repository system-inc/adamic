package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Family 17, Adamic's ruling of Nov 22 (#1atvfp5): Object.keys, values and entries, and encodeJson,
// stop as "Object enumeration with getter or accessor descriptors" only for a get or set accessor
// whose object can reach the enumerated operand by the type graph, no longer for any accessor anywhere
// in the program. The premise: a value at a typed slot is something assignable to that slot's type, so
// an accessor whose owner isn't assignable to any type the enumeration reads can't be on the operand,
// nor on anything encodeJson reads under it. An accessor that does reach stays refused (rulings rows
// 306 and 309 stand), and every doubt reaches: an operand whose types can't all be classified keeps
// the program-wide scan of before, and so does an accessor whose owner can't be named.

// enumerationReachDepth bounds encodeJson's walk under its operand; a type past it makes reach unknown.
const enumerationReachDepth = 16

// enumerationAccessorReach is one enumeration's question to every accessor in the program: the operand,
// whether the enumeration reads under it (encodeJson does, at any depth; Object.keys, values and entries
// read the operand's own properties), and what's worked out from them on the first accessor met.
type enumerationAccessorReach struct {
	operand *ast.Node
	deep    bool
	modules []*ast.SourceFile
	// read is set once the types below are collected.
	read    bool
	unknown bool
	types   []*checker.Type
	// classes are the program's class declarations and expressions, collected on the first class
	// accessor met.
	classes []*ast.Node
}

// enumerationAccessorReaches reports whether a get or set accessor can be on a value the enumeration
// reads: when what it reads is unknown, when the accessor's owner can't be named, or when an owner is
// assignable (the checker's IsTypeAssignableTo) to a type it reads. Subclasses and interface views come
// out of the assignability: a class whose getter implements interface I reaches an operand typed I.
func (l *lowering) enumerationAccessorReaches(accessor *ast.Node, reach *enumerationAccessorReach) bool {
	if !reach.read {
		reach.read = true
		reach.unknown, reach.types = l.enumerationReadTypes(reach.operand, reach.deep)
	}
	if reach.unknown {
		return true
	}
	owners, known := l.enumerationAccessorOwners(accessor, reach)
	if !known {
		return true
	}
	// Related either way: an owner assignable to a read type is an upcast, and a read type assignable to
	// the owner is a checked downcast in .ts, which keeps the same object, getter and all, at a slot its
	// owner isn't assignable to ({ a, get g } seen as { a, g, c }).
	for _, owner := range owners {
		for _, read := range reach.types {
			if l.checker.IsTypeAssignableTo(owner, read) || l.checker.IsTypeAssignableTo(read, owner) {
				return true
			}
		}
	}
	return false
}

// enumerationReadTypes is every object type a value the enumeration reads can sit at, or unknown. The
// operand's own type and its union members are read by every enumeration; for a deep one, so are,
// recursively, each one's property types, type arguments and array and tuple element types. Any,
// unknown, object, {}, a type parameter or another instantiable type, an index signature whose values
// can be objects, a type nested past enumerationReachDepth, and anything else that isn't an object
// type, a union or intersection of them, or a primitive, make it unknown.
func (l *lowering) enumerationReadTypes(operand *ast.Node, deep bool) (bool, []*checker.Type) {
	if operand == nil {
		return true, nil
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(operand))
	// A type that mentions a type parameter anywhere it holds relates by assignability for no
	// instantiation in particular, so what can reach it can't be told.
	if l.eachReached(proven, 0, map[*checker.Type]bool{}, func(reached *checker.Type) bool {
		return reached.Flags()&checker.TypeFlagsInstantiable != 0
	}) {
		return true, nil
	}
	type step struct {
		proven *checker.Type
		depth  int
	}
	var types []*checker.Type
	seen := map[*checker.Type]bool{}
	queue := []step{{proven: proven}}
	for len(queue) != 0 {
		next := queue[0]
		queue = queue[1:]
		if next.proven == nil || seen[next.proven] {
			continue
		}
		seen[next.proven] = true
		if next.depth > enumerationReachDepth {
			return true, nil
		}
		flags := next.proven.Flags()
		switch {
		case flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive|checker.TypeFlagsInstantiable) != 0:
			return true, nil
		case flags&checker.TypeFlagsUnion != 0:
			// A union member is the type the value is, so it's read at the union's own depth.
			for _, member := range next.proven.Types() {
				queue = append(queue, step{proven: member, depth: next.depth})
			}
			continue
		case flags&checker.TypeFlagsIntersection != 0:
			// The value is all its members at once: the intersection is read, and so is each member,
			// which only ever lets more owners reach.
			types = append(types, next.proven)
			for _, member := range next.proven.Types() {
				queue = append(queue, step{proven: member, depth: next.depth})
			}
			continue
		case flags&(checker.TypeFlagsPrimitive|checker.TypeFlagsNever) != 0:
			// A primitive holds no accessor.
			continue
		case flags&checker.TypeFlagsObject == 0:
			return true, nil
		}
		properties := l.checker.GetPropertiesOfType(next.proven)
		indexes := l.checker.GetIndexInfosOfType(next.proven)
		// {} and an empty interface are any object at all.
		if len(properties) == 0 && len(indexes) == 0 && len(l.checker.GetSignaturesOfType(next.proven, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(next.proven, checker.SignatureKindConstruct)) == 0 {
			return true, nil
		}
		for _, index := range indexes {
			if !l.enumerationPrimitive(index.ValueType()) {
				return true, nil
			}
		}
		types = append(types, next.proven)
		if !deep {
			continue
		}
		for _, argument := range l.typeArguments(next.proven) {
			queue = append(queue, step{proven: argument, depth: next.depth + 1})
		}
		if l.isArrayType(next.proven) || checker.IsTupleType(next.proven) {
			continue
		}
		for _, property := range properties {
			queue = append(queue, step{proven: l.checker.GetTypeOfSymbol(property), depth: next.depth + 1})
		}
	}
	return false, types
}

// enumerationPrimitive reports whether every value of a type is a primitive, which holds no accessor.
func (l *lowering) enumerationPrimitive(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		return !slices.ContainsFunc(proven.Types(), func(member *checker.Type) bool { return !l.enumerationPrimitive(member) })
	}
	return proven.Flags()&(checker.TypeFlagsPrimitive|checker.TypeFlagsNever) != 0
}

// enumerationAccessorOwners is every type an object carrying the accessor is, or not known. An object
// literal's accessor is on that literal, of its regular type. An interface's or type literal's accessor
// declares no getter at run time, but its declared type stands as an owner, failing closed. A class's
// accessor is on every instance of the class and of each class extending it, or on the class itself and
// each class extending it for a static one: every class in the program whose own side's property by the
// accessor's name is declared by it. Inside type parameters an owner relates for no instantiation in
// particular, and an accessor held anywhere else has no owner to name: both are not known.
func (l *lowering) enumerationAccessorOwners(accessor *ast.Node, reach *enumerationAccessorReach) ([]*checker.Type, bool) {
	holder := accessor.Parent
	if holder == nil || enumerationGenericContext(holder) {
		return nil, false
	}
	switch holder.Kind {
	case ast.KindObjectLiteralExpression:
		return []*checker.Type{regularTypeOfObjectLiteral(l.checker, l.checker.GetTypeAtLocation(holder))}, true
	case ast.KindInterfaceDeclaration:
		symbol := l.checker.GetSymbolOfDeclaration(holder)
		if symbol == nil {
			return nil, false
		}
		return []*checker.Type{l.checker.GetDeclaredTypeOfSymbol(symbol)}, true
	case ast.KindTypeLiteral:
		return []*checker.Type{l.checker.GetTypeAtLocation(holder)}, true
	case ast.KindClassDeclaration, ast.KindClassExpression:
	default:
		return nil, false
	}
	symbol := l.checker.GetSymbolOfDeclaration(accessor)
	if symbol == nil {
		return nil, false
	}
	if reach.classes == nil {
		reach.classes = []*ast.Node{}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
				reach.classes = append(reach.classes, node)
			}
			return node.ForEachChild(visit)
		}
		for _, module := range reach.modules {
			module.AsNode().ForEachChild(visit)
		}
	}
	static := ast.HasSyntacticModifier(accessor, ast.ModifierFlagsStatic)
	var owners []*checker.Type
	own := false
	for _, class := range reach.classes {
		classSymbol := l.checker.GetSymbolOfDeclaration(class)
		if classSymbol == nil {
			continue
		}
		side := l.checker.GetDeclaredTypeOfSymbol(classSymbol)
		if static {
			side = l.checker.GetTypeOfSymbol(classSymbol)
		}
		property := l.checker.GetPropertyOfType(side, symbol.Name)
		if property == nil || !slices.Contains(property.Declarations, accessor) {
			continue
		}
		if enumerationGenericContext(class) {
			return nil, false
		}
		owners = append(owners, side)
		own = own || class == holder
	}
	// The accessor's own class must be found among them, or its name doesn't lead back to it (a private or
	// computed name), and its owners aren't known.
	return owners, own
}

// enumerationGenericContext reports whether a declaration, or anything it's inside, has type parameters.
func enumerationGenericContext(node *ast.Node) bool {
	for around := node; around != nil; around = around.Parent {
		if ast.IsFunctionLike(around) && around.FunctionLikeData() != nil || around.Kind == ast.KindClassDeclaration || around.Kind == ast.KindClassExpression || around.Kind == ast.KindInterfaceDeclaration || around.Kind == ast.KindTypeAliasDeclaration {
			if len(around.TypeParameters()) != 0 {
				return true
			}
		}
	}
	return false
}
