package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) baseInstance(declaration *ast.Node, classType *checker.Type) (*instance, error) {
	for _, clause := range nodesOf(declaration.AsClassDeclaration().HeritageClauses) {
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
			continue
		}
		types := clause.AsHeritageClause().Types.Nodes
		if len(types) != 1 || !ast.IsIdentifier(ast.SkipParentheses(types[0].AsExpressionWithTypeArguments().Expression)) {
			return nil, l.notYet(clause, "a computed class base; name the base class directly")
		}
		bases := l.classBases(classType)
		if len(bases) != 1 {
			return nil, l.notYet(clause, "a class base whose type isn't known (name a class declared in this program as the base)")
		}
		baseType := bases[0]
		symbol := baseType.Symbol()
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindClassDeclaration {
			return nil, l.notYet(clause, "a base that isn't a declared class (extend a declared class; use implements for an interface)")
		}
		base := symbol.Declarations[0]
		return l.instantiate(base, baseType, clause)
	}
	return nil, nil
}

// TypeScript compares methods bivariantly and mutable properties covariantly. Neither is safe
// through a base reference: arguments flow into the override, and fields can be written through it.
func (l *lowering) checkOverrides(declaration *ast.Node, classType *checker.Type) error {
	bases := l.classBases(classType)
	if len(bases) == 0 {
		return nil
	}
	return l.checkMemberOverrides(declaration, classType, bases[0], false)
}

func (l *lowering) checkMemberOverrides(declaration *ast.Node, classType, base *checker.Type, static bool) error {
	checkABI := static || len(declaration.TypeParameters()) == 0 || classType != l.checker.GetTypeAtLocation(declaration.Name())
	for _, member := range classMembersWithParameters(declaration) {
		if member.Name() == nil || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) != static {
			continue
		}
		inherited := l.checker.GetPropertyOfType(base, member.Name().Text())
		if inherited == nil {
			continue
		}
		own := l.checker.GetTypeAtLocation(member.Name())
		if property := l.checker.GetPropertyOfType(classType, member.Name().Text()); property != nil {
			own = l.checker.GetTypeOfSymbol(property)
		}
		previous := l.checker.GetTypeOfSymbol(inherited)
		refuse := func(what, fix string) error { return &Refused{Where: l.program.Where(member), What: what, Fix: fix} }
		if (member.Kind == ast.KindMethodDeclaration) != (inherited.Flags&ast.SymbolFlagsMethod != 0) {
			return refuse("an inherited method replaced by a field, or a field replaced by a method", "keep the inherited member kind; use a different name for the new member (adamic/override-member-kind)")
		}
		if accessorMember(member) != accessorSymbol(inherited) {
			return refuse("an inherited data property replaced by an accessor, or an accessor replaced by data", "keep the inherited member kind; use another name for the new property (adamic/override-member-kind)")
		}
		if accessorMember(member) {
			ownGet, ownSet, baseGet, baseSet := false, false, false, false
			for _, candidate := range declaration.Members() {
				if candidate.Name() != nil && candidate.Name().Text() == member.Name().Text() && ast.HasSyntacticModifier(candidate, ast.ModifierFlagsStatic) == static {
					ownGet = ownGet || candidate.Kind == ast.KindGetAccessor
					ownSet = ownSet || candidate.Kind == ast.KindSetAccessor
				}
			}
			for _, candidate := range inherited.Declarations {
				baseGet = baseGet || candidate.Kind == ast.KindGetAccessor
				baseSet = baseSet || candidate.Kind == ast.KindSetAccessor
			}
			if (baseGet && !ownGet) || (baseSet && !ownSet) {
				return refuse("an accessor override that hides the inherited getter or setter", "override both halves of the inherited descriptor; delegate an unchanged half to super (adamic/override-accessor-pair)")
			}
			if member.Kind == ast.KindSetAccessor {
				for _, candidate := range inherited.Declarations {
					if candidate.Kind != ast.KindSetAccessor {
						continue
					}
					old := l.checker.GetSignatureFromDeclaration(candidate)
					next := l.checker.GetSignatureFromDeclaration(member)
					accepts := l.checker.GetTypeOfSymbol(old.Parameters()[0])
					override := l.checker.GetTypeOfSymbol(next.Parameters()[0])
					if !static {
						accepts = instantiateType(l.checker, accepts, l.typeMapperOf(candidate.Parent, base))
						override = instantiateType(l.checker, override, l.typeMapperOf(declaration, classType))
					}
					if !l.classAssignable(accepts, override) || l.widened(accepts, override, map[[2]*checker.Type]bool{}) != nil {
						return refuse("an accessor override that narrows a setter parameter (adamic/contravariant-override)", "accept the base setter's parameter type or a wider type; narrow it inside the setter")
					}
				}
			}
			if ownGet && (!l.classAssignable(own, previous) || l.widened(own, previous, map[[2]*checker.Type]bool{}) != nil) {
				return refuse("an accessor override with an unsafe read or write type (adamic/invariant-mutable)", "keep the inherited accessor type; narrow values inside the accessor")
			}
		}
		if member.Kind == ast.KindPropertyDeclaration || parameterProperty(member) {
			if !l.checker.IsReadonlySymbol(inherited) && (!l.classAssignable(previous, own) || !l.classAssignable(own, previous) || l.widened(previous, own, map[[2]*checker.Type]bool{}) != nil || l.widened(own, previous, map[[2]*checker.Type]bool{}) != nil) {
				return refuse("a mutable inherited field redeclared with a different type (adamic/invariant-mutable)", "keep the base field's type; narrow a local after reading it, or make the field readonly in the base")
			}
			if l.checker.IsReadonlySymbol(inherited) && (l.widened(own, previous, map[[2]*checker.Type]bool{}) != nil || !l.classAssignable(own, previous)) {
				return refuse("an inherited readonly field whose mutable contents are narrowed (adamic/invariant-mutable)", "keep mutable contents at the base type, or make the contents readonly in the base too")
			}
			if l.checker.IsReadonlySymbol(inherited) && !ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly) {
				return refuse("a readonly inherited field redeclared mutable", "keep the inherited field readonly (adamic/readonly-override)")
			}
			from, okFrom := l.representation(previous)
			to, okTo := l.representation(own)
			if checkABI && (!okFrom || !okTo || from != to) {
				return l.notYet(member, "an inherited field override with a different native representation (keep the inherited field type unchanged and narrow a local after reading it)")
			}
		}
		if member.Kind != ast.KindMethodDeclaration {
			continue
		}
		oldSignatures := l.checker.GetSignaturesOfType(previous, checker.SignatureKindCall)
		newSignatures := l.checker.GetSignaturesOfType(own, checker.SignatureKindCall)
		if len(oldSignatures) != 1 || len(newSignatures) != 1 {
			return l.notYet(member, "an overloaded override (use one override signature with union parameters and narrow them inside the method)")
		}
		old, next := oldSignatures[0], newSignatures[0]
		if len(old.Parameters()) != len(next.Parameters()) {
			return l.notYet(member, "an override with a different parameter count; keep the base method's parameter count")
		}
		for index, parameter := range old.Parameters() {
			accepts := l.checker.GetTypeOfSymbol(parameter)
			override := l.checker.GetTypeOfSymbol(next.Parameters()[index])
			a, knownA := l.overrideParameterRepresentation(parameter)
			b, knownB := l.overrideParameterRepresentation(next.Parameters()[index])
			if checkABI && (!knownA || !knownB || a != b) {
				return l.notYet(member, "an override with a different native representation for parameter "+strconv.Quote(next.Parameters()[index].Name)+"; keep the base method's parameter form")
			}
			if !l.classAssignable(accepts, override) || l.widened(accepts, override, map[[2]*checker.Type]bool{}) != nil {
				return refuse("an override that narrows a method parameter (adamic/contravariant-override)", "accept the base method's parameter type or a wider type; narrow it inside the method")
			}
		}
		oldResult := l.checker.GetReturnTypeOfSignature(old)
		newResult := l.checker.GetReturnTypeOfSignature(next)
		a, knownA := l.overrideResultRepresentation(oldResult)
		b, knownB := l.overrideResultRepresentation(newResult)
		if checkABI && (!knownA || !knownB || a != b) {
			return l.notYet(member, "an override with a different native result representation; keep the base method's result form")
		}
		if !l.classAssignable(newResult, oldResult) || l.widened(newResult, oldResult, map[[2]*checker.Type]bool{}) != nil {
			return refuse("an override that widens its return type", "return the base method's result type or a subtype (adamic/covariant-override)")
		}
	}
	return nil
}

// Match signature lowering, including the incoming optional form of a defaulted parameter.
// The signature's symbol carries the instantiated type; its declaration carries the default.
func (l *lowering) overrideParameterRepresentation(parameter *ast.Symbol) (ir.Type, bool) {
	if len(parameter.Declarations) != 1 || parameter.Declarations[0].Kind != ast.KindParameter {
		return 0, false
	}
	declaration := parameter.Declarations[0]
	declared := declaration.AsParameterDeclaration()
	if declared.DotDotDotToken != nil {
		return 0, false
	}
	name := declaration.Name()
	if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
		return ir.Object, declared.Initializer == nil && declared.QuestionToken == nil
	}
	representation, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
	if declared.Initializer != nil || declared.QuestionToken != nil {
		representation = ir.Maybe(representation)
	}
	return representation, known
}

// Named methods returning void or never have no C result, as signature lowering does.
func (l *lowering) overrideResultRepresentation(result *checker.Type) (ir.Type, bool) {
	if result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		return 0, true
	}
	return l.representation(result)
}

// Constructors initialize an already allocated object. A separate allocator fixes its dynamic
// identity before even the base constructor runs, so base method calls can dispatch dynamically.
func (l *lowering) inheritanceConstructor(index int, declaration *ast.Node) error {
	instance := l.instance
	var constructor *ast.Node
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindConstructor {
			constructor = member
		}
	}
	instance.initializer = len(l.result.Functions)
	initializer := instance.initializer
	l.result.Functions = append(l.result.Functions, ir.Function{Name: l.result.Functions[index].Name + "_initialize"})
	this := l.thisLocal(initializer)
	outerThis, outerIndex, outerFunction := l.this, l.functionIndex, l.function
	context := l.result.Functions[initializer]
	l.this, l.functionIndex, l.function = this, initializer, &context
	defer func() { l.this, l.functionIndex, l.function = outerThis, outerIndex, outerFunction }()
	if constructor != nil {
		if err := l.signature(initializer, constructor, this); err != nil {
			return err
		}
		l.result.Functions[initializer].Parameters = append([]int{this}, l.result.Functions[initializer].Parameters...)
	} else {
		l.result.Functions[initializer].Parameters = []int{this}
		if instance.base != nil {
			for _, parameter := range l.result.Functions[instance.base.initializer].Parameters[1:] {
				local := len(l.result.Locals)
				incoming := l.result.Locals[parameter]
				incoming.Function, incoming.Captured, incoming.Borrowed = initializer, false, false
				l.result.Locals = append(l.result.Locals, incoming)
				l.result.Functions[initializer].Parameters = append(l.result.Functions[initializer].Parameters, local)
			}
		}
	}
	// The layout contains inherited fields first, but only this class's initializers run here.
	fields := append([]ir.Field{}, l.result.Classes[instance.class-1].Fields...)
	for _, member := range classMembersWithParameters(declaration) {
		if member.Kind != ast.KindPropertyDeclaration && !parameterProperty(member) {
			continue
		}
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract) {
			return l.notYet(member, "a declare or abstract class field")
		}
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if !ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier {
			return l.notYet(member, "a field with a computed name (declare the field with a fixed identifier name)")
		}
		of, err := l.typeOf(member.Name())
		if err != nil {
			return err
		}
		if slotless(of) {
			return l.notYet(member, "a field of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(member.Name())))
		}
		field := ir.Field{Name: l.fieldName(member.Name()), Value: zeroValue(of), Private: member.Name().Kind == ast.KindPrivateIdentifier, Uninitialized: l.uninitializedDeclaration(member) || (member.Kind == ast.KindPropertyDeclaration && assertionInitializer(member.AsPropertyDeclaration().Initializer))}
		// An uninitialized reference still has its declared representation for the shape bitmap.
		if of.IsReference() {
			field.Value = ir.Undefined{Of: of}
		}
		found := false
		for i := range fields {
			if fields[i].Name == field.Name {
				fields[i] = field
				found = true
				break
			}
		}
		if !found {
			fields = append(fields, field)
		}
	}
	l.result.Classes[instance.class-1].Fields = fields
	if instance.base == nil {
		initialized, err := l.fieldInitializers(declaration, this)
		if err != nil {
			return err
		}
		l.result.Functions[initializer].Body = initialized
	}
	if constructor != nil {
		instance.constructing = true
		defer func() { instance.constructing = false }()
		var returnValue *ast.Node
		var returns ast.Visitor
		returns = func(node *ast.Node) bool {
			if ast.IsFunctionLike(node) {
				return false
			}
			if node.Kind == ast.KindReturnStatement && node.AsReturnStatement().Expression != nil {
				returnValue = node
				return true
			}
			return node.ForEachChild(returns)
		}
		constructor.Body().ForEachChild(returns)
		if returnValue != nil {
			return l.notYet(returnValue, "a constructor returning a replacement value (use a module-level factory function when construction must return a different object)")
		}
		state := uint8(0)
		if instance.base != nil {
			var err error
			state, err = l.prepareSuper(constructor, initializer)
			if err != nil {
				return err
			}
		}

		outerUnset := l.unsetUntil
		l.unsetUntil = lastFieldAssignment(declaration, constructor)
		err := l.lowerFunction(initializer, constructor, this)
		l.unsetUntil = outerUnset
		if err != nil {
			return err
		}
		if state&superUninitialized != 0 {
			l.result.Functions[initializer].Body = append(l.result.Functions[initializer].Body, l.superEnd())
		}
	} else if instance.base != nil {
		arguments := []ir.Expression{}
		for _, parameter := range l.result.Functions[initializer].Parameters {
			arguments = append(arguments, ir.Read{Local: parameter, Of: l.result.Locals[parameter].Type})
		}
		l.result.Functions[initializer].Body = []ir.Statement{ir.Evaluate{Value: ir.Call{Function: instance.base.initializer, Arguments: arguments}}}
		initialized, err := l.fieldInitializers(declaration, this)
		if err != nil {
			return err
		}
		l.result.Functions[initializer].Body = append(l.result.Functions[initializer].Body, initialized...)
	}
	if instance.base == nil && !instance.hasDescendants {
		// Keep creation and writes in one function when no derived constructor needs this
		// initializer. Freshness then knows this is new at each write, not just at return.
		body := l.result.Functions[initializer]
		for local := range l.result.Locals {
			if l.result.Locals[local].Function == initializer {
				l.result.Locals[local].Function = index
			}
		}
		wrapper := l.result.Functions[index]
		wrapper.Parameters = append([]int{}, body.Parameters[1:]...)
		wrapper.Body = append([]ir.Statement{ir.Declare{Local: this, Value: ir.ObjectLiteral{Fields: fields, Class: instance.class, Methods: instance.methodList()}}}, body.Body...)
		wrapper.Body = append(wrapper.Body, ir.Return{Value: ir.Read{Local: this, Of: ir.Object}})
		l.result.Functions[index] = wrapper
		l.result.Functions[initializer] = ir.Function{Name: body.Name}
		return nil
	}
	// The public constructor owns its object, including when an initializer throws.
	wrapper := l.result.Functions[index]
	arguments := []ir.Expression{}
	object := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: index})
	l.noteLocal(object, l.classType, declaration.Name())
	arguments = append(arguments, ir.Read{Local: object, Of: ir.Object})
	for _, parameter := range l.result.Functions[initializer].Parameters[1:] {
		incoming := l.result.Locals[parameter]
		incoming.Function, incoming.Captured, incoming.Borrowed = index, false, false
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, incoming)
		wrapper.Parameters = append(wrapper.Parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: incoming.Type})
	}
	wrapper.Body = []ir.Statement{
		ir.Declare{Local: object, Value: ir.ObjectLiteral{Fields: fields, Class: instance.class, Methods: instance.methodList()}},
		ir.Evaluate{Value: ir.Call{Function: initializer, Arguments: arguments}},
		ir.Return{Value: ir.Read{Local: object, Of: ir.Object}},
	}
	l.result.Functions[index] = wrapper
	return nil
}

func (l *lowering) fieldInitializers(declaration *ast.Node, this int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	available := map[string]bool{}
	if l.instance.base != nil {
		for _, field := range l.result.Classes[l.instance.base.class-1].Fields {
			available[field.Name] = true
		}
	}
	resets, err := l.parameterPropertyResets(declaration, this, available)
	if err != nil {
		return nil, err
	}
	statements = append(statements, resets...)
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindPropertyDeclaration || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		of, err := l.typeOf(member.Name())
		if err != nil {
			return nil, err
		}
		if l.uninitializedDeclaration(member) {
			value := zeroValue(of)
			if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			statements = append(statements, ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: l.fieldName(member.Name()), Value: value, Uninitialized: true, Site: l.writeSite(declaration.Name())})
			available[l.fieldName(member.Name())] = true
			continue
		}

		if initializer := member.AsPropertyDeclaration().Initializer; assertionInitializer(initializer) {
			if err := l.initializerReads(initializer, available); err != nil {
				return nil, err
			}
			prefix, present, value, err := l.lazyAssertion(initializer, of)
			if err != nil {
				return nil, err
			}
			empty := ir.Expression(zeroValue(of))
			if of.IsReference() {
				empty = ir.Undefined{Of: of}
			}
			statements = append(statements, prefix...)
			statements = append(statements, ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: l.fieldName(member.Name()), Value: empty, Uninitialized: true, Site: l.writeSite(declaration.Name())}, ir.If{Condition: present, Then: []ir.Statement{ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: l.fieldName(member.Name()), Value: value, Site: l.writeSite(declaration.Name())}}})
			available[l.fieldName(member.Name())] = true
			continue
		}
		value := zeroValue(of)
		if member.AsPropertyDeclaration().Initializer != nil {
			if err := l.initializerReads(member.AsPropertyDeclaration().Initializer, available); err != nil {
				return nil, err
			}
			value, err = l.expression(member.AsPropertyDeclaration().Initializer)
			if err != nil {
				return nil, err
			}
			value = fit(value, of)
		} else if of.IsReference() {
			value = ir.Undefined{Of: of}
		}
		statements = append(statements, ir.SetProperty{Object: ir.Read{Local: this, Of: ir.Object}, Name: l.fieldName(member.Name()), Value: value, Site: l.writeSite(declaration.Name())})
		available[l.fieldName(member.Name())] = true
	}
	return statements, nil
}

func (l *lowering) superCall(node *ast.Node) (ir.Expression, error) {
	if l.instance == nil || l.instance.base == nil || l.this < 0 {
		return nil, l.notYet(node, "super outside a derived constructor")
	}
	arguments := []ir.Expression{ir.Read{Local: l.this, Of: ir.Object}}
	for _, argument := range nodesOf(node.AsCallExpression().Arguments) {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}
	return ir.Call{Function: l.instance.base.initializer, Arguments: arguments}, nil
}

func (l *lowering) superStatement(node *ast.Node) ([]ir.Statement, error) {
	call, err := l.superCall(node)
	if err != nil {
		return nil, err
	}
	statements := []ir.Statement{}
	if l.instance.superStates[node]&superInitialized != 0 {
		// JavaScript constructs another base object before rejecting a second binding.
		// Do not overwrite the already initialized object's fields on this path.
		base := call.(ir.Call)
		second := ir.Call{Function: l.instance.base.constructor, Arguments: base.Arguments[1:], Returns: ir.Object}
		statements = append(statements, ir.If{Condition: ir.Read{Local: l.instance.superReady, Of: ir.Boolean}, Then: []ir.Statement{ir.Evaluate{Value: second}, l.superError("Super constructor may only be called once")}})
	}
	statements = append(statements, ir.Evaluate{Value: call}, ir.Assign{Local: l.instance.superReady, Value: ir.BooleanConstant{Value: true}})
	// Fields run only once super has returned, and before the next derived-body statement.
	declaration := l.classNode.Parent
	initialized, err := l.fieldInitializers(declaration, l.this)
	if err != nil {
		return nil, err
	}
	statements = append(statements, initialized...)
	assigned, err := l.parameterPropertyStores(declaration, l.this)
	return append(statements, assigned...), err
}

func (l *lowering) classInstanceOf(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	declaration := l.classes[l.symbol(ast.SkipParentheses(binary.Right))]
	if declaration == nil {
		return nil, l.notYet(node, "instanceof against a value that isn't a declared class (test against a declared class name, or use an explicit discriminant)")
	}
	identity := 0
	if len(declaration.TypeParameters()) > 0 {
		// JavaScript erases type arguments for instanceof, even when native layouts differ.
		identity = l.classIdentity(declaration)
	} else {
		instance, err := l.instantiate(declaration, l.checker.GetTypeAtLocation(declaration.Name()), node)
		if err != nil {
			return nil, err
		}
		identity = instance.class
	}
	value, err := l.expression(binary.Left)
	if err != nil {
		return nil, err
	}
	return ir.InstanceOf{Value: value, Class: identity}, nil
}

// Finish calls after lowering discovers every instantiated descendant. The analyses see all
// implementations; code generation still uses the receiver's single dynamic method table.
func (l *lowering) finishClassCalls() {
	l.result.MethodTargets = map[int][]int{}
	for _, class := range l.result.Classes {
		for slot, implementation := range class.Methods {
			for parent := &class; parent != nil; {
				if slot < len(parent.Methods) {
					function := parent.Methods[slot]
					if !slices.Contains(l.result.MethodTargets[function], implementation) {
						l.result.MethodTargets[function] = append(l.result.MethodTargets[function], implementation)
					}
				}
				if parent.Base == 0 {
					break
				}
				parent = &l.result.Classes[parent.Base-1]
			}
		}
	}
}

func (l *lowering) noteInheritance(modules []*ast.SourceFile) {
	l.derivedAncestors = map[*ast.Symbol]bool{}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindClassDeclaration && node.Name() != nil {
				typ := l.checker.GetTypeAtLocation(node.Name())
				for _, base := range l.checker.GetBaseTypes(typ) {
					l.derivedAncestors[base.Symbol()] = true
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
}

// A class view needs nominal ancestry, even when TypeScript's structural relation accepts two
// unrelated classes or a plain object. Otherwise a virtual call would trust a table that isn't there.
func (l *lowering) classViewRefusal(node *ast.Node) error {
	if !viewSite(node) || node.Kind == ast.KindParenthesizedExpression {
		return nil
	}
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.Name().Kind == ast.KindArrayBindingPattern {
		return l.classDestructuredView(node, parent.Name())
	}
	target := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if target == nil {
		return nil
	}
	// Fresh literals are built as their contextual type; their explicit values are
	// checked at their own sites. Spreads still need the whole inherited shape checked.
	if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindArrayLiteralExpression {
		spread := false
		for _, child := range literalParts(node) {
			if child.Kind == ast.KindSpreadAssignment || child.Kind == ast.KindSpreadElement {
				spread = true
			}
		}
		if !spread && !l.nominalLiteralTarget(target) {
			return nil
		}
	}
	source := l.checker.GetTypeAtLocation(node)
	mismatch := l.nominalMismatch(source, target, map[[2]*checker.Type]bool{})
	if mismatch == nil {
		return nil
	}
	target = mismatch
	return &Refused{Where: l.program.Where(node), What: "a value without nominal ancestry seen as " + l.checker.TypeToString(target), Fix: "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"}
}

func (l *lowering) nominalAncestor(source, target *checker.Type, seen map[[2]*checker.Type]bool) bool {
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(source); constraint != nil {
			source = constraint
		}
	}
	if weak := l.weakTarget(source); weak != nil {
		source = weak
	}
	if !isClassInstance(source) {
		return false
	}
	if source.Symbol() == target.Symbol() {
		from, to := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		if !l.sameClassArguments(from, to) {
			return false
		}
		for index := range from {
			if !l.enumAssignable(from[index], to[index]) || !l.enumAssignable(to[index], from[index]) || l.nominalMismatch(from[index], to[index], seen) != nil || l.nominalMismatch(to[index], from[index], seen) != nil {
				return false
			}
		}
		return true
	}
	for _, base := range l.classBases(source) {
		if l.nominalAncestor(base, target, seen) {
			return true
		}
	}
	return false
}

// Initializers may read fields already initialized, including the base's after super. A future
// field would hold undefined despite its declared type; a method may read future derived fields.
func (l *lowering) initializerReads(node *ast.Node, available map[string]bool) error {
	seen := map[*ast.Node]bool{}
	var examine func(*ast.Node, []string) error
	examine = func(body *ast.Node, path []string) error {
		if body == nil || seen[body] {
			return nil
		}
		seen[body] = true
		var refused error
		var visit ast.Visitor
		visit = func(child *ast.Node) bool {
			if refused != nil {
				return true
			}
			if ast.IsPartOfTypeNode(child) {
				return false
			}
			if child.Kind == ast.KindPropertyAccessExpression {
				object := ast.SkipParentheses(child.AsPropertyAccessExpression().Expression)
				if object.Kind == ast.KindThisKeyword || object.Kind == ast.KindSuperKeyword {
					symbol := l.checker.GetSymbolAtLocation(child.Name())
					if object.Kind == ast.KindThisKeyword && child.Name().Kind != ast.KindPrivateIdentifier {
						symbol = l.checker.GetPropertyOfType(l.classType, child.Name().Text())
					}
					if symbol != nil {
						for _, member := range symbol.Declarations {
							step := child.Name().Text()
							if member.Parent != nil && member.Parent.Name() != nil {
								step = member.Parent.Name().Text() + "." + step
							}
							next := append(append([]string{}, path...), step)
							if member.Kind == ast.KindMethodDeclaration || member.Kind == ast.KindGetAccessor {
								refused = examine(member.Body(), next)
							} else if member.Kind == ast.KindPropertyDeclaration && !available[l.fieldName(child.Name())] {
								refused = &Refused{Where: l.program.Where(child), What: "a field initializer reading an uninitialized field through " + strings.Join(next, " -> "), Fix: "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set (adamic/initialized-this)"}
							}
							if refused != nil {
								return true
							}
						}
						return false
					}
				}
			}
			if child.Kind == ast.KindThisKeyword || child.Kind == ast.KindSuperKeyword {
				refused = &Refused{Where: l.program.Where(child), What: "this in a field initializer before the fields it reads are initialized", Fix: "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set (adamic/initialized-this)"}
				return true
			}
			return child.ForEachChild(visit)
		}
		visit(body)
		return refused
	}
	return examine(node, nil)
}

// Private names with the same spelling in a base and a derived class are separate JavaScript slots.
func (l *lowering) fieldName(name *ast.Node) string {
	if name.Kind != ast.KindPrivateIdentifier {
		return name.Text()
	}
	symbol := l.checker.GetSymbolAtLocation(name)
	if symbol != nil && len(symbol.Declarations) > 0 {
		class := symbol.Declarations[0].Parent
		if class != nil && ast.HasSyntacticModifier(symbol.Declarations[0], ast.ModifierFlagsStatic) {
			if lowered := l.statics[l.symbol(class.Name())]; lowered != nil {
				return memberKey(name, lowered.class)
			}
		}
		if class != nil && class.Kind == ast.KindClassDeclaration {
			if l.instance != nil && l.classNode != nil && l.classNode.Parent == class {
				return name.Text() + "@" + strconv.Itoa(l.instance.class)
			}
			if instance := l.instances[l.classInstanceKey(class, nil)]; instance != nil {
				return name.Text() + "@" + strconv.Itoa(instance.class)
			}
			return l.program.Where(class) + ":" + name.Text()
		}
	}
	return name.Text()
}

// Assignment to a nominal class also requires ancestry; structural compatibility alone cannot
// provide the dynamic table an override's caller expects.
func (l *lowering) classAssignable(from, to *checker.Type) bool {
	return l.checker.IsTypeAssignableTo(from, to) && l.nominalMismatch(from, to, map[[2]*checker.Type]bool{}) == nil
}

// Follow nominal class slots inside structural containers too. A plain object inside A[] lacks
// the same method table as a plain object assigned directly to A. Mutable slots need both views.
func (l *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type {
	if from == nil || to == nil || from == to || from.Flags()&(checker.TypeFlagsNever|checker.TypeFlagsUndefined) != 0 || seen[[2]*checker.Type{from, to}] {
		return nil
	}
	seen[[2]*checker.Type{from, to}] = true
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if found := l.nominalMismatch(member, to, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		var failure *checker.Type
		for _, member := range to.Types() {
			if !l.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			if found := l.nominalMismatch(from, member, seen); found != nil {
				failure = found
			} else {
				return nil
			}
		}
		return failure
	}
	if weak := l.weakTarget(from); weak != nil {
		from = weak
	}
	if weak := l.weakTarget(to); weak != nil {
		to = weak
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(from); constraint != nil {
			return l.nominalMismatch(constraint, to, seen)
		}
		return nil
	}
	if l.isStaticType(to) {
		wanted := l.staticClass(to, nil)
		actual := l.staticClass(from, nil)
		for actual != nil {
			if actual == wanted {
				return nil
			}
			actual = l.staticBase(actual)
		}
		return to
	}
	if isClassInstance(to) {
		if symbol := to.Symbol(); symbol != nil && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			return nil
		}
		if !l.nominalAncestor(from, to, seen) {
			return to
		}
		return nil
	}
	if !l.structured(from) || !l.structured(to) {
		return nil
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		for index, parameter := range toSignatures[0].Parameters() {
			if index >= len(fromSignatures[0].Parameters()) {
				break
			}
			if found := l.nominalMismatch(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(fromSignatures[0].Parameters()[index]), seen); found != nil {
				return found
			}
		}
		return l.nominalMismatch(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), seen)
	}
	for _, target := range l.containers(to) {
		for _, source := range l.containers(from) {
			if !l.sameContainer(source, target) {
				continue
			}
			fromArguments, toArguments := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			mutable := !l.isLibraryType(target, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") && !(checker.IsTupleType(target) && target.TargetTupleType().IsReadonly())
			for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
				if found := l.nominalMismatch(fromArguments[index], toArguments[index], seen); found != nil {
					return found
				}
				if mutable {
					if found := l.nominalMismatch(toArguments[index], fromArguments[index], seen); found != nil {
						return found
					}
				}
			}
		}
	}
	if len(l.containers(to)) > 0 {
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(to) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		inside := l.checker.GetPropertyOfType(from, property.Name)
		if inside == nil {
			continue
		}
		source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(property)
		if found := l.nominalMismatch(source, target, seen); found != nil {
			return found
		}
		if !l.checker.IsReadonlySymbol(property) {
			if found := l.nominalMismatch(target, source, seen); found != nil {
				return found
			}
		}
	}
	return nil
}

// An inferred destructuring context can be the union of the tuple's elements. Compare the
// positional values with their bound names instead of treating that union as a writable array.
func (l *lowering) classDestructuredView(node, pattern *ast.Node) error {
	own := l.checker.GetTypeAtLocation(node)
	elements := l.checker.GetTypeArguments(own)
	tuple := checker.IsTupleType(own)
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		if !tuple {
			index = 0
		}
		if index >= len(elements) {
			continue
		}
		target := l.checker.GetTypeAtLocation(binding.Name())
		if mismatch := l.nominalMismatch(elements[index], target, map[[2]*checker.Type]bool{}); mismatch != nil {
			return &Refused{Where: l.program.Where(binding), What: "a destructured value without nominal ancestry seen as " + l.checker.TypeToString(mismatch), Fix: "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"}
		}
	}
	return nil
}

// The checker mangles private symbol names; the IR qualifies them by class ID. Cycle proofs
// must match the same declaration in both names, including inherited and generic private slots.
func (l *lowering) cycleFieldMatches(holder *checker.Type, field, written string) bool {
	if field == written {
		return true
	}
	property := l.checker.GetPropertyOfType(holder, field)
	if property == nil || len(property.Declarations) != 1 {
		return false
	}
	name := property.Declarations[0].Name()
	if name == nil || name.Kind != ast.KindPrivateIdentifier {
		return false
	}
	class := property.Declarations[0].Parent
	if ast.HasSyntacticModifier(property.Declarations[0], ast.ModifierFlagsStatic) {
		if lowered := l.statics[l.symbol(class.Name())]; lowered != nil {
			return written == memberKey(name, lowered.class)
		}
	}
	for key, instance := range l.instances {
		if !l.classKeyMatches(key, class) {
			continue
		}
		if written == name.Text()+"@"+strconv.Itoa(instance.class) {
			return true
		}
	}
	return false
}

func (l *lowering) cycleFieldName(field *ast.Symbol) string {
	if len(field.Declarations) == 1 {
		if name := field.Declarations[0].Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			return name.Text()
		}
	}
	return field.Name
}

// All monomorphizations of a source class have one erased identity. An identity-only
// descriptor lets instanceof name a generic class before any concrete instance is made.
func (l *lowering) classDefinition(declaration *ast.Node) int {
	for key, instance := range l.instances {
		if l.classKeyMatches(key, declaration) {
			return l.result.Classes[instance.class-1].Definition
		}
	}
	return len(l.result.Classes) + 1
}

func (l *lowering) classIdentity(declaration *ast.Node) int {
	definition := l.classDefinition(declaration)
	for index, class := range l.result.Classes {
		if class.Definition == definition {
			return index + 1
		}
	}
	identity := len(l.result.Classes) + 1
	l.result.Classes = append(l.result.Classes, ir.Class{Name: declaration.Name().Text() + "_identity", Definition: definition, Constructor: -1})
	if l.instances == nil {
		l.instances = map[string]*instance{}
	}
	key := l.classKeyPrefix(declaration) + ",identity"
	l.instances[key] = &instance{class: identity, constructor: -1, initializer: -1}
	return identity
}

// An inherited method's receiver can be a descendant with different type arguments.
// Find the declaration's actual instantiated view instead of substituting by position.
func (l *lowering) classView(proven *checker.Type, declaration *ast.Node) *checker.Type {
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	if weak := l.weakTarget(proven); weak != nil {
		proven = weak
	}
	if proven.Symbol() == l.symbol(declaration.Name()) {
		return proven
	}
	for _, base := range l.classBases(proven) {
		if found := l.classView(base, declaration); found != nil {
			return found
		}
	}
	return nil
}

// The checker exposes generic bases on the declaration. Substitute the receiver's
// actual arguments before choosing a native base layout or proving nominal ancestry.
func (l *lowering) classBases(proven *checker.Type) []*checker.Type {
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	bases := l.checker.GetBaseTypes(target)
	declaration := l.classNodeFor(proven)
	if declaration == nil || declaration.Kind != ast.KindClassDeclaration {
		return bases
	}
	mapper := l.typeMapperOf(declaration, proven)
	if mapper == nil {
		return bases
	}
	concrete := make([]*checker.Type, len(bases))
	for index, base := range bases {
		concrete[index] = instantiateType(l.checker, base, mapper)
	}
	return concrete
}

func literalParts(node *ast.Node) []*ast.Node {
	if node.Kind == ast.KindObjectLiteralExpression {
		return node.AsObjectLiteralExpression().Properties.Nodes
	}
	return node.AsArrayLiteralExpression().Elements.Nodes
}

func (l *lowering) nominalLiteralTarget(proven *checker.Type) bool {
	if weak := l.weakTarget(proven); weak != nil {
		proven = weak
	}
	if isClassInstance(proven) {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.nominalLiteralTarget(member) {
				return true
			}
		}
	}
	return false
}
