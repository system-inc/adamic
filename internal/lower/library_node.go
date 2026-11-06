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
	return nodeLibraryName(l.nodeLibrarySymbol(node))
}

func (l *lowering) nodeLibrarySymbol(node *ast.Node) *ast.Symbol {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		node = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if node.Kind == ast.KindNewExpression {
		node = ast.SkipParentheses(node.AsNewExpression().Expression)
	}
	return l.symbol(node)
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
// declares it, or by the host type itself, the owner's name and the member's. A declaration many
// host types inherit is registered by its owner only: FSWatcher's on is InternalEventEmitter's, as
// ChildProcess's is, whose shape has no on (Library on #j71fn59). The one place that answers it, for
// the view check here and Workers' census.
func (l *lowering) nodeHostObject(proven *checker.Type) (registered func(member *ast.Symbol) bool, host bool) {
	if proven == nil || proven.Symbol() == nil {
		return nil, false
	}
	owner := nodeLibraryName(proven.Symbol())
	if owner == "" {
		return nil, false
	}
	return func(member *ast.Symbol) bool {
		if nodeLibraryMemberImplemented(member) {
			return true
		}
		// The owner's key is for a member the host type inherits, named by a declaration elsewhere. A
		// member it declares itself already has that name, and nodeLibraryMemberImplemented has judged
		// it with its namespace: inspector's Debugger.CallFrame.url and Runtime.CallFrame.url share one
		// flat name, told apart only there.
		key := owner + "." + member.Name
		return key != nodeLibraryName(member) && implementedNodeMembers[key]
	}, true
}

// nodeHostOwnedMember names a member read off a host object by the host type it's read off, as
// nodeHostObject's registration does, node:fs.FSWatcher.on for an FSWatcher's inherited on: for a
// call through the member, the member's property access, or its name. It's "" for anything else.
func (l *lowering) nodeHostOwnedMember(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		node = ast.SkipParentheses(node.AsCallExpression().Expression)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		node = node.Parent
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	owner := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
	if _, host := l.nodeHostObject(owner); !host {
		return ""
	}
	return nodeLibraryName(owner.Symbol()) + "." + node.Name().Text()
}

// nodeHostViewDepth bounds how far nodeHostViewReach follows a view into its members' types, and
// nodeHostViewTooDeep starts what it reaches when it stops there, naming the type it stopped at.
const (
	nodeHostViewDepth   = 8
	nodeHostViewTooDeep = "past the depth the view check follows, at "
	nodeHostViewOpaque  = "an opaque view of "
)

// parameter distinguishes the opaque-view policy of a callback's incoming values from the
// existing result policy. Keep both in the seen key: visiting one must not suppress the other.
type nodeHostViewPair struct {
	own, view *checker.Type
	parameter bool
}

// nodeHostViewReach is the first member of a host object that a type it's seen as reaches and no
// registration names, or "" when every member reached is registered. It follows each property of
// the view, each object member of a union view, the property's type and a method's result and
// contravariant parameters, with a seen-set. Past nodeHostViewDepth it fails closed, naming the
// type it stopped at. A value
// that isn't a host object is walked together with its view to every host object inside it: an
// instantiated library generic (an array, a tuple, a Map) through its type arguments paired with
// the view's, and any other object through its properties, index signatures and methods' results,
// and anything callable through its call and construct results. So a container holding one, or a
// function giving one, is the host object seen through the container's view. A type parameter in
// the view is seen as its constraint, which is all its body can name.
func (l *lowering) nodeHostViewReach(own, view *checker.Type, depth int, seen map[nodeHostViewPair]bool, parameter bool) string {
	if own == nil || view == nil || own == view {
		return ""
	}
	if parameter && own.Flags()&checker.TypeFlagsInstantiableNonPrimitive != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(own); constraint != nil && constraint != own {
			return l.nodeHostViewReach(constraint, view, depth, seen, parameter)
		}
	}
	if own.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range own.Types() {
			if reached := l.nodeHostViewReach(member, view, depth, seen, parameter); reached != "" {
				return reached
			}
		}
		return ""
	}
	if own.Flags()&checker.TypeFlagsObject == 0 {
		return ""
	}
	if parameter && view.Flags()&checker.TypeFlagsInstantiableNonPrimitive != 0 && l.nodeHoldsHost(own) {
		return nodeHostViewOpaque + l.checker.TypeToString(own)
	}
	if view.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := l.checker.GetBaseConstraintOfType(view)
		if constraint == nil || constraint == view || constraint.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			if parameter && l.nodeHoldsHost(own) {
				return nodeHostViewOpaque + l.checker.TypeToString(own)
			}
			return ""
		}
		view = l.checker.GetNonNullableType(constraint)
	}
	if parameter && view.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 && l.nodeHoldsHost(own) {
		return nodeHostViewOpaque + l.checker.TypeToString(own)
	}
	view = l.checker.GetNonNullableType(view)
	registered, host := l.nodeHostObject(own)
	// A primitive view can't hold an object at all, a host object or one holding it.
	if own == view || seen[nodeHostViewPair{own, view, parameter}] || !host && !l.nodeHoldsHost(own) && !l.nodeHoldsHost(view) || view.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection) == 0 {
		return ""
	}
	// A union's properties are only those its members share, so each object member the value fits
	// is judged on its own: the program narrows to one and reaches all of its members. A member it
	// doesn't fit, a primitive or another host type among them, can't hold it. An intersection's
	// properties are already every member's.
	if view.Flags()&checker.TypeFlagsUnion != 0 {
		seen[nodeHostViewPair{own, view, parameter}] = true
		for _, member := range view.Types() {
			if member.Flags()&checker.TypeFlagsObject == 0 || !l.checker.IsTypeAssignableTo(own, member) {
				continue
			}
			if reached := l.nodeHostViewReach(own, member, depth, seen, parameter); reached != "" {
				return reached
			}
		}
		return ""
	}
	if depth > nodeHostViewDepth {
		// Past the bound it fails closed: on a host object, and on a container nodeHoldsHost hasn't
		// proved holds none (one that it has never gets this far).
		return nodeHostViewTooDeep + l.checker.TypeToString(own)
	}
	seen[nodeHostViewPair{own, view, parameter}] = true
	// A function or a class seen as a callable view gives what its calls give: each call or construct
	// signature's result is paired with the view's, as a callable property's are below.
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if reached := l.nodeHostSignaturesReach(l.checker.GetSignaturesOfType(own, kind), l.checker.GetSignaturesOfType(view, kind), depth, seen, parameter); reached != "" {
			return reached
		}
	}
	// A library container that's an instantiated generic (an array, a tuple, a Map, a Promise) holds
	// what its type arguments are, so they're paired with the view's, and its methods aren't walked.
	if !host && l.nodeLibraryContainer(own) && view.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		held, seenAs := l.checker.GetTypeArguments(own), l.checker.GetTypeArguments(view)
		if len(held) == len(seenAs) {
			for index := range held {
				if reached := l.nodeHostViewReach(held[index], seenAs[index], depth+1, seen, parameter); reached != "" {
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
		if reached := l.nodeHostViewReach(held, seenAs, depth+1, seen, parameter); reached != "" {
			return reached
		}
		if reached := l.nodeHostSignaturesReach(l.checker.GetSignaturesOfType(held, checker.SignatureKindCall), l.checker.GetSignaturesOfType(seenAs, checker.SignatureKindCall), depth, seen, parameter); reached != "" {
			return reached
		}
	}
	for _, index := range l.checker.GetIndexInfosOfType(view) {
		for _, held := range l.checker.GetIndexInfosOfType(own) {
			if held.KeyType() == index.KeyType() {
				if reached := l.nodeHostViewReach(held.ValueType(), index.ValueType(), depth+1, seen, parameter); reached != "" {
					return reached
				}
			}
		}
	}
	return ""
}

// nodeHoldsHost says whether a value of a type may hold a host object anywhere inside it, along the
// edges nodeHostViewReach walks: a union's members, a library container's type arguments, and any
// other object's properties, index signatures and call and construct results and parameters. It's remembered for the
// lowering, so the walk passes over the many program types that hold none: a search from one type
// that meets no host object proves every type it met holds none, and one that meets a host object
// proves the types on the way to it hold one.
func (l *lowering) nodeHoldsHost(root *checker.Type) bool {
	if l.nodeHostHolders == nil {
		l.nodeHostHolders = map[*checker.Type]bool{}
	}
	if holds, known := l.nodeHostHolders[root]; known {
		return holds
	}
	visited := map[*checker.Type]bool{}
	var holds func(*checker.Type) bool
	holds = func(proven *checker.Type) bool {
		if proven == nil {
			return false
		}
		if known, cached := l.nodeHostHolders[proven]; cached {
			return known
		}
		if visited[proven] {
			return false
		}
		visited[proven] = true
		if _, host := l.nodeHostObject(proven); host {
			l.nodeHostHolders[proven] = true
			return true
		}
		var next []*checker.Type
		switch {
		case proven.Flags()&checker.TypeFlagsInstantiableNonPrimitive != 0:
			if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != proven {
				next = append(next, constraint)
			}
		case proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
			next = proven.Types()
		case proven.Flags()&checker.TypeFlagsObject == 0:
		case l.nodeLibraryContainer(proven):
			next = l.checker.GetTypeArguments(proven)
		default:
			for _, property := range l.checker.GetPropertiesOfType(proven) {
				next = append(next, l.checker.GetTypeOfSymbol(property))
			}
			for _, index := range l.checker.GetIndexInfosOfType(proven) {
				next = append(next, index.ValueType())
			}
			for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
				for _, signature := range l.checker.GetSignaturesOfType(proven, kind) {
					next = append(next, l.checker.GetReturnTypeOfSignature(signature))
					for _, parameter := range signature.Parameters() {
						next = append(next, l.checker.GetTypeOfSymbol(parameter))
					}
				}
			}
		}
		for _, inner := range next {
			if holds(inner) {
				l.nodeHostHolders[proven] = true
				return true
			}
		}
		return false
	}
	if holds(root) {
		return true
	}
	for proven := range visited {
		l.nodeHostHolders[proven] = false
	}
	return false
}

// nodeHostSignaturesReach judges results covariantly and parameters contravariantly. A host
// callback hands its declared parameter types to the program's listener, and another function
// layer swaps that direction again. All overload pairs and every supplied position (including
// rest elements and optional parameters) are judged, not just a signature selected by its result.
func (l *lowering) nodeHostSignaturesReach(own, view []*checker.Signature, depth int, seen map[nodeHostViewPair]bool, parameter bool) string {
	for _, called := range view {
		for _, call := range own {
			called, call := checker.Checker_getCanonicalSignature(l.checker, called), checker.Checker_getCanonicalSignature(l.checker, call)
			result, held := l.checker.GetReturnTypeOfSignature(called), l.checker.GetReturnTypeOfSignature(call)
			if l.checker.IsTypeAssignableTo(held, result) {
				if reached := l.nodeHostViewReach(held, result, depth+1, seen, parameter); reached != "" {
					return reached
				}
			}
			for position := range l.nodeHostParameterPositions(call, called) {
				handed, declared := l.checker.GetTypeAtPosition(called, position), l.checker.GetTypeAtPosition(call, position)
				if declared.Flags()&checker.TypeFlagsInstantiableNonPrimitive != 0 && l.nodeHoldsHost(handed) {
					return nodeHostViewOpaque + l.checker.TypeToString(handed)
				}
				if reached := l.nodeHostViewReach(handed, declared, depth+1, seen, true); reached != "" {
					return reached
				}
			}
		}
	}
	return ""
}

// nodeHostParameterPositions expands fixed tuple rest parameters too. A declared rest slot can
// stand for several positions; inspecting just len(Parameters) would miss a later host object.
func (l *lowering) nodeHostParameterPositions(from, to *checker.Signature) int {
	positions := l.nodeHostParameterCount(from)
	if l.checker.HasEffectiveRestParameter(from) {
		positions = max(positions, l.nodeHostParameterCount(to))
	}
	if !l.checker.HasEffectiveRestParameter(to) {
		positions = min(positions, l.nodeHostParameterCount(to))
	}
	return positions
}

func (l *lowering) nodeHostParameterCount(signature *checker.Signature) int {
	n := len(signature.Parameters())
	if signature.HasRestParameter() && n != 0 {
		rest := l.checker.GetTypeOfSymbol(signature.Parameters()[n-1])
		if checker.IsTupleType(rest) {
			return n - 1 + len(l.checker.GetTypeArguments(rest))
		}
	}
	return n
}

// nodeLibraryContainer says whether a type is a tuple or an instantiated generic the library
// declares (an array, a Map, a Set, a Promise), which holds only what its type arguments are. A
// program's class or interface is a reference too, to itself with this, and holds what its members
// hold, so it's walked through them instead.
func (l *lowering) nodeLibraryContainer(proven *checker.Type) bool {
	if checker.IsTupleType(proven) {
		return true
	}
	symbol := proven.Symbol()
	return proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && symbol != nil && len(symbol.Declarations) != 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// nodeHostViewKey is a value type and the view it's seen through, as nodeHostViews keys them. It's a
// named struct, not a [2]*checker.Type, so the latent census's state copy (stage3/census/latent/
// statecopy) copies it field by field, as it copies the lowering's other state.
type nodeHostViewKey struct{ own, view *checker.Type }

// nodeHostViewAt is nodeHostViewReach from a relation, remembered for the lowering: the same value
// type is seen through the same view at many sites.
func (l *lowering) nodeHostViewAt(own, view *checker.Type) string {
	key := nodeHostViewKey{own, view}
	if reached, known := l.nodeHostViews[key]; known {
		return reached
	}
	reached := l.nodeHostViewReach(own, view, 0, map[nodeHostViewPair]bool{}, false)
	if l.nodeHostViews == nil {
		l.nodeHostViews = map[nodeHostViewKey]string{}
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
	if strings.HasPrefix(reached, nodeHostViewOpaque) {
		return l.notYet(node, subject+" seen as "+l.checker.TypeToString(view)+", which hides a host object behind any, unknown or an unbound type parameter")
	}
	if strings.HasPrefix(reached, nodeHostViewTooDeep) {
		return l.notYet(node, subject+" seen as "+l.checker.TypeToString(view)+", which goes "+reached+", where it may hold a host object")
	}
	return l.notYet(node, subject+" seen as "+l.checker.TypeToString(view)+", which reaches "+reached+", a member no host registration names")
}

// nodeHostCallbackArgument judges a registered call's callback before lowering the argument.
// GetTypeAtPosition expands rest slots; each library overload the call may use is checked.
func (l *lowering) nodeHostCallbackArgument(node *ast.Node) error {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression && parent.Kind != ast.KindNewExpression || !nodeLibraryMemberImplemented(l.nodeLibrarySymbol(parent)) {
		return nil
	}
	for position, argument := range parent.Arguments() {
		if argument != outer {
			continue
		}
		own := l.checker.GetTypeAtLocation(node)
		if len(l.checker.GetSignaturesOfType(own, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(own, checker.SignatureKindConstruct)) == 0 {
			return nil
		}
		kind := checker.SignatureKindCall
		if parent.Kind == ast.KindNewExpression {
			kind = checker.SignatureKindConstruct
		}
		for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(parent.Expression()), kind) {
			if position >= len(signature.Parameters()) && !signature.HasRestParameter() {
				continue
			}
			if !l.nodeHostCallbackOverload(parent, signature) {
				continue
			}
			view := l.checker.GetTypeAtPosition(signature, position)
			if reached := l.nodeHostViewAt(own, view); reached != "" {
				return l.nodeHostViewRefusal(node, own, view, reached)
			}
		}
	}
	return nil
}

// nodeHostCallbackOverload excludes signatures whose arguments cannot fit. Every overload
// of each applicable callback is still judged by the walk. Generic signatures
// stay in: an unbound slot cannot prove that the host hands out no object.
func (l *lowering) nodeHostCallbackOverload(call *ast.Node, signature *checker.Signature) bool {
	arguments := call.Arguments()
	if len(arguments) < signature.MinArgumentCount() || len(arguments) > l.nodeHostParameterCount(signature) && !l.checker.HasEffectiveRestParameter(signature) {
		return false
	}
	if len(signature.TypeParameters()) != 0 {
		return true
	}
	for index, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			continue
		}
		if !l.checker.IsTypeAssignableTo(l.checker.GetTypeAtLocation(argument), l.checker.GetTypeAtPosition(signature, index)) {
			return false
		}
	}
	return true
}

// nodeHostDeclaredArgument is the type an argument's parameter is declared as, before the call
// instantiates it: a generic body names the members of its constraint, not of what the call
// infers, so the argument is judged against that too. Nil for what isn't a call's argument. A value
// given a slot by a caller (slotSet) is judged against that slot, not the call's parameter (#b6m00y7).
func (l *lowering) nodeHostDeclaredArgument(node *ast.Node) *checker.Type {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if l.slotSet(outer) {
		return l.slotType(outer)
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
	if err := l.nodeHostCallbackArgument(node); err != nil {
		return err
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
	if found, err := l.nodeStackTraceLimitUseRefusal(node); found {
		return err
	}
	if found, err := l.nodeGlobalsUseRefusal(node); found {
		return err
	}
	if err := l.nodeProcessSignature(node); err != nil {
		return err
	}
	if found, err := l.nodeCommonJSUseRefusal(node); found {
		return err
	}
	if found, err := l.nodeWatchUseRefusal(node); found {
		return err
	}
	name := l.nodeLibraryMember(node)
	if name != "" && !nodeLibraryMemberImplemented(l.nodeLibrarySymbol(node)) {
		return l.notYet(node, name)
	}
	return nil
}

func init() { refusalRules.register(800, "nodeLibraryRefusal", (*lowering).nodeLibraryRefusal) }
