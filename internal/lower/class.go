package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// instance is one instantiation of a class, lowered: its constructor, and its methods by name.
//
// Classes in 0.1 are nominal, invariant and without inheritance (docs/0.1.md), so every method call
// resolves at compile time: a method is a function whose first parameter is this. A generic class is
// lowered once per instantiation, its type parameters standing for what that instantiation made them,
// so Stack<string> and Stack<number> are two classes in C.
type instance struct {
	constructor int
	methods     map[string]int
}

// instantiate lowers a class for the type arguments of a type of it, once per distinct set.
func (l *lowering) instantiate(declaration *ast.Node, classType *checker.Type, where *ast.Node) (*instance, error) {
	clauses := declaration.AsClassDeclaration().HeritageClauses
	for _, clause := range nodesOf(clauses) {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
			return nil, &Refused{Where: l.program.Where(clause), What: "class inheritance (extends)", Fix: "implement an interface instead; 0.1's classes don't inherit"}
		}
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
	key := declaration.Name().Text()
	for _, argument := range arguments {
		key += "," + typeName(argument)
	}
	key = l.program.Where(declaration) + ":" + key
	if existing, isLowered := l.instances[key]; isLowered {
		return existing, nil
	}

	// Register every function of the instantiation first, so methods can call each other.
	name := declaration.Name().Text()
	for _, argument := range arguments {
		name += "_" + typeName(argument)
	}
	lowered := &instance{constructor: len(l.result.Functions), methods: map[string]int{}}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_new", Returns: ir.Object})
	members := declaration.Members()
	for _, member := range members {
		switch member.Kind {
		case ast.KindMethodDeclaration:
			if !ast.IsIdentifier(member.Name()) {
				return nil, l.notYet(member, "a method with a computed name")
			}
			lowered.methods[member.Name().Text()] = len(l.result.Functions)
			l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_" + member.Name().Text()})
		case ast.KindPropertyDeclaration, ast.KindConstructor:
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
	l.substitution, l.instance = substitution, lowered
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	defer func() { l.substitution, l.locals, l.instance = outerSubstitution, outerLocals, outerInstance }()

	if err := l.constructor(lowered.constructor, declaration); err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.Kind != ast.KindMethodDeclaration {
			continue
		}
		this := l.thisLocal()
		if err := l.lowerFunction(lowered.methods[member.Name().Text()], member, this); err != nil {
			return nil, err
		}
	}
	return lowered, nil
}

// thisLocal makes a fresh local for a method's this.
func (l *lowering) thisLocal() int {
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object})
	return len(l.result.Locals) - 1
}

// constructor lowers a class's construction: an object with the class's fields, each from its
// initializer (or its type's zero until the constructor body assigns it, which the checker requires),
// then the constructor's body, which returns the object.
func (l *lowering) constructor(index int, declaration *ast.Node) error {
	this := l.thisLocal()
	fields := []ir.Field{}
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindPropertyDeclaration {
			continue
		}
		property := member.AsPropertyDeclaration()
		if !ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier {
			return l.notYet(member, "a field with a computed name")
		}
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			return l.notYet(member, "a static field")
		}
		of, err := l.typeOf(member.Name())
		if err != nil {
			return err
		}
		var value ir.Expression
		if property.Initializer != nil {
			if containsThis(property.Initializer) {
				return l.notYet(property.Initializer, "a field initializer that reads this")
			}
			if value, err = l.expression(property.Initializer); err != nil {
				return err
			}
		} else {
			value = zeroValue(of)
		}
		fields = append(fields, ir.Field{Name: member.Name().Text(), Value: value})
	}
	function := l.result.Functions[index]
	function.Body = []ir.Statement{ir.Declare{Local: this, Value: ir.ObjectLiteral{Fields: fields}}}
	l.result.Functions[index] = function
	var body *ast.Node
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindConstructor {
			body = member
		}
	}
	if body != nil {
		if err := l.lowerFunction(index, body, this); err != nil {
			return err
		}
	}
	function = l.result.Functions[index]
	function.Body = append(function.Body, ir.Return{Value: ir.Read{Local: this, Of: ir.Object}})
	l.result.Functions[index] = function
	return nil
}

// zeroValue is a field's value before its constructor assigns it.
func zeroValue(of ir.Type) ir.Expression {
	switch of {
	case ir.Number:
		return ir.NumberConstant{}
	case ir.Boolean:
		return ir.BooleanConstant{}
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
	return ir.Call{Function: lowered.constructor, Arguments: arguments, Returns: ir.Object}, nil
}

// callOrMethod lowers a call to one of the module's functions, or to a method of one of its classes.
func (l *lowering) callOrMethod(node *ast.Node) (ir.Expression, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return l.call(node)
	}
	method := l.checker.GetSymbolAtLocation(callee)
	if method == nil || len(method.Declarations) == 0 || method.Declarations[0].Kind != ast.KindMethodDeclaration {
		return l.call(node)
	}
	class := method.Declarations[0].Parent
	declaration, isClass := l.classes[l.symbol(class.Name())]
	if !isClass {
		return nil, l.notYet(node, "a method of a class stage 0 doesn't have")
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	// this.method() is the instantiation being lowered: this is the polymorphic this type there,
	// which carries no type arguments of its own.
	lowered := l.instance
	if ast.SkipParentheses(receiver).Kind != ast.KindThisKeyword || lowered == nil {
		var err error
		if lowered, err = l.instantiate(declaration, l.checker.GetTypeAtLocation(receiver), callee); err != nil {
			return nil, err
		}
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, err
	}
	arguments := []ir.Expression{object}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}
	function := lowered.methods[callee.Name().Text()]
	return ir.Call{Function: function, Arguments: arguments, Returns: l.result.Functions[function].Returns}, nil
}

// setProperty lowers object.name = value, as a statement.
func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	object, err := l.expression(target.AsPropertyAccessExpression().Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(target, "assigning a field of a "+typeName(object.Type()))
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	// A #private field is stored under its name, # and all, which nothing else can spell.
	return []ir.Statement{ir.SetProperty{Object: object, Name: target.Name().Text(), Value: value}}, nil
}

func nodesOf(list *ast.NodeList) []*ast.Node {
	if list == nil {
		return nil
	}
	return list.Nodes
}
