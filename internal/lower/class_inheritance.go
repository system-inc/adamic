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

func (lowering *lowering) baseInstance(declaration *ast.Node, classType *checker.Type) (*instance, error) {
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
			continue
		}
		types := clause.AsHeritageClause().Types.Nodes
		if len(types) != 1 ||
			!ast.IsIdentifier(ast.SkipParentheses(types[0].AsExpressionWithTypeArguments().Expression)) {
			return nil, lowering.notYet(clause, "a computed class base; name the base class directly")
		}
		bases := lowering.classBases(classType)
		if len(bases) != 1 {
			return nil, lowering.notYet(clause, "a class base whose type isn't known")
		}
		baseType := bases[0]
		symbol := baseType.Symbol()
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindClassDeclaration {
			return nil, lowering.notYet(clause, "a base that isn't a declared class")
		}
		base := symbol.Declarations[0]
		return lowering.instantiate(base, baseType, clause)
	}
	return nil, nil
}

// classOverrideRefusal checks a class declaration's overrides against its bases.
func (lowering *lowering) classOverrideRefusal(node *ast.Node) error {
	if node.Kind == ast.KindClassDeclaration {
		if checkOverridesError := lowering.checkOverrides(
			node,
			lowering.checker.GetTypeAtLocation(node.Name()),
		); checkOverridesError != nil {
			return checkOverridesError
		}
	}
	return nil
}

func init() { refusalRules.register(3000, "classOverrideRefusal", (*lowering).classOverrideRefusal) }

// TypeScript compares methods bivariantly and mutable properties covariantly. Neither is safe
// through a base reference: arguments flow into the override, and fields can be written through it.
func (lowering *lowering) checkOverrides(declaration *ast.Node, classType *checker.Type) error {
	bases := lowering.classBases(classType)
	if len(bases) == 0 {
		return nil
	}
	return lowering.checkMemberOverrides(declaration, classType, bases[0], false)
}

func (lowering *lowering) checkMemberOverrides(
	declaration *ast.Node,
	classType, base *checker.Type,
	static bool,
) error {
	checkABI := static ||
		len(declaration.TypeParameters()) == 0 ||
		classType != lowering.checker.GetTypeAtLocation(declaration.Name())
	for _, member := range classMembersWithParameters(declaration) {
		if member.Name() == nil || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) != static {
			continue
		}
		// Computed names are checker property identities, not textual AST names.
		// In particular Symbol.iterator must use its unique-symbol key.
		symbol := lowering.checker.GetSymbolOfDeclaration(member)
		if symbol == nil {
			return lowering.notYet(member, "a class member without a resolved property identity")
		}
		key := symbol.Name
		inherited := lowering.checker.GetPropertyOfType(base, key)
		if inherited == nil {
			continue
		}
		own := lowering.checker.GetTypeAtLocation(member.Name())
		if property := lowering.checker.GetPropertyOfType(classType, key); property != nil {
			own = lowering.checker.GetTypeOfSymbol(property)
		}
		previous := lowering.checker.GetTypeOfSymbol(inherited)
		refuse := func(what, fix string) error {
			return &Refused{Where: lowering.program.Where(member), What: what, Fix: fix}
		}
		if (member.Kind == ast.KindMethodDeclaration) != (inherited.Flags&ast.SymbolFlagsMethod != 0) {
			return refuse(
				"an inherited method replaced by a field, or a field replaced by a method",
				"keep the inherited member kind; use a different name for the new member",
			)
		}
		if accessorMember(member) != accessorSymbol(inherited) {
			return refuse(
				"an inherited data property replaced by an accessor, or an accessor replaced by data",
				"keep the inherited member kind; use another name for the new property",
			)
		}
		if accessorMember(member) {
			ownGet, ownSet, baseGet, baseSet := false, false, false, false
			for _, candidate := range declaration.Members() {
				if candidate.Name() != nil &&
					lowering.checker.GetSymbolOfDeclaration(candidate) != nil &&
					lowering.checker.GetSymbolOfDeclaration(candidate).Name == key &&
					ast.HasSyntacticModifier(candidate, ast.ModifierFlagsStatic) == static {
					ownGet = ownGet || candidate.Kind == ast.KindGetAccessor
					ownSet = ownSet || candidate.Kind == ast.KindSetAccessor
				}
			}
			for _, candidate := range inherited.Declarations {
				baseGet = baseGet || candidate.Kind == ast.KindGetAccessor
				baseSet = baseSet || candidate.Kind == ast.KindSetAccessor
			}
			if (baseGet && !ownGet) || (baseSet && !ownSet) {
				return refuse(
					"an accessor override that hides the inherited getter or setter",
					"override both halves of the inherited descriptor; delegate an unchanged half to super",
				)
			}
			if member.Kind == ast.KindSetAccessor {
				for _, candidate := range inherited.Declarations {
					if candidate.Kind != ast.KindSetAccessor {
						continue
					}
					old := lowering.checker.GetSignatureFromDeclaration(candidate)
					next := lowering.checker.GetSignatureFromDeclaration(member)
					accepts := lowering.checker.GetTypeOfSymbol(old.Parameters()[0])
					override := lowering.checker.GetTypeOfSymbol(next.Parameters()[0])
					if !static {
						accepts = instantiateType(
							lowering.checker,
							accepts,
							lowering.typeMapperOf(candidate.Parent, base),
						)
						override = instantiateType(
							lowering.checker,
							override,
							lowering.typeMapperOf(declaration, classType),
						)
					}
					if !lowering.classAssignable(accepts, override) ||
						lowering.widened(accepts, override, map[[2]*checker.Type]bool{}) != nil {
						return refuse(
							"an accessor override that narrows a setter parameter (adamic/contravariant-override)",
							"accept the base setter's parameter type or a wider type; narrow it inside the setter",
						)
					}
				}
			}
			if ownGet &&
				(!lowering.classAssignable(own, previous) ||
					lowering.widened(own, previous, map[[2]*checker.Type]bool{}) != nil) {
				return refuse(
					"an accessor override with an unsafe read or write type (adamic/invariant-mutable)",
					"keep the inherited accessor type; narrow values inside the accessor",
				)
			}
		}
		if member.Kind == ast.KindPropertyDeclaration || parameterProperty(member) {
			if !lowering.checker.IsReadonlySymbol(inherited) && (!lowering.classAssignable(previous, own) ||
				!lowering.classAssignable(own, previous) ||
				lowering.widened(previous, own, map[[2]*checker.Type]bool{}) != nil ||
				lowering.widened(own, previous, map[[2]*checker.Type]bool{}) != nil) {
				return refuse(
					"a mutable inherited field redeclared with a different type (adamic/invariant-mutable)",
					"keep the base field's type; narrow a local after reading it, or make the field readonly in the base",
				)
			}
			if lowering.checker.IsReadonlySymbol(inherited) &&
				(lowering.widened(own, previous, map[[2]*checker.Type]bool{}) != nil ||
					!lowering.classAssignable(own, previous)) {
				return refuse(
					"an inherited readonly field whose mutable contents are narrowed (adamic/invariant-mutable)",
					"keep mutable contents at the base type, or make the contents readonly in the base too",
				)
			}
			if lowering.checker.IsReadonlySymbol(inherited) &&
				!ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly) {
				return refuse("a readonly inherited field redeclared mutable", "keep the inherited field readonly")
			}
			from, okFrom := lowering.representation(previous)
			to, okTo := lowering.representation(own)
			if checkABI && (!okFrom || !okTo || from != to) {
				return lowering.notYet(member, "an inherited field override with a different native representation")
			}
		}
		if member.Kind != ast.KindMethodDeclaration {
			continue
		}
		oldSignatures := lowering.checker.GetSignaturesOfType(previous, checker.SignatureKindCall)
		newSignatures := lowering.checker.GetSignaturesOfType(own, checker.SignatureKindCall)
		if len(oldSignatures) != 1 || len(newSignatures) != 1 {
			return lowering.notYet(member, "an overloaded override")
		}
		old, next := oldSignatures[0], newSignatures[0]
		if len(old.Parameters()) != len(next.Parameters()) {
			return lowering.notYet(
				member,
				"an override with a different parameter count; keep the base method's parameter count",
			)
		}
		for index, parameter := range old.Parameters() {
			accepts := lowering.checker.GetTypeOfSymbol(parameter)
			override := lowering.checker.GetTypeOfSymbol(next.Parameters()[index])
			a, knownA := lowering.overrideParameterRepresentation(parameter)
			b, knownB := lowering.overrideParameterRepresentation(next.Parameters()[index])
			if checkABI && (!knownA || !knownB || a != b) {
				return lowering.notYet(member, "an override with a different native representation for parameter "+
					strconv.Quote(next.Parameters()[index].Name)+
					"; keep the base method's parameter form")
			}
			if !lowering.classAssignable(accepts, override) ||
				lowering.widened(accepts, override, map[[2]*checker.Type]bool{}) != nil {
				return refuse(
					"an override that narrows a method parameter (adamic/contravariant-override)",
					"accept the base method's parameter type or a wider type; narrow it inside the method",
				)
			}
		}
		oldResult := lowering.checker.GetReturnTypeOfSignature(old)
		newResult := lowering.checker.GetReturnTypeOfSignature(next)
		a, knownA := lowering.overrideResultRepresentation(oldResult)
		b, knownB := lowering.overrideResultRepresentation(newResult)
		if checkABI && (!knownA || !knownB || a != b) {
			return lowering.notYet(
				member,
				"an override with a different native result representation; keep the base method's result form",
			)
		}
		if !lowering.classAssignable(newResult, oldResult) ||
			lowering.widened(newResult, oldResult, map[[2]*checker.Type]bool{}) != nil {
			return refuse(
				"an override that widens its return type",
				"return the base method's result type or a subtype",
			)
		}
	}
	return nil
}

// Match signature lowering, including the incoming optional form of a defaulted parameter.
// The signature's symbol carries the instantiated type; its declaration carries the default.
func (lowering *lowering) overrideParameterRepresentation(parameter *ast.Symbol) (ir.Type, bool) {
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
	representation, known := lowering.representation(lowering.checker.GetTypeOfSymbol(parameter))
	if declared.Initializer != nil || declared.QuestionToken != nil {
		representation = ir.Maybe(representation)
	}
	return representation, known
}

// Named methods returning void or never have no C result, as signature lowering does.
func (lowering *lowering) overrideResultRepresentation(result *checker.Type) (ir.Type, bool) {
	if result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		return 0, true
	}
	return lowering.representation(result)
}

// Constructors initialize an already allocated object. A separate allocator fixes its dynamic
// identity before even the base constructor runs, so base method calls can dispatch dynamically.
func (lowering *lowering) inheritanceConstructor(index int, declaration *ast.Node) error {
	// Class initializers have their own receiver frame, even when instantiation
	// is discovered while lowering an unrelated closure. Only arrows created
	// inside this constructor may capture its locals.
	outerClosures := lowering.closures
	lowering.closures = nil
	defer func() { lowering.closures = outerClosures }()
	instance := lowering.instance
	var constructor *ast.Node
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindConstructor {
			constructor = member
		}
	}
	instance.initializer = len(lowering.result.Functions)
	initializer := instance.initializer
	source := lowering.functionSource(declaration)
	source.Role = "initialize"
	source.Declaration = append(append([]string{}, source.Declaration...), "initialize")
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: lowering.result.Functions[index].Name + "_initialize", Source: source},
	)
	this := lowering.thisLocal(initializer)
	outerThis, outerIndex, outerFunction := lowering.this, lowering.functionIndex, lowering.function
	context := lowering.result.Functions[initializer]
	lowering.this, lowering.functionIndex, lowering.function = this, initializer, &context
	defer func() {
		lowering.this, lowering.functionIndex, lowering.function = outerThis, outerIndex, outerFunction
	}()
	if constructor != nil {
		if signatureError := lowering.signature(initializer, constructor, this); signatureError != nil {
			return signatureError
		}
		lowering.result.Functions[initializer].Parameters = append(
			[]int{this},
			lowering.result.Functions[initializer].Parameters...,
		)
		lowering.result.Functions[initializer].Boundary.Parameters = append(
			[]ir.BoundaryParameter{{}},
			lowering.result.Functions[initializer].Boundary.Parameters...,
		)
	} else {
		lowering.result.Functions[initializer].Parameters = []int{this}
		lowering.result.Functions[initializer].Boundary.Parameters = []ir.BoundaryParameter{{}}
		if instance.base != nil {
			for _, parameter := range lowering.result.Functions[instance.base.initializer].Parameters[1:] {
				local := len(lowering.result.Locals)
				incoming := lowering.result.Locals[parameter]
				incoming.Function, incoming.Captured, incoming.Borrowed = initializer, false, false
				lowering.result.Locals = append(lowering.result.Locals, incoming)
				lowering.result.Functions[initializer].Parameters = append(
					lowering.result.Functions[initializer].Parameters,
					local,
				)
				lowering.result.Functions[initializer].Boundary.Parameters = append(
					lowering.result.Functions[initializer].Boundary.Parameters,
					ir.BoundaryParameter{},
				)
			}
		}
	}
	lowering.result.Functions[initializer].Receiver = true
	if constructor == nil && instance.base != nil {
		lowering.result.Functions[initializer].RestElement = lowering.result.Functions[instance.base.initializer].RestElement
		lowering.result.Functions[initializer].ForwardsArguments = instance.base.initializer + 1
	}
	// The layout contains inherited fields first, but only this class's initializers run here.
	fields := append([]ir.Field{}, lowering.result.Classes[instance.class-1].Fields...)
	for _, member := range classMembersWithParameters(declaration) {
		if member.Kind != ast.KindPropertyDeclaration && !parameterProperty(member) {
			continue
		}
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract) {
			return lowering.notYet(member, "a declare or abstract class field")
		}
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if !ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier {
			return lowering.notYet(member, "a field with a computed name")
		}
		of, typeOfError := lowering.typeOf(member.Name())
		if typeOfError != nil {
			return typeOfError
		}
		// Required boolean-or-undefined fields have a packed scalar slot.
		// Optional presence and boxed union class layouts remain separate work.
		if slotless(of) &&
			(of != ir.MaybeBoolean || member.PostfixToken() != nil) &&
			lowering.placeholderOrigin(member.Name()) == "" {
			return lowering.notYet(
				member,
				"a field of type "+lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(member.Name())),
			)
		}
		field := ir.Field{
			Name:          lowering.fieldName(member.Name()),
			Value:         zeroValue(of),
			Private:       member.Name().Kind == ast.KindPrivateIdentifier,
			Uninitialized: lowering.uninitializedDeclaration(member),
			Unset:         lowering.placeholderOrigin(member.Name()) != "",
		}
		// An uninitialized reference still has its declared representation for the shape bitmap.
		if of.IsReference() {
			field.Value = ir.Undefined{Of: of}
		}
		found := false
		for index := range fields {
			if fields[index].Name == field.Name {
				fields[index] = field
				found = true
				break
			}
		}
		if !found {
			fields = append(fields, field)
		}
	}
	lowering.result.Classes[instance.class-1].Fields = fields
	if instance.base == nil {
		initialized, fieldInitializersError := lowering.fieldInitializers(declaration, this)
		if fieldInitializersError != nil {
			return fieldInitializersError
		}
		lowering.result.Functions[initializer].Body = initialized
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
			return lowering.notYet(returnValue, "a constructor returning a replacement value")
		}
		state := uint8(0)
		if instance.base != nil {
			var prepareSuperError error
			state, prepareSuperError = lowering.prepareSuper(constructor, initializer)
			if prepareSuperError != nil {
				return prepareSuperError
			}
		}

		outerUnset := lowering.unsetUntil
		lowering.unsetUntil = lastFieldAssignment(declaration, constructor)
		lowerFunctionError := lowering.lowerFunction(initializer, constructor, this)
		lowering.unsetUntil = outerUnset
		if lowerFunctionError != nil {
			return lowerFunctionError
		}
		if state&superUninitialized != 0 {
			lowering.result.Functions[initializer].Body = append(
				lowering.result.Functions[initializer].Body,
				lowering.superEnd(),
			)
		}
	} else if instance.base != nil {
		arguments := []ir.Expression{}
		for _, parameter := range lowering.result.Functions[initializer].Parameters {
			arguments = append(arguments, ir.Read{Local: parameter, Of: lowering.result.Locals[parameter].Type})
		}
		lowering.result.Functions[initializer].Body = []ir.Statement{ir.Evaluate{Value: ir.Call{
			Function:     instance.base.initializer,
			Arguments:    arguments,
			ForwardCount: true,
			RestPacked:   true,
		}}}
		initialized, fieldInitializersError := lowering.fieldInitializers(declaration, this)
		if fieldInitializersError != nil {
			return fieldInitializersError
		}
		lowering.result.Functions[initializer].Body = append(
			lowering.result.Functions[initializer].Body,
			initialized...,
		)
	}
	if instance.base == nil && !instance.hasDescendants {
		// Keep creation and writes in one function when no derived constructor needs this
		// initializer. Freshness then knows this is new at each write, not just at return.
		body := lowering.result.Functions[initializer]
		for local := range lowering.result.Locals {
			if lowering.result.Locals[local].Function == initializer {
				lowering.result.Locals[local].Function = index
			}
		}
		wrapper := lowering.result.Functions[index]
		wrapper.Parameters = append([]int{}, body.Parameters[1:]...)
		wrapper.ArgumentsCount = body.ArgumentsCount
		wrapper.ForwardsArguments = body.ForwardsArguments
		wrapper.RestElement = body.RestElement
		wrapper.Body = append([]ir.Statement{ir.Declare{Local: this, Value: ir.ObjectLiteral{
			ProgramRegion: lowering.programClass(instance.class),
			Fields:        fields,
			Class:         instance.class,
			Methods:       instance.methodList(),
		}}}, body.Body...)
		wrapper.Body = append(wrapper.Body, ir.Return{Value: ir.Read{Local: this, Of: ir.Object}})
		lowering.result.Functions[index] = wrapper
		lowering.result.Functions[initializer] = ir.Function{Name: body.Name, Source: body.Source}
		return nil
	}
	// The public constructor owns its object, including when an initializer throws.
	wrapper := lowering.result.Functions[index]
	wrapper.ForwardsArguments = initializer + 1
	wrapper.RestElement = lowering.result.Functions[initializer].RestElement
	arguments := []ir.Expression{}
	object := len(lowering.result.Locals)
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: index})
	lowering.noteLocal(object, lowering.classType, declaration.Name())
	arguments = append(arguments, ir.Read{Local: object, Of: ir.Object})
	for _, parameter := range lowering.result.Functions[initializer].Parameters[1:] {
		incoming := lowering.result.Locals[parameter]
		incoming.Function, incoming.Captured, incoming.Borrowed = index, false, false
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(lowering.result.Locals, incoming)
		wrapper.Parameters = append(wrapper.Parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: incoming.Type})
	}
	wrapper.Body = []ir.Statement{
		ir.Declare{Local: object, Value: ir.ObjectLiteral{
			ProgramRegion: lowering.programClass(instance.class),
			Fields:        fields,
			Class:         instance.class,
			Methods:       instance.methodList(),
		}},
		ir.Evaluate{Value: ir.Call{Function: initializer, Arguments: arguments, ForwardCount: true, RestPacked: true}},
		ir.Return{Value: ir.Read{Local: object, Of: ir.Object}},
	}
	lowering.result.Functions[index] = wrapper
	return nil
}

func (lowering *lowering) fieldInitializers(declaration *ast.Node, this int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	available := map[string]bool{}
	if lowering.instance.base != nil {
		for _, field := range lowering.result.Classes[lowering.instance.base.class-1].Fields {
			available[field.Name] = true
		}
	}
	resets, parameterPropertyResetsError := lowering.parameterPropertyResets(declaration, this, available)
	if parameterPropertyResetsError != nil {
		return nil, parameterPropertyResetsError
	}
	statements = append(statements, resets...)
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindPropertyDeclaration || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		of, typeOfError := lowering.typeOf(member.Name())
		if typeOfError != nil {
			return nil, typeOfError
		}
		if lowering.uninitializedDeclaration(member) {
			value := zeroValue(of)
			if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			if lowering.placeholderOrigin(member.Name()) != "" {
				value = lowering.placeholderInitialValue(member.AsPropertyDeclaration().Initializer, of)
			}
			statements = append(statements, ir.SetProperty{
				Object:        ir.Read{Local: this, Of: ir.Object},
				Name:          lowering.fieldName(member.Name()),
				Value:         value,
				Uninitialized: true,
				Unset:         lowering.placeholderOrigin(member.Name()) != "",
				Site:          lowering.writeSite(declaration.Name()),
			})
			available[lowering.fieldName(member.Name())] = true
			continue
		}

		value := zeroValue(of)
		if member.AsPropertyDeclaration().Initializer != nil {
			if initializerReadsError := lowering.initializerReads(
				member.AsPropertyDeclaration().Initializer,
				available,
			); initializerReadsError != nil {
				return nil, initializerReadsError
			}
			var expressionError error
			value, expressionError = lowering.expression(member.AsPropertyDeclaration().Initializer)
			if expressionError != nil {
				return nil, expressionError
			}
			value = fit(value, of)
		} else if of.IsReference() {
			value = ir.Undefined{Of: of}
		}
		statements = append(statements, ir.SetProperty{
			Object: ir.Read{Local: this, Of: ir.Object},
			Name:   lowering.fieldName(member.Name()),
			Value:  value,
			Site:   lowering.writeSite(declaration.Name()),
		})
		available[lowering.fieldName(member.Name())] = true
	}
	return statements, nil
}

func (lowering *lowering) superCall(node *ast.Node) (ir.Expression, error) {
	if lowering.instance == nil || lowering.instance.base == nil || lowering.this < 0 {
		return nil, lowering.notYet(node, "super outside a derived constructor")
	}
	arguments := []ir.Expression{ir.Read{Local: lowering.this, Of: ir.Object}}
	for _, argument := range nodesOf(node.AsCallExpression().Arguments) {
		value, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, expressionError
		}
		arguments = append(arguments, value)
	}
	return ir.Call{Function: lowering.instance.base.initializer, Arguments: arguments}, nil
}

func (lowering *lowering) superStatement(node *ast.Node) ([]ir.Statement, error) {
	call, superCallError := lowering.superCall(node)
	if superCallError != nil {
		return nil, superCallError
	}
	statements := []ir.Statement{}
	if lowering.instance.superStates[node]&superInitialized != 0 {
		// JavaScript constructs another base object before rejecting a second binding.
		// Do not overwrite the already initialized object's fields on this path.
		base := call.(ir.Call)
		second := ir.Call{
			Function:  lowering.instance.base.constructor,
			Arguments: base.Arguments[1:],
			Returns:   ir.Object,
		}
		statements = append(statements, ir.If{
			Condition: ir.Read{Local: lowering.instance.superReady, Of: ir.Boolean},
			Then: []ir.Statement{
				ir.Evaluate{Value: second},
				lowering.superError("Super constructor may only be called once"),
			},
		})
	}
	statements = append(
		statements,
		ir.Evaluate{Value: call},
		ir.Assign{Local: lowering.instance.superReady, Value: ir.BooleanConstant{Value: true}},
	)
	// Fields run only once super has returned, and before the next derived-body statement.
	declaration := lowering.classNode.Parent
	initialized, fieldInitializersError := lowering.fieldInitializers(declaration, lowering.this)
	if fieldInitializersError != nil {
		return nil, fieldInitializersError
	}
	statements = append(statements, initialized...)
	assigned, parameterPropertyStoresError := lowering.parameterPropertyStores(declaration, lowering.this)
	return append(statements, assigned...), parameterPropertyStoresError
}

func (lowering *lowering) classInstanceOf(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	declaration := lowering.classes[lowering.symbol(ast.SkipParentheses(binary.Right))]
	if declaration == nil {
		return nil, lowering.notYet(node, "instanceof against a value that isn't a declared class")
	}
	identity := 0
	if len(declaration.TypeParameters()) > 0 {
		// JavaScript erases type arguments for instanceof, even when native layouts differ.
		identity = lowering.classIdentity(declaration)
	} else {
		instance, instantiateError := lowering.instantiate(
			declaration,
			lowering.checker.GetTypeAtLocation(declaration.Name()),
			node,
		)
		if instantiateError != nil {
			return nil, instantiateError
		}
		identity = instance.class
	}
	value, expressionError := lowering.expression(binary.Left)
	if expressionError != nil {
		return nil, expressionError
	}
	return ir.InstanceOf{Value: value, Class: identity}, nil
}

// Finish calls after lowering discovers every instantiated descendant. The analyses see all
// implementations; code generation still uses the receiver's single dynamic method table.
func (lowering *lowering) finishClassCalls() {
	lowering.result.MethodTargets = map[int][]int{}
	for _, class := range lowering.result.Classes {
		for slot, implementation := range class.Methods {
			for parent := &class; parent != nil; {
				if slot < len(parent.Methods) {
					function := parent.Methods[slot]
					if !slices.Contains(lowering.result.MethodTargets[function], implementation) {
						lowering.result.MethodTargets[function] = append(
							lowering.result.MethodTargets[function],
							implementation,
						)
					}
				}
				if parent.Base == 0 {
					break
				}
				parent = &lowering.result.Classes[parent.Base-1]
			}
		}
	}
}

func (lowering *lowering) noteInheritance(modules []*ast.SourceFile) {
	lowering.derivedAncestors = map[*ast.Symbol]bool{}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindClassDeclaration && node.Name() != nil {
				typ := lowering.checker.GetTypeAtLocation(node.Name())
				for _, base := range lowering.checker.GetBaseTypes(typ) {
					lowering.derivedAncestors[base.Symbol()] = true
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
}

// A class view needs nominal ancestry, even when TypeScript's structural relation accepts two
// unrelated classes or a plain object. Otherwise a virtual call would trust a table that isn't there.
func (lowering *lowering) classViewRefusal(node *ast.Node) error {
	if !viewSite(node) || node.Kind == ast.KindParenthesizedExpression {
		return nil
	}
	if parent := node.Parent; parent != nil &&
		parent.Kind == ast.KindVariableDeclaration &&
		parent.Name().Kind == ast.KindArrayBindingPattern {
		return lowering.classDestructuredView(node, parent.Name())
	}
	target := lowering.slotType(node)
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
		if !spread && !lowering.nominalLiteralTarget(target) {
			return nil
		}
	}
	source := lowering.checker.GetTypeAtLocation(node)
	mismatch := lowering.nominalMismatch(source, target, map[[2]*checker.Type]bool{})
	if mismatch == nil {
		return nil
	}
	return lowering.nominalViewRefusal(node, mismatch)
}

// nominalViewRefusal is classViewRefusal's refusal of a value seen as target, a class it has no
// nominal ancestry in.
func (lowering *lowering) nominalViewRefusal(node *ast.Node, target *checker.Type) error {
	return &Refused{
		Where: lowering.program.Where(node),
		What:  "a value without nominal ancestry seen as " + lowering.checker.TypeToString(target),
		Fix:   "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)",
	}
}

func init() { refusalRules.register(3400, "classViewRefusal", (*lowering).classViewRefusal) }

func (lowering *lowering) nominalAncestor(source, target *checker.Type, seen map[[2]*checker.Type]bool) bool {
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := lowering.checker.GetBaseConstraintOfType(source); constraint != nil {
			source = constraint
		}
	}
	if weak := lowering.weakTarget(source); weak != nil {
		source = weak
	}
	if !isClassInstance(source) {
		return false
	}
	if source.Symbol() == target.Symbol() {
		from, to := lowering.checker.GetTypeArguments(source), lowering.checker.GetTypeArguments(target)
		if !lowering.sameClassArguments(from, to) {
			return false
		}
		for index := range from {
			if !lowering.enumAssignable(from[index], to[index]) ||
				!lowering.enumAssignable(to[index], from[index]) ||
				lowering.nominalMismatch(from[index], to[index], seen) != nil ||
				lowering.nominalMismatch(to[index], from[index], seen) != nil {
				return false
			}
		}
		return true
	}
	for _, base := range lowering.classBases(source) {
		if lowering.nominalAncestor(base, target, seen) {
			return true
		}
	}
	return false
}

// Initializers may read fields already initialized, including the base's after super. A future
// field would hold undefined despite its declared type; a method may read future derived fields.
func (lowering *lowering) initializerReads(node *ast.Node, available map[string]bool) error {
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
					symbol := lowering.checker.GetSymbolAtLocation(child.Name())
					if object.Kind == ast.KindThisKeyword && child.Name().Kind != ast.KindPrivateIdentifier {
						symbol = lowering.checker.GetPropertyOfType(lowering.classType, child.Name().Text())
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
							} else if (member.Kind == ast.KindPropertyDeclaration || parameterProperty(member)) &&
								!available[lowering.fieldName(child.Name())] {
								refused = &Refused{
									Where: lowering.program.Where(child),
									What: "a field initializer reading an uninitialized field through " +
										strings.Join(next, " -> "),
									Fix: "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set",
								}
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
				refused = &Refused{
					Where: lowering.program.Where(child),
					What:  "this in a field initializer before the fields it reads are initialized",
					Fix:   "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set",
				}
				return true
			}
			return child.ForEachChild(visit)
		}
		visit(body)
		return refused
	}
	return examine(node, nil)
}

// Private names with the same spelling in a base and a derived class are separate JavaScript slots,
// so each is qualified by the class that declares it (privateName), the same spelling memberKey gives
// its definition, for fields, methods, accessors and statics alike.
func (lowering *lowering) fieldName(name *ast.Node) string {
	if name.Kind != ast.KindPrivateIdentifier {
		return name.Text()
	}
	if declaring := lowering.declaringClass(name); declaring != nil {
		return lowering.privateName(name.Text(), declaring)
	}
	return name.Text()
}

// Assignment to a nominal class also requires ancestry; structural compatibility alone cannot
// provide the dynamic table an override's caller expects.
func (lowering *lowering) classAssignable(from, to *checker.Type) bool {
	return lowering.checker.IsTypeAssignableTo(from, to) &&
		lowering.nominalMismatch(from, to, map[[2]*checker.Type]bool{}) == nil
}

// Follow nominal class slots inside structural containers too. A plain object inside A[] lacks
// the same method table as a plain object assigned directly to A. Mutable slots need both views.
func (lowering *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type {
	if from == nil ||
		to == nil ||
		from == to ||
		from.Flags()&(checker.TypeFlagsNever|checker.TypeFlagsUndefined) != 0 ||
		seen[[2]*checker.Type{from, to}] {
		return nil
	}
	seen[[2]*checker.Type{from, to}] = true
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if found := lowering.nominalMismatch(member, to, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		var failure *checker.Type
		for _, member := range to.Types() {
			if !lowering.checker.IsTypeAssignableTo(from, member) {
				continue
			}
			if found := lowering.nominalMismatch(from, member, seen); found != nil {
				failure = found
			} else {
				return nil
			}
		}
		return failure
	}
	if weak := lowering.weakTarget(from); weak != nil {
		from = weak
	}
	if weak := lowering.weakTarget(to); weak != nil {
		to = weak
	}
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := lowering.checker.GetBaseConstraintOfType(from); constraint != nil {
			return lowering.nominalMismatch(constraint, to, seen)
		}
		return nil
	}
	if lowering.isStaticType(to) {
		wanted := lowering.staticClass(to, nil)
		actual := lowering.staticClass(from, nil)
		for actual != nil {
			if actual == wanted {
				return nil
			}
			actual = lowering.staticBase(actual)
		}
		return to
	}
	if isClassInstance(to) {
		if symbol := to.Symbol(); symbol != nil &&
			len(symbol.Declarations) > 0 &&
			load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			return nil
		}
		if !lowering.nominalAncestor(from, to, seen) {
			return to
		}
		return nil
	}
	if !lowering.structured(from) || !lowering.structured(to) {
		return nil
	}
	fromSignatures := lowering.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := lowering.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		parameters := toSignatures[0].Parameters()
		if lowering.censusNeverRestSignature(toSignatures[0]) {
			parameters = nil
		}
		for index, parameter := range parameters {
			if index >= len(fromSignatures[0].Parameters()) {
				break
			}
			if found := lowering.nominalMismatch(
				lowering.checker.GetTypeOfSymbol(parameter),
				lowering.checker.GetTypeOfSymbol(fromSignatures[0].Parameters()[index]),
				seen,
			); found != nil {
				return found
			}
		}
		return lowering.nominalMismatch(
			lowering.checker.GetReturnTypeOfSignature(fromSignatures[0]),
			lowering.checker.GetReturnTypeOfSignature(toSignatures[0]),
			seen,
		)
	}
	for _, target := range lowering.containers(to) {
		for _, source := range lowering.containers(from) {
			if !lowering.sameContainer(source, target) {
				continue
			}
			fromArguments, toArguments := lowering.checker.GetTypeArguments(
				source,
			), lowering.checker.GetTypeArguments(target)
			mutable := !lowering.isLibraryType(target, "ReadonlyArray", "ReadonlyMap", "ReadonlySet") &&
				!(checker.IsTupleType(target) && target.TargetTupleType().IsReadonly())
			for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
				if found := lowering.nominalMismatch(fromArguments[index], toArguments[index], seen); found != nil {
					return found
				}
				if mutable {
					if found := lowering.nominalMismatch(toArguments[index], fromArguments[index], seen); found != nil {
						return found
					}
				}
			}
		}
	}
	if len(lowering.containers(to)) > 0 {
		return nil
	}
	for _, property := range lowering.checker.GetPropertiesOfType(to) {
		if property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		inside := lowering.checker.GetPropertyOfType(from, property.Name)
		if inside == nil {
			continue
		}
		source, target := lowering.checker.GetTypeOfSymbol(inside), lowering.checker.GetTypeOfSymbol(property)
		if found := lowering.nominalMismatch(source, target, seen); found != nil {
			return found
		}
		if !lowering.checker.IsReadonlySymbol(property) {
			if found := lowering.nominalMismatch(target, source, seen); found != nil {
				return found
			}
		}
	}
	return nil
}

// An inferred destructuring context can be the union of the tuple's elements. Compare the
// positional values with their bound names instead of treating that union as a writable array.
//
// A value typed as a union is judged member by member, and refused when any member is. Only a
// reference has type arguments: the checker dereferences nil asking anything else for them (#egxccg8).
// A string's elements are strings, which no class view can be, so a primitive member has nothing to
// judge. Any other member that isn't a reference (an intersection, an iterable interface, a type
// parameter) has elements this can't read, so it's NotYet, never passed unjudged.
func (lowering *lowering) classDestructuredView(node, pattern *ast.Node) error {
	own := lowering.checker.GetTypeAtLocation(node)
	members := []*checker.Type{own}
	if own.Flags()&checker.TypeFlagsUnion != 0 {
		members = own.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		if member.Flags()&checker.TypeFlagsObject == 0 || member.ObjectFlags()&checker.ObjectFlagsReference == 0 {
			return lowering.notYet(pattern, "a destructured "+
				lowering.checker.TypeToString(member)+
				", whose elements can't be judged for nominal ancestry")
		}
		elements := lowering.checker.GetTypeArguments(member)
		tuple := checker.IsTupleType(member)
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
			target := lowering.checker.GetTypeAtLocation(binding.Name())
			if mismatch := lowering.nominalMismatch(
				elements[index],
				target,
				map[[2]*checker.Type]bool{},
			); mismatch != nil {
				return &Refused{
					Where: lowering.program.Where(binding),
					What: "a destructured value without nominal ancestry seen as " +
						lowering.checker.TypeToString(mismatch),
					Fix: "construct that class or a subclass; use an interface for structural values (adamic/nominal-class)",
				}
			}
		}
	}
	return nil
}

// The checker mangles private symbol names; the IR qualifies them by class ID. Cycle proofs
// must match the same declaration in both names, including inherited and generic private slots.
func (lowering *lowering) cycleFieldMatches(holder *checker.Type, field, written string) bool {
	if field == written {
		return true
	}
	property := lowering.checker.GetPropertyOfType(holder, field)
	if property == nil || len(property.Declarations) != 1 {
		return false
	}
	name := property.Declarations[0].Name()
	if name == nil || name.Kind != ast.KindPrivateIdentifier {
		return false
	}
	if class := property.Declarations[0].Parent; class != nil && class.Kind == ast.KindClassDeclaration {
		return written == lowering.privateName(name.Text(), class)
	}
	return written == name.Text()
}

func (lowering *lowering) cycleFieldName(field *ast.Symbol) string {
	if len(field.Declarations) == 1 {
		if name := field.Declarations[0].Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			return name.Text()
		}
	}
	return field.Name
}

// All monomorphizations of a source class have one erased identity. An identity-only
// descriptor lets instanceof name a generic class before any concrete instance is made.
func (lowering *lowering) classDefinition(declaration *ast.Node) int {
	for key, instance := range lowering.instances {
		if lowering.classKeyMatches(key, declaration) {
			return lowering.result.Classes[instance.class-1].Definition
		}
	}
	return len(lowering.result.Classes) + 1
}

func (lowering *lowering) classIdentity(declaration *ast.Node) int {
	definition := lowering.classDefinition(declaration)
	for index, class := range lowering.result.Classes {
		if class.Definition == definition {
			return index + 1
		}
	}
	identity := len(lowering.result.Classes) + 1
	lowering.result.Classes = append(lowering.result.Classes, ir.Class{
		Source:      lowering.functionSource(declaration),
		Name:        declaration.Name().Text() + "_identity",
		Definition:  definition,
		Constructor: -1,
	})
	if lowering.instances == nil {
		lowering.instances = map[string]*instance{}
	}
	key := lowering.classKeyPrefix(declaration) + ",identity"
	lowering.instances[key] = &instance{class: identity, constructor: -1, initializer: -1}
	return identity
}

// An inherited method's receiver can be a descendant with different type arguments.
// Find the declaration's actual instantiated view instead of substituting by position.
func (lowering *lowering) classView(proven *checker.Type, declaration *ast.Node) *checker.Type {
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := lowering.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	if weak := lowering.weakTarget(proven); weak != nil {
		proven = weak
	}
	if proven.Symbol() == lowering.symbol(declaration.Name()) {
		return proven
	}
	for _, base := range lowering.classBases(proven) {
		if found := lowering.classView(base, declaration); found != nil {
			return found
		}
	}
	return nil
}

// The checker exposes generic bases on the declaration. Substitute the receiver's
// actual arguments before choosing a native base layout or proving nominal ancestry.
func (lowering *lowering) classBases(proven *checker.Type) []*checker.Type {
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	bases := lowering.checker.GetBaseTypes(target)
	declaration := lowering.classNodeFor(proven)
	if declaration == nil || declaration.Kind != ast.KindClassDeclaration {
		return bases
	}
	mapper := lowering.typeMapperOf(declaration, proven)
	if mapper == nil {
		return bases
	}
	concrete := make([]*checker.Type, len(bases))
	for index, base := range bases {
		concrete[index] = instantiateType(lowering.checker, base, mapper)
	}
	return concrete
}

func literalParts(node *ast.Node) []*ast.Node {
	if node.Kind == ast.KindObjectLiteralExpression {
		return node.AsObjectLiteralExpression().Properties.Nodes
	}
	return node.AsArrayLiteralExpression().Elements.Nodes
}

func (lowering *lowering) nominalLiteralTarget(proven *checker.Type) bool {
	if weak := lowering.weakTarget(proven); weak != nil {
		proven = weak
	}
	if isClassInstance(proven) {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if lowering.nominalLiteralTarget(member) {
				return true
			}
		}
	}
	return false
}
