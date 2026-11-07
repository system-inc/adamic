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

// methodList is this side's methods by name in order, so its named dispatch map is stable.
// Static instances keep a separate map; methods on the other side never enter it.
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
func (lowering *lowering) instantiate(
	declaration *ast.Node,
	classType *checker.Type,
	where *ast.Node,
) (*instance, error) {
	// Box<T> inside Maker<Node> is a Box<Node>: what it is, for the cycle finder.
	classType = lowering.concrete(classType)
	if view := lowering.classView(classType, declaration); view != nil {
		classType = view
	} else {
		return nil, lowering.notYet(where, "a method receiver without a nominal class view")
	}
	arguments := []ir.Type{}
	substitution := map[*checker.Type]ir.Type{}
	parameters := declaration.TypeParameters()
	var typeArguments []*checker.Type
	if len(parameters) > 0 {
		typeArguments = lowering.checker.GetTypeArguments(classType)
	}
	if len(parameters) != len(typeArguments) {
		return nil, lowering.notYet(where, "a class whose type arguments aren't known here")
	}
	if len(parameters) > 0 {
		if refuseGenericVoidReadsError := lowering.refuseGenericVoidReads(
			where,
			declaration,
			lowering.genericVoidBinding(declaration, typeArguments),
			nil,
			nil,
		); refuseGenericVoidReadsError != nil {
			return nil, refuseGenericVoidReadsError
		}
	}
	for index, parameter := range parameters {
		representation, isKnown := lowering.representation(typeArguments[index])
		if !isKnown {
			return nil, lowering.notYet(
				where,
				"a class instantiated with "+lowering.checker.TypeToString(typeArguments[index]),
			)
		}
		substitution[lowering.checker.GetTypeAtLocation(parameter.Name())] = representation
		arguments = append(arguments, representation)
	}
	key := lowering.classInstanceKey(declaration, typeArguments)
	mapper := lowering.typeMapperOf(declaration, classType)
	if nominalTypeArgumentsError := lowering.nominalTypeArguments(
		declaration,
		typeArguments,
		mapper,
		where,
	); nominalTypeArgumentsError != nil {
		return nil, nominalTypeArgumentsError
	}
	lowering.instantiated = append(lowering.instantiated, classType)
	// Every request is judged, hit or miss, before the table is read (predicates_branded.go).
	if brandedRequestError := lowering.brandedRequest("class:"+key, declaration, mapper); brandedRequestError != nil {
		return nil, brandedRequestError
	}
	if existing, isLowered := lowering.instances[key]; isLowered {
		for _, this := range existing.thisLocals {
			lowering.noteAlso(this, classType)
		}
		// The instance is shared, and so are its locals and function values: each is also what this
		// instantiation makes it.
		outerMapper := lowering.typeMapper
		lowering.typeMapper = mapper
		for _, shared := range existing.templates {
			if shared.isClosure {
				lowering.closureRecords = append(lowering.closureRecords, closureRecord{
					proven:   lowering.concrete(shared.proven),
					function: shared.closure,
					node:     lowering.closureNodeOf(shared.closure),
				})
			} else {
				lowering.noteAlso(shared.local, lowering.concrete(shared.proven))
			}
		}
		lowering.typeMapper = outerMapper
		return existing, nil
	}

	if len(parameters) > 0 {
		if lowering.genericDepth >= maximumGenericDepth {
			return nil, &Refused{
				Where: lowering.program.Where(where),
				What:  "a generic class instantiated without end (polymorphic recursion)",
				Fix:   "keep recursive type arguments unchanged, or write a class per type",
			}
		}
		lowering.genericDepth++
		defer func() { lowering.genericDepth-- }()
	}

	base, baseInstanceError := lowering.baseInstance(declaration, classType)
	if baseInstanceError != nil {
		return nil, baseInstanceError
	}
	if checkOverridesError := lowering.checkOverrides(declaration, classType); checkOverridesError != nil {
		return nil, checkOverridesError
	}

	// Register every function of the instantiation first, so methods can find each other.
	name := declaration.Name().Text()
	for _, argument := range arguments {
		name += "_" + typeName(argument)
	}
	lowered := &instance{
		constructor:   len(lowering.result.Functions),
		methods:       map[string]int{},
		staticMethods: map[string]bool{},
		slots:         map[string]int{},
		initializer:   -1,
		superReady:    -1,
	}
	source := lowering.functionSource(declaration)
	source.Specialization = lowering.sourceTypeKey(classType, map[*checker.Type]bool{})
	source.Role = "allocate"
	source.Declaration = append(append([]string{}, source.Declaration...), "allocate")
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: name + "_new", Source: source, Returns: ir.Object},
	)
	lowered.hasDescendants = lowering.derivedAncestors[lowering.symbol(declaration.Name())]
	lowered.base = base
	lowered.class = len(lowering.result.Classes) + 1
	metadata := ir.Class{
		Source:      lowering.functionSource(declaration),
		Name:        name,
		Constructor: lowered.constructor,
		Definition:  lowering.classDefinition(declaration),
	}
	if base != nil {
		metadata.Base = base.class
		metadata.Fields = append(metadata.Fields, lowering.result.Classes[base.class-1].Fields...)
		metadata.Methods = append(metadata.Methods, lowering.result.Classes[base.class-1].Methods...)
		for method, function := range base.methods {
			lowered.methods[method] = function
			lowered.staticMethods[method] = base.staticMethods[method]
		}
		for method, slot := range base.slots {
			lowered.slots[method] = slot
		}
	}
	metadata.OwnStart = len(metadata.Fields)
	lowering.result.Classes = append(lowering.result.Classes, metadata)
	members := declaration.Members()
	for _, member := range members {
		switch member.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if !ast.IsIdentifier(member.Name()) &&
				member.Name().Kind != ast.KindPrivateIdentifier &&
				!(member.Kind == ast.KindMethodDeclaration &&
					member.Name().Kind == ast.KindComputedPropertyName &&
					lowering.symbolIterator(member.Name().AsComputedPropertyName().Expression)) {
				return nil, lowering.notYet(member, "a method with a computed name")
			}
			if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				continue
			}
			methodName := lowering.methodKey(member, lowered.class)
			if accessorMember(member) {
				lowering.registerAccessor(lowered.class, member, len(lowering.result.Functions))
			}
			lowered.methods[methodName] = len(lowering.result.Functions)
			lowered.staticMethods[methodName] = ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic)
			if !lowered.staticMethods[methodName] {
				meta := &lowering.result.Classes[lowered.class-1]
				slot, exists := lowered.slots[methodName]
				if !exists {
					slot = len(meta.Methods)
					lowered.slots[methodName] = slot
					meta.Methods = append(meta.Methods, -1)
				}
				meta.Methods[slot] = len(lowering.result.Functions)
			}
			lowering.result.Functions = append(
				lowering.result.Functions,
				ir.Function{Name: name + "_" + methodName, MethodName: methodName},
			)
		case ast.KindPropertyDeclaration, ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		default:
			return nil, lowering.notYet(member, describe(member)+" in a class")
		}
	}
	if lowering.instances == nil {
		lowering.instances = map[string]*instance{}
	}
	lowering.instances[key] = lowered

	// Lower with this instantiation's meaning of each type parameter, and locals of its own: the
	// same parameter is a string here and may be a number in another instantiation.
	outerSubstitution, outerLocals, outerInstance := lowering.substitution, lowering.locals, lowering.instance
	outerClassType, outerClassNode, outerMapper := lowering.classType, lowering.classNode, lowering.typeMapper
	lowering.substitution, lowering.instance = substitution, lowered
	lowering.classType, lowering.classNode, lowering.typeMapper = classType, declaration.Name(), mapper
	defer func() {
		lowering.classType, lowering.classNode, lowering.typeMapper = outerClassType, outerClassNode, outerMapper
	}()
	lowering.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if lowering.result.Locals[local].Global {
			lowering.locals[symbol] = local
		}
	}
	defer func() {
		lowering.substitution, lowering.locals, lowering.instance = outerSubstitution, outerLocals, outerInstance
	}()

	// Every method's signature is written before any body is lowered, the constructor's included, so a
	// method can call one declared below it, and two can call each other.
	for _, member := range members {
		if !classFunction(member) || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		method := lowered.methods[lowering.methodKey(member, lowered.class)]
		if signatureError := lowering.signature(method, member, lowering.thisLocal(method)); signatureError != nil {
			return nil, signatureError
		}
		if validateAccessorSignatureError := lowering.validateAccessorSignature(
			member,
			method,
		); validateAccessorSignatureError != nil {
			return nil, validateAccessorSignatureError
		}
	}
	if constructorError := lowering.constructor(lowered.constructor, declaration); constructorError != nil {
		return nil, constructorError
	}
	for _, member := range members {
		if !classFunction(member) || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		method := lowered.methods[lowering.methodKey(member, lowered.class)]
		if member.Body() == nil {
			continue
		}
		if lowerFunctionError := lowering.lowerFunction(method, member, -1); lowerFunctionError != nil {
			return nil, lowerFunctionError
		}
	}
	return lowered, nil
}

// thisLocal makes a fresh local for the this of the method or constructor at owner.
func (lowering *lowering) thisLocal(owner int) int {
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: owner})
	lowering.noteLocal(len(lowering.result.Locals)-1, lowering.classType, lowering.classNode)
	lowering.instance.thisLocals = append(lowering.instance.thisLocals, len(lowering.result.Locals)-1)
	return len(lowering.result.Locals) - 1
}

// constructor lowers a class's construction: an object with the class's fields, each from its
// initializer (or its type's zero until the constructor body assigns it, which the checker requires),
// then the constructor's body, which returns the object.
func (lowering *lowering) constructor(index int, declaration *ast.Node) error {
	return lowering.inheritanceConstructor(index, declaration)
}

// zeroValue is a field's value before its constructor assigns it.
func zeroValue(of ir.Type) ir.Expression {
	switch of {
	case ir.Number:
		return ir.NumberConstant{}
	case ir.Boolean:
		return ir.BooleanConstant{}
	case ir.MaybeNumber, ir.MaybeBoolean:
		// A field of number | undefined left without a value is undefined, as JavaScript leaves it.
		return ir.MaybeOf{Of: of}
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
func (lowering *lowering) construct(node *ast.Node, declaration *ast.Node) (ir.Expression, error) {
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract) {
		return nil, &Refused{
			Where: lowering.program.Where(node),
			What:  "new of an abstract class",
			Fix:   "construct a concrete subclass that implements its abstract methods",
		}
	}
	lowered, instantiateError := lowering.instantiate(declaration, lowering.checker.GetTypeAtLocation(node), node)
	if instantiateError != nil {
		return nil, instantiateError
	}
	arguments := []ir.Expression{}
	if node.AsNewExpression().Arguments != nil {
		for _, argument := range node.AsNewExpression().Arguments.Nodes {
			value, expressionError := lowering.expression(argument)
			if expressionError != nil {
				return nil, expressionError
			}
			arguments = append(arguments, value)
		}
	}
	return lowering.staticConstruct(node, declaration, lowered.constructor, arguments), nil
}

// callOrMethod lowers a call to one of the module's functions, or to a method of one of its classes.
func (lowering *lowering) callOrMethod(node *ast.Node) (ir.Expression, error) {
	if value, handled, classGenericCallError := lowering.classGenericCall(node); handled {
		return value, classGenericCallError
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindSuperKeyword {
		return lowering.superCall(node)
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return lowering.call(node)
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	receiverType := lowering.concrete(lowering.checker.GetTypeAtLocation(receiver))
	method := lowering.checker.GetSymbolAtLocation(callee)
	if isClassInstance(receiverType) {
		if actual := lowering.checker.GetPropertyOfType(receiverType, callee.Name().Text()); actual != nil {
			method = actual
		}
	}
	// A structural signature identifies neither the static side nor an instance.
	// Leave it to callClosure's receiver lookup, using that object's named method map.
	if method == nil || len(method.Declarations) == 0 || method.Declarations[0].Kind != ast.KindMethodDeclaration {
		return lowering.call(node)
	}
	class := method.Declarations[0].Parent
	declaration, isClass := lowering.classes[lowering.symbol(class.Name())]
	if !isClass {
		return nil, lowering.notYet(node, "a method of a class stage 0 doesn't have")
	}
	if len(lowering.definedMembers(receiverType)) > 1 {
		return nil, lowering.notYet(
			receiver,
			"a method call through a union of class types; narrow with instanceof first",
		)
	}
	if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		// object?.method(...), the object perhaps undefined: the class is instantiated for what the
		// object is when it's there, and the call is made through the object's methods, which stops
		// at undefined as JavaScript's chain does (ir.Property's Method).
		if _, instantiateError := lowering.instantiate(
			declaration,
			lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(receiver)),
			callee,
		); instantiateError != nil {
			return nil, instantiateError
		}
		return lowering.callClosure(node)
	}
	// this.method() is the instantiation being lowered: this is the polymorphic this type there,
	// which carries no type arguments of its own.
	lowered := lowering.instance
	if ast.HasSyntacticModifier(method.Declarations[0], ast.ModifierFlagsStatic) {
		var staticInstanceError error
		if lowered == nil || !lowered.static || (ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword &&
			ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword) {
			lowered, staticInstanceError = lowering.staticInstance(lowering.staticClass(receiverType, declaration))
			if staticInstanceError != nil {
				return nil, staticInstanceError
			}
		}
	} else if (ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword &&
		ast.SkipParentheses(receiver).Kind != ast.KindSuperKeyword) || lowered == nil {
		var instantiateError error
		if lowered, instantiateError = lowering.instantiate(
			declaration,
			receiverType,
			callee,
		); instantiateError != nil {
			return nil, instantiateError
		}
	}
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		if lowering.instance == nil || lowering.instance.base == nil {
			return nil, lowering.notYet(node, "super outside a derived class")
		}
		lowered = lowering.instance.base
	}
	var object ir.Expression
	var expressionError error
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		if useOfThisError := lowering.useOfThis(receiver); useOfThisError != nil {
			return nil, useOfThisError
		}
		// super.method() reads this as this.method() does, and inside an arrow that's a capture.
		lowering.touch(lowering.this)
		object = ir.Read{Local: lowering.this, Of: ir.Object}
	} else {
		object, expressionError = lowering.expression(receiver)
	}
	if expressionError != nil {
		return nil, expressionError
	}
	object = lowering.privateStaticReceiver(callee.Name(), object, false)
	values, spread, callArgumentsError := lowering.callArguments(node.AsCallExpression().Arguments.Nodes)
	if callArgumentsError != nil {
		return nil, callArgumentsError
	}
	arguments := append([]ir.Expression{object}, values...)
	if len(spread) > 0 {
		spread = append([]bool{false}, spread...)
	}
	function := lowered.methods[lowering.fieldName(callee.Name())]
	lowering.fitCallArguments(function, arguments, spread)
	virtual := lowered.slots[lowering.fieldName(callee.Name())] + 1
	if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		virtual = 0
	}
	return ir.Call{
		Function:  function,
		Arguments: arguments,
		Spread:    spread,
		Returns:   lowering.result.Functions[function].Returns,
		Virtual:   virtual,
	}, nil
}

// setProperty lowers object.name = value, as a statement.
func (lowering *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if lowering.uninitializedInitializer(valueNode) {
		if absentOptionalWriteError := lowering.absentOptionalWrite(target); absentOptionalWriteError != nil {
			return nil, absentOptionalWriteError
		}
		member := lowering.checker.GetSymbolAtLocation(target)
		if member == nil {
			return nil, lowering.notYet(target, "resetting an unknown placeholder field")
		}
		for _, declaration := range member.Declarations {
			if declaration.Kind == ast.KindGetAccessor || declaration.Kind == ast.KindSetAccessor {
				return nil, lowering.notYet(target, "resetting a placeholder through an accessor")
			}
		}
		if target.Name().Kind == ast.KindPrivateIdentifier {
			return nil, lowering.notYet(target, "resetting a private placeholder field")
		}
		object, expressionError := lowering.expression(target.AsPropertyAccessExpression().Expression)
		if expressionError != nil {
			return nil, expressionError
		}
		if object.Type() != ir.Object {
			return nil, lowering.notYet(target, "resetting a placeholder field without object storage")
		}
		of, typeOfSymbolError := lowering.typeOfSymbol(target, member)
		if typeOfSymbolError != nil || censusFieldSlotless(of) {
			return nil, lowering.notYet(target, "resetting a placeholder field without supported storage")
		}
		return []ir.Statement{ir.SetProperty{
			Object:        object,
			Name:          lowering.fieldName(target.Name()),
			Value:         lowering.placeholderInitialValue(valueNode, of),
			Uninitialized: true,
			Unset:         true,
			Class:         lowering.classOf(target),
			Site:          lowering.writeSite(target.AsPropertyAccessExpression().Expression),
		}}, nil
	}
	if call, handled, superAccessorError := lowering.superAccessor(target, valueNode); handled {
		if superAccessorError != nil {
			return nil, superAccessorError
		}
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	}
	if member := lowering.checker.GetSymbolAtLocation(
		target,
	); member != nil && member.Flags&ast.SymbolFlagsMethod != 0 {
		return nil, lowering.notYet(target, "replacing a represented method at runtime")
	}
	if absentOptionalWriteError := lowering.absentOptionalWrite(target); absentOptionalWriteError != nil {
		return nil, absentOptionalWriteError
	}
	object, err := lowering.expression(target.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	_, _, nodeArray := lowering.nodeArrayLayoutOf(
		lowering.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression),
	)
	if object.Type() != ir.Object && !(object.Type() == ir.Array && nodeArray) {
		return nil, lowering.notYet(target, "assigning a field of a "+typeName(object.Type()))
	}
	value, err := lowering.expression(valueNode)
	if err != nil {
		return nil, err
	}
	declared := lowering.checker.GetTypeAtLocation(target)
	of, err := lowering.typeOf(target)
	if field := lowering.checker.GetSymbolAtLocation(target.Name()); field != nil {
		declared = lowering.checker.GetTypeOfSymbol(field)
		// What the field is declared to keep, not what the checker narrowed this write to.
		of, err = lowering.typeOfSymbol(target, field)
	}
	if err != nil || censusFieldSlotless(of) || censusFieldSlotless(value.Type()) {
		return nil, lowering.notYet(
			target,
			"storing "+lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(target))+" in a field",
		)
	}
	// A field of number | undefined is given a packed word, whatever it's assigned.
	value = fit(value, of)
	if call, handled := lowering.privateStaticStore(target, object, value); handled {
		return []ir.Statement{ir.Evaluate{Value: call}}, nil
	}
	// A #private field is stored under its name, # and all, which nothing else can spell.
	return []ir.Statement{ir.SetProperty{
		Object:       object,
		Name:         lowering.fieldName(target.Name()),
		Value:        value,
		DeclaredType: lowering.checker.TypeToString(declared),
		Class:        lowering.classOf(target),
		Site:         lowering.writeSite(target.AsPropertyAccessExpression().Expression),
		ViewReceiverTypeID: lowering.viewWriteReceiverTypeID(
			target,
			target.AsPropertyAccessExpression().Expression,
			lowering.fieldName(target.Name()),
		),
	}}, nil
}

// absentOptionalWrite refuses a write to an optional own field of a plain object when some literal left
// that field out: the literal's object has no slot for it, and the write would find none ('a field the
// checker proved is there is missing'), whatever the field's representation. A field held in a Union is
// refused wherever it's written, as before. A class instance declares every field, so it has the slot.
// The field is matched by its declarations (fieldDeclarations), never by the type that spelled it, so an
// interface, an intersection, an alias of either, an upcast view and a generic's constraint all reach the
// same field (#krv1hgh). Optional presence lifts this (#r3chqza).
func (lowering *lowering) absentOptionalWrite(target *ast.Node) error {
	member := lowering.checker.GetSymbolAtLocation(target)
	if member == nil ||
		member.Flags&ast.SymbolFlagsOptional == 0 ||
		isClassInstance(lowering.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)) {
		return nil
	}
	if stored, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(member)); stored == ir.Union {
		return lowering.notYet(target, "writing a possibly absent optional own field")
	}
	omitted := false
	for _, field := range fieldDeclarations(member) {
		omitted = omitted || lowering.omittedOptionals[field]
	}
	if omitted && !presentBeforeWrite(target) && !lowering.literalGivesField(target) {
		return lowering.notYet(target, "writing a possibly absent optional own field")
	}
	return nil
}

// fieldDeclarations is a property's symbol and the symbols of the declarations it was made from. An
// interface's or a type literal's property is its own declaration's. An intersection's, or an
// instantiated or mapped type's, is a symbol the checker makes for that type alone, so Mut & { brand }
// and Mut name one field by two symbols, and only the declarations' tie them.
func fieldDeclarations(property *ast.Symbol) []*ast.Symbol {
	symbols := []*ast.Symbol{property}
	for _, declaration := range property.Declarations {
		if declared := declaration.Symbol(); declared != nil && declared != property {
			symbols = append(symbols, declared)
		}
	}
	return symbols
}

// presentBeforeWrite is a write to object.name inside the then branch of an if that tested
// object.name !== undefined (or != null), the same spelling, directly or as a term of &&: a value was
// there, so the object has the slot.
func presentBeforeWrite(target *ast.Node) bool {
	spelled := accessPath(target)
	if spelled == "" {
		return false
	}
	for child, parent := target, target.Parent; parent != nil; child, parent = parent, parent.Parent {
		if parent.Kind == ast.KindFunctionDeclaration ||
			parent.Kind == ast.KindArrowFunction ||
			parent.Kind == ast.KindFunctionExpression ||
			parent.Kind == ast.KindMethodDeclaration {
			return false
		}
		if parent.Kind != ast.KindIfStatement || parent.AsIfStatement().ThenStatement != child {
			continue
		}
		if testsPresence(parent.AsIfStatement().Expression, spelled) {
			return true
		}
	}
	return false
}

func testsPresence(condition *ast.Node, spelled string) bool {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := condition.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken:
		return testsPresence(binary.Left, spelled) || testsPresence(binary.Right, spelled)
	case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken:
		left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
		absent := right.Kind == ast.KindUndefinedKeyword ||
			right.Kind == ast.KindNullKeyword ||
			(right.Kind == ast.KindIdentifier && right.Text() == "undefined")
		return absent && accessPath(left) == spelled
	}
	return false
}

// plainPropertyName is a literal member's name when it's spelled as an identifier, a string or a
// number, and "" for a computed name, which gives no field this can see.
func plainPropertyName(property *ast.Node) string {
	name := property.Name()
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
		return name.Text()
	}
	return ""
}

// accessPath spells a chain of names (this.a.b, node.flow), and gives "" for anything else, which
// matches nothing.
func accessPath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text()
	case ast.KindThisKeyword:
		return "this"
	case ast.KindPropertyAccessExpression:
		if inner := accessPath(node.AsPropertyAccessExpression().Expression); inner != "" {
			return inner + "." + node.AsPropertyAccessExpression().Name().Text()
		}
	}
	return ""
}

// literalGivesField is a write through a const whose initializer is an object literal naming the field.
func (lowering *lowering) literalGivesField(target *ast.Node) bool {
	receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
	if receiver.Kind != ast.KindIdentifier {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(receiver)
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := symbol.ValueDeclaration
	if declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || ast.SkipParentheses(initializer).Kind != ast.KindObjectLiteralExpression {
		return false
	}
	name := target.AsPropertyAccessExpression().Name().Text()
	for _, property := range ast.SkipParentheses(initializer).AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindSpreadAssignment && plainPropertyName(property) == name {
			return true
		}
	}
	return false
}

// noteOmittedOptionals finds every object literal whose context types it and records the optional
// properties of that type it leaves out, by their declarations (fieldDeclarations). A spread counts as
// giving a property its source declares required. A context that's an intersection of object types is
// one object, whose optional properties the literal leaves out as an interface's (#krv1hgh).
func (lowering *lowering) noteOmittedOptionals(modules []*ast.SourceFile) {
	lowering.omittedOptionals = map[*ast.Symbol]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			lowering.noteOmittedOptionalsOf(node)
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
}

func (lowering *lowering) noteOmittedOptionalsOf(literal *ast.Node) {
	contextual := lowering.slotType(literal)
	if contextual == nil {
		return
	}
	given := map[string]bool{}
	var spreads []*checker.Type
	for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindSpreadAssignment {
			spreads = append(spreads, lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression))
			continue
		}
		if name := plainPropertyName(property); name != "" {
			given[name] = true
		}
	}
	members := []*checker.Type{contextual}
	if contextual.Flags()&checker.TypeFlagsUnion != 0 {
		members = contextual.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
			continue
		}
		for _, property := range lowering.checker.GetPropertiesOfType(member) {
			if property.Flags&ast.SymbolFlagsOptional == 0 || given[property.Name] {
				continue
			}
			spread := false
			for _, source := range spreads {
				if inside := lowering.checker.GetPropertyOfType(
					source,
					property.Name,
				); inside != nil && inside.Flags&ast.SymbolFlagsOptional == 0 {
					spread = true
				}
			}
			if !spread {
				for _, field := range fieldDeclarations(property) {
					lowering.omittedOptionals[field] = true
				}
			}
		}
	}
}

// updateProperty lowers object.name op= value, and object.name++ and -- (a nil value, a step of 1).
// JavaScript evaluates the object once, reads the field, then evaluates the value: a variable may be
// read twice, since nothing runs between the two reads, and anything else is held in a local of its
// own for the statement, so make().count++ makes one object.
func (lowering *lowering) updateProperty(
	node *ast.Node,
	target *ast.Node,
	operator ast.Kind,
	valueNode *ast.Node,
) ([]ir.Statement, error) {
	if absentOptionalWriteError := lowering.absentOptionalWrite(target); absentOptionalWriteError != nil {
		return nil, absentOptionalWriteError
	}
	object, expressionError := lowering.expression(target.AsPropertyAccessExpression().Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	if object.Type() != ir.Object {
		return nil, lowering.notYet(target, "assigning a field of a "+typeName(object.Type()))
	}
	of, typeOfError := lowering.typeOf(target)
	if typeOfError != nil {
		return nil, typeOfError
	}
	object = lowering.privateStaticReceiver(target.Name(), object, false)
	statements := []ir.Statement{}
	field := lowering.checker.GetSymbolAtLocation(target.Name())
	_, isRead := object.(ir.Read)
	if !isRead || (field != nil && accessorSymbol(field)) {
		held := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "object", Type: ir.Object, Function: lowering.functionIndex},
		)
		lowering.noteLocal(
			held,
			lowering.concrete(lowering.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)),
			target,
		)
		statements = append(statements, ir.Declare{Local: held, Value: object})
		object = ir.Read{Local: held, Of: ir.Object}
	}
	value := ir.Expression(ir.NumberConstant{Value: 1})
	if valueNode != nil {
		if value, expressionError = lowering.expression(valueNode); expressionError != nil {
			return nil, expressionError
		}
	}
	name := lowering.fieldName(target.Name())
	current := lowering.readObjectField(
		target,
		ir.Property{Object: object, Name: name, Of: of, Class: lowering.classOf(target)},
	)
	if operator == ast.KindPlusToken && valueNode != nil {
		current, value = lowering.spelled(target, current), lowering.spelled(valueNode, value)
	}
	updated, combineError := lowering.combine(node, operator, current, value)
	if combineError != nil {
		return nil, combineError
	}
	statements = append(statements, ir.SetProperty{
		Object: object,
		Name:   name,
		Value:  updated,
		Class:  lowering.classOf(target),
		Site:   lowering.writeSite(target.AsPropertyAccessExpression().Expression),
		ViewReceiverTypeID: lowering.viewWriteReceiverTypeID(
			target,
			target.AsPropertyAccessExpression().Expression,
			name,
		),
	})
	if len(statements) == 1 {
		return statements, nil
	}
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

// classOf is, for an access to a class's field, that class's constructor plus one, and 0 for any other
// access, so the emitter can find the field where the class's layout puts it. Only a class without
// type parameters, already lowered, is looked up: one with them has a layout per instantiation, and
// looking up never lowers a class, so a program that compiled before can't stop compiling here.
func (lowering *lowering) classOf(access *ast.Node) int {
	field := lowering.checker.GetSymbolAtLocation(access.Name())
	if field == nil ||
		len(field.Declarations) != 1 ||
		(field.Declarations[0].Kind != ast.KindPropertyDeclaration && !parameterProperty(field.Declarations[0])) {
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
	if lowered, isLowered := lowering.instances[lowering.classInstanceKey(class, nil)]; isLowered {
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
		if member.Kind == ast.KindPropertyDeclaration &&
			member.AsPropertyDeclaration().Initializer == nil &&
			!(member.PostfixToken() != nil && member.PostfixToken().Kind == ast.KindExclamationToken) &&
			!ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
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
		if assignment.Kind != ast.KindBinaryExpression ||
			assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
			continue
		}
		target := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
		if target.Kind == ast.KindPropertyAccessExpression &&
			ast.SkipParentheses(target.AsPropertyAccessExpression().Expression).Kind == ast.KindThisKeyword &&
			unset[target.Name().Text()] {
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
func (lowering *lowering) useOfThis(node *ast.Node) error {
	if lowering.instance != nil && lowering.instance.unreadyThis[node] {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "this before super returns",
			Fix:   "call super(...) before using this",
		}
	}
	if lowering.instance != nil && lowering.instance.constructing && lowering.instance.hasDescendants {
		if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
			if field := lowering.checker.GetSymbolAtLocation(
				parent.Name(),
			); field != nil && field.Flags&ast.SymbolFlagsProperty != 0 {
				return nil
			}
		}
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "this escaping a base constructor before derived fields are initialized",
			Fix:   "use this only to read or write initialized base fields; call methods and publish the object after construction",
		}
	}
	if lowering.unsetUntil == 0 || node.Pos() >= lowering.unsetUntil {
		return nil
	}
	if parent := node.Parent; parent != nil &&
		parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Expression == node {
		if field := lowering.checker.GetSymbolAtLocation(parent.Name()); field != nil &&
			len(field.Declarations) > 0 &&
			(field.Declarations[0].Kind == ast.KindPropertyDeclaration || parameterProperty(field.Declarations[0])) {
			return nil
		}
	}
	return &Refused{
		Where: lowering.program.Where(node),
		What:  "this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise)",
		Fix:   "assign every field first, then use this",
	}
}

// The checker owns type identity, including recursive structural types. Reuse the
// first representative's key rather than treating a fresh checker object as a new type.
//
//go:linkname identicalTypes github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).isTypeIdenticalTo
func identicalTypes(checker *checker.Checker, source, target *checker.Type) bool

func (lowering *lowering) classKeyPrefix(declaration *ast.Node) string {
	return lowering.program.Where(declaration) + ":" + declaration.Name().Text()
}

func (lowering *lowering) sameClassArguments(from, to []*checker.Type) bool {
	if len(from) != len(to) {
		return false
	}
	for index := range from {
		if !identicalTypes(lowering.checker, from[index], to[index]) {
			return false
		}
	}
	return true
}

func (lowering *lowering) classInstanceKey(declaration *ast.Node, arguments []*checker.Type) string {
	prefix := lowering.classKeyPrefix(declaration)
	key := prefix
	for _, argument := range arguments {
		key += "," + lowering.genericTypeKey(argument)
	}
	return key
}

func (lowering *lowering) classKeyMatches(key string, declaration *ast.Node) bool {
	prefix := lowering.classKeyPrefix(declaration)
	return key == prefix || strings.HasPrefix(key, prefix+",")
}
