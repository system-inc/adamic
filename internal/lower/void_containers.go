package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// refuseVoidContainerRead refuses a use of a value that holds void at a position, unless the use is a
// move (#p7t8tmh, ruled Oct 28 on Language's attack). A discard releases what a void call returns
// where Node keeps it, so `[1, 2].map((i) => mark(i))` holds 1 and 2 on Node and undefined here, and
// `all.includes(undefined)` printed true where Node prints false: the library read the elements. H2
// refuses a read of one void value; this refuses a read of many, made by whatever the container is
// handed to. It judges expressions, not bindings: a call's result, a field or element read, a callback
// parameter, every expression whose type holds void in an element, a type argument, a property or an
// index signature, at any depth. A use not on the list of moves is a read, so a member nobody named
// fails closed, which is how includes got through a list of readers.
//
// The moves: length and any other field read, an element read (each yields a value judged in turn by
// its own type), === and !== on the container, entering a slot whose type keeps every void position
// exactly void, push, pop, shift, unshift, splice, slice and forEach called on it, for...of binding
// its element to a name, flowing on where the whole expression stands, and a discard. A promise adds
// two (#tfycfvw, ruled Oct 28): await, and then and catch with callbacks written in place. Promise.all,
// allSettled and race need no case of their own: the tuple they take keeps each void exactly, so the
// promises enter it as a slot, and the container they yield is judged in turn.
func (l *lowering) refuseVoidContainerRead(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || ast.IsDeclarationName(node) || ast.IsAssignmentTarget(node) {
		return nil
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Name() == node {
		return nil // The access is the use, never the name it reads.
	}
	source := l.checker.GetTypeAtLocation(node)
	if source == nil || !l.holdsVoid(source) {
		return nil
	}
	role := l.voidContainerUse(node, source)
	if role == "" {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "a read of a void container's elements, " + role, Fix: "keep the values with a non-void return type, or discard them with forEach (adamic/no-void-read)"}
}

// voidContainerMethods are the members that read no element of the container they're called on: they
// add values, remove them, copy a run of them into another container of the same type, or hand each to
// a callback whose parameter is judged as any void value is.
var voidContainerMethods = map[string]bool{"push": true, "pop": true, "shift": true, "unshift": true, "splice": true, "slice": true, "forEach": true}

// voidContainerUse is what the place node stands in does with the container it holds, or "" for a move.
func (l *lowering) voidContainerUse(node *ast.Node, source *checker.Type) string {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return "as a value"
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return ""
	case ast.KindForStatement:
		statement := parent.AsForStatement()
		if statement.Initializer == site || statement.Incrementor == site {
			return ""
		}
		return "as a loop condition"
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		return l.voidContainerMember(parent, source, l.checker.GetSymbolAtLocation(parent), access.Name().Text())
	case ast.KindElementAccessExpression:
		if parent.AsElementAccessExpression().Expression != site {
			return "as an index"
		}
		argument := parent.AsElementAccessExpression().ArgumentExpression
		if l.checker.GetTypeAtLocation(argument).Flags()&checker.TypeFlagsNumberLike != 0 {
			return l.voidElementRead(parent, "an element read")
		}
		symbol := l.checker.GetSymbolAtLocation(argument)
		name := "a computed member"
		if symbol != nil {
			name = symbol.Name
		}
		return l.voidContainerMember(parent, source, symbol, name)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return "" // Identity reads no element.
		case ast.KindCommaToken:
			if binary.Left == site {
				return ""
			}
			return l.voidContainerFlows(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			if binary.Right == site {
				return l.voidContainerFlows(parent)
			}
		case ast.KindEqualsToken:
			return l.voidContainerSlot(source, l.checker.GetTypeAtLocation(binary.Left), "an assigned value")
		}
		return "as an operand of " + scanner.TokenToString(binary.OperatorToken.Kind)
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == site {
			return "as a condition"
		}
		return l.voidContainerFlows(parent)
	case ast.KindArrayLiteralExpression, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		// The literal holding it is judged in turn by its own type.
		literal := parent
		if parent.Kind != ast.KindArrayLiteralExpression {
			literal = parent.Parent
		}
		return l.voidContainerFlows(literal)
	case ast.KindAwaitExpression, ast.KindNonNullExpression:
		return l.voidContainerFlows(parent)
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		if role := l.voidContainerSlot(source, l.checker.GetTypeAtLocation(parent), "an assertion's operand"); role != "" {
			return role
		}
		return l.voidContainerFlows(parent)
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration, ast.KindBindingElement:
		name := parent.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "as a destructured value"
		}
		return l.voidContainerSlot(source, l.checker.GetTypeAtLocation(name), "a binding's initializer")
	case ast.KindReturnStatement:
		return l.voidContainerReturn(source, ast.GetContainingFunction(parent))
	case ast.KindArrowFunction:
		if parent.Body() == site {
			return l.voidContainerReturn(source, parent)
		}
	case ast.KindForOfStatement:
		statement := parent.AsForInOrOfStatement()
		if statement.Expression != site {
			break
		}
		// Each element is bound to a name typed as the element is, and judged there.
		if initializer := statement.Initializer; initializer != nil && initializer.Kind == ast.KindVariableDeclarationList {
			declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
			if len(declarations) == 1 && declarations[0].Name().Kind == ast.KindIdentifier {
				return ""
			}
		}
		return "as a for...of source bound to an existing name or a pattern"
	case ast.KindCallExpression, ast.KindNewExpression:
		if parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == site {
			break
		}
		return l.voidContainerSlot(source, l.argumentSlot(parent, site), "an argument")
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return "as a spread"
	case ast.KindTemplateSpan:
		return "as a template literal part"
	case ast.KindTypeOfExpression:
		return "as typeof's operand"
	}
	return "as a value"
}

// voidContainerMember judges a member read off a container: a field or an accessor is a read of that
// one value, judged by its own type, and a method is a read of the elements unless it is on the list
// and called.
func (l *lowering) voidContainerMember(access *ast.Node, receiver *checker.Type, symbol *ast.Symbol, name string) string {
	if symbol == nil {
		return "through " + name
	}
	if symbol.Flags&ast.SymbolFlagsMethod == 0 {
		return l.voidElementRead(access, "a field read")
	}
	promised := (name == "then" || name == "catch") && l.isLibraryType(receiver, "Promise")
	if !voidContainerMethods[name] && !promised || !called(access) {
		return "through " + name
	}
	call := access
	for call.Parent.Kind == ast.KindParenthesizedExpression {
		call = call.Parent
	}
	call = call.Parent
	switch {
	case name == "forEach":
		return l.voidCallbackParameters(call, name, 1)
	case promised:
		return l.voidCallbackParameters(call, name, 2)
	}
	if name == "pop" || name == "shift" {
		return l.voidElementRead(call, name)
	}
	return ""
}

// voidElementRead judges what a field read, an element read, pop or shift hands out. A container or an
// exactly void value is judged in turn, by this rule or by H2's. A union that holds void came from a
// void position (a tuple read at a computed index, or one member of a union of containers), so it is a
// void read wherever it's read, even beside undefined, and only a discard keeps it.
func (l *lowering) voidElementRead(read *ast.Node, what string) string {
	element := l.checker.GetTypeAtLocation(read)
	if element == nil || element.Flags()&checker.TypeFlagsUnion == 0 {
		return ""
	}
	for _, member := range element.Types() {
		if member.Flags() == checker.TypeFlagsVoid {
			if role := l.voidReadRole(read); role != "" {
				return "through " + what + " whose value is read as " + role
			}
			return ""
		}
	}
	return ""
}

// voidCallbackParameters keeps forEach, then and catch moves only for callbacks written in place whose
// parameters keep the types the method hands them: each is then judged where it's read. A function
// named instead of written, or a parameter typed wider, would read the elements unseen. The first
// callbacks arguments are the callbacks, and then's and catch's may be passed as undefined or null.
func (l *lowering) voidCallbackParameters(call *ast.Node, method string, callbacks int) string {
	for position, argument := range call.AsCallExpression().Arguments.Nodes {
		if position >= callbacks {
			break
		}
		callback := ast.SkipParentheses(argument)
		if method != "forEach" && l.checker.GetTypeAtLocation(callback).Flags()&^(checker.TypeFlagsUndefined|checker.TypeFlagsNull) == 0 {
			continue
		}
		if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
			return "through " + method + " with a callback not written in place"
		}
		expected := l.argumentSlot(call, argument)
		if expected != nil {
			expected = l.checker.GetNonNullableType(expected)
		}
		var signatures []*checker.Signature
		if expected != nil {
			signatures = l.checker.GetSignaturesOfType(expected, checker.SignatureKindCall)
		}
		if len(signatures) != 1 {
			return "through " + method + " with a callback whose type can't be read"
		}
		for index, parameter := range callback.Parameters() {
			name := parameter.Name()
			if name == nil || name.Kind != ast.KindIdentifier || parameter.AsParameterDeclaration().DotDotDotToken != nil {
				return "through " + method + "'s callback parameter " + strconv.Itoa(index) + ", destructured or rest"
			}
			if !l.voidKeepsPosition(l.checker.GetTypeAtPosition(signatures[0], index), l.checker.GetTypeAtLocation(name), 0, map[[2]*checker.Type]bool{}) {
				return "through " + method + "'s callback parameter " + name.Text() + ", typed wider"
			}
		}
	}
	return ""
}

// voidContainerFlows is a move when the expression the container flows into is judged in turn: it
// holds void itself, or it is a single void value H2 judges.
func (l *lowering) voidContainerFlows(whole *ast.Node) string {
	if whole != nil {
		if flowed := l.checker.GetTypeAtLocation(whole); flowed != nil && (flowed.Flags() == checker.TypeFlagsVoid || l.holdsVoid(flowed)) {
			return ""
		}
	}
	return "as a value seen without its void"
}

func (l *lowering) voidContainerSlot(source, target *checker.Type, role string) string {
	if target == nil || !l.voidKeeps(source, target, 0, map[[2]*checker.Type]bool{}) {
		return "as " + role
	}
	return ""
}

// voidContainerReturn compares what a function returns with its return type, or for an async function
// with either its promise or what that resolves to.
func (l *lowering) voidContainerReturn(source *checker.Type, function *ast.Node) string {
	if function == nil || !ast.IsFunctionLike(function) {
		return "as a returned value"
	}
	signature := l.checker.GetSignatureFromDeclaration(function)
	if signature == nil {
		return "as a returned value"
	}
	returns := l.checker.GetReturnTypeOfSignature(signature)
	if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) && l.voidContainerSlot(source, checker.Checker_getAwaitedType(l.checker, returns), "") == "" {
		return ""
	}
	return l.voidContainerSlot(source, returns, "a returned value")
}

// argumentSlot is the instantiated parameter type argument enters, or nil.
func (l *lowering) argumentSlot(call, argument *ast.Node) *checker.Type {
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil {
		return nil
	}
	var arguments []*ast.Node
	if call.Kind == ast.KindCallExpression {
		arguments = call.AsCallExpression().Arguments.Nodes
	} else if call.AsNewExpression().Arguments != nil {
		arguments = call.AsNewExpression().Arguments.Nodes
	}
	for index, each := range arguments {
		if each == argument {
			return l.checker.GetTypeAtPosition(signature, index)
		}
	}
	return nil
}

// holdsVoid is whether a value of type holds void at a position: an element, a type argument, a
// property or an index signature, at any depth. A function's return isn't a position (calling one is
// H2's), and neither is a member of a union: void | X is the union ruling's. A primitive brand
// (`string & { __escapedIdentifier: void }`) holds nothing, since a primitive has no own property.
func (l *lowering) holdsVoid(t *checker.Type) bool {
	if held, known := l.voidHolders[t]; known {
		return held
	}
	visited := map[*checker.Type]bool{}
	held := l.holdsVoidWalk(t, visited)
	if l.voidHolders == nil {
		l.voidHolders = map[*checker.Type]bool{}
	}
	if held {
		l.voidHolders[t] = true
	} else {
		// Nothing the walk reached holds void, or it would have said so: every type it visited is settled.
		for visit := range visited {
			l.voidHolders[visit] = false
		}
	}
	return held
}

func (l *lowering) holdsVoidWalk(t *checker.Type, visited map[*checker.Type]bool) bool {
	if t == nil || visited[t] {
		return false
	}
	if held, known := l.voidHolders[t]; known {
		return held
	}
	visited[t] = true
	position := func(inside *checker.Type) bool {
		return inside != nil && (inside.Flags() == checker.TypeFlagsVoid || l.holdsVoidWalk(inside, visited))
	}
	switch {
	case t.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range t.Types() {
			if l.holdsVoidWalk(member, visited) {
				return true
			}
		}
		return false
	case t.Flags()&checker.TypeFlagsIntersection != 0:
		for _, member := range t.Types() {
			if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsESSymbolLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0 {
				return false
			}
		}
		for _, member := range t.Types() {
			if l.holdsVoidWalk(member, visited) {
				return true
			}
		}
		return false
	case t.Flags()&checker.TypeFlagsObject == 0:
		return false
	}
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil {
		for _, argument := range l.checker.GetTypeArguments(t) {
			if position(argument) {
				return true
			}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(t) {
		if position(checker.Checker_getIndexTypeOfType(l.checker, t, info.KeyType())) {
			return true
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		// A method's type is a function, and a function's return isn't a position.
		if property.Flags&ast.SymbolFlagsMethod == 0 && position(l.checker.GetTypeOfSymbol(property)) {
			return true
		}
	}
	return false
}

// voidKeeps is whether a slot of type target keeps every position where source holds void exactly
// void, so that what enters it is still judged as a void container wherever it's used. Anything it
// can't show keeps them is false: a union of more than one container, a property the target drops, a
// type parameter, any, unknown.
func (l *lowering) voidKeeps(source, target *checker.Type, depth int, seen map[[2]*checker.Type]bool) bool {
	if source == target {
		return true
	}
	if source == nil || target == nil || depth >= 32 {
		return false
	}
	if !l.holdsVoid(source) {
		return true
	}
	pair := [2]*checker.Type{source, target}
	if seen[pair] {
		return true // A cycle that has kept every position so far keeps them.
	}
	seen[pair] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if !l.voidKeeps(member, target, depth+1, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		// An optional slot is the one union that keeps the container whole.
		var kept *checker.Type
		for _, member := range target.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
			if kept != nil {
				return false
			}
			kept = member
		}
		return kept != nil && l.voidKeeps(source, kept, depth+1, seen)
	}
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if checker.Checker_isArrayOrTupleType(l.checker, source) || checker.Checker_isArrayOrTupleType(l.checker, target) {
		if !checker.Checker_isArrayOrTupleType(l.checker, source) || !checker.Checker_isArrayOrTupleType(l.checker, target) {
			return false
		}
		if checker.IsTupleType(source) {
			targets := l.checker.GetTypeArguments(target)
			for index, inside := range l.checker.GetTypeArguments(source) {
				to := l.checker.GetNumberIndexType(target)
				if checker.IsTupleType(target) {
					if index >= len(targets) {
						return false
					}
					to = targets[index]
				}
				if !l.voidKeepsPosition(inside, to, depth+1, seen) {
					return false
				}
			}
			return true
		}
		return !checker.IsTupleType(target) && l.voidKeepsPosition(l.checker.GetNumberIndexType(source), l.checker.GetNumberIndexType(target), depth+1, seen)
	}
	if source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && source.Target() != nil && source.Target() == target.Target() {
		targets := l.checker.GetTypeArguments(target)
		for index, inside := range l.checker.GetTypeArguments(source) {
			if index >= len(targets) || !l.voidKeepsPosition(inside, targets[index], depth+1, seen) {
				return false
			}
		}
		return true
	}
	for _, info := range l.checker.GetIndexInfosOfType(source) {
		if from := checker.Checker_getIndexTypeOfType(l.checker, source, info.KeyType()); l.holdsPosition(from) {
			to := checker.Checker_getIndexTypeOfType(l.checker, target, info.KeyType())
			if to == nil || !l.voidKeepsPosition(from, to, depth+1, seen) {
				return false
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(source) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		from := l.checker.GetTypeOfSymbol(property)
		if !l.holdsPosition(from) {
			continue
		}
		// A target that drops the property still hands the whole value to whatever reads it.
		to := l.checker.GetTypeOfPropertyOfType(target, property.Name)
		if to == nil || !l.voidKeepsPosition(from, to, depth+1, seen) {
			return false
		}
	}
	return true
}

// holdsPosition is whether a position's type is void or holds void.
func (l *lowering) holdsPosition(inside *checker.Type) bool {
	return inside != nil && (inside.Flags() == checker.TypeFlagsVoid || l.holdsVoid(inside))
}

// voidKeepsPosition is voidKeeps for one position: exactly void stays exactly void.
func (l *lowering) voidKeepsPosition(from, to *checker.Type, depth int, seen map[[2]*checker.Type]bool) bool {
	if from == nil || !l.holdsPosition(from) {
		return true
	}
	if from.Flags() == checker.TypeFlagsVoid {
		return to != nil && to.Flags() == checker.TypeFlagsVoid
	}
	return l.voidKeeps(from, to, depth, seen)
}
