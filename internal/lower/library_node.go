package lower

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Register only members with real lowering and runtime support. Host units add
// their names during init; the default is a named NotYet, never erased code.
var implementedNodeMembers = map[string]bool{"node:console.Console.log": true, "node:console.Console.error": true, "node:console.console": true}

func RegisterNodeLibraryMembers(names ...string) {
	for _, name := range names {
		implementedNodeMembers[name] = true
	}
}

func (l *lowering) nodeLibraryMember(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		node = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if node.Kind == ast.KindNewExpression {
		node = ast.SkipParentheses(node.AsNewExpression().Expression)
	}
	return nodeLibraryName(l.symbol(node))
}

// nodeLibraryName is the name a Node library declaration is registered by: its module, the classes
// and interfaces it's declared in, and its own name, as node:perf_hooks.Performance.timeOrigin. It's
// empty for a symbol no Node library file declares, or a module.
func nodeLibraryName(symbol *ast.Symbol) string {
	if symbol == nil {
		return ""
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if !load.IsNodeLibrary(file) {
			continue
		}
		if declaration.Kind == ast.KindModuleDeclaration {
			return ""
		}
		module := ""
		owners := []string{}
		for parent := declaration.Parent; parent != nil; parent = parent.Parent {
			if (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration) && parent.Name() != nil {
				owners = append([]string{parent.Name().Text()}, owners...)
			}
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil && parent.Name().Kind == ast.KindStringLiteral {
				module = parent.Name().Text()
				break
			}
		}
		if module == "" {
			module = strings.TrimSuffix(path.Base(file.FileName().AsString()), ".d.ts")
		}
		if !strings.HasPrefix(module, "node:") {
			module = "node:" + module
		}
		return strings.Join(append([]string{module}, append(owners, symbol.Name)...), ".")
	}
	return ""
}

// nodeHostObject says whether a type is a host object's, a class or interface a Node library file
// declares, and which of its members its registration names: implementedNodeMembers keys each one
// by its own declaration's nodeLibraryName, so a member it inherits is named by the interface that
// declares it. The one place that answers it, for the view check here and Workers' census.
func (l *lowering) nodeHostObject(proven *checker.Type) (registered func(member *ast.Symbol) bool, host bool) {
	if proven == nil || proven.Symbol() == nil || nodeLibraryName(proven.Symbol()) == "" {
		return nil, false
	}
	return func(member *ast.Symbol) bool { return implementedNodeMembers[nodeLibraryName(member)] }, true
}

// nodeHostViewDepth bounds how far nodeHostViewReach follows a view into its members' types.
const nodeHostViewDepth = 8

// nodeHostViewReach is the first member of a host object that a type it's seen as reaches and no
// registration names, or "" when every member reached is registered. It follows each property of
// the view, each object member of a union view, the property's type and a method's result, with a
// seen-set, and past nodeHostViewDepth it fails closed, naming the host type it stopped at. A value
// that isn't a host object is walked together with its view to every host object inside it: an
// instantiated generic (an array, a tuple, a Map) through its type arguments paired with the
// view's, and any other object through its properties, index signatures and methods' results. So a
// container holding one is the host object seen through the container's view. A type parameter in
// the view is seen as its constraint, which is all its body can name.
func (l *lowering) nodeHostViewReach(own, view *checker.Type, depth int, seen map[[2]*checker.Type]bool) string {
	if own == nil || view == nil || own == view {
		return ""
	}
	if own.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range own.Types() {
			if reached := l.nodeHostViewReach(member, view, depth, seen); reached != "" {
				return reached
			}
		}
		return ""
	}
	if own.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	view = l.checker.GetNonNullableType(view)
	if view.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := l.checker.GetBaseConstraintOfType(view)
		if constraint == nil || constraint == view {
			return ""
		}
		view = l.checker.GetNonNullableType(constraint)
	}
	registered, host := l.nodeHostObject(own)
	if own == view || seen[[2]*checker.Type{own, view}] {
		return ""
	}
	// A union's properties are only those its members share, so each object member the value fits
	// is judged on its own: the program narrows to one and reaches all of its members. A member it
	// doesn't fit, a primitive or another host type among them, can't hold it. An intersection's
	// properties are already every member's.
	if view.Flags()&checker.TypeFlagsUnion != 0 {
		seen[[2]*checker.Type{own, view}] = true
		for _, member := range view.Types() {
			if member.Flags()&checker.TypeFlagsObject == 0 || !l.checker.IsTypeAssignableTo(own, member) {
				continue
			}
			if reached := l.nodeHostViewReach(own, member, depth, seen); reached != "" {
				return reached
			}
		}
		return ""
	}
	if depth > nodeHostViewDepth {
		// A container this deep holds no host object the walk has met; a host object here fails closed.
		if !host {
			return ""
		}
		return l.checker.TypeToString(own) + ", past the depth the view check follows"
	}
	seen[[2]*checker.Type{own, view}] = true
	// A container that's an instantiated generic (an array, a tuple, a Map, a Promise) holds what its
	// type arguments are, so they're paired with the view's, and its library methods aren't walked.
	if !host && own.ObjectFlags()&checker.ObjectFlagsReference != 0 && view.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		held, seenAs := l.checker.GetTypeArguments(own), l.checker.GetTypeArguments(view)
		if len(held) == len(seenAs) {
			for index := range held {
				if reached := l.nodeHostViewReach(held[index], seenAs[index], depth+1, seen); reached != "" {
					return reached
				}
			}
			return ""
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(view) {
		member := l.checker.GetPropertyOfType(own, property.Name)
		if host && (member == nil || !registered(member)) {
			if name := nodeLibraryName(member); name != "" {
				return name
			}
			return l.checker.TypeToString(own) + "." + property.Name
		}
		if member == nil {
			continue
		}
		held, seenAs := l.checker.GetTypeOfSymbol(member), l.checker.GetTypeOfSymbol(property)
		if reached := l.nodeHostViewReach(held, seenAs, depth+1, seen); reached != "" {
			return reached
		}
		// Only a method's result is followed. A parameter would be the contravariant road, a host
		// method handing the program a host object through a callback; no registered member does
		// (TestNodeRegisteredCallbacksHandNoHostObject holds it), so none is followed yet.
		for _, called := range l.checker.GetSignaturesOfType(seenAs, checker.SignatureKindCall) {
			for _, call := range l.checker.GetSignaturesOfType(held, checker.SignatureKindCall) {
				if reached := l.nodeHostViewReach(l.checker.GetReturnTypeOfSignature(call), l.checker.GetReturnTypeOfSignature(called), depth+1, seen); reached != "" {
					return reached
				}
			}
		}
	}
	for _, index := range l.checker.GetIndexInfosOfType(view) {
		for _, held := range l.checker.GetIndexInfosOfType(own) {
			if held.KeyType() == index.KeyType() {
				if reached := l.nodeHostViewReach(held.ValueType(), index.ValueType(), depth+1, seen); reached != "" {
					return reached
				}
			}
		}
	}
	return ""
}

// nodeHostViewAt is nodeHostViewReach from a relation, remembered for the lowering: the same value
// type is seen through the same view at many sites.
func (l *lowering) nodeHostViewAt(own, view *checker.Type) string {
	key := [2]*checker.Type{own, view}
	if reached, known := l.nodeHostViews[key]; known {
		return reached
	}
	reached := l.nodeHostViewReach(own, view, 0, map[[2]*checker.Type]bool{})
	if l.nodeHostViews == nil {
		l.nodeHostViews = map[[2]*checker.Type]string{}
	}
	l.nodeHostViews[key] = reached
	return reached
}

// nodeHostViewRefusal is the NotYet for a value whose view reaches an unregistered member: the host
// object itself, or a value holding one (an object, an array, a tuple or a Map of them).
func (l *lowering) nodeHostViewRefusal(node *ast.Node, own, view *checker.Type, reached string) error {
	subject := "a host object of type " + l.checker.TypeToString(own)
	if _, host := l.nodeHostObject(l.checker.GetNonNullableType(own)); !host {
		subject = "a value of type " + l.checker.TypeToString(own) + " holding a host object"
	}
	return l.notYet(node, subject+" seen as "+l.checker.TypeToString(view)+", which reaches "+reached+", a member no host registration names")
}

// nodeHostDeclaredArgument is the type an argument's parameter is declared as, before the call
// instantiates it: a generic body names the members of its constraint, not of what the call
// infers, so the argument is judged against that too. Nil for what isn't a call's argument.
func (l *lowering) nodeHostDeclaredArgument(node *ast.Node) *checker.Type {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	call := outer.Parent
	if call == nil || call.Kind != ast.KindCallExpression && call.Kind != ast.KindNewExpression {
		return nil
	}
	resolved := l.checker.GetResolvedSignature(call)
	if resolved == nil || resolved.Declaration() == nil {
		return nil
	}
	arguments := call.Arguments()
	parameters := resolved.Declaration().Parameters()
	for index, argument := range arguments {
		if argument != outer || index >= len(parameters) || parameters[index].AsParameterDeclaration().DotDotDotToken != nil {
			continue
		}
		return l.checker.GetTypeAtLocation(parameters[index].Name())
	}
	return nil
}

func (l *lowering) nodeLibraryRefusal(node *ast.Node) error {
	if err := l.nodeFSDirectorySignature(node); err != nil {
		return err
	}
	if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) {
		return nil
	}
	if found, err := l.nodeStdoutUseRefusal(node); found {
		return err
	}
	if found, err := l.nodeEnvironmentUseRefusal(node); found {
		return err
	}
	if found, err := l.nodeRealpathUseRefusal(node); found {
		return err
	}
	if found, err := l.nodePlatformUseRefusal(node); found {
		return err
	}
	if err := l.nodeProcessSignature(node); err != nil {
		return err
	}
	name := l.nodeLibraryMember(node)
	if name != "" && !implementedNodeMembers[name] {
		return l.notYet(node, name)
	}
	return nil
}

func init() { refusalRules.register(800, "nodeLibraryRefusal", (*lowering).nodeLibraryRefusal) }
