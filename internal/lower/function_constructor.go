package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// tsc builds its nodes with function constructors handed out by objectAllocator `as any` (utilities.ts's
// Node, Token and Identifier, `function Node(this: Mutable<Node>, kind, pos, end)`, newed by baseNodeFactory.ts
// and the parser through a cache filled from a getter). Adamic's ruling on #837zjkt lowers that road, and no
// other:
//
//  1. A function constructor becomes a synthesized class. Its layout is its body's this writes, in body order,
//     and its body is the class's constructor, so the new lowers to a call of it (functionConstructorEligible,
//     buildFunctionConstructor). It has a class of its own and a static identity class, which is what the
//     constructor value is at run time.
//  2. The `as any` on a getter's operand is erased only as a road to a new callee: the arrow body of an
//     allocator literal's getter (allocatorGetterShape). Its value is a fresh object of the static identity
//     class, made by nothing else. Every other use of an any keeps its refusal; this one's stops are deferred
//     (deferAnyRoad) and given back when the road doesn't close.
//  3. The callee is proven by reach once lowering is done (checkFunctionConstructorRoads): every allocation
//     reaching a new's constructor value is one of its candidate function constructors, and an unknown reach
//     is NotYet at the new. newClassValue dispatches over the candidates by their identity classes.
//  4. The value is born as a checked view of the construct signature's return type, so every read through it
//     is checked. In .a the whole road stays refused.
//  5. A constructor whose prototype the program reads or writes is no candidate: that use makes it not new-only
//     (closedKindNewOnlyFunctions), so its cast keeps its refusal.
//  6. The allocator is sealed: a program that writes, aliases or hands on the allocator object, as
//     setObjectAllocator's Object.assign does when it's called, turns every such new NotYet.

// functionConstructor is a function declaration lowered as a synthesized class: its declaration, the type its
// this is declared as, its this writes in body order, its place in functionConstructorRoads' order, and once
// built, its class, its static identity class and its constructor function, -1 until then.
type functionConstructor struct {
	declaration *ast.Node
	thisType    *checker.Type
	writes      []*ast.Node
	index       int
	class       int
	static      int
	constructor int
}

// allocatorRoads is the program's admitted roads, found once: each eligible function constructor some
// admitted getter cast names, in the order the program's files name them, and the casts, getter arrows and
// allocator literals that road admits.
type allocatorRoads struct {
	constructors map[*ast.Symbol]*functionConstructor
	ordered      []*functionConstructor
	casts        map[*ast.Node]*functionConstructor
	arrows       map[*ast.Node]*functionConstructor
	literals     map[*ast.Node]bool
}

// functionConstructorNew is a new lowered through newClassValue with a function constructor among its candidates:
// the new, its dispatch function, whose first argument is the constructor value, and the candidates by their
// index in functionConstructorRoads' order.
type functionConstructorNew struct {
	node       *ast.Node
	dispatch   int
	candidates []int
}

// anyRoad is a stop an admitted road deferred: the cast's own refusal, or a relation's any, and whether it's the
// cast's, which is the stop main gave before the road was admitted.
type anyRoad struct {
	node *ast.Node
	stop error
	cast bool
}

// functionConstructorRoads walks the program's own files once for admitted getter casts.
func (l *lowering) functionConstructorRoads() *allocatorRoads {
	if l.functionConstructorScan != nil {
		return l.functionConstructorScan
	}
	roads := &allocatorRoads{constructors: map[*ast.Symbol]*functionConstructor{}, casts: map[*ast.Node]*functionConstructor{}, arrows: map[*ast.Node]*functionConstructor{}, literals: map[*ast.Node]bool{}}
	l.functionConstructorScan = roads
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindAsExpression {
			if declaration := l.allocatorGetterShape(node); declaration != nil {
				symbol := l.closedKindFunctionSymbol(declaration)
				constructor := roads.constructors[symbol]
				if constructor == nil {
					constructor = l.functionConstructorEligible(declaration, symbol)
					if constructor != nil {
						constructor.index = len(roads.ordered)
						roads.constructors[symbol] = constructor
						roads.ordered = append(roads.ordered, constructor)
					}
				}
				if constructor != nil {
					roads.casts[node] = constructor
					roads.arrows[allocatorGetterArrowOf(node)] = constructor
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if neverRestProgramFile(file) {
			file.AsNode().ForEachChild(visit)
		}
	}
	// An allocator literal is admitted whole or not at all: a literal holding any member but an admitted getter
	// keeps the any relation's stop at the literal.
	for arrow := range roads.arrows {
		literal := arrow.Parent.Parent
		if _, judged := roads.literals[literal]; judged {
			continue
		}
		whole := true
		for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindPropertyAssignment || roads.arrows[property.AsPropertyAssignment().Initializer] == nil {
				whole = false
			}
		}
		roads.literals[literal] = whole
	}
	return roads
}

// allocatorGetterArrowOf is the arrow an admitted cast is the expression body of.
func allocatorGetterArrowOf(cast *ast.Node) *ast.Node {
	site := cast
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	return site.Parent
}

// allocatorGetterShape is the function declaration a cast to any names when the cast is an allocator getter's
// whole result, as utilities.ts's `getNodeConstructor: () => Node as any` is, and nil otherwise. The cast is
// written in .ts, its operand names a function declaration, and it's the expression body of an arrow with no
// parameters that is a property of an object literal initializing a typed const at a module's top level. The
// const's type gives that property one call signature with no parameters, returning a type with exactly one
// construct signature, not generic, and no call signature, whose parameters are the function's after its this.
func (l *lowering) allocatorGetterShape(cast *ast.Node) *ast.Node {
	as := cast.AsAsExpression()
	if as.Type == nil || as.Type.Kind != ast.KindAnyKeyword || strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(cast)), ".a") {
		return nil
	}
	operand := ast.SkipParentheses(as.Expression)
	if !ast.IsIdentifier(operand) {
		return nil
	}
	arrow := allocatorGetterArrowOf(cast)
	if arrow == nil || arrow.Kind != ast.KindArrowFunction || len(arrow.Parameters()) != 0 || len(arrow.TypeParameters()) != 0 || ast.HasSyntacticModifier(arrow, ast.ModifierFlagsAsync) {
		return nil
	}
	if body := arrow.Body(); body == nil || ast.SkipParentheses(body) != cast {
		return nil
	}
	property := arrow.Parent
	if property == nil || property.Kind != ast.KindPropertyAssignment || property.AsPropertyAssignment().Initializer != arrow || !ast.IsIdentifier(property.Name()) {
		return nil
	}
	literal := property.Parent
	if literal == nil || literal.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	variable := literal.Parent
	if variable == nil || variable.Kind != ast.KindVariableDeclaration || variable.AsVariableDeclaration().Initializer != literal || variable.Type() == nil || !ast.IsIdentifier(variable.Name()) {
		return nil
	}
	list := variable.Parent
	if list == nil || list.Flags&ast.NodeFlagsConst == 0 || list.Parent == nil || list.Parent.Kind != ast.KindVariableStatement || list.Parent.Parent == nil || list.Parent.Parent.Kind != ast.KindSourceFile {
		return nil
	}
	member := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(variable.Name()), property.Name().Text())
	if member == nil {
		return nil
	}
	getters := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(member), checker.SignatureKindCall)
	if len(getters) != 1 || len(getters[0].Parameters()) != 0 {
		return nil
	}
	made := l.checker.GetReturnTypeOfSignature(getters[0])
	constructs := l.checker.GetSignaturesOfType(made, checker.SignatureKindConstruct)
	if len(constructs) != 1 || len(constructs[0].TypeParameters()) != 0 || len(l.checker.GetSignaturesOfType(made, checker.SignatureKindCall)) != 0 {
		return nil
	}
	symbol := l.valueSymbol(operand)
	if symbol == nil {
		return nil
	}
	var declaration *ast.Node
	for _, candidate := range symbol.Declarations {
		if candidate.Kind == ast.KindFunctionDeclaration {
			if declaration != nil {
				return nil
			}
			declaration = candidate
		}
	}
	if declaration == nil || l.closedKindFunctionSymbol(declaration) != symbol || len(constructs[0].Parameters()) != len(declaration.Parameters())-1 {
		return nil
	}
	return declaration
}

// functionConstructorEligible is the function constructor a declaration lowers as, or nil when it isn't one:
// its this is a declared parameter, every use of it is a new callee or a cast to any (closedKindNewOnlyFunctions,
// so a read or write of its prototype rules it out), it isn't generic, async or a generator, its other
// parameters are plain required names, it returns void, and its body is nothing but `this.name = value`
// statements whose values don't hold this.
//
// That's stricter than the ruling, which asks only that the body never read this before writing it and never
// pass this to a call. Slice 1 refuses every read of this and every write but a top-level one: a read after a
// write, or a write on one path, needs the slot's state at that point proven (an Absent slot for a field one
// path leaves unwritten, #837zjkt's slice 2), and a body that only writes needs no such proof, since nothing in
// it can see a slot before its write. A default, optional or rest parameter is refused for the same cause and
// one more, never because it's unsound: a default's initializer runs before the body's first write and may read
// this, and the dispatch maps a new's arguments onto the parameters one to one, by equal arity (Language's f09;
// tsc's constructors take none).
func (l *lowering) functionConstructorEligible(declaration *ast.Node, symbol *ast.Symbol) *functionConstructor {
	if symbol == nil || declaration.Name() == nil || !closedKindDeclaresThis(declaration) || !l.closedKindNewOnlyFunctions()[symbol] {
		return nil
	}
	if len(declaration.TypeParameters()) != 0 || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) || generatorDeclaration(declaration) {
		return nil
	}
	body := declaration.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	parameters := declaration.Parameters()
	for _, parameter := range parameters[1:] {
		declared := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declared.Initializer != nil || declared.QuestionToken != nil || declared.DotDotDotToken != nil {
			return nil
		}
	}
	if returns := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(declaration)); returns.Flags()&checker.TypeFlagsVoid == 0 {
		return nil
	}
	writes := []*ast.Node{}
	for _, statement := range body.AsBlock().Statements.Nodes {
		if statement.Kind != ast.KindExpressionStatement {
			return nil
		}
		assignment := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
		if assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
			return nil
		}
		target := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
		if target.Kind != ast.KindPropertyAccessExpression || ast.SkipParentheses(target.AsPropertyAccessExpression().Expression).Kind != ast.KindThisKeyword || !ast.IsIdentifier(target.Name()) {
			return nil
		}
		if containsThis(assignment.AsBinaryExpression().Right) {
			return nil
		}
		writes = append(writes, target)
	}
	return &functionConstructor{declaration: declaration, thisType: l.checker.GetTypeAtLocation(parameters[0].Name()), writes: writes, constructor: -1}
}

// allocatorGetterCast is the function constructor an admitted cast names, nil for any other node.
func (l *lowering) allocatorGetterCast(node *ast.Node) *functionConstructor {
	if node == nil || node.Kind != ast.KindAsExpression {
		return nil
	}
	return l.functionConstructorRoads().casts[node]
}

// allocatorGetterArrow is the function constructor an admitted getter arrow returns, nil for any other node.
func (l *lowering) allocatorGetterArrow(node *ast.Node) *functionConstructor {
	if node == nil || node.Kind != ast.KindArrowFunction {
		return nil
	}
	return l.functionConstructorRoads().arrows[node]
}

// allocatorRelationSite is whether a relation site is an admitted getter arrow, or an allocator literal whose
// every member is one: the any refuseAnyRelation finds there is the road's, deferred with its casts.
func (l *lowering) allocatorRelationSite(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	roads := l.functionConstructorRoads()
	return roads.arrows[node] != nil || roads.literals[node]
}

// functionConstructorAt is whether the function at index is a built function constructor's, whose this
// parameter names the object it makes and takes no argument.
func (l *lowering) functionConstructorAt(index int, declaration *ast.Node) bool {
	if declaration.Kind != ast.KindFunctionDeclaration || l.functionConstructorScan == nil {
		return false
	}
	for _, constructor := range l.functionConstructorScan.ordered {
		if constructor.constructor == index {
			return true
		}
	}
	return false
}

// deferAnyRoad keeps an admitted road's stop, given back by checkFunctionConstructorRoads where the road doesn't
// close.
func (l *lowering) deferAnyRoad(node *ast.Node, stop error, cast bool) {
	for _, road := range l.anyRoads {
		if road.node == node {
			return
		}
	}
	l.anyRoads = append(l.anyRoads, anyRoad{node: node, stop: stop, cast: cast})
}

// anyRoadStop is the deferred stop main gave first: a cast's refusal before a relation's, since the refusal walk
// returns the cast's at once and keeps a relation's for last.
func (l *lowering) anyRoadStop() error {
	for _, road := range l.anyRoads {
		if road.cast {
			return road.stop
		}
	}
	if len(l.anyRoads) != 0 {
		return l.anyRoads[0].stop
	}
	return nil
}

// functionConstructorIdentity is an admitted cast's value: a fresh object of its function constructor's static
// identity class. The road lets the value go nowhere but a new's callee, so its identity is never observed, and a
// fresh object needs no hoisted binding and no readiness check.
func (l *lowering) functionConstructorIdentity(constructor *functionConstructor) (ir.Expression, error) {
	if err := l.buildFunctionConstructor(constructor); err != nil {
		return nil, err
	}
	return ir.ObjectLiteral{ProgramRegion: l.programClass(constructor.static), Class: constructor.static}, nil
}

// buildFunctionConstructor lowers a function constructor once: its static identity class, its class, whose fields
// are its this writes in first-write order at their declared types' storage, and its constructor function, which
// makes the object, runs the body with this as the object, and returns it, as a class's constructor does
// (inheritanceConstructor's single-function form).
func (l *lowering) buildFunctionConstructor(constructor *functionConstructor) error {
	if constructor.constructor >= 0 {
		return nil
	}
	declaration := constructor.declaration
	name := declaration.Name().Text()
	source := l.functionSource(declaration)
	constructor.static = len(l.result.Classes) + 1
	l.result.Classes = append(l.result.Classes, ir.Class{Source: source, Name: name + "_static", Constructor: -1, Static: true})
	constructor.constructor = len(l.result.Functions)
	allocate := source
	allocate.Role = "allocate"
	allocate.Declaration = append(append([]string{}, source.Declaration...), "allocate")
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_new", Source: allocate, Returns: ir.Object})
	constructor.class = len(l.result.Classes) + 1
	l.classHolder(constructor.class, constructor.thisType)
	fields := []ir.Field{}
	written := map[string]bool{}
	for _, target := range constructor.writes {
		field := target.Name().Text()
		if written[field] {
			continue
		}
		written[field] = true
		symbol := l.checker.GetSymbolAtLocation(target.Name())
		if symbol == nil {
			return l.notYet(target, "a function constructor's write of a field its this doesn't declare")
		}
		of, err := l.typeOfSymbol(target, symbol)
		if err != nil {
			return err
		}
		if slotless(of) {
			return l.notYet(target, "a function constructor field of type "+l.checker.TypeToString(l.checker.GetTypeOfSymbol(symbol)))
		}
		value := zeroValue(of)
		if of.IsReference() {
			value = ir.Undefined{Of: of}
		}
		fields = append(fields, ir.Field{Name: field, Value: value})
	}
	l.result.Classes = append(l.result.Classes, ir.Class{Source: source, Name: name, Definition: constructor.class, Constructor: constructor.constructor, Fields: fields})

	// The body is lowered as a class's constructor is: with locals of its own beside the globals, and no class,
	// instantiation or mapper of whatever was being lowered when the first new needed it.
	this := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: constructor.constructor})
	outerLocals, outerInstance, outerSubstitution := l.locals, l.instance, l.substitution
	outerClassType, outerClassNode, outerMapper, outerUnset := l.classType, l.classNode, l.typeMapper, l.unsetUntil
	l.instance, l.substitution, l.classType, l.classNode, l.typeMapper, l.unsetUntil = nil, nil, nil, nil, nil, 0
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	l.noteLocal(this, constructor.thisType, declaration.Name())
	err := l.lowerFunction(constructor.constructor, declaration, this)
	l.locals, l.instance, l.substitution = outerLocals, outerInstance, outerSubstitution
	l.classType, l.classNode, l.typeMapper, l.unsetUntil = outerClassType, outerClassNode, outerMapper, outerUnset
	if err != nil {
		return err
	}
	function := l.result.Functions[constructor.constructor]
	function.Body = append([]ir.Statement{ir.Declare{Local: this, Value: ir.ObjectLiteral{ProgramRegion: l.programClass(constructor.class), Fields: fields, Class: constructor.class}}}, function.Body...)
	function.Body = append(function.Body, ir.Return{Value: ir.Read{Local: this, Of: ir.Object}})
	function.Returns = ir.Object
	l.result.Functions[constructor.constructor] = function
	return nil
}

// functionConstructorCandidates is the function constructors a new through a construct signature may make,
// built: those taking the signature's count of arguments whose this is related to its result either way. The set
// only trims the dispatch; the reach is the proof (checkFunctionConstructorRoads). A generic construct signature
// with such a candidate is NotYet.
func (l *lowering) functionConstructorCandidates(node *ast.Node, target *checker.Type) ([]*functionConstructor, error) {
	roads := l.functionConstructorRoads()
	if len(roads.ordered) == 0 {
		return nil, nil
	}
	signature := l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)[0]
	made := l.checker.GetReturnTypeOfSignature(signature)
	candidates := []*functionConstructor{}
	for _, constructor := range roads.ordered {
		if len(constructor.declaration.Parameters())-1 != len(signature.Parameters()) {
			continue
		}
		if !l.checker.IsTypeAssignableTo(constructor.thisType, made) && !l.checker.IsTypeAssignableTo(made, constructor.thisType) {
			continue
		}
		if len(signature.TypeParameters()) != 0 {
			return nil, l.notYet(node, "a new through a generic construct signature a function constructor may answer")
		}
		if err := l.buildFunctionConstructor(constructor); err != nil {
			return nil, err
		}
		candidates = append(candidates, constructor)
	}
	return candidates, nil
}

// allocatorGetterCall is, for a call through an allocator's getter, `allocator.getX()` with no arguments, never
// ?. or bound, the allocator's local and the index of the function constructor its getter's identity object is
// of, by allocators' fields. The callee is taken apart by ir.MethodCallee, so a getter call a narrowing or a
// non-null assertion checks is one too. It reads the callee's property spelling, never its targets: the graph
// can't bound a call through a property (ClosureTargets is Unknown there), and the allocator's seal is what
// names the getter that runs.
func allocatorGetterCall(value ir.Expression, allocators map[int]map[string]int) (int, int, bool) {
	call, isCall := value.(ir.CallClosure)
	if !isCall || len(call.Arguments) != 0 || call.Optional || call.Bound {
		return 0, 0, false
	}
	property, _, isMethod := ir.MethodCallee(call.Closure)
	if !isMethod || property.Optional {
		return 0, 0, false
	}
	read, isRead := property.Object.(ir.Read)
	if !isRead {
		return 0, 0, false
	}
	constructor, isField := allocators[read.Local][property.Name]
	return read.Local, constructor, isField
}

// functionConstructorUnsealed is the NotYet a new through an allocator the program writes, aliases or hands on
// takes: what reaches its constructor value isn't known.
const functionConstructorUnsealed = "a new of a function constructor through an allocator the program writes, aliases or hands on, so what reaches its constructor isn't known"

// checkFunctionConstructorRoads judges each admitted road once lowering is done, on the whole program's allocation
// graph, as checkSlotWrites judges its writes:
//
//  1. Seal: an allocator, a global declared once with an object literal whose every field is a getter closure
//     returning an identity object, is never assigned, and every read of it in code that can run is the object
//     of a getter's call with no arguments. An alias, an Object.assign into it, a store through it and a pass of
//     it are each another read, and break the seal (Adamic's premise pin).
//  2. Uses: every getter call through a sealed allocator is stored into a local, its cache, or is inside a new's
//     constructor value, and every read of a cache is inside a new's constructor value or is the result of the
//     cache's initializing helper (classConstructorValue). Anything else, a comparison, a pass, a store
//     elsewhere, gives back the deferred stop: the any keeps its refusal.
//  3. Reach: every allocation reaching a new's constructor value is an identity object of one of its candidates,
//     the entry hook naming each identity object and each sealed getter's call (the graph calls a call through a
//     property opaque); an unknown or other reach is NotYet at the new.
//
// A road whose cast reaches no new gives back its deferred stop too. A new in a function nothing runs makes
// nothing, and isn't judged.
func (l *lowering) checkFunctionConstructorRoads() error {
	if len(l.anyRoads) == 0 && len(l.functionConstructorNews) == 0 {
		return nil
	}
	if len(l.functionConstructorNews) == 0 {
		return l.anyRoadStop()
	}
	assignAllocationSites(l.result)
	graph := l.flowGraph(l.result)
	roads := l.functionConstructorRoads()
	statics := map[int]int{}
	for _, constructor := range roads.ordered {
		if constructor.constructor >= 0 {
			statics[constructor.static] = constructor.index
		}
	}
	running := func(visit func(node any) bool) {
		walk(l.result.Main, visit)
		for index, function := range l.result.Functions {
			if index < len(graph.running) && graph.running[index] {
				walk(function.Body, visit)
			}
		}
	}
	// A getter is a closure whose whole body returns an identity object.
	getters := map[int]int{}
	for index, function := range l.result.Functions {
		if !function.Closure || len(function.Body) != 1 {
			continue
		}
		if returned, isReturn := function.Body[0].(ir.Return); isReturn {
			if literal, isLiteral := returned.Value.(ir.ObjectLiteral); isLiteral {
				if constructor, isIdentity := statics[literal.Class]; isIdentity {
					getters[index] = constructor
				}
			}
		}
	}
	// An allocator is a global declared with a literal of getters, each field naming the constructor its getter's
	// identity object is of.
	allocators := map[int]map[string]int{}
	running(func(node any) bool {
		declared, isDeclare := node.(ir.Declare)
		if !isDeclare || declared.Local < 0 || declared.Local >= len(l.result.Locals) || !l.result.Locals[declared.Local].Global {
			return true
		}
		literal, isLiteral := declared.Value.(ir.ObjectLiteral)
		if !isLiteral || literal.Record || literal.Spread != nil || len(literal.Fields) == 0 {
			return true
		}
		fields := map[string]int{}
		for _, field := range literal.Fields {
			made, isClosure := field.Value.(ir.MakeClosure)
			if !isClosure {
				return true
			}
			constructor, isGetter := getters[made.Function]
			if !isGetter {
				return true
			}
			fields[field.Name] = constructor
		}
		allocators[declared.Local] = fields
		return true
	})
	getterCall := func(value ir.Expression) (int, int, bool) {
		return allocatorGetterCall(value, allocators)
	}
	declares, assigns, reads, calls := map[int]int{}, map[int]int{}, map[int]int{}, map[int]int{}
	running(func(node any) bool {
		switch value := node.(type) {
		case ir.Declare:
			declares[value.Local]++
		case ir.Assign:
			assigns[value.Local]++
		case ir.Read:
			reads[value.Local]++
		case ir.CallClosure:
			if allocator, _, isGetter := getterCall(value); isGetter {
				calls[allocator]++
			}
		}
		return true
	})
	sealed := map[int]bool{}
	unsealed := false
	for allocator := range allocators {
		sealed[allocator] = declares[allocator] == 1 && assigns[allocator] == 0 && reads[allocator] == calls[allocator]
		unsealed = unsealed || !sealed[allocator]
	}
	sealedGetter := func(value ir.Expression) (int, bool) {
		allocator, constructor, isGetter := getterCall(value)
		return constructor, isGetter && sealed[allocator]
	}

	// Each new's constructor value, the first argument of its dispatch, where code that can run calls it.
	dispatches := map[int]int{}
	for index, record := range l.functionConstructorNews {
		dispatches[record.dispatch] = index
	}
	constructorValues := make([][]ir.Expression, len(l.functionConstructorNews))
	running(func(node any) bool {
		if call, isCall := node.(ir.Call); isCall && len(call.Arguments) > 0 {
			// A dispatch is a direct call, whose one target is its function.
			if targets := l.result.CallTargets(call); len(targets) == 1 {
				if index, isDispatch := dispatches[targets[0]]; isDispatch {
					constructorValues[index] = append(constructorValues[index], call.Arguments[0])
				}
			}
		}
		return true
	})

	// Uses: a cache is a local a sealed getter's call is stored into.
	caches := map[int]bool{}
	totalCalls, admittedCalls := 0, 0
	running(func(node any) bool {
		switch value := node.(type) {
		case ir.Declare:
			if _, isGetter := sealedGetter(value.Value); isGetter {
				caches[value.Local] = true
				admittedCalls++
			}
		case ir.Assign:
			if _, isGetter := sealedGetter(value.Value); isGetter {
				caches[value.Local] = true
				admittedCalls++
			}
		case ir.CallClosure:
			if _, isGetter := sealedGetter(value); isGetter {
				totalCalls++
			}
		}
		return true
	})
	admittedReads := map[int]int{}
	for _, values := range constructorValues {
		for _, value := range values {
			walk(value, func(node any) bool {
				switch inner := node.(type) {
				case ir.Read:
					if caches[inner.Local] {
						admittedReads[inner.Local]++
					}
				case ir.CallClosure:
					if _, isGetter := sealedGetter(inner); isGetter {
						admittedCalls++
					}
				}
				return true
			})
		}
	}
	for index, function := range l.result.Functions {
		if function.Name != "class_constructor_cache" || index >= len(graph.running) || !graph.running[index] {
			continue
		}
		walk(function.Body, func(node any) bool {
			if returned, isReturn := node.(ir.Return); isReturn {
				if read, isRead := returned.Value.(ir.Read); isRead && caches[read.Local] {
					admittedReads[read.Local]++
				}
			}
			return true
		})
	}
	used := totalCalls == admittedCalls
	for cache := range caches {
		used = used && reads[cache] == admittedReads[cache]
	}
	if !used {
		if stop := l.anyRoadStop(); stop != nil {
			return stop
		}
		return l.notYet(l.functionConstructorNews[0].node, "a function constructor's value used other than as a new's constructor")
	}

	graph.entry = func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.ObjectLiteral:
			if constructor, isIdentity := statics[value.Class]; isIdentity {
				return constructor + 1
			}
		case ir.CallClosure:
			if constructor, isGetter := sealedGetter(value); isGetter {
				return constructor + 1
			}
		}
		return 0
	}
	for index, record := range l.functionConstructorNews {
		candidates := map[int]bool{}
		for _, candidate := range record.candidates {
			candidates[candidate+1] = true
		}
		for _, value := range constructorValues[index] {
			reaches := graph.ReachingAllocations(value)
			if reaches.Unknown {
				if unsealed {
					return l.notYet(record.node, functionConstructorUnsealed)
				}
				return l.notYet(record.node, "a new of a function constructor whose constructor value's reach isn't known ("+strings.Join(reaches.Reasons, "; ")+")")
			}
			// A value no allocation reaches would meet the dispatch's panic, where Node throws a TypeError a program
			// can catch, so it stops here instead.
			if len(reaches.Sites) == 0 {
				return l.notYet(record.node, "a new whose constructor value no function constructor reaches")
			}
			for _, site := range reaches.Sites {
				if !candidates[site] {
					return l.notYet(record.node, "a new whose constructor value may be something other than its function constructors")
				}
			}
		}
	}
	return nil
}
