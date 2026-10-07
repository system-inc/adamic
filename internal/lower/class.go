package lower

import (
	"sort"
	"strings"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// instance is one instantiation of a class, lowered: its constructor, and its methods by name.
//
// Each instantiation has one nominal identity and one prefix layout. A virtual call keeps its
// static signature while selecting the implementation from the object's dynamic table.
type instance struct {
	static         bool
	constructor    int
	methods        map[string]int
	slots          map[string]int
	class          int
	base           *instance
	initializer    int
	superReady     int
	unreadyThis    map[*ast.Node]bool
	superStates    map[*ast.Node]uint8
	constructing   bool
	hasDescendants bool
	// staticMethods are the methods declared static, which no object of the class has.
	staticMethods map[string]bool

	// thisLocals are the this of its constructor and methods, which every instantiation sharing it
	// records at its call sites, for the cycle finder.
	thisLocals []int

	// templates are the instance's locals and function values with their types as the source writes
	// them, so an instantiation that shares the instance notes each as its own type makes it.
	templates []template
}

// methodList is the instance's methods, static ones aside, by name in order, so the layouts made of them are the same
// every time.
func (lowered *instance) methodList() []ir.Method {
	names := []string{}
	for name := range lowered.methods {
		if !lowered.staticMethods[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	methods := []ir.Method{}
	for _, name := range names {
		methods = append(methods, ir.Method{Name: name, Function: lowered.methods[name]})
	}
	return methods
}

// template is a local, or a function value (closure set), and its type with type parameters left in.
type template struct {
	local, closure int
	proven         *checker.Type
	isClosure      bool
}

// instantiate lowers a class for the type arguments of a type of it, once per distinct set.
func (l *lowering) instantiate(declaration *ast.Node, classType *checker.Type, where *ast.Node) (*instance, error) {
	// Box<T> inside Maker<Node> is a Box<Node>: what it is, for the cycle finder.
	classType = l.concrete(classType)
	if classType == nil || !isClassInstance(classType) {
		return nil, l.notYet(where, "instantiating a class from its constructor type instead of an instance type")
	}
	if view := l.classView(classType, declaration); view != nil {
		classType = view
	} else {
		return nil, l.notYet(where, "a method receiver without a nominal class view")
	}
	arguments := []ir.Type{}
	substitution := map[*checker.Type]ir.Type{}
	parameters := declaration.TypeParameters()
	var typeArguments []*checker.Type
	if len(parameters) > 0 {
		typeArguments = l.checker.GetTypeArguments(classType)
	}
	if len(parameters) != len(typeArguments) {
		return nil, l.notYet(where, "a class whose type arguments aren't known here")
	}
	for index, parameter := range parameters {
		representation, isKnown := l.representation(typeArguments[index])
		if !isKnown {
			return nil, l.notYet(where, "a class instantiated with "+l.checker.TypeToString(typeArguments[index]))
		}
		substitution[l.checker.GetTypeAtLocation(parameter.Name())] = representation
		arguments = append(arguments, representation)
	}
	key := l.classInstanceKey(declaration, typeArguments)
	mapper := l.typeMapperOf(declaration, classType)
	if err := l.nominalTypeArguments(declaration, typeArguments, mapper, where); err != nil {
		return nil, err
	}
	l.instantiated = append(l.instantiated, classType)
	if existing, isLowered := l.instances[key]; isLowered {
		for _, this := range existing.thisLocals {
			l.noteAlso(this, classType)
		}
		// The instance is shared, and so are its locals and function values: each is also what this
		// instantiation makes it.
		outerMapper := l.typeMapper
		l.typeMapper = mapper
		for _, shared := range existing.templates {
			if shared.isClosure {
				l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(shared.proven), function: shared.closure, node: l.closureNodeOf(shared.closure)})
			} else {
				l.noteAlso(shared.local, l.concrete(shared.proven))
			}
		}
		l.typeMapper = outerMapper
		return existing, nil
	}

	if len(parameters) > 0 {
		if l.genericDepth >= maximumGenericDepth {
			return nil, &Refused{Where: l.program.Where(where), What: "a generic class instantiated without end (polymorphic recursion)", Fix: "keep recursive type arguments unchanged, or write a class per type (adamic/polymorphic-recursion)"}
		}
		l.genericDepth++
		defer func() { l.genericDepth-- }()
	}

	base, err := l.baseInstance(declaration, classType)
	if err != nil {
		return nil, err
	}
	if err := l.checkOverrides(declaration, classType); err != nil {
		return nil, err
	}

	// Register every function of the instantiation first, so methods can find each other.
	name := declaration.Name().Text()
	for _, argument := range arguments {
		name += "_" + typeName(argument)
	}
	lowered := &instance{constructor: len(l.result.Functions), methods: map[string]int{}, staticMethods: map[string]bool{}, slots: map[string]int{}, initializer: -1, superReady: -1}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_new", Returns: ir.Object})
	lowered.hasDescendants = l.derivedAncestors[l.symbol(declaration.Name())]
	lowered.base = base
	lowered.class = len(l.result.Classes) + 1
	metadata := ir.Class{Name: name, Constructor: lowered.constructor, Definition: l.classDefinition(declaration)}
	if base != nil {
		metadata.Base = base.class
		metadata.Fields = append(metadata.Fields, l.result.Classes[base.class-1].Fields...)
		metadata.Methods = append(metadata.Methods, l.result.Classes[base.class-1].Methods...)
		for method, function := range base.methods {
			lowered.methods[method] = function
			lowered.staticMethods[method] = base.staticMethods[method]
		}
		for method, slot := range base.slots {
			lowered.slots[method] = slot
		}
	}
	metadata.OwnStart = len(metadata.Fields)
	l.result.Classes = append(l.result.Classes, metadata)
	members := declaration.Members()
	for _, member := range members {
		switch member.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if !ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier && !(member.Kind == ast.KindMethodDeclaration && member.Name().Kind == ast.KindComputedPropertyName && l.symbolIterator(member.Name().AsComputedPropertyName().Expression)) {
				return nil, l.notYet(member, "a method with a computed name (give the method a fixed identifier name)")
			}
			if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				continue
			}
			methodName := l.methodKey(member, lowered.class)
			if accessorMember(member) {
				l.registerAccessor(lowered.class, member, len(l.result.Functions))
			}
			lowered.methods[methodName] = len(l.result.Functions)
			lowered.staticMethods[methodName] = ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic)
			if !lowered.staticMethods[methodName] {
				meta := &l.result.Classes[lowered.class-1]
				slot, exists := lowered.slots[methodName]
				if !exists {
					slot = len(meta.Methods)
					lowered.slots[methodName] = slot
					meta.Methods = append(meta.Methods, -1)
				}
				meta.Methods[slot] = len(l.result.Functions)
			}
			l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_" + methodName})
		case ast.KindPropertyDeclaration, ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		default:
			return nil, l.notYet(member, describe(member)+" in a class")
		}
	}
	if l.instances == nil {
		l.instances = map[string]*instance{}
	}
	l.instances[key] = lowered

	// Lower with this instantiation's meaning of each type parameter, and locals of its own: the
	// same parameter is a string here and may be a number in another instantiation.
	outerSubstitution, outerLocals, outerInstance := l.substitution, l.locals, l.instance
	outerClassType, outerClassNode, outerMapper := l.classType, l.classNode, l.typeMapper
	l.substitution, l.instance = substitution, lowered
	l.classType, l.classNode, l.typeMapper = classType, declaration.Name(), mapper
	defer func() { l.classType, l.classNode, l.typeMapper = outerClassType, outerClassNode, outerMapper }()
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	defer func() { l.substitution, l.locals, l.instance = outerSubstitution, outerLocals, outerInstance }()

	// Every method's signature is written before any body is lowered, the constructor's included, so a
	// method can call one declared below it, and two can call each other.
	for _, member := range members {
		if !classFunction(member) || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		method := lowered.methods[l.methodKey(member, lowered.class)]
		if err := l.signature(method, member, l.thisLocal(method)); err != nil {
			return nil, err
		}
		if err := l.validateAccessorSignature(member, method); err != nil {
			return nil, err
		}
	}
	if err := l.constructor(lowered.constructor, declaration); err != nil {
		return nil, err
	}
	for _, member := range members {
		if !classFunction(member) || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		method := lowered.methods[l.methodKey(member, lowered.class)]
		if member.Body() == nil {
			continue
		}
		if err := l.lowerFunction(method, member, -1); err != nil {
			return nil, err
		}
	}
	return lowered, nil
}

// thisLocal makes a fresh local for the this of the method or constructor at owner.
func (l *lowering) thisLocal(owner int) int {
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: owner})
	l.noteLocal(len(l.result.Locals)-1, l.classType, l.classNode)
	l.instance.thisLocals = append(l.instance.thisLocals, len(l.result.Locals)-1)
	return len(l.result.Locals) - 1
}

// constructor lowers a class's construction: an object with the class's fields, each from its
// initializer (or its type's zero until the constructor body assigns it, which the checker requires),
// then the constructor's body, which returns the object.
func (l *lowering) constructor(index int, declaration *ast.Node) error {
	return l.inheritanceConstructor(index, declaration)
}

// zeroValue is a field's value before its constructor assigns it.
func zeroValue(of ir.Type) ir.Expression {
	switch of {
	case ir.Number:
		return ir.NumberConstant{}
	case ir.Boolean:
		return ir.BooleanConstant{}
	case ir.MaybeNumber:
		// A field of number | undefined left without a value is undefined, as JavaScript leaves it.
		return ir.MaybeOf{Of: ir.MaybeNumber}
	}
	return ir.Undefined{}
}

func containsThis(node *ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if child.Kind == ast.KindThisKeyword {
			found = true
			return true
		}
		return child.ForEachChild(visit)
	}
	if node.Kind == ast.KindThisKeyword {
		return true
	}
	node.ForEachChild(visit)
	return found
}

// construct lowers new Class(...).
func (l *lowering) construct(node *ast.Node, declaration *ast.Node) (ir.Expression, error) {
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract) {
		return nil, &Refused{Where: l.program.Where(node), What: "new of an abstract class", Fix: "construct a concrete subclass that implements its abstract methods (adamic/concrete-construction)"}
	}
	lowered, err := l.instantiate(declaration, l.checker.GetTypeAtLocation(node), node)
	if err != nil {
		return nil, err
	}
	arguments := []ir.Expression{}
	if node.AsNewExpression().Arguments != nil {
		for _, argument := range node.AsNewExpression().Arguments.Nodes {
			value, err := l.expression(argument)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, value)
		}
	}
	return l.staticConstruct(node, declaration, lowered.constructor, arguments), nil
}

// callOrMethod lowers a call to one of the module's functions, or to a method of one of its classes.
func (l *lowering) callOrMethod(node *ast.Node) (ir.Expression, error) {
	if value, handled, err := l.classGenericCall(node); handled {
		return value, err
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindSuperKeyword {
		return l.superCall(node)
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return l.call(node)
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	receiverType := l.concrete(l.checker.GetTypeAtLocation(receiver))
	method := l.checker.GetSymbolAtLocation(callee)
	if isClassInstance(receiverType) {
		if actual := l.checker.GetPropertyOfType(receiverType, callee.Name().Text()); actual != nil {
			method = actual
		}
	}
	if len(l.staticGlobals) > 0 && method != nil && len(method.Declarations) > 0 && method.Declarations[0].Kind == ast.KindMethodSignature {
		return nil, l.notYet(node, "a method call through a structural signature in a program with statics; use typeof the declaring class")
	}
	if method == nil || len(method.Declarations) == 0 || method.Declarations[0].Kind != ast.KindMethodDeclaration {
		return l.call(node)
	}
	class := method.Declarations[0].Parent
	if class == nil || class.Name() == nil {
		return nil, l.notYet(node, "a method of an anonymous class")
	}
	declaration, isClass := l.classes[l.symbol(class.Name())]
	if !isClass {
		return nil, l.notYet(node, "a method of a class stage 0 doesn't have")
	}
	if len(l.definedMembers(receiverType)) > 1 {
		return nil, l.notYet(receiver, "a method call through a union of class types; narrow with instanceof first")
	}
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		// object?.method(...), the object perhaps undefined: the class is instantiated for what the
		// object is when it's there, and the call is made through the object's methods, which stops
		// at undefined as JavaScript's chain does (ir.Property's Method).
		if _, err := l.instantiate(declaration, l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)), callee); err != nil {
			return nil, err
		}
		return l.callClosure(node)
	}
	// this.method() is the instantiation being lowered: this is the polymorphic this type there,
	// which carries no type arguments of its own.
	lowered := l.instance
	if ast.HasSyntacticModifier(method.Declarations[0], ast.ModifierFlagsStatic) {
		var err error
		if lowered == nil || !lowered.static || (ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword && ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword) {
			lowered, err = l.staticInstance(l.staticClass(receiverType, declaration))
			if err != nil {
				return nil, err
			}
		}
	} else if (ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword && ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword) || lowered == nil {
		var err error
		if lowered, err = l.instantiate(declaration, receiverType, callee); err != nil {
			return nil, err
		}
	}
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		if l.instance == nil || l.instance.base == nil {
			return nil, l.notYet(node, "super outside a derived class")
		}
		lowered = l.instance.base
	}
	var object ir.Expression
	var err error
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		if err = l.useOfThis(ast.SkipParentheses(receiver)); err == nil {
			l.touch(l.this)
			object = ir.Read{Local: l.this, Of: ir.Object}
		}
	} else {
		object, err = l.expression(receiver)
	}
	if err != nil {
		return nil, err
	}
	object = l.privateStaticReceiver(callee.Name(), object, false)
	arguments := []ir.Expression{object}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}
	function := lowered.methods[l.fieldName(callee.Name())]
	virtual := lowered.slots[l.fieldName(callee.Name())] + 1
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		virtual = 0
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns, Virtual: virtual}, nil
}

// setProperty lowers object.name = value, as a statement.
func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	uninitialized := l.uninitializedInitializer(valueNode)
	if member := l.checker.GetSymbolAtLocation(target.Name()); uninitialized && member != nil && accessorSymbol(member) {
		return nil, l.notYet(target, "deinitializing an accessor property")
	}
	if call, handled, err := l.superAccessor(target, valueNode); handled {
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	}
	if member := l.checker.GetSymbolAtLocation(target); member != nil && member.Flags&ast.SymbolFlagsMethod != 0 {
		return nil, l.notYet(target, "replacing a represented method at runtime (store a replaceable arrow function in a declared callback field instead)")
	}
	if member := l.checker.GetSymbolAtLocation(target); member != nil && member.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)) {
		if stored, _ := l.representation(l.checker.GetTypeOfSymbol(member)); stored == ir.Union || (stored == ir.MaybeBoolean && !l.presentLiteralField(target)) || (stored.IsReference() && l.absentLiteralField(target)) {
			return nil, l.notYet(target, "writing a possibly absent optional own field")
		}
	}
	object, err := l.expression(target.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(target, "assigning a field of a "+typeName(object.Type()))
	}
	of, err := l.typeOf(target)
	if field := l.checker.GetSymbolAtLocation(target.Name()); field != nil {
		// What the field is declared to keep, not what the checker narrowed this write to.
		of, err = l.typeOfSymbol(target, field)
	}
	if err != nil || (slotless(of) && of != ir.MaybeBoolean && of != ir.Union) {
		return nil, l.notYet(target, "storing "+l.checker.TypeToString(l.checker.GetTypeAtLocation(target))+" in a field")
	}
	value := uninitializedValue(of)
	if !uninitialized {
		value, err = l.expression(valueNode)
		if err != nil {
			return nil, err
		}
	}
	if slotless(value.Type()) && value.Type() != ir.MaybeBoolean && value.Type() != ir.Union {
		return nil, l.notYet(target, "storing "+l.checker.TypeToString(l.checker.GetTypeAtLocation(target))+" in a field")
	}
	// A field of number | undefined is given a packed word, whatever it's assigned.
	value = fit(value, of)
	if call, handled := l.privateStaticStore(target, object, value, uninitialized); handled {
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	}
	// A #private field is stored under its name, # and all, which nothing else can spell.
	return []ir.Statement{ir.SetProperty{Object: object, Name: l.fieldName(target.Name()), Value: value, Uninitialized: uninitialized, Class: l.classOf(target), Site: l.writeSite(target.AsPropertyAccessExpression().Expression)}}, nil
}

// updateProperty lowers object.name op= value, and object.name++ and -- (a nil value, a step of 1).
// JavaScript evaluates the object once, reads the field, then evaluates the value: a variable may be
// read twice, since nothing runs between the two reads, and anything else is held in a local of its
// own for the statement, so make().count++ makes one object.
func (l *lowering) updateProperty(node *ast.Node, target *ast.Node, operator ast.Kind, valueNode *ast.Node) ([]ir.Statement, error) {
	object, err := l.expression(target.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(target, "assigning a field of a "+typeName(object.Type()))
	}
	of, err := l.typeOf(target)
	if err != nil {
		return nil, err
	}
	object = l.privateStaticReceiver(target.Name(), object, false)
	statements := []ir.Statement{}
	field := l.checker.GetSymbolAtLocation(target.Name())
	_, isRead := object.(ir.Read)
	if !isRead || (field != nil && accessorSymbol(field)) {
		held := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "object", Type: ir.Object, Function: l.functionIndex})
		l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)), target)
		statements = append(statements, ir.Declare{Local: held, Value: object})
		object = ir.Read{Local: held, Of: ir.Object}
	}
	value := ir.Expression(ir.NumberConstant{Value: 1})
	if valueNode != nil {
		if value, err = l.expression(valueNode); err != nil {
			return nil, err
		}
	}
	name := l.fieldName(target.Name())
	current := ir.Expression(ir.Property{Object: object, Name: name, Of: of, Class: l.classOf(target)})
	if operator == ast.KindPlusToken && valueNode != nil {
		current, value = l.spelled(target, current), l.spelled(valueNode, value)
	}
	updated, err := l.combine(node, operator, current, value)
	if err != nil {
		return nil, err
	}
	statements = append(statements, ir.SetProperty{Object: object, Name: name, Value: updated, Class: l.classOf(target), Site: l.writeSite(target.AsPropertyAccessExpression().Expression)})
	if len(statements) == 1 {
		return statements, nil
	}
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

// classOf is, for an access to a class's field, that class's constructor plus one, and 0 for any other
// access, so the emitter can find the field where the class's layout puts it. Only a class without
// type parameters, already lowered, is looked up: one with them has a layout per instantiation, and
// looking up never lowers a class, so a program that compiled before can't stop compiling here.
func (l *lowering) classOf(access *ast.Node) int {
	field := l.checker.GetSymbolAtLocation(access.Name())
	if field == nil || len(field.Declarations) != 1 || (field.Declarations[0].Kind != ast.KindPropertyDeclaration && !parameterProperty(field.Declarations[0])) {
		return 0
	}
	if ast.HasSyntacticModifier(field.Declarations[0], ast.ModifierFlagsStatic) {
		return 0
	}
	class := field.Declarations[0].Parent
	if parameterProperty(field.Declarations[0]) {
		class = class.Parent
	}
	if class == nil || class.Kind != ast.KindClassDeclaration || len(class.TypeParameters()) > 0 {
		return 0
	}
	if lowered, isLowered := l.instances[l.classInstanceKey(class, nil)]; isLowered {
		return lowered.constructor + 1
	}
	return 0
}

func nodesOf(list *ast.NodeList) []*ast.Node {
	if list == nil {
		return nil
	}
	return list.Nodes
}

// lastFieldAssignment is where, in the constructor body, every field without an initializer has
// been assigned: the end of the latest of their first assignments, each a statement of the body
// itself, which run in order and unconditionally. A field first assigned anywhere else (in a branch,
// a loop or a closure) is set by no point the body can be sure of, so it's the body's end; 0 when
// there are no such fields. Before that point, JavaScript reads an unset field as undefined, which a
// field of number or string can't hold: the checker proves the constructor itself never reads one,
// but not a method it calls, nor anything it hands this to (useOfThis).
func lastFieldAssignment(declaration *ast.Node, constructor *ast.Node) int {
	unset := map[string]bool{}
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindPropertyDeclaration && member.AsPropertyDeclaration().Initializer == nil && !(member.PostfixToken() != nil && member.PostfixToken().Kind == ast.KindExclamationToken) && !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			unset[member.Name().Text()] = true
		}
	}
	if len(unset) == 0 || constructor.Body() == nil {
		return 0
	}
	last := 0
	for _, statement := range constructor.Body().AsBlock().Statements.Nodes {
		if len(unset) == 0 {
			break
		}
		if statement.Kind != ast.KindExpressionStatement {
			continue
		}
		assignment := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
		if assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
			continue
		}
		target := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
		if target.Kind == ast.KindPropertyAccessExpression && ast.SkipParentheses(target.AsPropertyAccessExpression().Expression).Kind == ast.KindThisKeyword && unset[target.Name().Text()] {
			delete(unset, target.Name().Text())
			last = statement.End()
		}
	}
	if len(unset) > 0 {
		return constructor.Body().End()
	}
	return last
}

// useOfThis refuses this, in a constructor before its last field is assigned, anywhere but as the
// object of a field it reads or writes: a method it calls, or a function it's handed to, could read a
// field not set yet.
func (l *lowering) useOfThis(node *ast.Node) error {
	path := "this"
	if node.Kind == ast.KindSuperKeyword {
		path = "super"
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
		path += "." + parent.Name().Text()
		if parent.Parent != nil && parent.Parent.Kind == ast.KindCallExpression {
			path += "(...)"
		}
	}

	if l.instance != nil && l.instance.unreadyThis[node] {
		return &Refused{Where: l.program.Where(node), What: "this before super returns", Fix: "call super(...) before using this (adamic/this-after-super)"}
	}
	// An arrow captures this when it is created, even if its body only reads a field.
	// Use the outermost arrow's position: nested arrows can run while fields are unset too.
	var closure *ast.Node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindArrowFunction {
			closure = parent
		}
		if parent.Kind == ast.KindConstructor || parent.Kind == ast.KindMethodDeclaration || parent.Kind == ast.KindFunctionDeclaration || parent.Kind == ast.KindFunctionExpression {
			break
		}
	}
	if closure != nil && ((l.unsetUntil != 0 && closure.Pos() < l.unsetUntil) || (l.instance != nil && l.instance.constructing && l.instance.hasDescendants)) {
		name := "an arrow closure"
		if parent := closure.Parent; parent != nil && parent.Kind == ast.KindVariableDeclaration {
			name = "closure " + parent.Name().Text()
		}
		return &Refused{Where: l.program.Where(closure), What: "this captured by " + name + " before every field is set", Fix: "create the closure after the fields are set, or pass the value in"}
	}
	if l.instance != nil && l.instance.constructing && l.instance.hasDescendants {
		if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
			if field := l.checker.GetSymbolAtLocation(parent.Name()); field != nil && field.Flags&ast.SymbolFlagsProperty != 0 {
				return nil
			}
		}
		return &Refused{Where: l.program.Where(node), What: path + " escaping a base constructor before derived fields are initialized", Fix: "use this only to read or write initialized base fields; call methods and publish the object after construction (adamic/initialized-this)"}
	}
	if l.unsetUntil == 0 || node.Pos() >= l.unsetUntil {
		return nil
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
		if field := l.checker.GetSymbolAtLocation(parent.Name()); field != nil && len(field.Declarations) > 0 && (field.Declarations[0].Kind == ast.KindPropertyDeclaration || parameterProperty(field.Declarations[0])) {
			return nil
		}
	}
	return &Refused{Where: l.program.Where(node), What: path + " escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise)", Fix: "assign every field first, then use this (adamic/initialized-this)"}
}

// The checker owns type identity, including recursive structural types. Reuse the
// first representative's key rather than treating a fresh checker object as a new type.
//
//go:linkname identicalTypes github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).isTypeIdenticalTo
func identicalTypes(checker *checker.Checker, source, target *checker.Type) bool

func (l *lowering) classKeyPrefix(declaration *ast.Node) string {
	return l.program.Where(declaration) + ":" + declaration.Name().Text()
}

func (l *lowering) sameClassArguments(from, to []*checker.Type) bool {
	if len(from) != len(to) {
		return false
	}
	for index := range from {
		if !identicalTypes(l.checker, from[index], to[index]) {
			return false
		}
	}
	return true
}

func (l *lowering) classInstanceKey(declaration *ast.Node, arguments []*checker.Type) string {
	prefix := l.classKeyPrefix(declaration)
	key := prefix
	for _, argument := range arguments {
		key += "," + l.genericTypeKey(argument)
	}
	return key
}

func (l *lowering) classKeyMatches(key string, declaration *ast.Node) bool {
	prefix := l.classKeyPrefix(declaration)
	return key == prefix || strings.HasPrefix(key, prefix+",")
}
