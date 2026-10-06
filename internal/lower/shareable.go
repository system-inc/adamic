package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// shareable uses the cycle finder's type graph, including its visited-type guard.
// Checking each node's own slots is enough: every reachable node must pass. Readonly
// describes a slot, never the contents of that slot. In particular it cannot bless a Map.
func (l *lowering) shareable(proven *checker.Type, path string, where *ast.Node) string {
	return l.shareableWith(proven, path, where, map[*ast.Node]bool{})
}

func (l *lowering) shareableWith(proven *checker.Type, path string, where *ast.Node, functions map[*ast.Node]bool) string {
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	finder.use(proven, where)
	types := append([]*checker.Type{proven}, finder.seen...)
	types = append(types, finder.shapes...)
	paths := map[*checker.Type]string{proven: path}
	for _, current := range types {
		if current == nil {
			return path + " has no complete type"
		}
		name := paths[current]
		if name == "" {
			name = path + " (" + l.checker.TypeToString(current) + ")"
		}
		child := func(of *checker.Type, suffix string) {
			if _, known := paths[of]; !known {
				paths[of] = name + suffix
			}
		}
		flags := current.Flags()
		switch {
		case finder.weak(current) || l.weakTarget(current) != nil:
			return name + " is a Weak"
		case flags&checker.TypeFlagsUnion != 0:
			for _, member := range current.Types() {
				child(member, "")
			}
		case flags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsNever|checker.TypeFlagsVoid) != 0:
		case flags&checker.TypeFlagsObject == 0 || finder.template(current):
			return name + " has a type the compiler cannot see the whole of"
		case l.isLibraryType(current, "Map", "Set"):
			kind := "Map"
			if l.isLibraryType(current, "Set") {
				kind = "Set"
			}
			return name + " is a mutable " + kind
		case l.checker.IsArrayType(current) || checker.IsTupleType(current):
			if !l.isLibraryType(current, "ReadonlyArray") && !(checker.IsTupleType(current) && current.TargetTupleType().IsReadonly()) {
				return name + " is a mutable " + l.checker.TypeToString(current)
			}
			for _, element := range l.checker.GetTypeArguments(current) {
				child(element, "[]")
			}
		case l.isLibraryType(current, "ReadonlyMap", "ReadonlySet"):
			for index, element := range l.checker.GetTypeArguments(current) {
				suffix := ".key"
				if index > 0 {
					suffix = ".value"
				}
				child(element, suffix)
			}
		case finder.isFunction(current):
			if why := l.shareableFunctionType(current, name, where, functions); why != "" {
				return why
			}
		default:
			if symbol := current.Symbol(); symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.GetSourceFileOfNode(declaration).IsDeclarationFile && current.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
						return name + " has a type the compiler cannot see the whole of"
					}
				}
			}
			for _, field := range finder.fields(current) {
				if !l.checker.IsReadonlySymbol(field) {
					return name + "." + field.Name + " is a mutable field"
				}
				child(l.checker.GetTypeOfSymbol(field), "."+field.Name)
			}
		}
	}
	return ""
}

// The program is closed: all function values are source bodies the compiler sees.
// Use the existing module order, and judge every source body assignable to this
// function type. A type with no visible producer cannot prove its captures.
// The shared visited set makes mutually recursive closure/type graphs finite.
func (l *lowering) shareableFunctionType(proven *checker.Type, path string, where *ast.Node, functions map[*ast.Node]bool) string {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return path + " is a function whose captures are not known"
	}
	proof := &parallelProof{l: l, active: functions, checked: map[*ast.Node]bool{}, shareableFunctions: functions}
	found := false
	why := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) || why != "" {
			return false
		}
		if ast.IsFunctionLike(node) && node.Body() != nil && l.checker.IsTypeAssignableTo(l.checker.GetTypeAtLocation(node), proven) {
			found = true
			if err := proof.function(node, nil); err != nil {
				why = path + " has unproven captures or effects: " + err.Error()
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
