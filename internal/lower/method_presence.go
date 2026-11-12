package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A method read whose only consumer tests for it is an observation, not a detach (#4q1b2er): the
// operand of typeof or of !, a condition (the test of if, while, do...while, for, and of ?:), and the
// left of &&, whose falsy left would be the whole value, while a method value is never falsy, so
// nothing detached escapes it. tsc probes its host this way (`typeof Object.setPrototypeOf ===
// "function"`, `if (array.fill)`, `if (ch.repeat)`). Every other consumer keeps the method value and
// stays refused as a method read as a value (methodValueRefusal): ||, ??, an arm of ?:, an assignment
// or an initializer, an argument, a return, an alias and a comma operand.
//
// Only a method each of whose declarations is a method declaration or signature with no question
// token is admitted, and its answer is a constant only where the receiver's construction proves the
// method there: a library global's library method (Object.setPrototypeOf), a library method on a value
// Adamic holds as its own array or string (array.fill, ch.repeat), a program class's method on a
// receiver of that class's type, and a method a host object's registration names on a receiver
// hostOrigin proves is that host object by how it was made (#q9nfg88, tsc's tryGetPerformanceHooks).
// Anything else, a method signature on a program's interface or object
// type above all, may be a checked view or a host value whose real object lacks the method, where Node
// says "undefined": it is not yet (methodPresenceUnproven), never folded. The library cases are true
// only because load pins every program to lib.es2024.d.ts (load.go's compilerOptions): a bundled
// declaration is then one Node 24 has, where lib.es2026's Math.sumPrecise and Uint8Array.fromBase64,
// say, are not (TestMethodPresenceLibraryStopsAtES2024 goes red if the pin moves). And all three rest
// on the stops that keep a method from being taken off such a receiver, Object.defineProperty,
// Object.setPrototypeOf, Object.assign, Reflect.set, a write and a delete
// (TestMethodPresenceRestsOnTheseStops). The constant comes once the
// receiver is evaluated, once, for its effects. Through ?. the receiver may be missing, and its
// presence is the answer. An optional method (m?(): T) is never a
// constant, since nothing proves it there: its presence test is read off its slot as it runs, the way
// an optional function-typed field's is (property), and stops where its receiver may hold anything but
// a plain object literal (optionalMemberPresenceTest, #p2n17c4 shape 1).

// methodPresenceTest reports whether node reads a method that is always there only to test for it.
// A read the library observations already admit (typeof String.raw, typeof Array.prototype.map,
// typeof performance.now) keeps the lowering it has.
func (l *lowering) methodPresenceTest(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression || !l.alwaysPresentMethod(node) || !presenceConsumer(node) {
		return false
	}
	return !l.stringMethodObservation(node) && !l.libraryArrayObservedMethod(node) && !l.nodeProcessMethodObservation(node)
}

// alwaysPresentMethod reports whether a property access reads a method every declaration of which is a
// method declaration or signature that isn't optional.
func (l *lowering) alwaysPresentMethod(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodDeclaration && declaration.Kind != ast.KindMethodSignature || ast.HasQuestionToken(declaration) {
			return false
		}
	}
	return true
}

// presenceConsumer reports whether what consumes node, through any parentheses around it, only tests
// it: typeof, !, a condition, or the left of &&.
func presenceConsumer(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindTypeOfExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindBinaryExpression:
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Left == node
	}
	return false
}

// methodPresenceRead is what testing for a method comes to: the receiver's evaluation, kept for its
// effects, and, read through ?., whether the receiver is there, which is then the answer. present is
// nil when the method is always there.
type methodPresenceRead struct {
	effects []ir.Statement
	present ir.Expression
}

// then is result after the receiver's effects.
func (read methodPresenceRead) then(result ir.Expression) ir.Expression {
	if len(read.effects) == 0 {
		return result
	}
	return ir.Effects{Body: read.effects, Result: result}
}

// methodPresence lowers operand's receiver when operand is a method read only to test for it; handled
// is false for any other operand, which its consumer lowers as it always has.
func (l *lowering) methodPresence(operand *ast.Node) (read methodPresenceRead, handled bool, err error) {
	node := ast.SkipParentheses(operand)
	if !l.methodPresenceTest(node) {
		return methodPresenceRead{}, false, nil
	}
	access := node.AsPropertyAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it, as property stops it.
		return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method at the end of an optional chain longer than one step")
	}
	if receiver.Kind == ast.KindSuperKeyword {
		return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method read through super")
	}
	library := l.libraryDeclaredMethod(node)
	if access.QuestionDotToken == nil && l.libraryGlobalReceiver(receiver) {
		if !library {
			return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
		}
		// Object, Array, Math: reading a library global has no effects, and Object isn't a value Adamic
		// holds (libraryGlobalValue), so there is nothing to evaluate.
		return methodPresenceRead{}, true, nil
	}
	class := l.programClassMethod(node)
	// A method declared on a program's interface is there when every value the receiver can hold is
	// a host object whose registration names it as a method, as a class's is on its instance.
	host := !class && !library && l.hostOriginMethod(node)
	if !class && !library && !host {
		return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return methodPresenceRead{}, true, err
	}
	if !class && !host && value.Type() != ir.Array && value.Type() != ir.String {
		// A library method on an object Adamic didn't build as its own array or string: a Map, a host
		// value, a view. Only the object itself can say whether it has the method.
		return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
	}
	if access.QuestionDotToken == nil {
		return methodPresenceRead{effects: []ir.Statement{ir.Evaluate{Value: value}}}, true, nil
	}
	switch value.Type() {
	case ir.Object, ir.Array:
		// An object or an array is falsy only when it's missing (combine's && says why), so its
		// truthiness is ?.'s nullish test, on the receiver's one evaluation.
		return methodPresenceRead{present: ir.Truthy{Value: value}}, true, nil
	}
	return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method read through ?. on a "+typeName(value.Type()))
}

// methodPresenceUnproven is what a presence test stops as where nothing about how the receiver was made
// proves the method there.
const methodPresenceUnproven = "a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it"

// libraryDeclaredMethod reports whether every declaration of the method node reads is TypeScript's
// bundled library's.
func (l *lowering) libraryDeclaredMethod(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return false
		}
	}
	return true
}

// programClassMethod reports whether node reads, off a receiver whose type is one of the program's own
// classes, a method each declaration of which is a class's method with a body: what callOrMethod calls
// directly on such a receiver, by the class's own method. A value of a class type is never a view (every
// view's target is an object type no class is, interfaceCast to speculation), and a cast to one is
// checked by the class's identity (checkedClassCast), so such a receiver is an instance, and the method
// is on its prototype. this is read as its class, the constraint of its polymorphic this type.
func (l *lowering) programClassMethod(node *ast.Node) bool {
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	proven := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	if receiver.Kind == ast.KindThisKeyword && proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	if !isClassInstance(proven) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		class := declaration.Parent
		if declaration.Kind != ast.KindMethodDeclaration || declaration.Body() == nil || class == nil || class.Name() == nil {
			return false
		}
		if _, known := l.classes[l.symbol(class.Name())]; !known {
			return false
		}
	}
	return true
}

// tsc's tryGetPerformanceHooks tests `typeof performance.now === "function"` on a local it destructures
// from tryGetPerformance's literal, which holds the routed performance binding or node:perf_hooks'
// export, while the checker types it by tsc's own Performance interface (#q9nfg88). That type proves
// nothing, since any program object can have it, so hostOrigin walks back from the receiver along the
// few roads that can only carry the host object, and every other road is unknown, which keeps
// methodPresenceUnproven. Whether the method is there is read from the model's registration, never
// assumed from the type (hostOriginMethod).
//
// The fold also rests on nothing taking the method off the host object, as the class case rests on its
// stops: a write to now through the binding or any local typed with it is a method read as a value
// (methodValueRefusal), Object.defineProperty on it is refused (objectCallArguments; the one form
// admitted is a record entry's, recordDefineShape), and a write through a structural view stops at
// nodeProcessStorageABI.

// hostOriginMethod reports whether node, a property access, reads off a receiver hostOrigin proves is a
// registered host object a member that object's registration names and that is a method there, with
// call signatures: never a field such as timeOrigin, whose typeof Node answers "number" (Language's
// condition on #q9nfg88). An optional member may be missing, so it isn't one.
func (l *lowering) hostOriginMethod(node *ast.Node) bool {
	model := l.hostOrigin(node.AsPropertyAccessExpression().Expression, nil, map[*ast.Node]bool{})
	registered, host := l.nodeHostObject(model)
	if !host {
		return false
	}
	member := l.checker.GetPropertyOfType(model, node.Name().Text())
	return member != nil && member.Flags&ast.SymbolFlagsOptional == 0 && registered(member) && len(l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(member), checker.SignatureKindCall)) > 0
}

// hostOrigin is the model export's type every value node can hold is, once the property names in path
// are read off it in turn, or nil, unknown, where the walk can't prove one. It's syntactic and bounded:
// each node is met once (visited), so a cycle through consts or a recursive function answers unknown.
// The roads, each failing closed:
//  1. a name that resolves, by its symbol, to a routed host binding: its model export. A parameter or
//     local of the same name is that, never the binding;
//  2. a namespace import of an ambient module a Node library file declares, read by one name whose
//     export is registered and a host object: node:perf_hooks' performance, through
//     `import * as perfHooks` and `const { performance } = perfHooks` (hostOriginExport);
//  3. a const, or a plain element of a const's object destructuring: its initializer, or that
//     initializer's property by the element's property name, so a rename follows the property
//     (hostOriginBinding);
//  4. an object literal read by a name it writes exactly once, plainly: that property's value
//     (hostOriginProperty);
//  5. a call, read by a name, to a plain function declaration: each of its own returns, nullish or a
//     literal, all agreeing (hostOriginReturned);
//  6. anything else is unknown: a let, a parameter, a cast, a call through a value, an await, a
//     property of a value the walk didn't follow to a literal.
func (l *lowering) hostOrigin(node *ast.Node, path []string, visited map[*ast.Node]bool) *checker.Type {
	node = ast.SkipParentheses(node)
	if visited[node] {
		return nil
	}
	visited[node] = true
	switch node.Kind {
	case ast.KindIdentifier:
		if row, _ := l.hostBindingOf(node); row != nil && l.hostBindingRouted(*row) {
			// The walk never follows a host object's own properties.
			if model := l.hostBindingModel(*row); len(path) == 0 {
				if _, host := l.nodeHostObject(model); host {
					return model
				}
			}
			break
		}
		if model := l.hostOriginExport(node, path); model != nil {
			return model
		}
		if value, read := l.hostOriginBinding(node, path); value != nil {
			return l.hostOrigin(value, read, visited)
		}
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access.QuestionDotToken == nil && ast.IsIdentifier(access.Name()) {
			return l.hostOrigin(access.Expression, append([]string{access.Name().Text()}, path...), visited)
		}
	case ast.KindObjectLiteralExpression:
		if value := hostOriginProperty(node, path); value != nil {
			return l.hostOrigin(value, path[1:], visited)
		}
	case ast.KindCallExpression:
		if model := l.hostOriginReturned(node, path, visited); model != nil {
			return model
		}
	}
	return nil
}

// hostOriginSymbol is the symbol a name reads: for a shorthand property, the value it's written from,
// not the property it writes.
func (l *lowering) hostOriginSymbol(node *ast.Node) *ast.Symbol {
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
		return l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	return l.checker.GetSymbolAtLocation(node)
}

// hostOriginExport is hostOrigin's second road: node names a namespace import of a Node library's
// ambient module, and path is one name, whose export is registered and a host object. It's that
// export's type, or nil.
func (l *lowering) hostOriginExport(node *ast.Node, path []string) *checker.Type {
	if len(path) != 1 {
		return nil
	}
	symbol := l.hostOriginSymbol(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias == 0 || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindNamespaceImport {
		return nil
	}
	exported := l.nodeModuleExport(l.checker.GetAliasedSymbol(symbol), path[0])
	if exported == nil || !nodeLibraryMemberImplemented(exported) {
		return nil
	}
	model := l.checker.GetTypeOfSymbol(exported)
	if _, host := l.nodeHostObject(model); !host {
		return nil
	}
	return model
}

// hostOriginBinding is hostOrigin's third road: what a const's name holds, as the expression it was
// initialized with and the names to read off that, or nil. A let or a var can be written, an element
// with a default can hand over the default instead, and a rest element is a new object, so each is
// unknown.
func (l *lowering) hostOriginBinding(node *ast.Node, path []string) (*ast.Node, []string) {
	symbol := l.hostOriginSymbol(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil
	}
	declaration := symbol.Declarations[0]
	var value *ast.Node
	read := path
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		if hostOriginConst(declaration) && ast.IsIdentifier(declaration.Name()) {
			value = declaration.Initializer()
		}
	case ast.KindBindingElement:
		element := declaration.AsBindingElement()
		pattern := declaration.Parent
		key := hostOriginKey(declaration.PropertyNameOrName())
		if element.Initializer == nil && element.DotDotDotToken == nil && ast.IsIdentifier(declaration.Name()) && key != "" && pattern.Kind == ast.KindObjectBindingPattern && hostOriginConst(pattern.Parent) {
			value, read = pattern.Parent.Initializer(), append([]string{key}, path...)
		}
	}
	if value == nil {
		return nil, nil
	}
	// The const guards the binding: nothing writes it, so it holds its initializer's value for good. It
	// doesn't guard a property of what it holds, which another reference to that object could write
	// (Compiler's must on #q9nfg88: `p.performance = own` between the call and the destructuring). The
	// no-escape rule guards that, for every object the walk reads a property off (hostOriginContained).
	// The value the walk ends at is the host object, whose methods its own stops guard (see above).
	if len(path) > 0 && !l.hostOriginContained(declaration, symbol) {
		return nil, nil
	}
	return value, read
}

// hostOriginConst reports whether a declaration is a const's with an initializer, in a variable
// statement: not a let's, a var's, a using's or a loop's.
func hostOriginConst(declaration *ast.Node) bool {
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() == nil {
		return false
	}
	list := declaration.Parent
	return list != nil && list.Kind == ast.KindVariableDeclarationList && list.Flags&ast.NodeFlagsBlockScoped == ast.NodeFlagsConst && list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement
}

// hostOriginKey is a property name written out, as an identifier or a string, or "" for a computed or
// numeric one, which the walk doesn't read.
func hostOriginKey(name *ast.Node) string {
	if name != nil && (name.Kind == ast.KindIdentifier || name.Kind == ast.KindStringLiteral) {
		return name.Text()
	}
	return ""
}

// hostOriginContained is the no-escape rule (Lower's, for Compiler's must on #q9nfg88): a const holding
// an object the walk reads a property off counts only if every use of it in its file is one that can't
// write that property or hand the object to something that could: a truthiness or nullish test, a
// const's plain object destructuring of it, or a property read that is the whole of a const's
// initializer. Anything else escapes: an argument, a return, an assignment, a property write, a spread,
// an alias, a method call, a template, typeof, a loose ==, for...in, an export, or a capture by a
// closure, which can run when anything has happened to it. Every const the walk reads through is
// fresh from a literal the walk follows, so these are its only references.
func (l *lowering) hostOriginContained(declaration *ast.Node, symbol *ast.Symbol) bool {
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return false
	}
	scope := hostOriginScope(declaration)
	name := declaration.Name()
	escaped := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node != name && node.Text() == name.Text() && !ast.IsPartOfTypeNode(node) && l.hostOriginReads(node, symbol) && !hostOriginContainedUse(node, scope) {
			escaped = true
			return true
		}
		return node.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return !escaped
}

// hostOriginReads reports whether an identifier reads symbol: by name, as a shorthand property's value,
// or as an export specifier's local.
func (l *lowering) hostOriginReads(node *ast.Node, symbol *ast.Symbol) bool {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindExportSpecifier && l.checker.GetExportSpecifierLocalTargetSymbol(parent) == symbol {
		return true
	}
	return l.hostOriginSymbol(node) == symbol || l.checker.GetSymbolAtLocation(node) == symbol
}

// hostOriginScope is the function a node is in, or its file.
func hostOriginScope(node *ast.Node) *ast.Node {
	for node = node.Parent; node != nil && node.Kind != ast.KindSourceFile && !ast.IsFunctionLike(node) && node.Kind != ast.KindClassStaticBlockDeclaration; node = node.Parent {
	}
	return node
}

// hostOriginContainedUse reports whether one reference to a const is a use hostOriginContained allows.
func hostOriginContainedUse(reference, scope *ast.Node) bool {
	if hostOriginScope(reference) != scope {
		return false
	}
	use := reference
	for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
		use = use.Parent
	}
	parent := use.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == use
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == use
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == use
	case ast.KindBinaryExpression:
		return comparedWithUndefined(reference)
	case ast.KindVariableDeclaration:
		return parent.Initializer() == use && hostOriginConst(parent) && hostOriginDestructuring(parent.Name())
	case ast.KindPropertyAccessExpression:
		if parent.AsPropertyAccessExpression().Expression != use || parent.AsPropertyAccessExpression().QuestionDotToken != nil {
			return false
		}
		read := parent
		for read.Parent != nil && read.Parent.Kind == ast.KindParenthesizedExpression {
			read = read.Parent
		}
		return read.Parent != nil && read.Parent.Kind == ast.KindVariableDeclaration && read.Parent.Initializer() == read && hostOriginConst(read.Parent) && ast.IsIdentifier(read.Parent.Name())
	}
	return false
}

// hostOriginDestructuring reports whether a binding name is an object pattern each element of which
// reads one property by a name written out, with no default, rest or nested pattern: it only reads.
func hostOriginDestructuring(name *ast.Node) bool {
	if name.Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, element := range name.Elements() {
		binding := element.AsBindingElement()
		if binding.Initializer != nil || binding.DotDotDotToken != nil || !ast.IsIdentifier(element.Name()) || hostOriginKey(element.PropertyNameOrName()) == "" {
			return false
		}
	}
	return true
}

// hostOriginProperty is hostOrigin's fourth road: the value an object literal writes for path's first
// name, or nil. That property must be written exactly once, as a plain property or a shorthand, since a
// later write of the same name wins. And the literal must hold no spread, which can override it, no
// computed key, which can name it, and no accessor on any property whatever its name: destructuring
// reads properties in order, so a getter on another one runs first and can write this one through this
// (Language's must on #q9nfg88, run on Node). A method is fine: it runs only when called on the object,
// which is an escape.
func hostOriginProperty(literal *ast.Node, path []string) *ast.Node {
	if len(path) == 0 {
		return nil
	}
	var value *ast.Node
	written := 0
	for _, member := range literal.Properties() {
		switch member.Kind {
		case ast.KindSpreadAssignment, ast.KindGetAccessor, ast.KindSetAccessor:
			return nil
		}
		key := hostOriginKey(member.Name())
		if key == "" {
			return nil
		}
		if key != path[0] {
			continue
		}
		written++
		value = nil
		switch member.Kind {
		case ast.KindPropertyAssignment:
			value = member.Initializer()
		case ast.KindShorthandPropertyAssignment:
			if member.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == nil {
				value = member.Name()
			}
		}
	}
	if written != 1 {
		return nil
	}
	return value
}

// hostOriginReturned is hostOrigin's fifth road: a call read by a name, to a plain function declaration
// (hostOriginFunction), is what each of the function's own returns gives, nullish or a literal, all
// agreeing on one model. ForEachReturnStatement doesn't enter a nested function or arrow, whose returns
// are its own, and reads a return in a finally, which overrides the try's. A nullish return is passed
// over because a name is read off the call's result, which throws there. It's nil where any return is
// anything else, or none is a literal.
func (l *lowering) hostOriginReturned(node *ast.Node, path []string, visited map[*ast.Node]bool) *checker.Type {
	function := l.hostOriginFunction(node)
	if function == nil || len(path) == 0 {
		return nil
	}
	var model *checker.Type
	agreed := true
	ast.ForEachReturnStatement(function.Body(), func(statement *ast.Node) bool {
		returned := statement.Expression()
		if returned == nil || l.hostOriginNullish(returned) {
			return false
		}
		var origin *checker.Type
		if ast.SkipParentheses(returned).Kind == ast.KindObjectLiteralExpression {
			origin = l.hostOrigin(returned, path, visited)
		}
		if origin == nil || model != nil && origin != model {
			agreed = false
			return true
		}
		model = origin
		return false
	})
	if !agreed {
		return nil
	}
	return model
}

// hostOriginFunction is the declaration a call calls by name when it's a plain function: one
// declaration, so not overloaded, with a body, no type parameters, and neither async, whose result is a
// promise, nor a generator, whose result is an iterator, never its return value. A method, an arrow or a
// function expression is a value, which the walk doesn't follow. Nil for anything else.
func (l *lowering) hostOriginFunction(node *ast.Node) *ast.Node {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || call.TypeArguments != nil || !ast.IsIdentifier(callee) {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(callee)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	function := symbol.Declarations[0]
	if function.Kind != ast.KindFunctionDeclaration || function.Body() == nil || len(function.TypeParameters()) != 0 || ast.GetFunctionFlags(function) != ast.FunctionFlagsNormal {
		return nil
	}
	return function
}

// hostOriginNullish reports whether an expression is null or the global undefined, not a local that
// shadows its name.
func (l *lowering) hostOriginNullish(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return true
	}
	if !ast.IsIdentifier(node) || node.Text() != "undefined" {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol == nil || len(symbol.Declarations) == 0
}

// libraryGlobalReceiver reports whether node names a library global, as isLibraryGlobal does for one
// name.
func (l *lowering) libraryGlobalReceiver(node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// methodPresenceTypeOf lowers typeof of a method read only to test for it: "function", or through ?.
// "undefined" where the receiver is missing.
func (l *lowering) methodPresenceTypeOf(node *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(node.AsTypeOfExpression().Expression)
	if !handled || err != nil {
		return nil, handled, err
	}
	function := ir.StringConstant{Index: l.constant("function")}
	if read.present == nil {
		return read.then(function), true, nil
	}
	return ir.Conditional{Condition: read.present, WhenTrue: function, WhenNot: ir.StringConstant{Index: l.constant("undefined")}}, true, nil
}

// methodPresenceCondition lowers a condition that is a method read only to test for it: true, or
// through ?. whether the receiver is there.
func (l *lowering) methodPresenceCondition(node *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(node)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present == nil {
		return read.then(ir.BooleanConstant{Value: true}), true, nil
	}
	return read.present, true, nil
}

// methodPresenceNot lowers ! of a method read only to test for it: false, or through ?. whether the
// receiver is missing.
func (l *lowering) methodPresenceNot(operand *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(operand)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present == nil {
		return read.then(ir.BooleanConstant{Value: false}), true, nil
	}
	return ir.Unary{Operator: ir.Not, Operand: read.present}, true, nil
}

// methodPresenceAnd lowers left && right whose left is a method read only to test for it: right's
// value, after the receiver's effects.
func (l *lowering) methodPresenceAnd(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		// || and ?? keep the method value: it is their whole value, never right's.
		return nil, false, nil
	}
	read, handled, err := l.methodPresence(binary.Left)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present != nil {
		// The receiver's undefined would be the whole value, a slot right's representation may not hold.
		return nil, true, l.notYet(node, "the left of && reading a method through ?., whose missing receiver would be the value")
	}
	right, err := l.expression(binary.Right)
	if err != nil {
		return nil, true, err
	}
	return read.then(right), true, nil
}
