package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// shareable uses the cycle finder's type graph, including its visited-type guard.
// Checking each node's own slots is enough: every reachable node must pass. Readonly
// describes a slot, never the contents of that slot. In particular it cannot bless a Map.
func (lowering *lowering) shareable(proven *checker.Type, path string, where *ast.Node) string {
	return lowering.shareableWith(proven, path, where, map[*ast.Node]bool{})
}

func (lowering *lowering) shareableWith(
	proven *checker.Type,
	path string,
	where *ast.Node,
	functions map[*ast.Node]bool,
) string {
	finder := &cycleFinder{lowering: lowering, where: map[*checker.Type]*ast.Node{}}
	finder.use(proven, where)
	shapes := lowering.parallelShapes()
	types := append([]*checker.Type{proven}, finder.seen...)
	types = append(types, finder.shapes...)
	paths := map[*checker.Type]string{proven: path}
	for index := 0; index < len(types); index++ {
		current := types[index]
		if current == nil {
			return path + " has no complete type"
		}
		name := paths[current]
		if name == "" {
			name = path + " (" + lowering.checker.TypeToString(current) + ")"
		}
		child := func(of *checker.Type, suffix string) {
			if _, known := paths[of]; !known {
				paths[of] = name + suffix
			}
		}
		flags := current.Flags()
		switch {
		case finder.weak(current) || lowering.weakTarget(current) != nil:
			return name + " is a Weak"
		case flags&checker.TypeFlagsUnion != 0:
			for _, member := range current.Types() {
				child(member, "")
			}
		case flags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsNever|checker.TypeFlagsVoid) != 0:
		case flags&checker.TypeFlagsObject == 0 || finder.template(current):
			return name + " has a type the compiler cannot see the whole of"
		case lowering.isLibraryType(current, "Map", "Set"):
			kind := "Map"
			if lowering.isLibraryType(current, "Set") {
				kind = "Set"
			}
			return name + " is a mutable " + kind
		case lowering.isArrayType(current) || checker.IsTupleType(current):
			if !lowering.isLibraryType(current, "ReadonlyArray") &&
				!(checker.IsTupleType(current) && current.TargetTupleType().IsReadonly()) {
				return name + " is a mutable " + lowering.checker.TypeToString(current)
			}
			for _, element := range lowering.checker.GetTypeArguments(current) {
				child(element, "[]")
			}
		case lowering.isLibraryType(current, "ReadonlyMap", "ReadonlySet"):
			for index, element := range lowering.checker.GetTypeArguments(current) {
				suffix := ".key"
				if index > 0 {
					suffix = ".value"
				}
				child(element, suffix)
			}
		case finder.isFunction(current):
			if why := lowering.shareableFunctionType(current, name, where, functions); why != "" {
				return why
			}
		default:
			// Mapped collection APIs can still expose mutable Map/Set methods. They
			// must not turn into empty immutable records when methods are omitted.
			for _, property := range lowering.checker.GetPropertiesOfType(current) {
				for _, declaration := range property.Declarations {
					parent := declaration.Parent
					if parent == nil ||
						parent.Kind != ast.KindInterfaceDeclaration ||
						!ast.GetSourceFileOfNode(parent).IsDeclarationFile {
						continue
					}
					if parent.Name().Text() == "Map" || parent.Name().Text() == "Set" {
						return name + " is a mutable " + parent.Name().Text()
					}
					if parent.Name().Text() == "ReadonlyMap" || parent.Name().Text() == "ReadonlySet" {
						return name + " has a collection type the compiler cannot see the whole of"
					}
				}
			}
			if symbol := current.Symbol(); symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.GetSourceFileOfNode(declaration).IsDeclarationFile &&
						current.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
						return name + " has a type the compiler cannot see the whole of"
					}
				}
			}
			for _, field := range finder.fields(current) {
				if !lowering.checker.IsReadonlySymbol(field) {
					return name + "." + field.Name + " is a mutable field"
				}
				child(lowering.checker.GetTypeOfSymbol(field), "."+field.Name)
			}

			// A readonly interface or base class may hide a subtype's extra fields.
			// Reuse cycleFinder's shapes and use graph; do not add another type walker.
			visible := map[string]bool{}
			for _, field := range finder.fields(current) {
				visible[field.Name] = true
			}
			for _, shape := range shapes {
				if shape == current || !lowering.checker.IsTypeAssignableTo(shape, current) {
					continue
				}
				class := shape.Symbol() != nil && shape.Symbol().Flags&ast.SymbolFlagsClass != 0
				for _, field := range finder.fields(shape) {
					if visible[field.Name] && !class {
						continue
					}
					fieldPath := name + "." + field.Name
					if !lowering.checker.IsReadonlySymbol(field) {
						return fieldPath + " is a mutable field hidden by " + lowering.checker.TypeToString(current)
					}
					held := lowering.checker.GetTypeOfSymbol(field)
					child(held, "."+field.Name)
					beforeTypes, beforeShapes := len(finder.seen), len(finder.shapes)
					finder.use(held, where)
					types = append(types, finder.seen[beforeTypes:]...)
					types = append(types, finder.shapes[beforeShapes:]...)
				}
			}
		}
	}
	return ""
}

// The program is closed: all function values are source bodies the compiler sees.
// Use the existing module order, and judge every source body assignable to this
// function type. A type with no visible producer cannot prove its captures.
// The shared visited set makes mutually recursive closure/type graphs finite.
func (lowering *lowering) shareableFunctionType(
	proven *checker.Type,
	path string,
	where *ast.Node,
	functions map[*ast.Node]bool,
) string {
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return path + " is a function whose captures are not known"
	}
	proof := &parallelProof{
		lowering:           lowering,
		active:             functions,
		checked:            map[*ast.Node]bool{},
		shareableFunctions: functions,
	}
	found := false
	why := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) || why != "" {
			return false
		}
		if ast.IsFunctionLike(node) &&
			node.Body() != nil &&
			lowering.checker.IsTypeAssignableTo(lowering.checker.GetTypeAtLocation(node), proven) {
			found = true
			if functionError := proof.function(node, nil); functionError != nil {
				if refused, ok := functionError.(*Refused); ok {
					why = path + " has unproven captures or effects: " + refused.What
				} else {
					why = path + " has captures or effects the compiler cannot prove"
				}
				return false
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if why != "" {
		return why
	}
	if !found {
		return path + " is a function whose captures are not known"
	}
	return ""
}

// The same shape discovery the cycle finder uses, restricted to source values.
// Contextually readonly literal slots are judged by the declared view above;
// extra slots and complete class layouts must also be immutable.
func (lowering *lowering) parallelShapes() []*checker.Type {
	finder := &cycleFinder{lowering: lowering, where: map[*checker.Type]*ast.Node{}}
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return nil
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindNewExpression {
			proven := lowering.checker.GetTypeAtLocation(node)
			if proven.Flags()&checker.TypeFlagsObject != 0 && !finder.template(proven) {
				finder.made(proven, node)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return finder.shapes
}
