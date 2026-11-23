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

// enumerationReachWhy is a scratch diagnostic (#1atvfp5): why the last accessor reached. Never lands.
var enumerationReachWhy string

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
		enumerationReachWhy = "owner unknown for accessor at " + l.program.Where(accessor)
		return true
	}
	// Related either way: an owner assignable to a read type is an upcast, and a read type assignable to
	// the owner is a checked downcast in .ts, which keeps the same object, getter and all, at a slot its
	// owner isn't assignable to ({ a, get g } seen as { a, g, c }).
	for _, owner := range owners {
		for _, read := range reach.types {
			if l.checker.IsTypeAssignableTo(owner, read) || l.checker.IsTypeAssignableTo(read, owner) {
				enumerationReachWhy = "related: accessor at " + l.program.Where(accessor) + " owner " + l.checker.TypeToString(owner) + " read " + l.checker.TypeToString(read)
				return true
			}
		}
	}
	return false
}

// enumerationReadTypes is every object type a value the enumeration reads can sit at, or unknown. The
// operand's own type, a union both as itself and by its members, is read by every enumeration; for a deep one, so are,
// recursively, each one's property types, type arguments and array and tuple element types. Any,
// unknown, object, {}, a type parameter or another instantiable type, an index signature whose values
// can be objects, a type nested past enumerationReachDepth, and anything else that isn't an object
// type, a union or intersection of them, or a primitive, make it unknown.
func (l *lowering) enumerationReadTypes(operand *ast.Node, deep bool) (bool, []*checker.Type) {
	if operand == nil {
		enumerationReachWhy = "unknown: no operand"
		return true, nil
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(operand))
	// A type that mentions a type parameter anywhere it holds relates by assignability for no
	// instantiation in particular, so what can reach it can't be told.
	if l.eachReached(proven, 0, map[*checker.Type]bool{}, func(reached *checker.Type) bool {
		return reached.Flags()&checker.TypeFlagsInstantiable != 0
	}) {
		enumerationReachWhy = "unknown: instantiable in " + l.checker.TypeToString(proven)
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
			enumerationReachWhy = "unknown: depth at " + l.checker.TypeToString(next.proven)
			return true, nil
		}
		flags := next.proven.Flags()
		switch {
		case flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive|checker.TypeFlagsInstantiable) != 0:
			enumerationReachWhy = "unknown: any, unknown, object or instantiable " + l.checker.TypeToString(next.proven)
			return true, nil
		case flags&checker.TypeFlagsUnion != 0:
			// A union member is the type the value is, so it's read at the union's own depth. The union is
			// read as itself too: tsc relates a source to a discriminated union as a whole, by its
			// discriminant, where the source relates to no one member either way (Lower's witness).
			types = append(types, next.proven) // the union itself
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
			enumerationReachWhy = "unknown: not an object type " + l.checker.TypeToString(next.proven)
			return true, nil
		}
		properties := l.checker.GetPropertiesOfType(next.proven)
		indexes := l.checker.GetIndexInfosOfType(next.proven)
		// {} and an empty interface are any object at all.
		if len(properties) == 0 && len(indexes) == 0 && len(l.checker.GetSignaturesOfType(next.proven, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(next.proven, checker.SignatureKindConstruct)) == 0 {
			enumerationReachWhy = "unknown: empty object type " + l.checker.TypeToString(next.proven)
			return true, nil
		}
		for _, index := range indexes {
			if !l.enumerationPrimitive(index.ValueType()) {
				enumerationReachWhy = "unknown: object-valued index in " + l.checker.TypeToString(next.proven)
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
// literal's accessor is on that literal, of its regular type, known unless that type mentions a type
// parameter: a literal built inside a generic function but typed without one, like tsc's
// createTransformationContext literal inside transformNodes<T>, has one owner for every instantiation
// (J-f3dgc1m found that literal alone kept all twelve of family 17's sites on the program-wide scan).
// An interface's or type literal's accessor
// declares no getter at run time, but its declared type stands as an owner, failing closed. A class's
// accessor is on every instance of the class and of each class extending it, or on the class itself and
// each class extending it for a static one: every class in the program whose own side's property by the
// accessor's name is declared by it. Inside type parameters an owner relates for no instantiation in
// particular, and an accessor held anywhere else has no owner to name: both are not known. A class's
// own type parameters are the exception, taken at any (enumerationAnyInstance).
func (l *lowering) enumerationAccessorOwners(accessor *ast.Node, reach *enumerationAccessorReach) ([]*checker.Type, bool) {
	holder := accessor.Parent
	if holder == nil {
		return nil, false
	}
	if holder.Kind == ast.KindObjectLiteralExpression {
		if declared, named := l.enumerationLiteralDeclaredOwner(holder); named {
			return []*checker.Type{declared}, true
		}
		owner := regularTypeOfObjectLiteral(l.checker, l.checker.GetTypeAtLocation(holder))
		if l.eachReached(owner, 0, map[*checker.Type]bool{}, func(reached *checker.Type) bool {
			return reached.Flags()&checker.TypeFlagsInstantiable != 0
		}) {
			return nil, false
		}
		return []*checker.Type{owner}, true
	}
	class := holder.Kind == ast.KindClassDeclaration || holder.Kind == ast.KindClassExpression
	// A class's own type parameters are taken at any below; type parameters around it are not.
	if class && enumerationGenericContext(holder.Parent) || !class && enumerationGenericContext(holder) {
		return nil, false
	}
	switch holder.Kind {
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
		if enumerationGenericContext(class.Parent) {
			return nil, false
		}
		owners = append(owners, l.enumerationAnyInstance(side, static))
		own = own || class == holder
	}
	// The accessor's own class must be found among them, or its name doesn't lead back to it, and its
	// owners aren't known. A #private name does lead back (J-t8pj70p's probe found Secret owning #hidden),
	// so an enumeration related to that class is refused although a #private getter is never an own
	// property: that only ever refuses more.
	return owners, own
}

// enumerationLiteralDeclaredOwner is the declared type of the variable a literal initializes, when the
// variable's annotation names no type parameter and no type declared inside type parameters: the literal
// is reached only through that variable, so every slot holding it relates to the declared type, and the
// declared type stands as its owner. A literal's own type is no use there: eachReached walks each member
// signature uninstantiated and stops past depth 16, so tsc's createTransformationContext literal, typed
// through NodeFactory's generic methods, read as mentioning a type parameter (J-6b476xq on 0b1ae257). A
// typeof query, a this type or an infer in the annotation, or a name that can't be resolved, leaves it
// unnamed, and the literal's own type is judged as before.
func (l *lowering) enumerationLiteralDeclaredOwner(literal *ast.Node) (*checker.Type, bool) {
	around := literal.Parent
	for around != nil && around.Kind == ast.KindParenthesizedExpression {
		around = around.Parent
	}
	if around == nil || around.Kind != ast.KindVariableDeclaration {
		return nil, false
	}
	declaration := around.AsVariableDeclaration()
	if declaration.Type == nil || declaration.Initializer == nil || ast.SkipParentheses(declaration.Initializer) != literal {
		return nil, false
	}
	open := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindTypeQuery, ast.KindThisType, ast.KindInferType:
			open = true
		case ast.KindTypeReference:
			symbol := l.symbol(node.AsTypeReferenceNode().TypeName)
			if symbol == nil || symbol.Flags&ast.SymbolFlagsTypeParameter != 0 {
				open = true
				break
			}
			for _, named := range symbol.Declarations {
				open = open || enumerationGenericContext(named)
			}
		}
		return open || node.ForEachChild(visit)
	}
	visit(declaration.Type)
	if open {
		return nil, false
	}
	return l.checker.GetTypeAtLocation(around.Name()), true
}

// enumerationAnyInstance is a class's instance side with each of its own type parameters taken at any, so
// it relates to a read type when some instantiation could (any is assignable to and from every type), and
// a static side as it is, since no static member can name the class's type parameters. tsc's
// IdentifierNameMap<V> (transformers/utilities.ts) is the case: its size getter kept family 17's sites
// unknown until the class was read at V = any (J-x8teca6).
func (l *lowering) enumerationAnyInstance(side *checker.Type, static bool) *checker.Type {
	if static || side.ObjectFlags()&checker.ObjectFlagsInterface == 0 || side.AsInterfaceType() == nil {
		return side
	}
	parameters := side.AsInterfaceType().TypeParameters()
	if len(parameters) == 0 {
		return side
	}
	anys := make([]*checker.Type, len(parameters))
	for index := range anys {
		anys[index] = l.checker.GetAnyType()
	}
	return instantiateType(l.checker, side, newTypeMapper(parameters, anys))
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
