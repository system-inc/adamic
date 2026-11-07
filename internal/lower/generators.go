package lower

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Generators, first slice (#qk8rztp S1c): a top-level function* declaration with plain parameters,
// whose body yields only as a statement, consumed by a for-of directly over a call to it or dropped
// by a call as a statement. The body runs on one counted frame (runtime/generator.h): flow's
// NormalizeAsync puts every local in a frame cell, as it does an async function's, and each yield is
// a suspension. No slot but the consuming loop's own holds the generator, so nothing else can see or
// keep one: every other road a generator object could take stops here as not yet built, and
// checkGeneratorRoads holds the lowered program to that.
//
// Since S3a, a function* declared directly in an ordinary function's body lowers too, as a nested
// function: hoisted, its call made through the closure of the function that declared it, which its
// frame keeps (runtime/generator.h's self), so its body reads its parent's cells, live, whenever it
// is stepped. Its consumers are a for-of over its call, the call as a statement, and a one-read
// Iterable<T> parameter's argument; a write inside it to a local it captured stops as not yet built.
//
// Since S3b, so does an anonymous function* written in place and called at once, as tsc's
// elaborateJsxComponents writes one (checker.ts:21781): a closure of its own, lowered once, the first
// time a consumer asks for it (inPlaceGenerator), whose call makes its frame through a closure made
// for the call, its carrier, which the frame keeps as a nested generator's does.

// generatorNotYet is a generator form this slice doesn't lower.
func (lowering *lowering) generatorNotYet(node *ast.Node, what string) error {
	return lowering.notYet(node, "a generator function "+what+" (#qk8rztp)")
}

// generatorTitle is where a stop about a generator's whole declaration points, and what it calls the
// generator: its name and "a generator function", or, for one written in place with no name, the
// expression itself and "a generator written in place" (Lower, on #qk8rztp S3b). Every such stop
// goes through it (generatorDeclarationNotYet), so an anonymous one reads not yet built, never a
// panic on a missing name.
func generatorTitle(declaration *ast.Node) (*ast.Node, string) {
	if name := declaration.Name(); name != nil {
		return name, "a generator function"
	}
	return declaration, "a generator written in place"
}

// generatorDeclarationNotYet is generatorNotYet for a stop about a generator's whole declaration,
// pointed at and named by generatorTitle.
func (lowering *lowering) generatorDeclarationNotYet(declaration *ast.Node, what string) error {
	where, title := generatorTitle(declaration)
	return lowering.notYet(where, title+" "+what+" (#qk8rztp)")
}

// inPlaceNotYet stops a generator written in place and called whose call is used any other way: a
// value no slot can hold yet, as a named generator's call used so is.
func (lowering *lowering) inPlaceNotYet(node *ast.Node) error {
	return lowering.notYet(
		node,
		"a generator written in place and called, used other than by a for-of over its call, a call as a statement or a one-read Iterable parameter (#qk8rztp)",
	)
}

// generatorDeclaration reports whether a function-like node is a synchronous generator.
func generatorDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// loweredGenerator is a generator declaration this slice lowers: a named function* declared at a
// module's top level, one nestedGenerator lowers, or one written in place (generatorInPlace). A
// generic one at the top level is lowered once per instantiation, where it's called, as every
// generic function is (generic.go), and its yielded type is judged at that instance, so one nothing
// instantiates costs nothing. Every other generator stops at lowering, wherever lowering reaches its
// declaration, in the words generatorForm gives it (Lower, open question 7 on #qk8rztp); only an
// async generator is stopped before lowering, refused for good by rule 2400. It reads syntax alone,
// so the refusal walk's void rules (generatorStepped) and lowering agree on which calls are a lowered
// generator's.
func loweredGenerator(node *ast.Node) bool {
	topLevel := node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().AsteriskToken != nil &&
		node.Parent != nil && node.Parent.Kind == ast.KindSourceFile && node.Name() != nil &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
	return topLevel || nestedGenerator(node) || generatorInPlace(node)
}

// generatorInPlace is a generator expression S3b lowers (#qk8rztp S3b): anonymous, not async, with no
// type parameter of its own or in scope, and called where it's written (calledInPlace). A name of its
// own would bind it inside its body, where a call through it makes a second frame through a carrier
// no rule here covers, so a named one stays not yet built (Lower). That the function it's written in
// doesn't suspend is a fact of lowering, so inPlaceGenerator says it there.
func generatorInPlace(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionExpression &&
		node.AsFunctionExpression().AsteriskToken != nil &&
		node.Name() == nil &&
		node.Body() != nil &&
		len(node.TypeParameters()) == 0 &&
		!genericScope(node.Parent) &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) &&
		calledInPlace(node) != nil
}

// calledInPlace is the call a function expression is the callee of, any parentheses around it
// skipped (Adamic's word on #qk8rztp S3b, as Lower read it), or nil. That call is then the
// expression's one use, so nothing else can step the frame it makes. A comma, a conditional's arm, a
// variable in between, .call, .apply or .bind, an optional call, a tag and new are each no such call.
func calledInPlace(node *ast.Node) *ast.Node {
	// The callee as the call writes it, its parentheses included.
	callee := node
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	if callee.Parent == nil || callee.Parent.Kind != ast.KindCallExpression {
		return nil
	}
	call := callee.Parent.AsCallExpression()
	if call.Expression != callee || call.QuestionDotToken != nil {
		return nil
	}
	return callee.Parent
}

// nestedGenerator is a generator declaration S3a lowers as a nested function (#qk8rztp S3a): named,
// not async, with no type parameters of its own and none in scope, declared directly in a function's
// body, where nestedDeclarations binds it with its siblings. No type parameter in scope means it is
// lowered once, with its parent, never per instance. That the parent is neither async nor a
// generator is a fact about the parent's signature, so nestedDeclarations says it there.
func nestedGenerator(node *ast.Node) bool {
	return nestedGeneratorShape(node) && len(node.TypeParameters()) == 0 && !genericScope(node.Parent) &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
}

// nestedGeneratorShape is a named function* declared directly in a function's body, the place
// nestedDeclarations binds a nested declaration, generic or not.
func nestedGeneratorShape(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration &&
		node.AsFunctionDeclaration().AsteriskToken != nil &&
		node.Name() != nil &&
		node.Body() != nil &&
		node.Parent != nil &&
		node.Parent.Kind == ast.KindBlock &&
		ast.IsFunctionLike(node.Parent.Parent)
}

// genericScope reports whether a type parameter is in scope at node: some function or class around
// it declares one. Only those kinds are asked, since the checker's TypeParameters panics on a node
// that can't have any.
func genericScope(node *ast.Node) bool {
	for scope := node; scope != nil; scope = scope.Parent {
		if (ast.IsFunctionLike(scope) ||
			scope.Kind == ast.KindClassDeclaration ||
			scope.Kind == ast.KindClassExpression) && len(scope.TypeParameters()) != 0 {
			return true
		}
	}
	return false
}

// generatorImplementation is the generator a method's overload signature signs for, when its
// implementation is one. tsc refuses `*` on an overload signature (TS1222), so only the
// implementation says it is a generator.
func generatorImplementation(declaration *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindMethodDeclaration || declaration.Body() != nil || declaration.Symbol() == nil {
		return nil
	}
	for _, candidate := range declaration.Symbol().Declarations {
		if candidate.Kind == ast.KindMethodDeclaration && candidate.Body() != nil && generatorDeclaration(candidate) {
			return candidate
		}
	}
	return nil
}

// generatorForm names the form of a generator declaration this slice doesn't lower, for its stop.
func generatorForm(node *ast.Node) string {
	switch {
	case node.Kind == ast.KindFunctionExpression && calledInPlace(node) != nil && len(node.TypeParameters()) != 0:
		// One written in place is lowered once, with the function it's written in, never per
		// instance (#qk8rztp S3b).
		return "written in place with type parameters of its own"
	case node.Kind == ast.KindFunctionExpression && calledInPlace(node) != nil && genericScope(node.Parent):
		return "written in place inside a generic function or class"
	case node.Kind == ast.KindFunctionExpression:
		return "written as an expression"
	case node.Kind == ast.KindMethodDeclaration &&
		node.Parent != nil &&
		node.Parent.Kind == ast.KindObjectLiteralExpression:
		return "written as an object literal's method"
	case node.Kind == ast.KindMethodDeclaration && ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic):
		return "written as a static method"
	case node.Kind == ast.KindMethodDeclaration:
		return "written as a method"
	case node.Name() == nil:
		return "declared without a name"
	case node.Parent != nil && node.Parent.Kind == ast.KindModuleBlock:
		return "declared inside a namespace"
	case nestedGeneratorShape(node) && len(node.TypeParameters()) != 0:
		// A nested generic function is put by to instantiate where it's called, which S3a doesn't
		// build for a generator (#qk8rztp S3a).
		return "declared generic inside a function"
	case nestedGeneratorShape(node) && genericScope(node.Parent):
		return "declared inside a generic function or class"
	case node.Parent != nil && node.Parent.Kind != ast.KindSourceFile:
		return "declared inside a function or a block"
	}
	return "of this form"
}

// generatorSignature writes a generator's result: a Generator, which yields one slot per value.
// Its parameters are written by signature, as any function's are, and must be plain names, since a
// default, a pattern or a rest would run or bind at the call, which isn't built yet.
func (lowering *lowering) generatorSignature(function *ir.Function, declaration *ast.Node) error {
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) ||
			declared.Initializer != nil ||
			declared.QuestionToken != nil ||
			declared.DotDotDotToken != nil ||
			parameter.Name().Text() == "this" {
			return lowering.generatorNotYet(parameter, "with a parameter that isn't a plain required name")
		}
	}
	// Under a generic instance's type mapper, concrete reads the result this instantiation made of it,
	// so the yielded type below is judged per instance: an unread type parameter stays itself, and
	// has no representation.
	returns := lowering.concrete(
		lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(declaration)),
	)
	// The checker's own Generator, inferred or written, or an iterator type a body may be declared to
	// return. Every consumer routes a generator's call by its origin, never by this declared type
	// (Compiler's risk 2), so a generator declared MapIterator or SetIterator is stepped as a generator,
	// never read as a collection's iterator.
	if !lowering.isLibraryType(returns, "Generator", "IterableIterator", "MapIterator", "SetIterator") {
		return lowering.generatorDeclarationNotYet(
			declaration,
			"declared to return "+lowering.checker.TypeToString(returns),
		)
	}
	arguments := lowering.checker.GetTypeArguments(returns)
	if len(arguments) == 0 {
		return lowering.generatorDeclarationNotYet(declaration, "without a yielded type")
	}
	yielded := arguments[0]
	if yielded.Flags()&checker.TypeFlagsNever != 0 {
		return lowering.generatorDeclarationNotYet(declaration, "that yields nothing")
	}
	of, known := lowering.representation(yielded)
	switch {
	case !known:
	case of == ir.Number, of == ir.Boolean, of == ir.String, of == ir.Object, of == ir.Array, of == ir.Map, of == ir.Record:
		function.Generator = true
		function.Returns = ir.Generator
		function.Yields = of
		function.AsyncSite = lowering.asyncSite(declaration)
		return nil
	}
	return lowering.generatorDeclarationNotYet(
		declaration,
		"yielding "+lowering.checker.TypeToString(yielded)+" (each yielded value must be one present slot)",
	)
}

// checkGeneratorBody stops what a lowered generator body holds that its frame can't yet: a local a
// closure captures, which would put a canonical closure's environment on a generator's cells
// (closure.c's canonical_owner panics on one), arguments, whose count no frame cell carries yet, and,
// in a nested generator or one written in place, each a closure, a write to a local it captured
// (nestedGeneratorWrites). Each stop names its declaration through generatorTitle.
func (lowering *lowering) checkGeneratorBody(index int, declaration *ast.Node) error {
	function := lowering.result.Functions[index]
	if !function.Generator {
		return nil
	}
	if function.ArgumentsCount != 0 || function.ReadsArguments {
		return lowering.generatorDeclarationNotYet(declaration, "reading arguments")
	}
	if function.NestedFrame || function.FrameIdentity != 0 {
		return lowering.generatorDeclarationNotYet(declaration, "declaring a nested function")
	}
	for local, declared := range lowering.result.Locals {
		if declared.Function == index && declared.Captured {
			where, title := generatorTitle(declaration)
			if node := lowering.localNodes[local]; node != nil {
				where = node
			}
			return lowering.notYet(where, title+
				" whose local '"+
				declared.Name+
				"' a closure captures (a frame cell can't be a closure's environment yet) (#qk8rztp)")
		}
	}
	if function.Closure {
		return lowering.nestedGeneratorWrites(index, declaration)
	}
	return nil
}

// nestedGeneratorWrites stops a nested generator whose body, or any function declared in it, writes
// a local its carrier holds a cell of, one some enclosing function declared (Lower's term 3 on
// #qk8rztp S3a). The writes are counted on the IR by every field localFields says writes a local, as
// the one-read rule counts them, so a compound, logical or ++ assignment, a destructuring target and
// a for-of head over the local each count. A write by the parent, or by a sibling the body calls, is
// no write inside the generator: the body's next step reads it live, through the carrier's cell.
func (lowering *lowering) nestedGeneratorWrites(index int, declaration *ast.Node) error {
	// Inside is every function whose declaration lies within the generator's, by the source's own
	// nesting (Compiler's rule on #qk8rztp S3a), so no form of function written in the body is a road
	// the count forgets: an arrow, a function expression, a nested declaration, an object literal's
	// method or accessor. A sibling declared beside the generator in its parent lies outside.
	inside := map[int]bool{index: true}
	for function, node := range lowering.functionNodes {
		if declaredWithin(node, declaration) {
			inside[function] = true
		}
	}
	bodies := slices.Sorted(maps.Keys(inside))
	for position := 0; position < len(bodies); position++ {
		walk(lowering.result.Functions[bodies[position]].Body, func(node any) bool {
			// A helper lowering makes with no declaration of its own, as a prefix update's, is
			// inside when a function inside makes it.
			if made, isMade := node.(ir.MakeClosure); isMade &&
				!inside[made.Function] &&
				lowering.functionNodes[made.Function] == nil {
				inside[made.Function] = true
				bodies = append(bodies, made.Function)
			}
			return true
		})
	}
	// A local the body or its closures write from outside is one some closure among them reads
	// through a cell, so it is in that closure's environment.
	checked := map[int]bool{}
	for _, body := range bodies {
		for _, local := range lowering.result.Functions[body].Environment {
			declared := lowering.result.Locals[local]
			if checked[local] || declared.Global || inside[declared.Function] {
				continue
			}
			checked[local] = true
			count := localUseCount{}
			for _, writer := range bodies {
				walk(lowering.result.Functions[writer].Body, func(node any) bool {
					count.node(node, local)
					return true
				})
			}
			if count.writes != 0 {
				// A generator written in place is named by generatorTitle already.
				writer := "a nested generator"
				if declaration.Kind == ast.KindFunctionExpression {
					writer = "it"
				}
				return lowering.generatorDeclarationNotYet(
					declaration,
					"where '"+declared.Name+"', a captured local, is written inside "+writer,
				)
			}
		}
	}
	return nil
}

// declaredWithin reports whether node is declaration or lies inside it, by the AST's parents.
func declaredWithin(node, declaration *ast.Node) bool {
	for at := node; at != nil; at = at.Parent {
		if at == declaration {
			return true
		}
	}
	return false
}

// yieldStatement lowers yield value; as a statement of a lowered generator's body.
func (lowering *lowering) yieldStatement(node *ast.Node) ([]ir.Statement, error) {
	yield := node.AsYieldExpression()
	if lowering.function == nil || !lowering.function.Generator {
		return nil, lowering.generatorNotYet(node, "yielding from a function this slice doesn't lower")
	}
	if yield.AsteriskToken != nil {
		return nil, lowering.generatorNotYet(node, "delegating with yield*")
	}
	if yield.Expression == nil {
		return nil, lowering.generatorNotYet(node, "yielding without a value")
	}
	for parent := node.Parent; parent != nil && !ast.IsFunctionLike(parent); parent = parent.Parent {
		if parent.Kind == ast.KindTryStatement {
			// A try around a yield needs a throw or return delivered at the yield (#qk8rztp S5).
			return nil, lowering.generatorNotYet(node, "yielding inside try")
		}
	}
	value, expressionError := lowering.expression(yield.Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	value = fit(value, lowering.function.Yields)
	if value.Type() != lowering.function.Yields {
		return nil, lowering.generatorNotYet(node, "yielding a value held differently from its other yields")
	}
	return []ir.Statement{ir.Evaluate{Value: ir.Yield{Value: value, Site: lowering.asyncSite(node)}}}, nil
}

// generatorCallee is the generator function a call calls, and whether it is one. A call by name is
// read from its callee; any other callee (a namespace's member, (0, gen)) from the signature the
// checker resolved, so a generator called another way still routes by its origin, never by its
// declared iterator type, and generatorCall stops it there. A generic generator's function is its
// instance for this call, lowered here if no call has made it yet, which is why this can fail. A
// nested generator is found by its binding in its parent's scope (nestedGeneratorCallee), and one
// written in place is the callee itself, lowered the first time any consumer asks
// (inPlaceGenerator).
func (lowering *lowering) generatorCallee(node *ast.Node) (int, bool, error) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return 0, false, nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if generatorInPlace(callee) {
		index, inPlaceGeneratorError := lowering.inPlaceGenerator(callee)
		return index, inPlaceGeneratorError == nil, inPlaceGeneratorError
	}
	named := callee
	if !ast.IsIdentifier(callee) {
		// A generator written in place is the callee itself, above, so a declaration with no name
		// reached here is one this call doesn't place, and it fails closed (Adamic on #qk8rztp S3b).
		signature := lowering.checker.GetResolvedSignature(node)
		if signature == nil ||
			signature.Declaration() == nil ||
			!loweredGenerator(signature.Declaration()) ||
			signature.Declaration().Name() == nil {
			return 0, false, nil
		}
		named = signature.Declaration().Name()
	}
	if index, isNested := lowering.nestedGeneratorCallee(named); isNested {
		return index, true, nil
	}
	if declaration, isGeneric := lowering.generics[lowering.symbol(named)]; isGeneric && loweredGenerator(declaration) {
		index, instantiateFunctionError := lowering.instantiateFunction(node, declaration)
		if instantiateFunctionError != nil {
			return 0, false, instantiateFunctionError
		}
		return index, lowering.result.Functions[index].Generator, nil
	}
	index, isFunction := lowering.functions[lowering.symbol(named)]
	if !isFunction || !lowering.result.Functions[index].Generator {
		return 0, false, nil
	}
	return index, true, nil
}

// generatorCall lowers a direct call to a generator function, which makes its frame and runs none
// of its body. Only the two consumers that keep the generator to themselves call it.
func (lowering *lowering) generatorCall(node *ast.Node, function int) (ir.Expression, error) {
	call := ast.SkipParentheses(node).AsCallExpression()
	if callee := ast.SkipParentheses(call.Expression); !ast.IsIdentifier(callee) && !generatorInPlace(callee) {
		// A callee other than the generator's name could do something before the call (a getter, a
		// comma's left), which no consumer here evaluates yet (Compiler, on #qk8rztp). One written in
		// place does nothing before it but make its closure, the carrier its call is made through.
		return nil, lowering.generatorNotYet(node, "called other than by its name")
	}
	if call.QuestionDotToken != nil {
		return nil, lowering.generatorNotYet(node, "called optionally")
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, lowering.generatorNotYet(argument, "called with a spread argument")
		}
	}
	if lowering.result.Functions[function].Closure {
		return lowering.nestedGeneratorCall(call, function)
	}
	return lowering.callFunction(call, function)
}

// nestedGeneratorCallee is the nested generator a name binds, and whether it binds one: a function*
// nestedDeclarations bound in an enclosing function's scope (#qk8rztp S3a).
func (lowering *lowering) nestedGeneratorCallee(name *ast.Node) (int, bool) {
	if !ast.IsIdentifier(name) {
		return 0, false
	}
	local, isLocal := lowering.locals[lowering.symbol(name)]
	if !isLocal {
		return 0, false
	}
	index := lowering.result.Locals[local].NestedFunction - 1
	if index < 0 || !lowering.result.Functions[index].Generator {
		return 0, false
	}
	return index, true
}

// nestedGeneratorCall lowers a call to a nested generator, which makes its frame and runs none of
// its body, through the closure its body's captured reads go through (Lower's term 1 on #qk8rztp
// S3a): from a sibling, the sibling's own closure, whose cells have the generator's layout
// (ClosureSelf); from an anonymous closure that forwards to the group, one made for the call
// (MakeClosure), as call does for any nested function; and from anywhere else in the parent's scope,
// the parent's own binding of it, or one made as nestedReference makes it. The frame keeps that
// carrier (its self), so whatever made it may let go of it before the first step. A generator
// written in place has no binding: its carrier is its own closure, made for the call with the
// expression's own cells, a temporary its statement lets go of, which the frame keeps all the same
// (#qk8rztp S3b, Memory's reading of S3a's must).
func (lowering *lowering) nestedGeneratorCall(call *ast.CallExpression, function int) (ir.Expression, error) {
	callee := ast.SkipParentheses(call.Expression)
	var carrier ir.Expression
	if generatorInPlace(callee) {
		carrier = ir.MakeClosure{Function: function}
	} else if lowering.nestedSibling(callee) == function {
		carrier = ir.ClosureSelf{}
		if lowering.function.NestedParent != lowering.result.Functions[function].NestedParent {
			carrier = ir.MakeClosure{Function: function}
		}
	} else {
		carrier = lowering.nestedValue(lowering.locals[lowering.symbol(callee)])
	}
	parameters := lowering.result.Functions[function].Parameters
	if len(call.Arguments.Nodes) != len(parameters) {
		return nil, lowering.generatorNotYet(call.AsNode(), "called with other than one argument per parameter")
	}
	arguments := []ir.Expression{}
	for position, argument := range call.Arguments.Nodes {
		value, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, expressionError
		}
		arguments = append(arguments, fit(value, lowering.result.Locals[parameters[position]].Type))
	}
	return ir.CallClosure{Closure: carrier, Direct: function + 1, Arguments: arguments, Returns: ir.Generator}, nil
}

// inPlaceGeneratorKey is a generator written in place, by its expression and the function it's
// lowered in (Lower's and Compiler's key on #qk8rztp S3b).
type inPlaceGeneratorKey struct {
	expression *ast.Node
	function   int
}

// inPlaceGenerator is the function a generator written in place lowers to: the one this lowering
// made the first time a consumer asked, or made now. A consumer asks before it decides (forOf, a
// statement, planIteration, iterated and an array literal's spread each read generatorCallee first,
// and the array literal asks again through iterated), so one call may be asked about more than once,
// and the expression must still be one function (TestInPlaceGeneratorIsLoweredOncePerSite).
func (lowering *lowering) inPlaceGenerator(expression *ast.Node) (int, error) {
	if index, lowered := lowering.inPlaceGenerators[inPlaceGeneratorKey{
		expression: expression,
		function:   lowering.functionIndex,
	}]; lowered {
		return index, nil
	}
	return lowering.lowerInPlaceGenerator(expression)
}

// lowerInPlaceGenerator lowers a generator written in place as the closure it is, with its own
// cells, once: a second lowering of the same expression in the same function would be a second
// function for one call, which only a missed cache read could ask for, so it is a compiler bug said
// out loud. It stops one at a module's top level or inside a function that suspends, whose locals
// live in its own frame's cells, which no closure's environment can be yet (nestedDeclarations'
// stop for a nested generator).
func (lowering *lowering) lowerInPlaceGenerator(expression *ast.Node) (int, error) {
	key := inPlaceGeneratorKey{expression: expression, function: lowering.functionIndex}
	if index, lowered := lowering.inPlaceGenerators[key]; lowered {
		return 0, fmt.Errorf(
			"lower: compiler bug: the generator written in place at %s lowered again, where function %d already is it",
			lowering.program.Where(expression),
			index,
		)
	}
	switch {
	case lowering.function == nil:
		return 0, lowering.generatorNotYet(expression, "written in place at a module's top level")
	case lowering.function.Suspends():
		return 0, lowering.generatorNotYet(expression, "written in place inside an async function or a generator")
	}
	made, closureExpressionError := lowering.closureExpression(expression)
	if closureExpressionError != nil {
		return 0, closureExpressionError
	}
	closure, isClosure := made.(ir.MakeClosure)
	if !isClosure || !lowering.result.Functions[closure.Function].Generator {
		return 0, fmt.Errorf(
			"lower: compiler bug: the generator written in place at %s lowered to no generator's closure",
			lowering.program.Where(expression),
		)
	}
	if lowering.inPlaceGenerators == nil {
		lowering.inPlaceGenerators = map[inPlaceGeneratorKey]int{}
	}
	lowering.inPlaceGenerators[key] = closure.Function
	return closure.Function, nil
}

// iterableArgument reports whether a generator's call, parentheses aside, is an argument passed
// straight into a parameter typed the library's Iterable<T> or IterableIterator<T>, or either
// `| undefined` (iterableParameter), of a callee resolved by its symbol to exactly one body: a
// module's own function declaration, an overload set judged at its implementation, or a generic
// one, judged per instance where it's lowered. Never a method, a function value, a closure, a
// library function, a generator or an async function, and never through a spread or an optional
// call (Adamic's ruling on #qk8rztp S2c-1, conditions 1 and 2). It is syntactic, so the refusal walk
// and lowering agree on it; that the parameter's one use is one for-of is proven after lowering
// (checkIterableViews), and a program this admits that the proof refuses stops there, not yet, which
// fails closed.
func (lowering *lowering) iterableArgument(node *ast.Node) bool {
	admitted, _ := lowering.iterableParameter(node)
	return admitted
}

// iterableParameter is iterableArgument, and whether the parameter it admits is nullable: typed
// `Iterable<T> | undefined` or `IterableIterator<T> | undefined`, a union of exactly one of the
// library's two and undefined, as tsc's firstOrUndefinedIterator's is (#qk8rztp S2c-3). Never
// `| null`, which a for-of's undefined test doesn't rule out, and never `?` or a default, which stay
// refused above. That a nullable one is held as a nullable object, never a union, is decided where
// the parameter's slot is known, after lowering (generatorParameter), since a generic instance's
// T isn't known to the refusal walk.
func (lowering *lowering) iterableParameter(node *ast.Node) (bool, bool) {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false, false
	}
	call := parent.AsCallExpression()
	if call.QuestionDotToken != nil || call.Expression == site {
		return false, false
	}
	position := -1
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false, false
		}
		if argument == site {
			position = index
		}
	}
	callee := ast.SkipParentheses(call.Expression)
	if position < 0 || !ast.IsIdentifier(callee) {
		return false, false
	}
	implementation := lowering.iterableCallee(lowering.symbol(callee))
	if implementation == nil {
		return false, false
	}
	parameters := implementation.Parameters()
	if position >= len(parameters) {
		return false, false
	}
	parameter := parameters[position]
	declared := parameter.AsParameterDeclaration()
	if !ast.IsIdentifier(parameter.Name()) ||
		parameter.Name().Text() == "this" ||
		declared.DotDotDotToken != nil ||
		declared.QuestionToken != nil ||
		declared.Initializer != nil {
		return false, false
	}
	parameterType := lowering.checker.GetTypeAtLocation(parameter.Name())
	if lowering.isLibraryType(parameterType, "Iterable", "IterableIterator") {
		return true, false
	}
	if parameterType.Flags()&checker.TypeFlagsUnion == 0 || len(parameterType.Types()) != 2 {
		return false, false
	}
	// Two members, one of them undefined, and what's left once it's taken away the library's own.
	missing := 0
	for _, member := range parameterType.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			missing++
		}
	}
	nullable := missing == 1 &&
		lowering.isLibraryType(lowering.checker.GetNonNullableType(parameterType), "Iterable", "IterableIterator")
	return nullable, nullable
}

// iterableCallee is the one body a symbol names, when it is a module-level function declaration
// that is neither a generator nor async: its only declaration, or the implementation of an overload
// set whose other declarations are its signatures.
func (lowering *lowering) iterableCallee(symbol *ast.Symbol) *ast.Node {
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return nil
	}
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration ||
			declaration.Parent == nil ||
			declaration.Parent.Kind != ast.KindSourceFile {
			return nil
		}
		if declaration.Body() == nil {
			continue
		}
		if implementation != nil {
			return nil
		}
		implementation = declaration
	}
	if implementation == nil ||
		generatorDeclaration(implementation) ||
		ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) {
		return nil
	}
	return implementation
}

// iterableOfGenerator lowers a generator's call passed into an Iterable<T> parameter
// (iterableArgument): the call makes the frame, IterableOf hands it over as the object the slot
// holds, and the site is kept for checkIterableViews to prove and to name.
func (lowering *lowering) iterableOfGenerator(node *ast.Node, function int) (ir.Expression, error) {
	if lowering.function != nil && lowering.function.Suspends() {
		return nil, lowering.generatorNotYet(
			node,
			"passed into an Iterable<T> parameter inside an async function or a generator",
		)
	}
	made, generatorCallError := lowering.generatorCall(node, function)
	if generatorCallError != nil {
		return nil, generatorCallError
	}
	argument := generatorArgument{node: node, generator: function}
	argument.key, argument.value, argument.pair = lowering.generatorPairTypes(node)
	_, argument.nullable = lowering.iterableParameter(node)
	lowering.generatorArguments = append(lowering.generatorArguments, argument)
	return ir.IterableOf{Value: made, Site: len(lowering.generatorArguments)}, nil
}

// generatorArgument is a generator's call iterableOfGenerator handed over, how its yields hold a
// [key, value] pair when they are one, and whether the parameter it goes into may be undefined
// (iterableParameter).
type generatorArgument struct {
	node       *ast.Node
	generator  int
	pair       bool
	key, value ir.Type
	nullable   bool
}

// forOfGenerator lowers for (const x of gen(a)) by the generator's origin, never by its declared
// iterator type: the call makes g, each pass steps it, a pass that finds it done leaves, and the
// binding is the step's value, a fresh one each pass. g is the block's local, so every way out of
// the loop, a break, a return, a throw or a jump to an outer loop, lets go of it once, and with it
// the frame and whatever the frame still holds (Memory, 04:42: no finally exists yet, so releasing
// g is the whole of IteratorClose).
func (lowering *lowering) forOfGenerator(node *ast.Node, function int, name *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForInOrOfStatement()
	if lowering.function != nil && lowering.function.Suspends() {
		return nil, lowering.generatorNotYet(
			statement.Expression,
			"consumed by a for-of inside an async function or a generator",
		)
	}
	if lowering.function == nil && moduleAwaits(ast.GetSourceFileOfNode(node).AsNode()) {
		return nil, lowering.generatorNotYet(
			statement.Expression,
			"consumed by a for-of in a module that awaits at its top level",
		)
	}
	var awaits ast.Visitor
	awaits = func(child *ast.Node) bool {
		if ast.IsFunctionLike(child) {
			return false
		}
		return child.Kind == ast.KindAwaitExpression || child.ForEachChild(awaits)
	}
	if awaits(node) {
		return nil, lowering.generatorNotYet(statement.Expression, "consumed by a for-of that awaits")
	}
	if !ast.IsIdentifier(name) {
		return nil, lowering.generatorNotYet(name, "consumed by a for-of that destructures")
	}
	made, generatorCallError := lowering.generatorCall(statement.Expression, function)
	if generatorCallError != nil {
		return nil, generatorCallError
	}
	yields := lowering.result.Functions[function].Yields
	generator := lowering.iterationLocal("generator", ir.Generator, lowering.functionIndex)
	step := lowering.iterationLocal("generator_step", yields, lowering.functionIndex)
	local, declareLocalError := lowering.declareLocal(name)
	if declareLocalError != nil {
		return nil, declareLocalError
	}
	if lowering.result.Locals[local].Type != yields {
		return nil, lowering.generatorNotYet(name, "consumed into a variable held differently from what it yields")
	}
	body, statementError := lowering.statement(statement.Statement)
	if statementError != nil {
		return nil, statementError
	}
	held := ir.Read{Local: generator, Of: ir.Generator}
	pass := []ir.Statement{
		// The step is tested for a throw before done is read (a step that threw also finished the
		// body): the native backend tests after a call that can throw, and this is one when
		// ResumeMayThrow says so.
		ir.Declare{Local: step, Value: ir.GeneratorNext{Generator: held, Function: function, Of: yields}},
		ir.If{Condition: ir.GeneratorDone{Generator: held}, Then: []ir.Statement{ir.Break{}}},
		ir.Declare{Local: local, Value: ir.Read{Local: step, Of: yields}},
	}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: append(pass, body...)}
	return []ir.Statement{ir.Block{Body: []ir.Statement{ir.Declare{Local: generator, Value: made}, loop}}}, nil
}

// checkGeneratorRoads holds the lowered program to this slice's one rule for generator objects: a
// generator's call is the value of its consuming loop's local or a statement's whole expression, and
// that local is read only by a step or a done test. A road a lowering missed would compile a frame
// into a slot no analysis knows holds one, so it is a compiler bug, said out loud.
func checkGeneratorRoads(program *ir.Program) error {
	counted, allowed := 0, 0
	held := map[int]bool{}
	// IterableOf is the one door from a generator to an object's slot, and only a call's own argument
	// goes through it; what a view's dispatch holds, no runtime call or field read takes, only its
	// kind test and its steps (Compiler's and Memory's conditions on #qk8rztp S2c-1).
	iterables, handed := 0, 0
	dispatched, stepped, nexts, reads := map[int]bool{}, map[int]int{}, map[int]int{}, map[int]int{}
	collectionTests := map[int]int{}
	var visit func(value reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				visit(value.Elem())
			}
			return
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
			return
		case reflect.Struct:
		default:
			return
		}
		switch node := value.Interface().(type) {
		case ir.Declare:
			if makesGenerator(node.Value) && program.Locals[node.Local].Type == ir.Generator {
				allowed++
				held[node.Local] = true
			}
		case ir.Evaluate:
			if makesGenerator(node.Value) {
				allowed++
			}
		case ir.Call:
			// A generator handed into an Iterable<T> parameter, as the call's own argument.
			for _, argument := range node.Arguments {
				if of, ok := argument.(ir.IterableOf); ok {
					handed++
					if makesGenerator(of.Value) {
						allowed++
					}
				}
			}
		case ir.IterableOf:
			iterables++
		case ir.Block:
			if node.Closes != 0 {
				dispatched[node.Closes-1] = true
			}
		case ir.IsGenerator:
			// The view's dispatch reads what it holds only to test its kind and to step it: by a
			// generator's step, or by a collection iterator's, which reads its next.
			if read, ok := node.Value.(ir.Read); ok && read.Of == ir.Object {
				stepped[read.Local]++
			}
		case ir.IsCollectionIterator:
			if read, ok := node.Value.(ir.Read); ok && read.Of == ir.Object {
				stepped[read.Local]++
				collectionTests[read.Local]++
			}
		case ir.Property:
			if read, ok := node.Object.(ir.Read); ok && read.Of == ir.Object && node.Name == "next" {
				nexts[read.Local]++
			}
		case ir.Read:
			reads[node.Local]++
		case ir.GeneratorNext:
			if read, ok := node.Generator.(ir.Read); ok && read.Of == ir.Generator {
				allowed++
			} else if ok && read.Of == ir.Object {
				stepped[read.Local]++
			}
		case ir.GeneratorDone:
			if read, ok := node.Generator.(ir.Read); ok && read.Of == ir.Generator {
				allowed++
			} else if ok && read.Of == ir.Object {
				stepped[read.Local]++
			}
		}
		if expression, ok := value.Interface().(ir.Expression); ok && expression.Type() == ir.Generator {
			counted++
		}
		for index := 0; index < value.NumField(); index++ {
			visit(value.Field(index))
		}
	}
	visit(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	if counted != allowed {
		return fmt.Errorf(
			"lower: compiler bug: %d generator values where only %d are a consuming loop's or a dropped call's",
			counted,
			allowed,
		)
	}
	if iterables != handed {
		return fmt.Errorf(
			"lower: compiler bug: %d generators seen as objects where only %d are a call's own argument",
			iterables,
			handed,
		)
	}
	for local := range dispatched {
		// One read of next per collection test, the arm's own step (Lower's must on #qk8rztp S2c-1,
		// her option (b)): any other next of the local isn't the dispatch's.
		if nexts[local] != collectionTests[local] || reads[local] != stepped[local]+nexts[local] {
			return fmt.Errorf(
				"lower: compiler bug: the iterable view's local %s is read %d times where only its dispatch's %d kind tests and steps may read it",
				program.Locals[local].Name,
				reads[local],
				stepped[local]+nexts[local],
			)
		}
	}
	for local, declared := range program.Locals {
		if declared.Type == ir.Generator && (!held[local] || declared.Global || declared.Captured) {
			return fmt.Errorf(
				"lower: compiler bug: the generator local %s is held outside its consuming loop",
				declared.Name,
			)
		}
	}
	return nil
}

// makesGenerator is a generator's own call, the one expression that makes its frame: a module
// generator's ir.Call, or a nested one's ir.CallClosure through its carrier, direct to its function
// (nestedGeneratorCall).
func makesGenerator(value ir.Expression) bool {
	switch made := value.(type) {
	case ir.Call:
		return made.Returns == ir.Generator
	case ir.CallClosure:
		return made.Returns == ir.Generator && made.Direct > 0
	}
	return false
}

// collectGenerator lowers [...gen(a)] and Array.from(gen(a)) by the generator's origin: a helper of
// its own per call site takes the call's arguments, makes the generator, steps it to its end pushing
// each value, and returns the array. Its arguments are evaluated where the spread or Array.from
// stands, before the generator is made, as JavaScript evaluates them, and the generator is the
// helper's local, held only by its loop, so the helper's way out, a return or a throw, lets go of it.
func (lowering *lowering) collectGenerator(node *ast.Node, function int) (ir.Expression, ir.Type, error) {
	if lowering.result.Functions[function].Closure {
		return nil, 0, lowering.nestedCollectNotYet(node, function)
	}
	if lowering.function != nil && lowering.function.Suspends() {
		return nil, 0, lowering.generatorNotYet(
			node,
			"consumed by spread or Array.from inside an async function or a generator",
		)
	}
	made, generatorCallError := lowering.generatorCall(node, function)
	if generatorCallError != nil {
		return nil, 0, generatorCallError
	}
	call := made.(ir.Call)
	yields := lowering.result.Functions[function].Yields
	helper := len(lowering.result.Functions)
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: "collect_generator", Returns: ir.Array},
	)
	var parameters []int
	var arguments []ir.Expression
	for _, parameter := range lowering.result.Functions[function].Parameters {
		of := lowering.result.Locals[parameter].Type
		local := lowering.iterationLocal(lowering.result.Locals[parameter].Name, of, helper)
		parameters = append(parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: of})
	}
	if len(arguments) != len(call.Arguments) {
		return nil, 0, lowering.generatorNotYet(node, "called with other than one argument per parameter")
	}
	generator := lowering.iterationLocal("generator", ir.Generator, helper)
	step := lowering.iterationLocal("generator_step", yields, helper)
	output := lowering.iterationLocal("collected", ir.Array, helper)
	held := ir.Read{Local: generator, Of: ir.Generator}
	collected := ir.Read{Local: output, Of: ir.Array}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: []ir.Statement{
		ir.Declare{Local: step, Value: ir.GeneratorNext{Generator: held, Function: function, Of: yields}},
		ir.If{Condition: ir.GeneratorDone{Generator: held}, Then: []ir.Statement{ir.Break{}}},
		ir.Evaluate{Value: ir.ArrayPush{
			Array:   collected,
			Value:   ir.Read{Local: step, Of: yields},
			Element: yields,
			Site:    lowering.iterationArrayWriteSite(node),
		}},
	}}
	body := []ir.Statement{
		ir.Declare{Local: output, Value: ir.ArrayLiteral{Element: yields}},
		ir.Block{Body: []ir.Statement{ir.Declare{
			Local: generator,
			Value: ir.Call{Function: function, Arguments: arguments, Returns: ir.Generator},
		}, loop}},
		ir.Return{Value: collected},
	}
	lowering.result.Functions[helper] = ir.Function{
		Name:       "collect_generator",
		Returns:    ir.Array,
		Parameters: parameters,
		Body:       body,
	}
	return ir.Call{Function: helper, Arguments: call.Arguments, Returns: ir.Array}, yields, nil
}

// nestedCollectNotYet stops a nested generator's call a collect helper would step. The helper makes
// the frame, so it would be handed the carrier, and it is an ordinary function, whose parameters
// region.go's environment proof doesn't count as escaping: the parent's environment could be put off
// the heap while the frame holds a closure into it (Memory's seam on #qk8rztp S3a, #s6tt5xt). A
// generator written in place is a closure too, with a carrier of its own, so it stops the same way,
// in its own words (#qk8rztp S3b).
func (lowering *lowering) nestedCollectNotYet(node *ast.Node, function int) error {
	if lowering.result.Functions[function].NestedParent == 0 {
		return lowering.generatorNotYet(node, "written in place and consumed by spread, Array.from, new Map or new Set")
	}
	return lowering.generatorNotYet(
		node,
		"declared inside a function and consumed by spread, Array.from, new Map or new Set",
	)
}

// newMapFromGenerator lowers new Map(gen(a)) by the generator's origin: collectGenerator steps it to
// its end, as a spread does, and the Map is made from the pairs it gave, each set in order. Node's
// constructor sets each pair before it takes the next step; setting them all after the last step is
// the same, since nothing can see the Map before the constructor returns it (no step reaches it) and
// set on a fresh built-in Map can't throw. A step that throws leaves the helper before any Map is
// made, and the pairs it had are let go of with it. The pair's types are read from what the
// generator yields, never from its declared iterator type.
func (lowering *lowering) newMapFromGenerator(
	source *ast.Node,
	function int,
	key ir.Type,
	value ir.Type,
) (ir.Expression, error) {
	if lowering.result.Functions[function].Closure {
		return nil, lowering.nestedCollectNotYet(source, function)
	}
	pairKey, pairValue, isPair := lowering.generatorPairTypes(source)
	if !isPair {
		// The words newMapFrom gives any other source that isn't pairs, and below any other pairs held
		// otherwise, so each stop is found beside the original (Lower, on #qk8rztp).
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	if pairKey != key || pairValue != value {
		return nil, lowering.notYet(source, "new Map from pairs held otherwise than the Map's keys and values")
	}
	pairs, yields, collectGeneratorError := lowering.collectGenerator(source, function)
	if collectGeneratorError != nil {
		return nil, collectGeneratorError
	}
	if yields != ir.Object {
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	return ir.MapNew{Key: key, Value: value, Pairs: pairs}, nil
}

// generatorPairTypes is how the [key, value] pairs a generator's call yields hold their keys and
// values: its yielded type, the first type argument of what it returns, must be a tuple of exactly
// two required elements, each held in one slot.
func (lowering *lowering) generatorPairTypes(source *ast.Node) (ir.Type, ir.Type, bool) {
	arguments := lowering.checker.GetTypeArguments(
		lowering.concrete(lowering.checker.GetTypeAtLocation(ast.SkipParentheses(source))),
	)
	if len(arguments) == 0 || !checker.IsTupleType(arguments[0]) {
		return 0, 0, false
	}
	elements := lowering.checker.GetTypeArguments(arguments[0])
	flags := arguments[0].TargetTupleType().ElementFlags()
	if len(elements) != 2 ||
		len(flags) != 2 ||
		flags[0] != checker.ElementFlagsRequired ||
		flags[1] != checker.ElementFlagsRequired {
		return 0, 0, false
	}
	key, keyKnown := lowering.representation(elements[0])
	value, valueKnown := lowering.representation(elements[1])
	return key, value, keyKnown && valueKnown
}

// generatorStepped is a direct call to a generator this slice lowers that a stepping consumer
// consumes: a spread, Array.from, new Map or new Set (steppedSite). Each steps it to its end and
// never reads its last result's value, so a TReturn of void, what a plain function* infers and what
// tsc's mapIterator declares, is never read there, and the void rules (adamic/no-void-read,
// no-void-widening) have nothing to refuse in it. It holds only when the yielded type holds no void
// at any depth (for new Map, neither element of the pair): every value a step hands over is read. A
// stored generator, a user iterable or an iterator object is no direct call, and stays judged. It
// is syntactic, since the refusal walk runs before any function is lowered. It applies rulings.md's
// void-container classifier (judged by expression) to a generator, as Adamic ruled on #qk8rztp, for
// spread and Array.from (S1c) and for new Map and new Set (S2b-1c). cohere's gate never refused the
// S1c programs: it has no twin of no-void-read, and its no-void-widening judges function-type
// relations (#z7zv50g).
//
// The callee is placed by its body, never by its name alone (Adamic on #qk8rztp S3b): a name's one
// declaration, which must be a generator this compiler lowers, or a generator written in place and
// called at once, whose body is the callee itself. Anything else it can't place, a comma, a
// conditional's arm or an ordinary function written in place among them, stays judged.
func (lowering *lowering) generatorStepped(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !generatorInPlace(callee) {
		if !ast.IsIdentifier(callee) {
			return false
		}
		symbol := lowering.symbol(callee)
		if symbol == nil || len(symbol.Declarations) != 1 || !loweredGenerator(symbol.Declarations[0]) {
			return false
		}
	}
	if !lowering.steppedSite(node) {
		return false
	}
	returns := lowering.checker.GetTypeAtLocation(node)
	if !lowering.isLibraryType(returns, "Generator") {
		return false
	}
	arguments := lowering.checker.GetTypeArguments(returns)
	return len(arguments) == 3 && !lowering.isVoidValue(arguments[0]) && !lowering.holdsVoid(arguments[0])
}

// steppedSite reports whether node, parentheses aside, stands where a consumer steps it to its end:
// an array literal's spread, the one argument of the library's Array.from, or the one argument of
// the library's new Map or new Set. The constructors are judged by the built-in's symbol, never by
// the name, so a user class called Map, or a subclass of the built-in, isn't one (Language, on
// #qk8rztp): a subclass's set is program code a step can run. And, since S2c-1, the argument for a
// one-body function's Iterable<T> parameter (iterableArgument, `| undefined` too since S2c-3), whose
// only use besides its undefined tests is one for-of, which never reads the completion value either
// (Adamic's ruling on #qk8rztp, rulings.md:294 one call deeper). That use is proven after lowering; a
// program this lets past the void rules that the proof then refuses stops there, not yet, so the arm
// fails closed.
func (lowering *lowering) steppedSite(node *ast.Node) bool {
	if lowering.iterableArgument(node) {
		return true
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	switch {
	case parent == nil:
		return false
	case parent.Kind == ast.KindSpreadElement &&
		parent.Parent != nil &&
		parent.Parent.Kind == ast.KindArrayLiteralExpression:
		return true
	case parent.Kind == ast.KindCallExpression &&
		lowering.arrayFromCall(parent) &&
		len(parent.AsCallExpression().Arguments.Nodes) == 1 &&
		parent.AsCallExpression().Arguments.Nodes[0] == site:
		return true
	case parent.Kind == ast.KindNewExpression &&
		lowering.collectionConstructor(parent) &&
		parent.AsNewExpression().Arguments != nil &&
		len(parent.AsNewExpression().Arguments.Nodes) == 1 &&
		parent.AsNewExpression().Arguments.Nodes[0] == site:
		return true
	}
	return false
}

// collectionConstructor is new Map or new Set, the library's own.
func (lowering *lowering) collectionConstructor(created *ast.Node) bool {
	callee := ast.SkipParentheses(created.AsNewExpression().Expression)
	return lowering.isLibraryGlobal(callee, "Map") || lowering.isLibraryGlobal(callee, "Set")
}

// arrayFromCall is a call of the library's Array.from.
func (lowering *lowering) arrayFromCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression &&
		callee.Name().Text() == "from" &&
		lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array")
}
