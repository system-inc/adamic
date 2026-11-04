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

// loweredGenerator is a generator declaration this slice lowers: a named, non-generic function*
// declared at a module's top level. Rule 2500 stops every other generator before lowering.
func loweredGenerator(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().AsteriskToken != nil &&
		node.Parent != nil && node.Parent.Kind == ast.KindSourceFile && node.Name() != nil && len(node.TypeParameters()) == 0 &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
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

// generatorCallee is the generator function a call calls directly by name, and whether it is one.
func (l *lowering) generatorCallee(node *ast.Node) (int, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return 0, false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !ast.IsIdentifier(callee) {
		return 0, false
	}
	index, isFunction := l.functions[l.symbol(callee)]
	if !isFunction || !l.result.Functions[index].Generator {
		return 0, false
	}
	return index, true
}

// generatorCall lowers a direct call to a generator function, which makes its frame and runs none
// of its body. Only the two consumers that keep the generator to themselves call it.
func (l *lowering) generatorCall(node *ast.Node, function int) (ir.Expression, error) {
	call := ast.SkipParentheses(node).AsCallExpression()
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
		case ir.GeneratorNext:
			if read, ok := node.Generator.(ir.Read); ok && read.Of == ir.Generator {
				allowed++
			}
		case ir.GeneratorDone:
			if read, ok := node.Generator.(ir.Read); ok && read.Of == ir.Generator {
				allowed++
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

// generatorStepped is a direct call to a generator this slice lowers that a spread or Array.from
// consumes. Both step it to its end and never read its last result's value, so a TReturn of void,
// what a plain function* infers, is never read there, and the void rules (adamic/no-void-read,
// no-void-widening) have nothing to refuse in it. It holds only when the yielded type holds no void:
// every value a step hands over is read. It is syntactic, since the refusal walk runs before any
// function is lowered. It applies rulings.md's void-container classifier (judged by expression) to a
// generator, as Adamic confirmed on #qk8rztp. cohere's gate never refused these programs: it has no
// twin of no-void-read, and its no-void-widening judges function-type relations (#z7zv50g).
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
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	switch {
	case parent == nil:
		return false
	case parent.Kind == ast.KindSpreadElement && parent.Parent != nil && parent.Parent.Kind == ast.KindArrayLiteralExpression:
	case parent.Kind == ast.KindCallExpression && l.arrayFromCall(parent) && len(parent.AsCallExpression().Arguments.Nodes) == 1 && parent.AsCallExpression().Arguments.Nodes[0] == site:
	default:
		return false
	}
	returns := l.checker.GetTypeAtLocation(node)
	if !l.isLibraryType(returns, "Generator") {
		return false
	}
	arguments := l.checker.GetTypeArguments(returns)
	return len(arguments) == 3 && !l.isVoidValue(arguments[0]) && !l.holdsVoid(arguments[0])
}

// arrayFromCall is a call of the library's Array.from.
func (l *lowering) arrayFromCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "from" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array")
}
