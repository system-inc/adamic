package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// hostMemberValue is a read recv.m, as a value, of a method a program's interface or type literal
// declares (hostMemberReadable), which lowering let through where main stops it and checkHostMemberValues
// judges once lowering is done: read is the property access, object its receiver as lowering read it, or
// nil while lowering hasn't, and stop main's refusal of a method read as a value, unchanged. tsc saves its
// host's methods so before it patches them (program.ts:543, `const originalReadFile = host.readFile`),
// and calls or restores what it saved (#fj0vqtf wave 2). The deferral may only admit: every road that
// doesn't end in a value checkHostMemberValues admits ends in stop, but for a use outside the closed list,
// whose stop is checkHostMemberValues' own.
type hostMemberValue struct {
	read   *ast.Node
	object ir.Expression
	stop   error
}

// hostMemberReadable says whether node reads a host member (hostMemberDeclared) as a value: never written,
// which hostMemberWritable takes where it's a plain `=` and main refuses otherwise, never deleted, and never
// called, which methodValueStop doesn't refuse.
func (l *lowering) hostMemberReadable(node *ast.Node) bool {
	if ast.IsAssignmentTarget(node) || hostMemberAssignment(node) != nil {
		return false
	}
	held := node
	for held.Parent != nil && held.Parent.Kind == ast.KindParenthesizedExpression {
		held = held.Parent
	}
	if held.Parent != nil && held.Parent.Kind == ast.KindDeleteExpression {
		return false
	}
	return l.hostMemberDeclared(node)
}

// deferHostMemberValue holds a read hostMemberReadable lets through to checkHostMemberValues, with stop,
// main's refusal of it.
func (l *lowering) deferHostMemberValue(read *ast.Node, stop error) {
	for _, value := range l.hostMemberValues {
		if value.read == read {
			return
		}
	}
	l.hostMemberValues = append(l.hostMemberValues, hostMemberValue{read: read, stop: stop})
	if l.hostMemberFirst == nil {
		l.hostMemberFirst = stop
	}
}

// hostMemberValueDeferred says whether node is a read deferHostMemberValue holds.
func (l *lowering) hostMemberValueDeferred(node *ast.Node) bool {
	for _, value := range l.hostMemberValues {
		if value.read == node {
			return true
		}
	}
	return false
}

// hostMemberValueRead marks the read at node as lowered off object. A read lowered again, as a retried
// lowering does, is marked with its latest receiver, the one the program holds.
func (l *lowering) hostMemberValueRead(node *ast.Node, object ir.Expression) {
	for index := range l.hostMemberValues {
		if l.hostMemberValues[index].read == node {
			l.hostMemberValues[index].object = object
		}
	}
}

// hostMemberValueJudge is what checkHostMemberValues judges one value's uses with: the value, the graph
// and stores checkHostMemberWrites built, the functions the member may hold as closures (known false where
// that can't be bounded), and the const bindings of the value already followed.
type hostMemberValueJudge struct {
	value     hostMemberValue
	graph     *allocationFlowGraph
	stores    map[int]ir.SetProperty
	functions []int
	known     bool
	seen      map[*ast.Symbol]bool
}

// checkHostMemberValues judges each read hostMemberReadable let through, after checkHostMemberWrites has
// judged the writes, on the same graph (Lower's design on #fj0vqtf, with Language's and Compiler's
// conditions):
//
//  1. Reach and shape, as for a write: every allocation reaching the receiver is a plain program object
//     literal that holds m as a field of its own and isn't made with a spread, and an unknown reach fails
//     closed. A literal's method-syntax member and its arrow property are both a field holding a closure,
//     so the read is a field read. Every read's reach is judged before any read's uses, and each road
//     here ends in main's stop.
//  2. The closed list, enforced on the value: a literal's method-syntax member takes its object as this,
//     so the value may only go where that this is handed to it as Node hands it. Every use of the read,
//     and of every const binding that holds it, through every alias, is judged (hostMemberValueUse), so
//     no use escapes the list: f.call(t, ...), where t reaches only plain literals, under the layout rule
//     (hostMemberLayout); a plain call, only where no function the member may hold takes a this; a write
//     back to a method of a plain literal, under the layout rule. Anything else,
//     an argument, a return, a field of another object, an array element, a comparison, a binding that
//     isn't const, is a NotYet at that use in its own words.
func (l *lowering) checkHostMemberValues(graph *allocationFlowGraph, stores map[int]ir.SetProperty) error {
	index := graph.projectionIndex()
	for _, value := range l.hostMemberValues {
		// A read lowering never made, or made off something other than a local, can't be asked about
		// here, so it fails closed.
		if _, isRead := value.object.(ir.Read); !isRead {
			return value.stop
		}
		reaches := graph.ReachingAllocations(value.object)
		if err := l.hostMemberReach(value.stop, reaches, index); err != nil {
			return err
		}
		for _, site := range reaches.Sites {
			literal := index.records[site]
			if literal.Spread != nil || hostMemberField(literal, value.read.Name().Text()) == nil {
				// The member may be a spread source's, copied, or isn't a field of the literal's own.
				return value.stop
			}
		}
	}
	for _, value := range l.hostMemberValues {
		judge := &hostMemberValueJudge{value: value, graph: graph, stores: stores, seen: map[*ast.Symbol]bool{}}
		judge.functions, judge.known = l.hostMemberClosures(ir.Property{Object: value.object, Name: value.read.Name().Text()}, graph, map[string]bool{})
		if err := l.hostMemberValueUse(judge, value.read); err != nil {
			return err
		}
	}
	return nil
}

// hostMemberValueNamed is how a stop at a use of judge's value names it.
func hostMemberValueNamed(judge *hostMemberValueJudge) string {
	return sourceExpression(judge.value.read) + ", a method read as a value,"
}

// hostMemberValueUse judges one use of judge's value: use is the read itself, or a reference to a const
// binding that holds it.
func (l *lowering) hostMemberValueUse(judge *hostMemberValueJudge, use *ast.Node) error {
	held := use
	for held.Parent != nil && held.Parent.Kind == ast.KindParenthesizedExpression {
		held = held.Parent
	}
	parent := held.Parent
	switch {
	case held.Kind == ast.KindShorthandPropertyAssignment || parent == nil:
		// { original } stores the value in a field, and a use with no parent is none of the list's.
	case parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == held:
		return l.hostMemberValueBinding(judge, parent)
	case parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == held && parent.AsCallExpression().QuestionDotToken == nil:
		return l.hostMemberValueCall(judge, use, parent)
	case parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == held && parent.Name().Text() == "call" && called(parent) && l.libraryMember(parent):
		call := parent
		for call.Parent != nil && call.Parent.Kind == ast.KindParenthesizedExpression {
			call = call.Parent
		}
		return l.hostMemberValueThrough(judge, call.Parent)
	case parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Right == held:
		return l.hostMemberValueRestore(judge, parent)
	}
	return l.notYet(use, hostMemberValueNamed(judge)+" "+hostMemberValueOther(held))
}

// hostMemberValueOther names a use outside the closed list, held the use through any parentheses.
func hostMemberValueOther(held *ast.Node) string {
	if held.Kind == ast.KindShorthandPropertyAssignment {
		return "stored in an object's field"
	}
	if held.Parent == nil {
		return "used other than called"
	}
	switch parent := held.Parent; parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == held {
			return "called through ?."
		}
		return "passed as an argument"
	case ast.KindNewExpression:
		return "passed as an argument or constructed"
	case ast.KindReturnStatement, ast.KindArrowFunction:
		return "returned"
	case ast.KindPropertyAssignment:
		return "stored in an object's field"
	case ast.KindArrayLiteralExpression:
		return "stored in an array"
	case ast.KindExportSpecifier:
		return "exported"
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
			return "compared"
		}
	}
	return "used other than called"
}

// hostMemberValueBinding follows a const binding of judge's value: every reference to it in its file is a
// use, judged as the read is. A binding that isn't one const name, or is exported, may be read where no
// reference shows it, so it stops. A binding of a binding (an alias) is followed the same way.
func (l *lowering) hostMemberValueBinding(judge *hostMemberValueJudge, declaration *ast.Node) error {
	name := declaration.Name()
	if !ast.IsIdentifier(name) || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return l.notYet(declaration, hostMemberValueNamed(judge)+" held in a binding that isn't one const name")
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return l.notYet(declaration, hostMemberValueNamed(judge)+" held in an exported const")
	}
	symbol := l.symbol(name)
	if symbol == nil {
		return judge.value.stop
	}
	if judge.seen[symbol] {
		return nil
	}
	judge.seen[symbol] = true
	references := []*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindShorthandPropertyAssignment && l.checker.GetShorthandAssignmentValueSymbol(node) == symbol {
			// { original } reads the binding through a name that resolves to the property.
			references = append(references, node)
			return false
		}
		if ast.IsIdentifier(node) && node != name && !ast.IsPartOfTypeNode(node) && l.symbol(node) == symbol {
			references = append(references, node)
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	for _, reference := range references {
		if err := l.hostMemberValueUse(judge, reference); err != nil {
			return err
		}
	}
	return nil
}

// hostMemberValueCall judges a plain call of a binding of judge's value, use the callee and call the call.
// It hands the function no this, so it's admitted only where no function the member may hold takes one: an
// arrow property, or a function written to it without a this of its own. A literal's method-syntax member
// would read its this from the call's first argument natively, where Node throws a TypeError reading off
// undefined (Language's h2), so it stops, and so does a member whose functions can't be bounded.
func (l *lowering) hostMemberValueCall(judge *hostMemberValueJudge, use, call *ast.Node) error {
	held := "a call of " + sourceExpression(use) + " with no object, which holds " + sourceExpression(judge.value.read) + ", a method"
	if !judge.known {
		return l.notYet(call, held+" whose functions can't all be known")
	}
	for _, function := range judge.functions {
		if l.result.Functions[function].Receiver {
			return l.notYet(call, held+" whose function takes its object as this")
		}
	}
	return nil
}

// hostMemberValueThrough judges f.call(t, ...) of judge's value, call the call, which hostMemberCall
// lowered: t must reach only plain program object literals, and the value's functions must read this's
// fields at one place across t's literals and the read's own (hostMemberLayout, Compiler's rule). tsc's
// `originalReadFile.call(host, fileName)` passes the very host the value was read off, so the two sets are
// one and the rule costs it nothing. A call lowering never made this way fails closed to main's stop.
func (l *lowering) hostMemberValueThrough(judge *hostMemberValueJudge, call *ast.Node) error {
	for _, through := range l.hostMemberCalls {
		if through.call != call {
			continue
		}
		calling := hostMemberValueNamed(judge) + " called through .call"
		if _, isRead := through.thisArg.(ir.Read); !isRead {
			return l.notYet(call, calling+" with a this that isn't a name")
		}
		index := judge.graph.projectionIndex()
		thisArg := judge.graph.ReachingAllocations(through.thisArg)
		if err := l.hostMemberReach(l.notYet(call, calling+" with a this that isn't known to be a plain object literal"), thisArg, index); err != nil {
			return err
		}
		if !judge.known {
			return l.notYet(call, calling+", where what it holds can't all be known")
		}
		source := judge.graph.ReachingAllocations(judge.value.object)
		if field, agrees := l.hostMemberLayout(judge.functions, append(append([]int{}, source.Sites...), thisArg.Sites...), index); !agrees {
			return l.notYet(call, calling+", "+hostMemberLayoutMisfit(field))
		}
		return nil
	}
	return judge.value.stop
}

// hostMemberValueRestore judges a write of judge's value back to a method of a plain literal, assignment
// the write: tsc restores what it saved so (program.ts:4956, `host.readFile = originalReadFile`). The write
// is one checkHostMemberWrites has judged, by reach, shape and holders. A call through the written object
// then hands that object to the value's function as this, so the layout rule holds the value's functions
// to the written object's literals and the read's together (hostMemberLayout).
func (l *lowering) hostMemberValueRestore(judge *hostMemberValueJudge, assignment *ast.Node) error {
	for _, write := range l.hostMemberWrites {
		if hostMemberAssignment(write.target) != assignment {
			continue
		}
		store, lowered := judge.stores[write.site]
		if write.site == 0 || !lowered {
			return write.stop
		}
		target := judge.graph.ReachingAllocations(store.Object)
		source := judge.graph.ReachingAllocations(judge.value.object)
		writing := "a write of " + hostMemberValueNamed(judge) + " to method " + store.Name + " of " + sourceExpression(write.target.AsPropertyAccessExpression().Expression)
		if target.Unknown || source.Unknown || !judge.known {
			return l.notYet(assignment, writing+", where what it's written to or what it holds can't all be known")
		}
		if field, agrees := l.hostMemberLayout(judge.functions, append(append([]int{}, source.Sites...), target.Sites...), judge.graph.projectionIndex()); !agrees {
			return l.notYet(assignment, writing+", "+hostMemberLayoutMisfit(field))
		}
		return nil
	}
	return l.notYet(assignment, hostMemberValueNamed(judge)+" written somewhere other than a method of a plain object")
}

// hostMemberLayoutMisfit says why hostMemberLayout didn't admit, by the field it gave.
func hostMemberLayoutMisfit(field string) string {
	if field == "" {
		return "whose function uses this other than to read or write its fields"
	}
	return "whose function reads this." + field + ", which the objects it's called on don't all hold at one place"
}

// hostMemberClosures is the functions value may hold as a closure, and false where that can't be bounded:
// a closure made of one, a local through every source the graph gives it, and a field of plain literals
// through each literal's own value and every store the graph sees to it. Anything else, a closure no source
// names, a parameter's, a call's result, a field an opaque store may write, is unbounded.
func (l *lowering) hostMemberClosures(value ir.Expression, graph *allocationFlowGraph, seen map[string]bool) ([]int, bool) {
	key := fmt.Sprintf("%#v", value)
	if seen[key] {
		// Already followed on this walk: it adds nothing new.
		return nil, true
	}
	seen[key] = true
	sources := []ir.Expression{}
	switch held := value.(type) {
	case ir.MakeClosure:
		return []int{held.Function}, true
	case ir.Read:
		if _, unknown := graph.unknownAt(held.Local + 1); unknown {
			return nil, false
		}
		sources = graph.sources[held.Local+1]
	case ir.Property:
		reaches := graph.ReachingAllocations(held.Object)
		index := graph.projectionIndex()
		if reaches.Unknown || index.opaque[held.Name] {
			return nil, false
		}
		for _, site := range reaches.Sites {
			literal, isLiteral := index.records[site]
			if !isLiteral || literal.Spread != nil {
				return nil, false
			}
			field := hostMemberField(literal, held.Name)
			if field == nil {
				return nil, false
			}
			sources = append(append(sources, field), index.stores[site][held.Name]...)
		}
	case ir.Effects:
		// Effects keep their result's identity, as ClosureTargets reads them.
		sources = []ir.Expression{held.Result}
	case ir.Conditional:
		sources = []ir.Expression{held.WhenTrue, held.WhenNot}
	case ir.Narrow:
		sources = []ir.Expression{held.Value}
	case ir.CheckedCast:
		sources = []ir.Expression{held.Value}
	default:
		return nil, false
	}
	if len(sources) == 0 {
		return nil, false
	}
	functions := []int{}
	for _, source := range sources {
		found, known := l.hostMemberClosures(source, graph, seen)
		if !known {
			return nil, false
		}
		functions = append(functions, found...)
	}
	return functions, true
}

// hostMemberLayout is Compiler's layout rule: a function that takes its object as this reads that object's
// fields at fixed places, so every field each of functions reads or writes off this must sit at one
// place, the same index, in the layout of every literal at sites, the objects it may be called on (the
// uniform offsets uniformFieldOffsets proves, over these literals alone). It gives the first field that
// doesn't, and false. A function whose this goes anywhere but a field's object can't be judged, so it
// fails closed, named as this.
func (l *lowering) hostMemberLayout(functions []int, sites []int, index *allocationProjection) (string, bool) {
	for _, function := range functions {
		if !l.result.Functions[function].Receiver {
			continue
		}
		names, known := l.hostMemberThisFields(function)
		if !known {
			return "", false
		}
		for _, name := range names {
			offset := -1
			for _, site := range sites {
				literal, isLiteral := index.records[site]
				if !isLiteral {
					return name, false
				}
				at := -1
				for position, field := range literal.Fields {
					if field.Name == name {
						at = position
					}
				}
				if at < 0 || offset >= 0 && at != offset {
					return name, false
				}
				offset = at
			}
		}
	}
	return "", true
}

// hostMemberThisFields is every field function, one that takes its object as this, reads or writes off
// that this, and false where its this is used any other way, in its own body or a closure's that captures
// it: passed on, stored, returned, compared.
func (l *lowering) hostMemberThisFields(function int) ([]string, bool) {
	parameters := l.result.Functions[function].Parameters
	if len(parameters) == 0 {
		return nil, false
	}
	this := parameters[0]
	names := []string{}
	reads, accounted := 0, 0
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.Read:
			if value.Local == this {
				reads++
			}
		case ir.Property:
			if read, isRead := value.Object.(ir.Read); isRead && read.Local == this {
				accounted++
				names = append(names, value.Name)
			}
		case ir.SetProperty:
			if read, isRead := value.Object.(ir.Read); isRead && read.Local == this {
				accounted++
				names = append(names, value.Name)
			}
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, body := range l.result.Functions {
		walk(body.Body, inspect)
	}
	return names, reads == accounted
}

// hostMemberThrough is a call f.call(t, ...) hostMemberCall lowered, call the call and thisArg t as
// lowering read it, for hostMemberValueThrough to judge.
type hostMemberThrough struct {
	call    *ast.Node
	thisArg ir.Expression
}

// hostMemberValueOf says whether node, through parentheses and const bindings, is a host member read as a
// value that checkHostMemberValues judges (deferHostMemberValue): the read itself, or a const name bound to
// it or to such a name.
func (l *lowering) hostMemberValueOf(node *ast.Node) bool {
	seen := map[*ast.Symbol]bool{}
	for {
		node = ast.SkipParentheses(node)
		if l.hostMemberValueDeferred(node) {
			return true
		}
		if !ast.IsIdentifier(node) {
			return false
		}
		symbol := l.symbol(node)
		if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
			return false
		}
		seen[symbol] = true
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || declaration.AsVariableDeclaration().Initializer == nil {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
	}
}

// hostMemberCall lowers f.call(t, ...args), where f is a host member read as a value (hostMemberValueOf),
// as bind's road does (functionBind): a factory makes a bound closure of f and t, whose function calls f
// through CallClosure{Bound: true} with its own arguments, and the call calls it at once. The bound call
// hands t to f's this where f's function takes one, as Node does (adamic_bound_closure_call natively,
// adamicBoundCall on JavaScript), and checkHostMemberValues admits it only under the layout rule
// (hostMemberValueThrough). One argument for each of f's declared parameters, none spread or rest, and a t
// held as an object, are what this lowers; any other .call of such a value stops. Any other .call isn't
// this road's.
func (l *lowering) hostMemberCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().QuestionDotToken != nil || callee.Name().Text() != "call" || !l.libraryMember(callee) {
		return nil, false, nil
	}
	target := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if !l.hostMemberValueOf(target) {
		return nil, false, nil
	}
	written := call.Arguments.Nodes
	if len(written) == 0 || hasSpread(node) {
		return nil, true, l.notYet(node, "a .call of a method read as a value without its this, or with a spread argument")
	}
	proven := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(target)))
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(node, "a .call of a method read as a value without one declared signature")
	}
	parameters := signatures[0].Parameters()
	if len(written)-1 != len(parameters) {
		return nil, true, l.notYet(node, "a .call of a method read as a value with other than one argument for each parameter")
	}
	for _, parameter := range parameters {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, true, l.notYet(node, "a .call of a method read as a value with a rest parameter")
			}
		}
	}
	// What JavaScript evaluates, in its order: f, t, then the arguments.
	function, err := l.expression(target)
	if err != nil {
		return nil, true, err
	}
	thisArg, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if thisArg.Type() != ir.Object {
		return nil, true, l.notYet(written[0], "a .call of a method read as a value whose this isn't held as an object")
	}
	types := []ir.Type{}
	arguments := []ir.Expression{}
	for index, parameter := range parameters {
		of, known := l.bindParameterType(parameter)
		if !known {
			return nil, true, l.notYet(written[index+1], "a .call argument without a value representation")
		}
		argument, err := l.expression(written[index+1])
		if err != nil {
			return nil, true, err
		}
		types = append(types, of)
		arguments = append(arguments, fit(argument, of))
	}
	var returns ir.Type
	if result := l.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		var known bool
		returns, known = l.representation(result)
		if !known {
			return nil, true, l.notYet(node, "a .call result without a value representation")
		}
	}
	maker := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "bind_factory", Returns: ir.Closure})
	wrapper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "bound_function", Closure: true})
	environment := []int{}
	reads := []ir.Expression{}
	for index, value := range []ir.Expression{function, thisArg} {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "bound_value", Type: value.Type(), Function: maker, Captured: true})
		if index == 0 {
			l.noteLocal(local, proven, target)
		} else {
			l.noteLocal(local, l.checker.GetTypeAtLocation(written[0]), written[0])
		}
		environment = append(environment, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	bound := l.result.Functions[wrapper]
	bound.Source.Role = "bound_function"
	bound.Environment = environment
	bound.Returns = returns
	forwarded := []ir.Expression{}
	for index, parameter := range parameters {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: types[index], Function: wrapper})
		l.noteLocal(local, l.checker.GetTypeOfSymbol(parameter), node)
		bound.Parameters = append(bound.Parameters, local)
		forwarded = append(forwarded, ir.Read{Local: local, Of: types[index]})
	}
	invoke := ir.CallClosure{Bound: true, Closure: reads[0], Arguments: forwarded, FunctionType: int(proven.Id()), Returns: returns}
	if returns == 0 {
		bound.Body = []ir.Statement{ir.Evaluate{Value: invoke}, ir.Return{}}
	} else {
		bound.Body = []ir.Statement{ir.Return{Value: invoke}}
	}
	l.result.Functions[wrapper] = bound
	l.closureRecords = append(l.closureRecords, closureRecord{proven: proven, function: wrapper, node: node})
	made, err := l.programAllocation(node, l.graphAllocation(ir.BoundClosure{Function: wrapper}, node))
	if err != nil {
		return nil, true, err
	}
	l.result.Functions[maker].Parameters = environment
	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: made}}
	recorded := false
	for index := range l.hostMemberCalls {
		if l.hostMemberCalls[index].call == node {
			// Lowered again, as a retried lowering does: the latest is the one the program holds.
			l.hostMemberCalls[index].thisArg = thisArg
			recorded = true
		}
	}
	if !recorded {
		l.hostMemberCalls = append(l.hostMemberCalls, hostMemberThrough{call: node, thisArg: thisArg})
	}
	return ir.CallClosure{Closure: ir.Call{Function: maker, Arguments: []ir.Expression{function, thisArg}, Returns: ir.Closure}, Arguments: arguments, FunctionType: int(proven.Id()), Returns: returns}, true, nil
}
