package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// A program namespace whose name is read as a value, tsc's `tracing = tracingEnabled`, has one object
// (#67jk7wq, docs/rulings.md's namespace value): a global holding a plain object literal with one field
// per exported function, in declaration order, each the function's canonical value (functionValue), so
// N.f === N.f and x.f === N.f. It's assigned once, at the start of the namespace's block, where the
// block's functions already exist, and a read of the name as a value is a read of that global behind
// the namespace's readiness, so no read sees it before the block ran. Only the narrow form is built: one
// block, at a file's top level, exporting functions, types and const enums alone, and seen only as its
// own type, typeof N or typeof N | undefined. Everything else stays NotYet in its own words, and a write
// to a member through the value is refused, since a qualified call is a direct call that wouldn't see it.

// noteNamespaceValues finds every program namespace whose name is read as a value, before the refusal
// pass and before any namespace block is lowered, since the block makes the object.
func (l *lowering) noteNamespaceValues(modules []*ast.SourceFile) {
	l.namespaceValues = map[*ast.Symbol]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && !l.namespaceQualifiedReceiver(node) && l.namespaceValueNode(node) {
			if declaration := l.namespaceDeclaration(node); declaration != nil && namespaceProgram(declaration) && !l.namespaceMergedFunction(node) {
				l.namespaceValues[l.symbol(declaration.Name())] = true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
}

// namespaceProgram reports whether a namespace declaration is the program's own executable block, not
// a host's ambient description (process and the like keep main's stop).
func namespaceProgram(declaration *ast.Node) bool {
	return !ambientDeclaration(declaration) && !ast.GetSourceFileOfNode(declaration).IsDeclarationFile
}

// namespaceQualifiedReceiver reports whether node is written as the receiver of a qualified read, N.f,
// with nothing between them; (N).f reads the namespace as a value.
func (l *lowering) namespaceQualifiedReceiver(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node
}

// namespaceHasValue reports whether a namespace declaration's object is built.
func (l *lowering) namespaceHasValue(declaration *ast.Node) bool {
	return l.namespaceValues[l.symbol(declaration.Name())]
}

// namespaceContainer is the global holding a namespace's object, declared undefined before anything runs.
func (l *lowering) namespaceContainer(declaration *ast.Node) int {
	symbol := l.symbol(declaration.Name())
	if local, found := l.namespaceContainers[symbol]; found {
		return local
	}
	if l.namespaceContainers == nil {
		l.namespaceContainers = map[*ast.Symbol]int{}
	}
	local := len(l.result.Locals)
	l.namespaceContainers[symbol] = local
	l.result.Locals = append(l.result.Locals, ir.Local{Name: declaration.Name().Text() + "_namespace", Type: ir.Object, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: local, Value: ir.Undefined{Of: ir.Object}})
	return local
}

// namespaceContainerLiteral is a namespace's object: each exported function's canonical value, in
// declaration order. namespaceValueAdmission has already stopped every other export that has a value.
func (l *lowering) namespaceContainerLiteral(declaration *ast.Node) (ir.Expression, error) {
	fields := []ir.Field{}
	for _, statement := range namespaceStatements(declaration) {
		if statement.Kind != ast.KindFunctionDeclaration || statement.Body() == nil || !ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			continue
		}
		function, found := l.functions[l.symbol(statement.Name())]
		if !found {
			return nil, l.notYet(statement, "an exported function in a namespace used as a value, without a lowered function")
		}
		value, err := l.functionValue(statement.Name(), function)
		if err != nil {
			return nil, err
		}
		fields = append(fields, ir.Field{Name: statement.Name().Text(), Value: value})
	}
	return ir.ObjectLiteral{Fields: fields}, nil
}

// namespaceContainerStatements make a namespace's object at the start of its block, when its name is
// read as a value anywhere.
func (l *lowering) namespaceContainerStatements(declaration *ast.Node) ([]ir.Statement, error) {
	if !l.namespaceHasValue(declaration) {
		return nil, nil
	}
	literal, err := l.namespaceContainerLiteral(declaration)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Assign{Local: l.namespaceContainer(declaration), Value: literal}}, nil
}

// namespaceValueRead is the namespace's name read as a value: its object, behind the namespace's
// readiness, which turns true only once the block has run. namespaceInitialization stops a read it can
// prove comes first; a read reached through an edge it can't follow is checked here as it runs.
func (l *lowering) namespaceValueRead(declaration *ast.Node) ir.Expression {
	value := ir.Expression(ir.Read{Local: l.namespaceContainer(declaration), Of: ir.Object})
	check := l.namespaceReadyCheck(declaration, "namespace "+declaration.Name().Text()+" read as a value before its block ran")
	b := l.libraryArrayBuilder([]ir.Expression{check, value})
	return b.finish("namespace_value", b.read(b.parameters[1]))
}

// namespaceValueAdmission judges a program namespace's name read as a value, node, other than as a
// qualified read's receiver. A host namespace, and any lowering that didn't note the program's namespace
// values first, keeps main's stop.
func (l *lowering) namespaceValueAdmission(node, declaration *ast.Node) error {
	if !ast.IsIdentifier(node) || !namespaceProgram(declaration) || l.namespaceValues == nil || !l.namespaceHasValue(declaration) {
		return l.notYet(node, "a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports")
	}
	// A namespace merged across blocks stops at its first block (namespaceRefusal), before any read.
	if body := declaration.AsModuleDeclaration().Body; declaration.Parent.Kind != ast.KindSourceFile || body == nil || body.Kind != ast.KindModuleBlock {
		return l.notYet(node, "a nested or dotted namespace used as a value")
	}
	for _, member := range namespaceStatements(declaration) {
		if err := l.namespaceValueExport(member); err != nil {
			return err
		}
	}
	return nil
}

// namespaceBlocks counts the namespace blocks declaring a namespace's symbol.
func (l *lowering) namespaceBlocks(declaration *ast.Node) int {
	blocks := 0
	for _, other := range l.symbol(declaration.Name()).Declarations {
		if other.Kind == ast.KindModuleDeclaration {
			blocks++
		}
	}
	return blocks
}

// namespaceValueExport stops an export the object doesn't hold: anything with a value but a non-generic
// function with a body and one declaration. Private statements aren't on the object.
func (l *lowering) namespaceValueExport(member *ast.Node) error {
	if !ast.HasSyntacticModifier(member, ast.ModifierFlagsExport) {
		return nil
	}
	switch member.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return nil
	case ast.KindEnumDeclaration:
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsConst) {
			return nil
		}
		return l.notYet(member, "an exported runtime enum in a namespace used as a value")
	case ast.KindVariableStatement:
		return l.notYet(member, "an exported variable in a namespace used as a value")
	case ast.KindClassDeclaration:
		return l.notYet(member, "an exported class in a namespace used as a value")
	case ast.KindModuleDeclaration:
		if namespaceRuntime(member) {
			return l.notYet(member, "an exported runtime namespace in a namespace used as a value")
		}
		return nil
	case ast.KindFunctionDeclaration:
		if member.Body() == nil || len(l.symbol(member.Name()).Declarations) != 1 {
			return l.notYet(member, "an exported overloaded or bodiless function in a namespace used as a value")
		}
		if len(member.TypeParameters()) > 0 || generatorDeclaration(member) {
			return l.notYet(member, "an exported generic or generator function in a namespace used as a value")
		}
		if member.Name().Text() == "__proto__" {
			return l.notYet(member, "an exported function named __proto__ in a namespace used as a value")
		}
		return nil
	}
	return l.notYet(member, "an export, "+describe(member)+", in a namespace used as a value")
}

// namespaceOfType is the program namespace whose object a value of the type is, undefined aside, or nil.
func (l *lowering) namespaceOfType(proven *checker.Type) *ast.Node {
	proven = l.checker.GetNonNullableType(proven)
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	symbol := proven.Symbol()
	if symbol == nil || symbol.Flags&ast.SymbolFlagsValueModule == 0 || symbol.Flags&(ast.SymbolFlagsFunction|ast.SymbolFlagsClass|ast.SymbolFlagsEnum) != 0 || l.checker.GetTypeOfSymbol(symbol) != proven {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindModuleDeclaration && ast.IsIdentifier(declaration.Name()) && namespaceProgram(declaration) {
			return declaration
		}
	}
	return nil
}

// namespaceSlotKept reports whether a slot holds a namespace's object as its own type: typeof N, or
// typeof N | undefined, whose undefined is an Object slot's missing reference. null is another tag, and
// any other member is a view of the object.
func (l *lowering) namespaceSlotKept(declaration *ast.Node, slot *checker.Type) bool {
	if slot == nil {
		return false
	}
	members := []*checker.Type{slot}
	if slot.Flags()&checker.TypeFlagsUnion != 0 {
		members = slot.Types()
	}
	kept := false
	for _, member := range members {
		switch {
		case member.Flags()&checker.TypeFlagsUndefined != 0:
		case l.namespaceOfType(member) == declaration && member.Flags()&checker.TypeFlagsObject != 0:
			kept = true
		default:
			return false
		}
	}
	return kept
}

// namespaceValueUse judges where a value of a program namespace's type goes. It may be read through,
// compared, tested and held in a variable, parameter or return of its own type; anything else that could
// see the object as another shape, reflect on it, or write to it stops here, failing closed, so the
// object is only ever read for its fixed fields. A qualified read's receiver is main's namespace road.
func (l *lowering) namespaceValueUse(node *ast.Node) error {
	// Only a namespace whose name is read as a value has an object; a value of another namespace's type
	// is always undefined.
	if len(l.namespaceValues) == 0 {
		return nil
	}
	// { N } reads N through a name namespaceValueNode counts as a declaration's.
	if node.Kind == ast.KindShorthandPropertyAssignment {
		value := l.checker.GetShorthandAssignmentValueSymbol(node)
		if value != nil && value.Flags&ast.SymbolFlagsAlias != 0 {
			value = l.checker.GetAliasedSymbol(value)
		}
		if value != nil && value.Flags&ast.SymbolFlagsValueModule != 0 {
			for _, declaration := range value.Declarations {
				if declaration.Kind == ast.KindModuleDeclaration && namespaceProgram(declaration) {
					return l.notYet(node, "a namespace value seen as a field or an element")
				}
			}
		}
		return nil
	}
	if node.Kind == ast.KindParenthesizedExpression || !l.namespaceValueNode(node) || ast.IsAssignmentTarget(node) {
		return nil
	}
	declaration := l.namespaceOfType(l.checker.GetTypeAtLocation(node))
	if declaration == nil || !l.namespaceHasValue(declaration) || l.namespaceDeclaration(node) != nil && l.namespaceQualifiedReceiver(node) {
		return nil
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil {
		return nil
	}
	seen := func(what string) error {
		return l.notYet(node, "a namespace value seen as "+what)
	}
	kept := func(slot *checker.Type) error {
		if l.namespaceSlotKept(declaration, slot) {
			return nil
		}
		if slot == nil {
			return seen("a value with no declared type to keep it as its own")
		}
		return seen(l.checker.TypeToString(slot))
	}
	// The whole of a logical or a conditional holding the value is judged as its own value where its
	// type is the namespace's; where it isn't, only a test may read it.
	whole := func(holder *ast.Node) error {
		if l.namespaceOfType(l.checker.GetTypeAtLocation(holder)) == declaration || l.namespaceTestOnly(holder) {
			return nil
		}
		return seen(l.checker.TypeToString(l.checker.GetTypeAtLocation(holder)))
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return l.namespaceValueMember(outer, parent)
	case ast.KindElementAccessExpression:
		return seen("a receiver read by a computed name")
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return nil
		case ast.KindAmpersandAmpersandToken:
			if binary.Left == outer {
				return nil
			}
			return whole(parent)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return whole(parent)
		case ast.KindCommaToken:
			if binary.Left == outer {
				return nil
			}
			return whole(parent)
		case ast.KindEqualsToken:
			target := ast.SkipParentheses(binary.Left)
			if binary.Right != outer || !ast.IsIdentifier(target) {
				return seen("a field or an element")
			}
			return kept(l.checker.GetTypeOfSymbol(l.symbol(target)))
		case ast.KindInKeyword:
			return seen("the right side of in")
		}
		return seen("an operand of " + scanner.TokenToString(binary.OperatorToken.Kind))
	case ast.KindPrefixUnaryExpression:
		if parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			return nil
		}
		return seen("an operand of " + scanner.TokenToString(parent.AsPrefixUnaryExpression().Operator))
	case ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindExpressionStatement, ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		return nil
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == outer {
			return nil
		}
		return whole(parent)
	case ast.KindVariableDeclaration:
		if !ast.IsIdentifier(parent.Name()) {
			return seen("a destructuring pattern")
		}
		return kept(l.checker.GetTypeOfSymbol(l.symbol(parent.Name())))
	case ast.KindCallExpression, ast.KindNewExpression:
		return l.namespaceValueArgument(node, outer, parent, declaration)
	case ast.KindReturnStatement, ast.KindArrowFunction:
		return kept(l.slotType(outer))
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return kept(l.checker.GetTypeAtLocation(parent))
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return seen("a spread source")
	case ast.KindTemplateSpan:
		return seen("a string")
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindArrayLiteralExpression, ast.KindPropertyDeclaration:
		return seen("a field or an element")
	}
	return seen("part of " + describe(parent))
}

// namespaceValueArgument judges a namespace value passed to a call: only into a parameter of its own type,
// of a call whose signature isn't generic and isn't the library's, whose parameters could see anything.
// A constructor's argument may become a field, a parameter property, which a view of its instance could
// see as another shape, so it stops.
func (l *lowering) namespaceValueArgument(node, outer, call, declaration *ast.Node) error {
	seen := func(what string) error {
		return l.notYet(node, "a namespace value seen as "+what)
	}
	if call.Expression() == outer {
		return seen("a callee")
	}
	if call.Kind == ast.KindNewExpression {
		return seen("a constructor argument")
	}
	// A value a caller set a slot for (refuseInSlot) is written into a slot this can't follow, so it stops.
	if l.slotSet(outer) {
		return seen("a value written into a slot a caller set")
	}
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil {
		return seen("an argument to a call whose signature isn't known")
	}
	if librarySignature(signature) {
		callee := ast.SkipParentheses(call.Expression())
		if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.Expression(), "Object") {
			return seen("an argument to Object." + callee.Name().Text())
		}
		return seen("an argument to a library function")
	}
	if callee := signature.Declaration(); callee != nil && ast.IsFunctionLike(callee) && len(callee.TypeParameters()) > 0 {
		return seen("a generic argument")
	}
	// The parameter is judged as declared, never as instantiated: a generic class's method, or an interface
	// method reached through I<typeof N>, has no type parameters of its own, and its parameter's slot is
	// typeof N here while its body sees a type parameter, where a write through the value would lower
	// unjudged (Lower's must).
	declared := l.namespaceDeclaredParameter(call, outer, signature)
	if declared == nil {
		return seen("an argument with no declared parameter of its own type")
	}
	if l.mentionsTypeParameter(declared, map[*checker.Type]bool{}) {
		return seen("an argument to a parameter declared with a type parameter")
	}
	if l.namespaceSlotKept(declaration, declared) {
		return nil
	}
	return seen(l.checker.TypeToString(declared))
}

// namespaceDeclaredParameter is the type a call's argument's parameter is declared with in the callee's own
// declaration, uninstantiated, or nil when there's no one plain parameter for it (no declaration, a rest
// parameter, or more arguments than parameters).
func (l *lowering) namespaceDeclaredParameter(call, argument *ast.Node, signature *checker.Signature) *checker.Type {
	callee := signature.Declaration()
	if callee == nil || !ast.IsFunctionLike(callee) {
		return nil
	}
	index := -1
	for position, written := range call.Arguments() {
		if written == argument {
			index = position
		}
	}
	parameters := []*ast.Node{}
	for _, parameter := range callee.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			continue
		}
		parameters = append(parameters, parameter)
	}
	if index < 0 || index >= len(parameters) || parameters[index].AsParameterDeclaration().DotDotDotToken != nil || !ast.IsIdentifier(parameters[index].Name()) {
		return nil
	}
	symbol := l.symbol(parameters[index].Name())
	if symbol == nil {
		return nil
	}
	return l.checker.GetTypeOfSymbol(symbol)
}

// namespaceValueMember judges a read through a namespace value, receiver.name. A write is refused: the
// exported functions have fixed identity, and a qualified call, a direct call, wouldn't see it. An exported
// const enum is read only as Enum.Member through a chain of names, which lowering inlines as tsc's emit
// does; the object holds no enum.
func (l *lowering) namespaceValueMember(receiver, access *ast.Node) error {
	if ast.IsAssignmentTarget(access) {
		return &Refused{Where: l.program.Where(access), What: "a write to a namespace member through the namespace's value; its exported functions have fixed identity", Fix: "keep the function to replace in a variable of its own, and call through that"}
	}
	member := l.symbol(access)
	if member == nil || member.Flags&ast.SymbolFlagsConstEnum == 0 {
		return nil
	}
	read := access.Parent
	if access.AsPropertyAccessExpression().QuestionDotToken != nil || access.Flags&ast.NodeFlagsOptionalChain != 0 || read == nil || read.Kind != ast.KindPropertyAccessExpression || read.AsPropertyAccessExpression().Expression != access || read.Flags&ast.NodeFlagsOptionalChain != 0 || l.enumMember(read) == nil || !l.namespaceValueNamed(receiver) {
		return l.notYet(access, "a const enum read through a namespace value other than as receiver.Enum.Member through names alone")
	}
	return nil
}

// namespaceValueNamed reports whether a receiver is names and plain property reads alone, which an
// inlined const enum member may skip evaluating: no call, no optional link, no accessor.
func (l *lowering) namespaceValueNamed(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindPropertyAccessExpression:
		if node.AsPropertyAccessExpression().QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return false
		}
		if member := l.checker.GetSymbolAtLocation(node.Name()); member == nil || accessorSymbol(member) {
			return false
		}
		return l.namespaceValueNamed(node.Expression())
	}
	return false
}

// namespaceTestOnly reports whether a value is only tested for truthiness: a condition, the operand of !,
// or an operand of a logical that is itself only tested.
func (l *lowering) namespaceTestOnly(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement:
		return true
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			return parent.AsBinaryExpression().Left == node || l.namespaceTestOnly(parent)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return l.namespaceTestOnly(parent)
		}
	}
	return false
}
