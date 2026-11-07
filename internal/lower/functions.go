// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// functionBody lowers a module function declaration's parameters and body.
func (lowering *lowering) functionBody(declaration *ast.Node) error {
	return lowering.lowerFunction(lowering.functions[lowering.symbol(declaration.Name())], declaration, -1)
}

// defaulted is a parameter with a default: the local the body declares, the one the argument arrives
// in, and the default.
type defaulted struct {
	local, incoming int
	initializer     *ast.Node
}

// signed is what a function whose signature is written still needs to lower its body: the local its
// this is, or -1, its parameters' defaults, and its destructured parameters.
type signed struct {
	this     int
	defaults []defaulted
	patterns []patterned
}

// patterned is a destructured parameter, ([key, value]) or ({ x, y }): the pattern, the parameter,
// and the local it arrives in whole.
type patterned struct {
	pattern   *ast.Node
	parameter *ast.Node
	incoming  int
}

// lowerFunction lowers a function-like declaration into the function at index: its signature, unless
// signature has written it already, then its body. this, when not -1, is the local a method's or
// constructor's this is, passed first.
//
// It works on a copy and writes it back by index: lowering a body can instantiate a class, which
// appends functions, and a pointer into the slice would be left pointing at the old one.
func (lowering *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {
	if implementation := lowering.censusImplementation(declaration); implementation != nil {
		declaration = implementation
	}
	if _, isSigned := lowering.signed[index]; !isSigned {
		if signatureError := lowering.signature(index, declaration, this); signatureError != nil {
			return signatureError
		}
	}
	pending := lowering.signed[index]
	delete(lowering.signed, index)
	return lowering.lowerBody(index, declaration, pending.this, pending.defaults, pending.patterns)
}

// signature writes the function at index's parameters and result, from the checker, without lowering
// its body, so a call to it lowers whether or not its body has been. this is as for lowerFunction.
func (lowering *lowering) signature(index int, declaration *ast.Node, this int) error {
	if implementation := lowering.censusImplementation(declaration); implementation != nil {
		declaration = implementation
	}
	if censusOverloadsError := lowering.censusOverloads(declaration); censusOverloadsError != nil {
		return censusOverloadsError
	}
	function := lowering.result.Functions[index]
	function.SourceLength = sourceFunctionLength(declaration)
	name := declaration.Name()
	if name == nil {
		name = declaration
	}
	source := ast.GetSourceFileOfNode(name)
	line, column := scanner.GetLineAndCharacterOfPosition(source, scanner.GetTokenPosOfNode(name, source, false))
	function.Position = ir.SourcePosition{File: lowering.program.FileName(source), Line: line + 1, Column: column + 1}
	if this >= 0 && declaration.Kind != ast.KindConstructor {
		// A method receives this; a constructor makes it.
		function.Parameters = append(function.Parameters, this)
		function.Boundary.Parameters = append(function.Boundary.Parameters, ir.BoundaryParameter{})
		function.Receiver = true
	}
	if implementation := generatorImplementation(declaration); implementation != nil {
		// A class writes each overload's signature into its implementation's function before the
		// implementation's own (class.go), so the gate reads the implementation from the first one,
		// before an overload could write it as an ordinary function's.
		return lowering.generatorNotYet(implementation, generatorForm(implementation))
	}
	if generatorDeclaration(declaration) {
		// Every function body is lowered after its signature is written, here, so this is where a
		// generator of a form this slice doesn't lower stops, before its body could be lowered as an
		// ordinary function's (a function* that only returns would return undefined, not a generator).
		if !loweredGenerator(declaration) {
			return lowering.generatorNotYet(declaration, generatorForm(declaration))
		}
		if generatorSignatureError := lowering.generatorSignature(
			&function,
			declaration,
		); generatorSignatureError != nil {
			return generatorSignatureError
		}
	} else if declaration.Kind != ast.KindConstructor {
		signature := lowering.checker.GetSignatureFromDeclaration(declaration)
		returns := lowering.concrete(lowering.checker.GetReturnTypeOfSignature(signature))
		function.Boundary.Return = lowering.boundarySchema(declaration, returns)
		// A function that never returns (it panics on every path, as (why) => panic(why) does) has no
		// result to hold, as one returning void hasn't. An arrow whose expression is never for another
		// reason, a variable the checker narrowed to nothing, isn't one.
		neverArrow := returns.Flags()&checker.TypeFlagsNever != 0 &&
			declaration.Body() != nil &&
			declaration.Body().Kind != ast.KindBlock &&
			declaration.Body().Kind != ast.KindCallExpression &&
			!lowering.isPanicCall(declaration.Body())
		if returns.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 || neverArrow {
			valueType, isKnown := lowering.representation(returns)
			if returns.Flags()&checker.TypeFlagsUndefined != 0 {
				valueType, isKnown = ir.Object, true
			}
			if function.Closure && lowering.regexReplacementArgument(declaration) {
				if returns.Flags()&checker.TypeFlagsUndefined != 0 {
					valueType, isKnown = ir.String, true
				}
				if returns.Flags()&checker.TypeFlagsNull != 0 {
					valueType, isKnown = ir.Object, true
				}
			}
			if !isKnown {
				valueType, isKnown = lowering.clockGenericReturnsT01(returns)
			}
			if !isKnown {
				// An arrow function has no name to point at, so it's pointed at whole.
				where := declaration.Name()
				if where == nil {
					where = declaration
				}
				return lowering.notYet(where, "a function returning "+lowering.checker.TypeToString(returns))
			}
			function.Returns = valueType
		}
	}
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) {
		returns := lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(declaration))
		if !lowering.isLibraryType(returns, "Promise") {
			return lowering.notYet(declaration, "an async result other than the library Promise")
		}
		arguments := lowering.checker.GetTypeArguments(returns)
		if len(arguments) != 1 {
			return lowering.notYet(declaration, "full Promise compatibility")
		}
		function.Async = true
		function.AsyncSite = lowering.asyncSite(declaration)
		function.Returns = ir.Promise
		if arguments[0].Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) == 0 {
			var known bool
			function.AsyncReturns, known = lowering.representation(arguments[0])
			if !known {
				return lowering.notYet(declaration, "an async payload of "+lowering.checker.TypeToString(arguments[0]))
			}
		}
	}
	outerIndexForParameters := lowering.functionIndex
	lowering.functionIndex = index
	defer func() { lowering.functionIndex = outerIndexForParameters }()
	// A parameter with a default arrives as what may be missing, and the body begins by declaring the
	// parameter itself: the argument, or the default when it's undefined, as JavaScript decides.
	defaults := []defaulted{}
	// A destructured parameter, ([key, value]) or ({ x, y }), arrives whole in a parameter of its own,
	// and its names are declared from it before the body runs.
	patterns := []patterned{}
	for position, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		function.Boundary.Parameters = append(
			function.Boundary.Parameters,
			ir.BoundaryParameter{
				Schema: lowering.boundarySchema(parameter, lowering.checker.GetTypeAtLocation(parameter)),
			},
		)
		if name := parameter.Name(); (name.Kind == ast.KindArrayBindingPattern ||
			name.Kind == ast.KindObjectBindingPattern) &&
			declared.DotDotDotToken == nil &&
			declared.Initializer == nil &&
			declared.QuestionToken == nil {
			incoming := len(lowering.result.Locals)
			heldAs := ir.Object
			if represented, _ := lowering.representation(
				lowering.checker.GetTypeAtLocation(parameter),
			); represented == ir.String {
				heldAs = ir.String
			}
			lowering.result.Locals = append(
				lowering.result.Locals,
				ir.Local{Name: "destructured", Type: heldAs, Function: index},
			)
			function.Parameters = append(function.Parameters, incoming)
			patterns = append(patterns, patterned{pattern: name, parameter: parameter, incoming: incoming})
			continue
		}
		if !ast.IsIdentifier(parameter.Name()) {
			return lowering.notYet(parameter, "a parameter that isn't a plain name")
		}
		// A this parameter names the receiver's type and takes no argument: f(this: void, n) is called
		// as f(1). Declared here as a local it shifts every argument after it by one, so a function
		// declaration printed NaN where Node prints 2, and a method, whose receiver is already its first
		// parameter, shifts the same way (#j8mt4pr). Every declaration's parameters pass through this
		// loop, so it stops here when another parameter follows, until this is dropped from the lowered
		// parameters, the real fix. With nothing after it nothing shifts, and the body's this is the
		// receiver, so an object literal's value(this: { flags: number }) lowers and agrees with Node
		// (namespace_method_receiver.a). A function expression, a nested declaration, a namespace
		// function and a function used as a value stop sooner, in their own words.
		if parameter.Name().Text() == "this" && position < len(declaration.Parameters())-1 {
			return lowering.notYet(
				parameter,
				"a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented)",
			)
		}
		local, declareLocalError := lowering.declareLocal(parameter.Name())
		if declareLocalError != nil {
			return declareLocalError
		}
		if parameterProperty(parameter) {
			// The binder gives a parameter property both a field symbol and a lexical parameter
			// symbol. Reads and assignments in the body refer to the latter.
			parameters := lowering.checker.GetSignatureFromDeclaration(declaration).Parameters()
			lowering.locals[parameters[position]] = local
		}
		if declared.DotDotDotToken != nil {
			proven := lowering.concrete(lowering.checker.GetTypeAtLocation(parameter.Name()))
			// Only an array reference has type arguments to read. The checker dereferences nil for
			// anything else, such as the union of tuples tsc's createConstructorTypeNode takes as
			// Parameters<typeof createConstructorTypeNode1 | typeof createConstructorTypeNode2> (#egxccg8).
			if !lowering.isArrayType(proven) {
				return lowering.notYet(parameter, "a rest parameter that isn't an array")
			}
			elements := lowering.checker.GetTypeArguments(proven)
			if len(elements) != 1 {
				return lowering.notYet(parameter, "a rest parameter that isn't an array")
			}
			element, known := lowering.representation(elements[0])
			if !known || slotless(element) && element != ir.Union {
				return lowering.notYet(parameter, "a rest parameter whose elements cannot be packed")
			}
			function.RestElement = element
		}
		if function.Closure &&
			censusCallableSlotless(lowering.result.Locals[local].Type) &&
			!(lowering.result.Locals[local].Type == ir.Union && lowering.regexReplacementArgument(declaration)) {
			// Its arguments are each one adamic_value.
			return lowering.notYet(
				parameter,
				"a function value taking "+
					lowering.checker.TypeToString(lowering.checker.GetTypeAtLocation(parameter.Name())),
			)
		}
		if declared.Initializer == nil {
			function.Parameters = append(function.Parameters, local)
			continue
		}
		missing := ir.Maybe(lowering.result.Locals[local].Type)
		if !missing.IsMaybe() && !missing.IsReference() {
			return lowering.notYet(parameter, "a default for a "+typeName(missing)+" parameter")
		}
		incoming := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: parameter.Name().Text(), Type: missing, Function: index},
		)
		function.Parameters = append(function.Parameters, incoming)
		defaults = append(defaults, defaulted{local: local, incoming: incoming, initializer: declared.Initializer})
	}
	if function.Closure &&
		censusCallableSlotless(function.Returns) &&
		!(function.Returns == ir.Union && lowering.regexReplacementArgument(declaration)) {
		// A function value's arguments and result must each fit one adamic_value.
		return lowering.notYet(declaration, "a function value returning "+typeName(function.Returns))
	}
	if declaration.Body() == nil && !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract) {
		return lowering.notYet(declaration, "a function without a body")
	}
	function.OptionalParameters = map[int]bool{}
	offset := 0
	if function.Receiver {
		offset = 1
	}
	for i, node := range declaration.Parameters() {
		parameter := node.AsParameterDeclaration()
		if parameter.QuestionToken != nil || parameter.Initializer != nil {
			function.OptionalParameters[function.Parameters[i+offset]] = true
		}
	}
	if function.Source.Module == "" {
		function.Source = lowering.functionSource(declaration)
	}
	lowering.result.Functions[index] = function
	if !function.Closure && declaration.Kind != ast.KindConstructor && declaration.Name() != nil {
		lowering.closureRecords = append(lowering.closureRecords, closureRecord{
			proven:   lowering.concrete(lowering.checker.GetTypeAtLocation(declaration.Name())),
			function: index,
			node:     declaration,
		})
	}
	// A declared constructor also supplies the public allocator's signature.
	// The initializer's receiver is internal and the class result is not data.
	if declaration.Kind == ast.KindConstructor && lowering.instance != nil && lowering.instance.constructor >= 0 {
		allocator := &lowering.result.Functions[lowering.instance.constructor]
		allocator.Position = function.Position
		allocator.Boundary.Parameters = append([]ir.BoundaryParameter(nil), function.Boundary.Parameters...)
	}
	if lowering.signed == nil {
		lowering.signed = map[int]signed{}
	}
	lowering.signed[index] = signed{this: this, defaults: defaults, patterns: patterns}
	return nil
}

// lowerBody lowers the body of the function at index, whose signature is written.
func (lowering *lowering) lowerBody(
	index int,
	declaration *ast.Node,
	this int,
	defaults []defaulted,
	patterns []patterned,
) error {
	if declaration.Body() == nil {
		return lowering.notYet(declaration, "a function without a body")
	}
	if lowering.functionNodes == nil {
		lowering.functionNodes = map[int]*ast.Node{}
	}
	lowering.functionNodes[index] = declaration
	// Every call in a body is judged when the body is lowered, under its own mapper (predicates_branded.go).
	if judgeBrandedCallsError := lowering.judgeBrandedCalls(declaration); judgeBrandedCallsError != nil {
		return judgeBrandedCallsError
	}
	function := lowering.result.Functions[index]
	if !function.Closure {
		// A function declaration, a method or a constructor has its body lowered where a use of it is first met,
		// a closure's body included, but it's declared at the top level and captures nothing from
		// the closures being lowered there: a local of its own read in a closure of its own must not
		// land in their environments.
		outerClosures := lowering.closures
		lowering.closures = nil
		defer func() { lowering.closures = outerClosures }()
	}
	body := declaration.Body()
	outer, outerThis, outerIndex := lowering.function, lowering.this, lowering.functionIndex
	lowering.function, lowering.functionIndex = &function, index
	if this >= 0 {
		lowering.this = this
	}
	// Defaults run before the body, in order, after a constructor's fields, as JavaScript runs them.
	var prologue, lowered []ir.Statement
	var err error
	// Signature allocates incoming locals in source parameter order. Merge the two
	// prologues by that order: a default can read an earlier destructured name,
	// and a later destructuring must observe any mutation made by a default.
	for patternIndex, defaultIndex := 0, 0; patternIndex < len(patterns) || defaultIndex < len(defaults); {
		if err != nil {
			break
		}
		if patternIndex < len(patterns) &&
			(defaultIndex == len(defaults) || patterns[patternIndex].incoming < defaults[defaultIndex].incoming) {
			parameter := patterns[patternIndex]
			patternIndex++
			patternType := lowering.checker.GetTypeAtLocation(parameter.parameter)
			heldAs, _ := lowering.representation(patternType)
			var destructured []ir.Statement
			destructured, err = lowering.destructureFrom(parameter.pattern, patternType, heldAs, parameter.incoming)
			prologue = append(prologue, destructured...)
			continue
		}
		parameter := defaults[defaultIndex]
		defaultIndex++
		if declaration.Kind == ast.KindConstructor {
			if err = lowering.parameterPropertyDefault(declaration, parameter.initializer); err != nil {
				break
			}
		}
		if lowering.uninitializedInitializer(parameter.initializer) {
			lowering.result.Locals[parameter.local].Uninitialized = true
			incoming := ir.Read{Local: parameter.incoming, Of: lowering.result.Locals[parameter.incoming].Type}
			prologue = append(prologue, ir.Declare{
				Local: parameter.local,
				Value: lowering.placeholderInitialValue(
					parameter.initializer,
					lowering.result.Locals[parameter.local].Type,
				),
				Uninitialized: true,
			}, ir.If{
				Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: incoming}},
				Then: []ir.Statement{
					ir.Assign{
						Local: parameter.local,
						Value: fit(incoming, lowering.result.Locals[parameter.local].Type),
					},
				},
			})
			continue
		}

		var fallback ir.Expression
		if fallback, err = lowering.expression(parameter.initializer); err != nil {
			break
		}
		of := lowering.result.Locals[parameter.local].Type
		if fallback = fit(fallback, of); fallback.Type() != of {
			err = lowering.notYet(parameter.initializer, "a default of another type than its parameter")
			break
		}
		incoming := ir.Read{Local: parameter.incoming, Of: lowering.result.Locals[parameter.incoming].Type}
		prologue = append(
			prologue,
			ir.Declare{Local: parameter.local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: of}},
		)
	}
	if err != nil {
		lowering.function, lowering.this, lowering.functionIndex = outer, outerThis, outerIndex
		return err
	}
	if declaration.Kind == ast.KindConstructor && lowering.instance.base == nil {
		assigned, assignError := lowering.parameterPropertyStores(declaration.Parent, this)
		if assignError != nil {
			lowering.function, lowering.this, lowering.functionIndex = outer, outerThis, outerIndex
			return assignError
		}
		prologue = append(prologue, assigned...)
	}
	if body.Kind == ast.KindBlock {
		lowered, err = lowering.statements(body.AsBlock().Statements.Nodes)
	} else if function.BodyReturns() == 0 || lowering.isPanicCall(body) {
		// An arrow function's expression body, when it returns nothing or never returns (() =>
		// panic('why')), is a statement.
		lowered, err = lowering.expressionStatement(body)
	} else {
		// Otherwise its value is what it returns.
		var value ir.Expression
		if value, err = lowering.expression(body); err == nil {
			if err = lowering.checkAsyncReturn(body, value); err == nil {
				lowered, err = lowering.checkedPredicateReturn(declaration, body, fit(value, function.BodyReturns()))
			}
		}
	}
	lowering.function, lowering.this, lowering.functionIndex = outer, outerThis, outerIndex
	if err != nil {
		return err
	}
	if body.Kind == ast.KindBlock &&
		declaration.Kind != ast.KindConstructor &&
		declaration.Flags&ast.NodeFlagsHasImplicitReturn != 0 &&
		lowering.permitsImplicitReturn(declaration) {
		// The checker permits an end that is reachable only when the result can be undefined
		// (or void). Put that exit in the IR so every backend and ownership pass sees it.
		returned := ir.Return{}
		if function.BodyReturns() != 0 {
			returned.Value = fit(ir.Undefined{}, function.BodyReturns())
		}
		lowered = append(lowered, returned)
	}
	// The environment may have grown while the body was lowered (captures are found as they're
	// read), so it's taken from what's recorded, not from this copy.
	function.Environment = lowering.result.Functions[index].Environment
	function.ForwardedNestedParent = lowering.result.Functions[index].ForwardedNestedParent
	function.ReferenceParents = lowering.result.Functions[index].ReferenceParents
	function.Body = append(function.Body, prologue...)
	function.Body = append(function.Body, lowered...)
	lowering.finishNestedEnvironment(&function, index)
	lowering.result.Functions[index] = function
	return lowering.checkGeneratorBody(index, declaration)
}

// Descriptor construction must not intern boundary-only literals into codegen's
// string table. The same builder runs against a private table and reports no errors.
func (lowering *lowering) boundarySchema(node *ast.Node, proven *checker.Type) *ir.JSONDecodeSchema {
	schema, jsonDecodeSchemaError := lowering.jsonDecodeSchema(
		node,
		proven,
		jsonSchemaBoundary,
		lowering.result.BoundaryStrings.Intern,
	)
	if jsonDecodeSchemaError != nil {
		return nil
	}
	return &schema
}

// The binder marks a syntactic end even for a switch the checker proved exhaustive. Only the
// checker's result type can authorize returning undefined at that end.
func (lowering *lowering) permitsImplicitReturn(declaration *ast.Node) bool {
	result := lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(declaration))
	if lowering.isLibraryType(result, "Promise") {
		arguments := lowering.checker.GetTypeArguments(result)
		if len(arguments) == 1 {
			result = arguments[0]
		}
	}
	if result.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range result.Types() {
			if member.Flags() == checker.TypeFlagsVoid {
				return true
			}
		}
	}
	return result.Flags()&checker.TypeFlagsVoid != 0 || lowering.includesUndefined(result)
}
