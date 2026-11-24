package lower

import (
	"slices"
	"strconv"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Bind arguments enter a normal factory call before any closure is allocated.
// Its captured parameter cells own the snapshots and participate in the cycle proof.
// thisArg is captured, and the kept call hands it to a target that takes a this
// (adamicBoundCall, adamic_bound_closure_call). A bind lowered here hands a this
// body only its own receiver, a class method bound to the object it's read off
// (BoundMethod): expression.go stops this outside methods, a function expression
// refuses it, bindReceiverRule stops a method bound to another object, and a bind
// of a host member value is a use checkHostMemberValues stops. A .call of a host
// member value (hostMemberValueCallThrough) hands thisArg to a literal's method,
// and checkHostMemberValues holds it to the literal the method was read off.
func (l *lowering) functionBind(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "bind" || !l.libraryMember(callee) {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	target := ast.SkipParentheses(access.Expression)
	if _, builtin := l.libraryMethod(target, map[*ast.Symbol]bool{}); builtin {
		return nil, false, nil
	}
	if (node.Flags&ast.NodeFlagsOptionalChain != 0 || access.QuestionDotToken != nil) && target.Kind != ast.KindPropertyAccessExpression {
		return nil, true, l.notYet(node, "bind through an optional chain")
	}
	proven := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(target)))
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if l.isLibraryType(proven, "Function") || len(signatures) != 1 {
		return nil, true, l.notYet(node, "bind without one declared function signature")
	}
	if len(call.Arguments.Nodes) == 0 || hasSpread(node) {
		return nil, true, l.notYet(node, "bind without an explicit receiver or with spread fixed arguments")
	}
	method := false
	if target.Kind == ast.KindPropertyAccessExpression {
		if err := l.bindReceiverRule(node); err != nil {
			return nil, true, err
		}
		receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
		argument := ast.SkipParentheses(call.Arguments.Nodes[0])
		method = ast.IsIdentifier(receiver) && ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument)
		if !method && !l.bindOwnPlainProperty(receiver) {
			return nil, true, l.notYet(node, "bind of a function property without a captured receiver lookup")
		}
	}
	if predicateOfSignature(l.checker, signatures[0]) != nil {
		return nil, true, l.notYet(node, "bind of a narrowing function signature")
	}
	parameters := signatures[0].Parameters()
	for _, parameter := range parameters {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, true, l.notYet(node, "bind of a rest-parameter function")
			}
		}
	}
	var value ir.Expression
	var err error
	if method {
		value = ir.Undefined{Of: ir.Closure}
	} else {
		value, err = l.expression(target)
	}
	if err != nil {
		return nil, true, err
	}
	values := []ir.Expression{value}
	types := []*checker.Type{proven}
	nodes := []*ast.Node{target}
	for index, written := range call.Arguments.Nodes {
		argument, err := l.expression(written)
		if err != nil {
			return nil, true, err
		}
		if index > 0 && index-1 < len(parameters) {
			if of, known := l.bindParameterType(parameters[index-1]); known {
				argument = fit(argument, of)
			}
		}
		if argument.Type() == 0 {
			argument = fit(argument, ir.Object)
		}
		values = append(values, argument)
		types = append(types, l.checker.GetTypeAtLocation(written))
		nodes = append(nodes, written)
	}
	boundName := ""
	if method {
		boundName = target.Name().Text()
	}
	maker, _, err := l.boundFactory(node, node, values, types, nodes, parameters, proven, l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))), boundName)
	if err != nil {
		return nil, true, err
	}
	if method {
		l.wholeObjectCopies = append(l.wholeObjectCopies, wholeObjectCopy{node: target, source: values[1], copied: l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression), reader: "bind lookup"})
		return ir.BoundMethod{Object: values[1], Name: target.Name().Text(), Optional: access.QuestionDotToken != nil, Fixed: values[2:], Factory: maker}, true, nil
	}
	if access.QuestionDotToken != nil {
		function := -1
		if l.function != nil {
			function = l.functionIndex
		}
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_bind_target", Type: value.Type(), Function: function, ExpressionAssigned: true})
		read := ir.Read{Local: local, Of: value.Type()}
		values[0] = read
		return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: value}}, Result: ir.Conditional{
			Condition: ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}},
			WhenTrue:  ir.Call{Function: maker, Arguments: values, Returns: ir.Closure}, WhenNot: ir.Undefined{Of: ir.Closure}, Of: ir.Closure,
		}}, true, nil
	}
	return ir.Call{Function: maker, Arguments: values, Returns: ir.Closure}, true, nil
}

// boundFactory makes the factory a bind lowers to and the wrapper its bound closure runs, for node, the
// call that binds: values are the target, the thisArg and the fixed arguments as lowered, types the
// checker's type of each and nodes the nodes that wrote them, parameters the target's declared ones and
// proven its type. bound is the bound function's type, whose one signature gives the wrapper's result, and
// site the node the bound closure is allocated as. boundName names a class method bound through its
// receiver, or is empty. It gives the factory, whose call with values makes the bound closure, and the
// wrapper's result. A .call of a host member value lowers through it too (hostMemberValueCallThrough).
func (l *lowering) boundFactory(node, site *ast.Node, values []ir.Expression, types []*checker.Type, nodes []*ast.Node, parameters []*ast.Symbol, proven, bound *checker.Type, boundName string) (int, ir.Type, error) {
	maker := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "bind_factory", Returns: ir.Closure})
	wrapper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "bound_function", Closure: true})
	environment := []int{}
	reads := []ir.Expression{}
	for index, value := range values {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "bound_value", Type: value.Type(), Function: maker, Captured: true})
		l.noteLocal(local, types[index], nodes[index])
		environment = append(environment, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	boundSignatures := l.checker.GetSignaturesOfType(bound, checker.SignatureKindCall)
	if len(boundSignatures) != 1 {
		return 0, 0, l.notYet(node, "bind result without one declared function signature")
	}
	function := l.result.Functions[wrapper]
	function.Source.Role = "bound_function"
	function.Environment = environment
	arguments := append([]ir.Expression{}, reads[2:]...)
	// The library models bind's remaining parameters as a generic rest tuple.
	// Use the declared target's slots after the fixed prefix, rather than treating
	// that tuple as a single object argument.
	for _, parameter := range parameters[min(len(values)-2, len(parameters)):] {
		proven := l.checker.GetTypeOfSymbol(parameter)
		of, known := l.bindParameterType(parameter)
		if !known {
			return 0, 0, l.notYet(node, "a bound function parameter without a value representation")
		}
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: of, Function: wrapper})
		l.noteLocal(local, proven, node)
		function.Parameters = append(function.Parameters, local)
		arguments = append(arguments, ir.Read{Local: local, Of: of})
	}
	returns := l.checker.GetReturnTypeOfSignature(boundSignatures[0])
	if returns.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		var known bool
		function.Returns, known = l.representation(returns)
		if !known {
			return 0, 0, l.notYet(node, "a bound function result without a value representation")
		}
	}
	invoke := ir.CallClosure{Bound: true, BoundName: boundName, Closure: reads[0], Arguments: arguments, FunctionType: int(proven.Id()), Returns: function.Returns}
	if function.Returns == 0 {
		function.Body = []ir.Statement{ir.Evaluate{Value: invoke}, ir.Return{}}
	} else {
		function.Body = []ir.Statement{ir.Return{Value: invoke}}
	}
	l.result.Functions[wrapper] = function
	l.closureRecords = append(l.closureRecords, closureRecord{proven: bound, function: wrapper, node: node})
	made, err := l.programAllocation(site, l.graphAllocation(ir.BoundClosure{Function: wrapper, Fixed: len(values) - 2}, site))
	if err != nil {
		return 0, 0, err
	}
	l.result.Functions[maker].Parameters = environment
	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: made}}
	return maker, function.Returns, nil
}

// Bound descriptors and rebinding are deliberately separate promises. Follow
// simple aliases before lowering, including underneath typeof, which may erase reads.
func (l *lowering) boundSource(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return l.boundSource(node.AsAsExpression().Expression, seen)
	}
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "bind" && l.libraryMember(callee)
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] {
		return false
	}
	seen[symbol] = true
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindVariableDeclaration && l.boundSource(declaration.AsVariableDeclaration().Initializer, seen) {
			return true
		}
	}
	return false
}

func (l *lowering) boundObservation(node *ast.Node) error {
	if node.Kind == ast.KindCallExpression {
		if l.boundKeyedLength(node) && l.boundLengthMethodEntry(node.AsCallExpression().Arguments.Nodes[0], map[*ast.Symbol]bool{}) {
			return l.notYet(node, "a bound function's length (bound length metadata)")
		}
		if err := l.bindReceiverRule(node); err != nil {
			return err
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if l.boundSource(access.Expression, map[*ast.Symbol]bool{}) {
			switch node.Name().Text() {
			case "length":
				if l.boundLengthMethodEntry(access.Expression, map[*ast.Symbol]bool{}) {
					return l.notYet(node, "a bound function's length (bound length metadata)")
				}
			case "name":
				return l.notYet(node, "a bound function's name (bound name metadata)")
			case "bind":
				return l.notYet(node, "bind of a bound function (keeping its first receiver)")
			}
		}
	}
	if node.Kind == ast.KindNewExpression && l.boundSource(node.AsNewExpression().Expression, map[*ast.Symbol]bool{}) {
		return l.notYet(node, "new on a bound function (bound construction)")
	}
	return nil
}

// Class method entries carry code, not a runtime source length. Their bound
// wrappers can be hidden in ordinary callable slots too; unproven origins stop.
func (l *lowering) boundLengthMethodEntry(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return l.boundLengthMethodEntry(node.AsAsExpression().Expression, seen)
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return true
		}
		target := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if target.Kind != ast.KindPropertyAccessExpression || len(call.Arguments.Nodes) == 0 {
			return !l.boundLengthOriginal(target, seen)
		}
		return true // Own slots can also be overwritten with entry-backed bound wrappers.
	}
	if !ast.IsIdentifier(node) {
		return true
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] || l.assignedSymbol(symbol) {
		return true
	}
	seen[symbol] = true
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindVariableDeclaration && declaration.AsVariableDeclaration().Initializer != nil {
			return l.boundLengthMethodEntry(declaration.AsVariableDeclaration().Initializer, seen)
		}
	}
	return true
}

// A closure's source length is available for original source functions. An
// erased bound method can itself occupy a closure slot while lacking that
// metadata, so unproven and reassigned sources keep the observation stop.
func (l *lowering) boundLengthOriginal(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		return l.boundLengthOriginal(node.AsAsExpression().Expression, seen)
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || seen[symbol] || l.assignedSymbol(symbol) {
			return false
		}
		seen[symbol] = true
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration {
				return true
			}
			if declaration.Kind == ast.KindVariableDeclaration {
				return l.boundLengthOriginal(declaration.AsVariableDeclaration().Initializer, seen)
			}
		}

	}
	return false
}

// Recognize literal length reads through the Reflect.get helper before the
// unknown-view guard loses the adapter's provenance.
// Match the helper's parameter symbols and read shape, never its written name.
func (l *lowering) boundKeyedLength(node *ast.Node) bool {
	call := node.AsCallExpression()
	if len(call.Arguments.Nodes) != 2 {
		return false
	}
	key := ast.SkipParentheses(call.Arguments.Nodes[1])
	if key.Kind != ast.KindStringLiteral || key.Text() != "length" || !l.boundSource(call.Arguments.Nodes[0], map[*ast.Symbol]bool{}) {
		return false
	}
	signature := l.checker.GetResolvedSignature(node)
	if signature == nil || signature.Declaration() == nil || !ast.IsFunctionLike(signature.Declaration()) {
		return false
	}
	declaration := signature.Declaration()
	parameters := declaration.Parameters()
	if parameters == nil || len(parameters) != 2 || declaration.Body() == nil {
		return false
	}
	valueSymbol, keySymbol := l.symbol(parameters[0].Name()), l.symbol(parameters[1].Name())
	if valueSymbol == nil || keySymbol == nil {
		return false
	}
	matches := func(read *ast.Node) bool {
		if read.Kind != ast.KindCallExpression {
			return false
		}
		invoke := read.AsCallExpression()
		callee := ast.SkipParentheses(invoke.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "get" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Reflect") || len(invoke.Arguments.Nodes) != 3 {
			return false
		}
		arguments := invoke.Arguments.Nodes
		if l.symbol(ast.SkipParentheses(arguments[1])) != keySymbol || l.symbol(ast.SkipParentheses(arguments[2])) != valueSymbol {
			return false
		}
		wrapped := ast.SkipParentheses(arguments[0])
		if wrapped.Kind != ast.KindCallExpression {
			return false
		}
		wrapper := wrapped.AsCallExpression()
		return l.isLibraryGlobal(wrapper.Expression, "Object") && len(wrapper.Arguments.Nodes) == 1 && l.symbol(ast.SkipParentheses(wrapper.Arguments.Nodes[0])) == valueSymbol
	}
	found := false
	var visit ast.Visitor
	visit = func(read *ast.Node) bool {
		found = matches(read)
		if found {
			return true
		}
		return read.ForEachChild(visit)
	}
	declaration.Body().ForEachChild(visit)
	return found
}

func init() { refusalRules.register(2890, "boundObservation", (*lowering).boundObservation) }

// declaresBeyond is whether candidate declares more parameters than a call passing passed arguments
// fills: a value held as a narrower type, `(shape: Shape, index?: number) => boolean` held as
// `(shape: Shape) => boolean`, is called with the type's arguments alone, while JavaScript hands it
// every argument its caller passed (#04f5s0s, #jqf1drp). Parameters holds the declared parameters in
// order (a destructured or defaulted one arrives as one incoming local), behind a method's leading
// receiver, which Receiver marks and no argument fills, so it's left out of the count. The environment
// is Environment, never Parameters. A lone this parameter is counted, which can only stop more.
func declaresBeyond(candidate ir.Function, passed int) bool {
	return declaredParameters(candidate) > passed
}

// declaredParameters is how many parameters candidate declares, its leading receiver left out
// (declaresBeyond).
func declaredParameters(candidate ir.Function) int {
	if candidate.Receiver {
		return len(candidate.Parameters) - 1
	}
	return len(candidate.Parameters)
}

// boundReachObserved, when a test sets it, hears each bound value's reach as checkBoundArguments asks
// it: the bind's place, the names of the functions it reached, and whether the reach was known. It's
// nil outside that test (function_bind_reach_test.go).
var boundReachObserved atomic.Pointer[func(where string, functions []string, known bool)]

// parameterCount is count parameters, in the singular for one.
func parameterCount(count int) string {
	if count == 1 {
		return "1 parameter"
	}
	return strconv.Itoa(count) + " parameters"
}

// boundValueReach is, for a bound closure's target value, the functions that can be it by the
// allocation flow, asked once lowering is done as checkHostMemberWrites asks it (ReachingAllocations):
// a MakeClosure or BoundClosure reaching the value names its function, and undefined or null names
// none. Anything else reaching it, or a reach the graph calls Unknown (a parameter with unseen
// callers, an opaque call, a host value), is unknown. checkBoundArguments then asks boundNameReach
// for a property target, retaining every function of the bound type when the name is unreadable,
// so unknown fails closed. The graph is built once, on the first bind that needs it, its entry
// naming those values as the iterable views' proof names its origins
// (iterable_views.go); every road the graph follows, and every allocation it names itself, is left
// to it.
func (l *lowering) boundValueReach() func(ir.Expression) ([]int, bool) {
	assignAllocationSites(l.result)
	graph := l.flowGraph(l.result)
	const nothing, other = -1, -2
	named := []int{}
	site := func(function int) int {
		named = append(named, function)
		return len(named)
	}
	graph.entry = func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.MakeClosure:
			return site(value.Function)
		case ir.BoundClosure:
			return site(value.Function)
		case ir.Undefined, ir.Null:
			return site(nothing)
		case ir.Read, ir.Call, ir.CallClosure, ir.Conditional, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.Defined, ir.Coalesce, ir.Logical, ir.Property, ir.ArrayIndex, ir.ObjectLiteral, ir.ArrayLiteral:
			return 0
		}
		return site(other)
	}
	return func(value ir.Expression) ([]int, bool) {
		reaches := graph.ReachingAllocations(value)
		if reaches.Unknown {
			return nil, false
		}
		functions := []int{}
		for _, reached := range reaches.Sites {
			if reached <= 0 || reached > len(named) || named[reached-1] == other {
				// An allocation the graph names itself, an object or an array, or a value made
				// another way.
				return nil, false
			}
			if named[reached-1] != nothing {
				functions = append(functions, named[reached-1])
			}
		}
		return functions, true
	}
}

// Until bound adapters forward an unbounded argument vector, fail closed when a
// possible implementation observes its exact count, collects a rest tail or
// declares more parameters than the type it's bound as names (declaresBeyond).
func (l *lowering) checkBoundArguments() error {
	var reach func(ir.Expression) ([]int, bool)
	var named func(*ast.Node) ([]int, bool)
	for _, record := range l.closureRecords {
		if l.firstLinkSkips(record.node) {
			continue
		}
		function := l.result.Functions[record.function]
		if function.Source.Role != "bound_function" {
			continue
		}
		var found error
		walk(function.Body, func(node any) bool {
			if call, ok := node.(ir.CallClosure); ok && call.Bound {
				typed := l.result.FunctionTypeTargets[call.FunctionType]
				candidates := append([]int{}, typed...)
				methods := []int{}
				if call.BoundName != "" {
					methods = l.result.MethodsNamed(call.BoundName)
					candidates = append(candidates, methods...)
				}
				for _, target := range candidates {
					if candidate := l.result.Functions[target]; candidate.ReadsArguments || candidate.RestElement != 0 {
						found = l.notYet(record.node, "bind of a function that observes its argument count or collects rest arguments")
					}
				}
				if found != nil {
					// The count's stop gives way to the arguments and rest one, whose words stay as pinned.
					return true
				}
				// The count asks only what can be the bound value: every function of its type counts
				// functions the program never binds here, such as one taking an optional parameter held as
				// () => number elsewhere, which stopped function_bind.a's honest absentProperty.f?.bind.
				// A method bind keeps every method of the name.
				counted := append([]int{}, methods...)
				if slices.ContainsFunc(typed, func(target int) bool { return declaresBeyond(l.result.Functions[target], len(call.Arguments)) }) {
					if reach == nil {
						reach = l.boundValueReach()
					}
					reached, known := reach(call.Closure)
					if observe := boundReachObserved.Load(); observe != nil {
						names := []string{}
						for _, function := range reached {
							names = append(names, l.result.Functions[function].Name)
						}
						(*observe)(l.program.Where(record.node), names, known)
					}
					if known {
						counted = append(counted, reached...)
					} else {
						if named == nil {
							named = l.boundNameReach()
						}
						if functions, readable := named(boundMemberTarget(record.node)); readable {
							counted = append(counted, functions...)
						} else {
							counted = append(counted, typed...)
						}
					}
				}
				for _, target := range counted {
					if candidate := l.result.Functions[target]; declaresBeyond(candidate, len(call.Arguments)) {
						found = l.notYet(record.node, "bind of a function that declares "+parameterCount(declaredParameters(candidate))+" where the type it's bound as names "+strconv.Itoa(len(call.Arguments)))
						break
					}
				}
			}
			return true
		})
		if found != nil {
			return found
		}
	}
	return nil
}

func (l *lowering) bindParameterType(parameter *ast.Symbol) (ir.Type, bool) {
	of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
	for _, declaration := range parameter.Declarations {
		if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer != nil {
			of = ir.Maybe(of)
		}
	}
	return of, known
}

// A function-typed field can be satisfied by a class method. Until form two's
// lookup is installed, only a literal's own data fields have a proven closure arm.
func (l *lowering) bindOwnPlainProperty(receiver *ast.Node) bool {
	receiver = ast.SkipParentheses(receiver)
	if ast.IsIdentifier(receiver) {
		symbol := l.symbol(receiver)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		receiver = declaration.AsVariableDeclaration().Initializer
		if receiver == nil {
			return false
		}
		receiver = ast.SkipParentheses(receiver)
	}
	if receiver.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, field := range receiver.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
	}
	return true
}

// Only the direct X.m.bind(X) head owns dispatch. All paths are stopped even
// when a type declares plain fields: an actual getter can satisfy that view.
func (l *lowering) bindReceiverRule(node *ast.Node) error {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "bind" || !l.libraryMember(callee) {
		return nil
	}
	target := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if target.Kind == ast.KindElementAccessExpression {
		return l.notYet(node, "bind through an indexed receiver")
	}
	if target.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	if _, builtin := l.libraryMethod(target, map[*ast.Symbol]bool{}); builtin {
		return nil
	}
	if l.inheritedLibraryMember(target) {
		return l.notYet(node, "bind of an inherited library function without a declared callback type")
	}
	receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		if receiver.Kind == ast.KindCallExpression {
			return l.notYet(node, "bind with a call as its method receiver")
		}
		return l.notYet(node, "bind through a receiver path (a getter can change the receiver)")
	}
	if len(call.Arguments.Nodes) == 0 {
		return l.notYet(node, "method bind without its receiver")
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[0])
	same := ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument)
	symbol := l.memberSymbol(target)
	if !same {
		if symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 {
			return l.notYet(node, "method bind with a different receiver")
		}
		return nil
	}
	if !l.stableBindReceiver(receiver) {
		return l.notYet(node, "method bind whose receiver "+sourceExpression(receiver)+" isn't a const or a parameter the function never assigns ("+l.bindReceiverKind(receiver)+")")
	}
	if target.AsPropertyAccessExpression().QuestionDotToken != nil || call.QuestionDotToken != nil {
		return l.notYet(node, "bind outside the direct method optional-chain head")
	}
	if l.bindGetters == nil {
		l.bindGetters = map[string]bool{}
		var visit ast.Visitor
		visit = func(current *ast.Node) bool {
			if name, named := memberKeyText(current); named && current.Kind == ast.KindGetAccessor {
				l.bindGetters[name] = true
			}
			current.ForEachChild(visit)
			return false
		}
		for _, file := range l.program.Files() {
			file.AsNode().ForEachChild(visit)
		}
	}
	if l.bindGetters[target.Name().Text()] {
		return l.notYet(node, "bind of a getter-backed method value")
	}
	return nil
}

func (l *lowering) stableBindReceiver(node *ast.Node) bool {
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindParameter {
		return !l.assignedSymbol(symbol)
	}
	return declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0
}

// bindReceiverKind says what a receiver stableBindReceiver turns down is, for its stop's words: the old
// words said "assigned" of every one, a destructured binding never assigned included (#d6jncyx, family 66).
func (l *lowering) bindReceiverKind(node *ast.Node) string {
	symbol := l.symbol(node)
	switch {
	case symbol == nil || len(symbol.Declarations) == 0:
		return "a name with no declaration"
	case len(symbol.Declarations) != 1:
		return "a name declared more than once"
	}
	declaration := symbol.Declarations[0]
	switch {
	case declaration.Kind == ast.KindParameter:
		return "a parameter the function assigns"
	case declaration.Kind == ast.KindBindingElement:
		return "a destructured binding"
	case declaration.Kind == ast.KindVariableDeclaration:
		return "a let or var"
	}
	return describe(declaration)
}

func (l *lowering) directBindMethodRead(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.Name().Text() != "bind" || !called(parent) || !l.libraryMember(parent) {
		return false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	outer := parent.Parent
	for outer != nil && outer.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer == nil || outer.Kind != ast.KindCallExpression || len(outer.AsCallExpression().Arguments.Nodes) == 0 {
		return false
	}
	argument := ast.SkipParentheses(outer.AsCallExpression().Arguments.Nodes[0])
	return ast.IsIdentifier(receiver) && ast.IsIdentifier(argument) && l.symbol(receiver) == l.symbol(argument) && l.stableBindReceiver(receiver) && l.bindReceiverRule(outer) == nil
}

// The links of ordinary and bound closures are the cells their environment owns.
// The cycle proof and Program membership must read the same kept edges;
// allocation and destruction continue to use the owning environment itself.
func (l *lowering) closureLinks(closure closureRecord) []int {
	return l.result.Functions[closure.function].Environment
}
