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
// seen-set, and past nodeHostViewDepth it fails closed, naming the host type it stopped at.
func (l *lowering) nodeHostViewReach(own, view *checker.Type, depth int, seen map[[2]*checker.Type]bool) string {
	if own == nil || view == nil {
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
	view = l.checker.GetNonNullableType(view)
	registered, host := l.nodeHostObject(own)
	if !host || own == view || seen[[2]*checker.Type{own, view}] {
		return ""
	}
	// A union's properties are only those its members share, so each object member is judged on
	// its own: the program narrows to one and reaches all of its members. A primitive member can't
	// hold the host object. An intersection's properties are already every member's.
	if view.Flags()&checker.TypeFlagsUnion != 0 {
		seen[[2]*checker.Type{own, view}] = true
		for _, member := range view.Types() {
			if member.Flags()&checker.TypeFlagsObject == 0 {
				continue
			}
			if reached := l.nodeHostViewReach(own, member, depth, seen); reached != "" {
				return reached
			}
		}
		return ""
	}
	if depth > nodeHostViewDepth {
		return l.checker.TypeToString(own) + ", past the depth the view check follows"
	}
	seen[[2]*checker.Type{own, view}] = true
	for _, property := range l.checker.GetPropertiesOfType(view) {
		member := l.checker.GetPropertyOfType(own, property.Name)
		if member == nil || !registered(member) {
			if name := nodeLibraryName(member); name != "" {
				return name
			}
			return l.checker.TypeToString(own) + "." + property.Name
		}
		held, seenAs := l.checker.GetTypeOfSymbol(member), l.checker.GetTypeOfSymbol(property)
		if reached := l.nodeHostViewReach(held, seenAs, depth+1, seen); reached != "" {
			return reached
		}
		// Only a method's result is followed. A parameter would be the contravariant road, a host
		// method handing the program a host object through a callback; no registered member does
		// (write's callback gets an Error, nextTick's its own arguments), so none is followed yet.
		for _, called := range l.checker.GetSignaturesOfType(seenAs, checker.SignatureKindCall) {
			for _, call := range l.checker.GetSignaturesOfType(held, checker.SignatureKindCall) {
				if reached := l.nodeHostViewReach(l.checker.GetReturnTypeOfSignature(call), l.checker.GetReturnTypeOfSignature(called), depth+1, seen); reached != "" {
					return reached
				}
			}
		}
	}
	return ""
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
