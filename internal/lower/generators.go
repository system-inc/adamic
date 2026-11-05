package lower

import (
	"fmt"
	"reflect"

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

// generatorNotYet is a generator form this slice doesn't lower.
func (l *lowering) generatorNotYet(node *ast.Node, what string) error {
	return l.notYet(node, "a generator function "+what+" (#qk8rztp)")
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
// module's top level. A generic one is lowered once per instantiation, where it's called, as every
// generic function is (generic.go), and its yielded type is judged at that instance, so one nothing
// instantiates costs nothing. Every other generator stops at lowering, wherever lowering reaches its
// declaration, in the words generatorForm gives it (Lower, open question 7 on #qk8rztp); only an
// async generator is stopped before lowering, refused for good by rule 2400.
func loweredGenerator(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().AsteriskToken != nil &&
		node.Parent != nil && node.Parent.Kind == ast.KindSourceFile && node.Name() != nil &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
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
	case node.Kind == ast.KindFunctionExpression:
		return "written as an expression"
	case node.Kind == ast.KindMethodDeclaration && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression:
		return "written as an object literal's method"
	case node.Kind == ast.KindMethodDeclaration && ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic):
		return "written as a static method"
	case node.Kind == ast.KindMethodDeclaration:
		return "written as a method"
	case node.Name() == nil:
		return "declared without a name"
	case node.Parent != nil && node.Parent.Kind == ast.KindModuleBlock:
		return "declared inside a namespace"
	case node.Parent != nil && node.Parent.Kind != ast.KindSourceFile:
		return "declared inside a function or a block"
	}
	return "of this form"
}

// generatorSignature writes a generator's result: a Generator, which yields one slot per value.
// Its parameters are written by signature, as any function's are, and must be plain names, since a
// default, a pattern or a rest would run or bind at the call, which isn't built yet.
func (l *lowering) generatorSignature(function *ir.Function, declaration *ast.Node) error {
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declared.Initializer != nil || declared.QuestionToken != nil || declared.DotDotDotToken != nil || parameter.Name().Text() == "this" {
			return l.generatorNotYet(parameter, "with a parameter that isn't a plain required name")
		}
	}
	// Under a generic instance's type mapper, concrete reads the result this instantiation made of it,
	// so the yielded type below is judged per instance: an unread type parameter stays itself, and
	// has no representation.
	returns := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(declaration)))
	// The checker's own Generator, inferred or written, or an iterator type a body may be declared to
	// return. Every consumer routes a generator's call by its origin, never by this declared type
	// (Compiler's risk 2), so a generator declared MapIterator or SetIterator is stepped as a generator,
	// never read as a collection's iterator.
	if !l.isLibraryType(returns, "Generator", "IterableIterator", "MapIterator", "SetIterator") {
		return l.generatorNotYet(declaration.Name(), "declared to return "+l.checker.TypeToString(returns))
	}
	arguments := l.checker.GetTypeArguments(returns)
	if len(arguments) == 0 {
		return l.generatorNotYet(declaration.Name(), "without a yielded type")
	}
	yielded := arguments[0]
	if yielded.Flags()&checker.TypeFlagsNever != 0 {
		return l.generatorNotYet(declaration.Name(), "that yields nothing")
	}
	of, known := l.representation(yielded)
	switch {
	case !known:
	case of == ir.Number, of == ir.Boolean, of == ir.String, of == ir.Object, of == ir.Array, of == ir.Map, of == ir.Record:
		function.Generator = true
		function.Returns = ir.Generator
		function.Yields = of
		function.AsyncSite = l.asyncSite(declaration)
		return nil
	}
	return l.generatorNotYet(declaration.Name(), "yielding "+l.checker.TypeToString(yielded)+" (each yielded value must be one present slot)")
}

// checkGeneratorBody stops what a lowered generator body holds that its frame can't yet: a local a
// closure captures, which would put a canonical closure's environment on a generator's cells
// (closure.c's canonical_owner panics on one), and arguments, whose count no frame cell carries yet.
func (l *lowering) checkGeneratorBody(index int, declaration *ast.Node) error {
	function := l.result.Functions[index]
	if !function.Generator {
		return nil
	}
	if function.ArgumentsCount != 0 || function.ReadsArguments {
		return l.generatorNotYet(declaration.Name(), "reading arguments")
	}
	if function.NestedFrame || function.FrameIdentity != 0 {
		return l.generatorNotYet(declaration.Name(), "declaring a nested function")
	}
	for local, declared := range l.result.Locals {
		if declared.Function == index && declared.Captured {
			where := declaration.Name()
			if node := l.localNodes[local]; node != nil {
				where = node
			}
			return l.generatorNotYet(where, "whose local '"+declared.Name+"' a closure captures (a frame cell can't be a closure's environment yet)")
		}
	}
	return nil
}

// yieldStatement lowers yield value; as a statement of a lowered generator's body.
func (l *lowering) yieldStatement(node *ast.Node) ([]ir.Statement, error) {
	yield := node.AsYieldExpression()
	if l.function == nil || !l.function.Generator {
		return nil, l.generatorNotYet(node, "yielding from a function this slice doesn't lower")
	}
	if yield.AsteriskToken != nil {
		return nil, l.generatorNotYet(node, "delegating with yield*")
	}
	if yield.Expression == nil {
		return nil, l.generatorNotYet(node, "yielding without a value")
	}
	for parent := node.Parent; parent != nil && !ast.IsFunctionLike(parent); parent = parent.Parent {
		if parent.Kind == ast.KindTryStatement {
			// A try around a yield needs a throw or return delivered at the yield (#qk8rztp S5).
			return nil, l.generatorNotYet(node, "yielding inside try")
		}
	}
	value, err := l.expression(yield.Expression)
	if err != nil {
		return nil, err
	}
	value = fit(value, l.function.Yields)
	if value.Type() != l.function.Yields {
		return nil, l.generatorNotYet(node, "yielding a value held differently from its other yields")
	}
	return []ir.Statement{ir.Evaluate{Value: ir.Yield{Value: value, Site: l.asyncSite(node)}}}, nil
}

// generatorCallee is the generator function a call calls, and whether it is one. A call by name is
// read from its callee; any other callee (a namespace's member, (0, gen)) from the signature the
// checker resolved, so a generator called another way still routes by its origin, never by its
// declared iterator type, and generatorCall stops it there. A generic generator's function is its
// instance for this call, lowered here if no call has made it yet, which is why this can fail.
func (l *lowering) generatorCallee(node *ast.Node) (int, bool, error) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return 0, false, nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	named := callee
	if !ast.IsIdentifier(callee) {
		signature := l.checker.GetResolvedSignature(node)
		if signature == nil || signature.Declaration() == nil || !loweredGenerator(signature.Declaration()) {
			return 0, false, nil
		}
		named = signature.Declaration().Name()
	}
	if declaration, isGeneric := l.generics[l.symbol(named)]; isGeneric && loweredGenerator(declaration) {
		index, err := l.instantiateFunction(node, declaration)
		if err != nil {
			return 0, false, err
		}
		return index, l.result.Functions[index].Generator, nil
	}
	index, isFunction := l.functions[l.symbol(named)]
	if !isFunction || !l.result.Functions[index].Generator {
		return 0, false, nil
	}
	return index, true, nil
}

// generatorCall lowers a direct call to a generator function, which makes its frame and runs none
// of its body. Only the two consumers that keep the generator to themselves call it.
func (l *lowering) generatorCall(node *ast.Node, function int) (ir.Expression, error) {
	call := ast.SkipParentheses(node).AsCallExpression()
	if !ast.IsIdentifier(ast.SkipParentheses(call.Expression)) {
		// A callee other than the generator's name could do something before the call (a getter, a
		// comma's left), which no consumer here evaluates yet (Compiler, on #qk8rztp).
		return nil, l.generatorNotYet(node, "called other than by its name")
	}
	if call.QuestionDotToken != nil {
		return nil, l.generatorNotYet(node, "called optionally")
	}
	for _, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.generatorNotYet(argument, "called with a spread argument")
		}
	}
	return l.callFunction(call, function)
}

// iterableArgument reports whether a generator's call, parentheses aside, is an argument passed
// straight into a parameter typed the library's Iterable<T> or IterableIterator<T>, of a callee
// resolved by its symbol to exactly one body: a module's own function declaration, an overload set
// judged at its implementation, or a generic one, judged per instance where it's lowered. Never a
// method, a function value, a closure, a library function, a generator or an async function, and
// never through a spread or an optional call (Adamic's ruling on #qk8rztp S2c-1, conditions 1 and
// 2). It is syntactic, so the refusal walk and lowering agree on it; that the parameter's one use is
// one for-of is proven after lowering (checkIterableViews), and a program this admits that the proof
// refuses stops there, not yet, which fails closed.
func (l *lowering) iterableArgument(node *ast.Node) bool {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call.QuestionDotToken != nil || call.Expression == site {
		return false
	}
	position := -1
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
		if argument == site {
			position = index
		}
	}
	callee := ast.SkipParentheses(call.Expression)
	if position < 0 || !ast.IsIdentifier(callee) {
		return false
	}
	implementation := l.iterableCallee(l.symbol(callee))
	if implementation == nil {
		return false
	}
	parameters := implementation.Parameters()
	if position >= len(parameters) {
		return false
	}
	parameter := parameters[position]
	declared := parameter.AsParameterDeclaration()
	if !ast.IsIdentifier(parameter.Name()) || parameter.Name().Text() == "this" || declared.DotDotDotToken != nil || declared.QuestionToken != nil || declared.Initializer != nil {
		return false
	}
	return l.isLibraryType(l.checker.GetTypeAtLocation(parameter.Name()), "Iterable", "IterableIterator")
}

// iterableCallee is the one body a symbol names, when it is a module-level function declaration
// that is neither a generator nor async: its only declaration, or the implementation of an overload
// set whose other declarations are its signatures.
func (l *lowering) iterableCallee(symbol *ast.Symbol) *ast.Node {
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return nil
	}
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindSourceFile {
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
	if implementation == nil || generatorDeclaration(implementation) || ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) {
		return nil
	}
	return implementation
}

// iterableOfGenerator lowers a generator's call passed into an Iterable<T> parameter
// (iterableArgument): the call makes the frame, IterableOf hands it over as the object the slot
// holds, and the site is kept for checkIterableViews to prove and to name.
func (l *lowering) iterableOfGenerator(node *ast.Node, function int) (ir.Expression, error) {
	if l.function != nil && l.function.Suspends() {
		return nil, l.generatorNotYet(node, "passed into an Iterable<T> parameter inside an async function or a generator")
	}
	made, err := l.generatorCall(node, function)
	if err != nil {
		return nil, err
	}
	argument := generatorArgument{node: node, generator: function}
	argument.key, argument.value, argument.pair = l.generatorPairTypes(node)
	l.generatorArguments = append(l.generatorArguments, argument)
	return ir.IterableOf{Value: made, Site: len(l.generatorArguments)}, nil
}

// generatorArgument is a generator's call iterableOfGenerator handed over, and how its yields hold
// a [key, value] pair when they are one.
type generatorArgument struct {
	node       *ast.Node
	generator  int
	pair       bool
	key, value ir.Type
}

// forOfGenerator lowers for (const x of gen(a)) by the generator's origin, never by its declared
// iterator type: the call makes g, each pass steps it, a pass that finds it done leaves, and the
// binding is the step's value, a fresh one each pass. g is the block's local, so every way out of
// the loop, a break, a return, a throw or a jump to an outer loop, lets go of it once, and with it
// the frame and whatever the frame still holds (Memory, 04:42: no finally exists yet, so releasing
// g is the whole of IteratorClose).
func (l *lowering) forOfGenerator(node *ast.Node, function int, name *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForInOrOfStatement()
	if l.function != nil && l.function.Suspends() {
		return nil, l.generatorNotYet(statement.Expression, "consumed by a for-of inside an async function or a generator")
	}
	if l.function == nil && moduleAwaits(ast.GetSourceFileOfNode(node).AsNode()) {
		return nil, l.generatorNotYet(statement.Expression, "consumed by a for-of in a module that awaits at its top level")
	}
	var awaits ast.Visitor
	awaits = func(child *ast.Node) bool {
		if ast.IsFunctionLike(child) {
			return false
		}
		return child.Kind == ast.KindAwaitExpression || child.ForEachChild(awaits)
	}
	if awaits(node) {
		return nil, l.generatorNotYet(statement.Expression, "consumed by a for-of that awaits")
	}
	if !ast.IsIdentifier(name) {
		return nil, l.generatorNotYet(name, "consumed by a for-of that destructures")
	}
	made, err := l.generatorCall(statement.Expression, function)
	if err != nil {
		return nil, err
	}
	yields := l.result.Functions[function].Yields
	generator := l.iterationLocal("generator", ir.Generator, l.functionIndex)
	step := l.iterationLocal("generator_step", yields, l.functionIndex)
	local, err := l.declareLocal(name)
	if err != nil {
		return nil, err
	}
	if l.result.Locals[local].Type != yields {
		return nil, l.generatorNotYet(name, "consumed into a variable held differently from what it yields")
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
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
			if call, ok := node.Value.(ir.Call); ok && call.Returns == ir.Generator && program.Locals[node.Local].Type == ir.Generator {
				allowed++
				held[node.Local] = true
			}
		case ir.Evaluate:
			if call, ok := node.Value.(ir.Call); ok && call.Returns == ir.Generator {
				allowed++
			}
		case ir.Call:
			// A generator handed into an Iterable<T> parameter, as the call's own argument.
			for _, argument := range node.Arguments {
				if of, ok := argument.(ir.IterableOf); ok {
					handed++
					if call, ok := of.Value.(ir.Call); ok && call.Returns == ir.Generator {
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
		return fmt.Errorf("lower: compiler bug: %d generator values where only %d are a consuming loop's or a dropped call's", counted, allowed)
	}
	if iterables != handed {
		return fmt.Errorf("lower: compiler bug: %d generators seen as objects where only %d are a call's own argument", iterables, handed)
	}
	for local := range dispatched {
		// One read of next per collection test, the arm's own step (Lower's must on #qk8rztp S2c-1,
		// her option (b)): any other next of the local isn't the dispatch's.
		if nexts[local] != collectionTests[local] || reads[local] != stepped[local]+nexts[local] {
			return fmt.Errorf("lower: compiler bug: the iterable view's local %s is read %d times where only its dispatch's %d kind tests and steps may read it", program.Locals[local].Name, reads[local], stepped[local]+nexts[local])
		}
	}
	for local, declared := range program.Locals {
		if declared.Type == ir.Generator && (!held[local] || declared.Global || declared.Captured) {
			return fmt.Errorf("lower: compiler bug: the generator local %s is held outside its consuming loop", declared.Name)
		}
	}
	return nil
}

// collectGenerator lowers [...gen(a)] and Array.from(gen(a)) by the generator's origin: a helper of
// its own per call site takes the call's arguments, makes the generator, steps it to its end pushing
// each value, and returns the array. Its arguments are evaluated where the spread or Array.from
// stands, before the generator is made, as JavaScript evaluates them, and the generator is the
// helper's local, held only by its loop, so the helper's way out, a return or a throw, lets go of it.
func (l *lowering) collectGenerator(node *ast.Node, function int) (ir.Expression, ir.Type, error) {
	if l.function != nil && l.function.Suspends() {
		return nil, 0, l.generatorNotYet(node, "consumed by spread or Array.from inside an async function or a generator")
	}
	made, err := l.generatorCall(node, function)
	if err != nil {
		return nil, 0, err
	}
	call := made.(ir.Call)
	yields := l.result.Functions[function].Yields
	helper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "collect_generator", Returns: ir.Array})
	var parameters []int
	var arguments []ir.Expression
	for _, parameter := range l.result.Functions[function].Parameters {
		of := l.result.Locals[parameter].Type
		local := l.iterationLocal(l.result.Locals[parameter].Name, of, helper)
		parameters = append(parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: of})
	}
	if len(arguments) != len(call.Arguments) {
		return nil, 0, l.generatorNotYet(node, "called with other than one argument per parameter")
	}
	generator := l.iterationLocal("generator", ir.Generator, helper)
	step := l.iterationLocal("generator_step", yields, helper)
	output := l.iterationLocal("collected", ir.Array, helper)
	held := ir.Read{Local: generator, Of: ir.Generator}
	collected := ir.Read{Local: output, Of: ir.Array}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: []ir.Statement{
		ir.Declare{Local: step, Value: ir.GeneratorNext{Generator: held, Function: function, Of: yields}},
		ir.If{Condition: ir.GeneratorDone{Generator: held}, Then: []ir.Statement{ir.Break{}}},
		ir.Evaluate{Value: ir.ArrayPush{Array: collected, Value: ir.Read{Local: step, Of: yields}, Element: yields, Site: l.iterationArrayWriteSite(node)}},
	}}
	body := []ir.Statement{
		ir.Declare{Local: output, Value: ir.ArrayLiteral{Element: yields}},
		ir.Block{Body: []ir.Statement{ir.Declare{Local: generator, Value: ir.Call{Function: function, Arguments: arguments, Returns: ir.Generator}}, loop}},
		ir.Return{Value: collected},
	}
	l.result.Functions[helper] = ir.Function{Name: "collect_generator", Returns: ir.Array, Parameters: parameters, Body: body}
	return ir.Call{Function: helper, Arguments: call.Arguments, Returns: ir.Array}, yields, nil
}

// newMapFromGenerator lowers new Map(gen(a)) by the generator's origin: collectGenerator steps it to
// its end, as a spread does, and the Map is made from the pairs it gave, each set in order. Node's
// constructor sets each pair before it takes the next step; setting them all after the last step is
// the same, since nothing can see the Map before the constructor returns it (no step reaches it) and
// set on a fresh built-in Map can't throw. A step that throws leaves the helper before any Map is
// made, and the pairs it had are let go of with it. The pair's types are read from what the
// generator yields, never from its declared iterator type.
func (l *lowering) newMapFromGenerator(source *ast.Node, function int, key ir.Type, value ir.Type) (ir.Expression, error) {
	pairKey, pairValue, isPair := l.generatorPairTypes(source)
	if !isPair {
		// The words newMapFrom gives any other source that isn't pairs, and below any other pairs held
		// otherwise, so each stop is found beside the original (Lower, on #qk8rztp).
		return nil, l.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	if pairKey != key || pairValue != value {
		return nil, l.notYet(source, "new Map from pairs held otherwise than the Map's keys and values")
	}
	pairs, yields, err := l.collectGenerator(source, function)
	if err != nil {
		return nil, err
	}
	if yields != ir.Object {
		return nil, l.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	return ir.MapNew{Key: key, Value: value, Pairs: pairs}, nil
}

// generatorPairTypes is how the [key, value] pairs a generator's call yields hold their keys and
// values: its yielded type, the first type argument of what it returns, must be a tuple of exactly
// two required elements, each held in one slot.
func (l *lowering) generatorPairTypes(source *ast.Node) (ir.Type, ir.Type, bool) {
	arguments := l.checker.GetTypeArguments(l.concrete(l.checker.GetTypeAtLocation(ast.SkipParentheses(source))))
	if len(arguments) == 0 || !checker.IsTupleType(arguments[0]) {
		return 0, 0, false
	}
	elements := l.checker.GetTypeArguments(arguments[0])
	flags := arguments[0].TargetTupleType().ElementFlags()
	if len(elements) != 2 || len(flags) != 2 || flags[0] != checker.ElementFlagsRequired || flags[1] != checker.ElementFlagsRequired {
		return 0, 0, false
	}
	key, keyKnown := l.representation(elements[0])
	value, valueKnown := l.representation(elements[1])
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
func (l *lowering) generatorStepped(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !ast.IsIdentifier(callee) {
		return false
	}
	symbol := l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || !loweredGenerator(symbol.Declarations[0]) {
		return false
	}
	if !l.steppedSite(node) {
		return false
	}
	returns := l.checker.GetTypeAtLocation(node)
	if !l.isLibraryType(returns, "Generator") {
		return false
	}
	arguments := l.checker.GetTypeArguments(returns)
	return len(arguments) == 3 && !l.isVoidValue(arguments[0]) && !l.holdsVoid(arguments[0])
}

// steppedSite reports whether node, parentheses aside, stands where a consumer steps it to its end:
// an array literal's spread, the one argument of the library's Array.from, or the one argument of
// the library's new Map or new Set. The constructors are judged by the built-in's symbol, never by
// the name, so a user class called Map, or a subclass of the built-in, isn't one (Language, on
// #qk8rztp): a subclass's set is program code a step can run. And, since S2c-1, the argument for a
// one-body function's Iterable<T> parameter (iterableArgument), whose only use is one for-of, which
// never reads the completion value either (Adamic's ruling on #qk8rztp, rulings.md:294 one call
// deeper). That use is proven after lowering; a program this lets past the void rules that the proof
// then refuses stops there, not yet, so the arm fails closed.
func (l *lowering) steppedSite(node *ast.Node) bool {
	if l.iterableArgument(node) {
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
	case parent.Kind == ast.KindSpreadElement && parent.Parent != nil && parent.Parent.Kind == ast.KindArrayLiteralExpression:
		return true
	case parent.Kind == ast.KindCallExpression && l.arrayFromCall(parent) && len(parent.AsCallExpression().Arguments.Nodes) == 1 && parent.AsCallExpression().Arguments.Nodes[0] == site:
		return true
	case parent.Kind == ast.KindNewExpression && l.collectionConstructor(parent) && parent.AsNewExpression().Arguments != nil && len(parent.AsNewExpression().Arguments.Nodes) == 1 && parent.AsNewExpression().Arguments.Nodes[0] == site:
		return true
	}
	return false
}

// collectionConstructor is new Map or new Set, the library's own.
func (l *lowering) collectionConstructor(created *ast.Node) bool {
	callee := ast.SkipParentheses(created.AsNewExpression().Expression)
	return l.isLibraryGlobal(callee, "Map") || l.isLibraryGlobal(callee, "Set")
}

// arrayFromCall is a call of the library's Array.from.
func (l *lowering) arrayFromCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "from" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array")
}
