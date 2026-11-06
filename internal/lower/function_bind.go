package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Bind arguments enter a normal factory call before any closure is allocated.
// Its captured parameter cells own the snapshots and participate in the cycle proof.
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
	if node.Flags&ast.NodeFlagsOptionalChain != 0 || access.QuestionDotToken != nil {
		return nil, true, l.notYet(node, "bind through an optional chain")
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(target))
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if l.isLibraryType(proven, "Function") || len(signatures) != 1 {
		return nil, true, l.notYet(node, "bind without one declared function signature")
	}
	if len(call.Arguments.Nodes) == 0 || hasSpread(node) {
		return nil, true, l.notYet(node, "bind without an explicit receiver or with spread fixed arguments")
	}
	if target.Kind == ast.KindPropertyAccessExpression {
		symbol := l.memberSymbol(target)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 {
			return nil, true, l.notYet(node, "bind of a method value before receiver dispatch is captured")
		}
		if !l.bindOwnPlainProperty(target.AsPropertyAccessExpression().Expression) {
			return nil, true, l.notYet(node, "bind of a function property before receiver dispatch is captured")
		}
	}
	parameters := signatures[0].Parameters()
	for _, parameter := range parameters {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, true, l.notYet(node, "bind of a rest-parameter function")
			}
		}
	}
	value, err := l.expression(target)
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
	boundType := l.concrete(l.checker.GetTypeAtLocation(node))
	boundSignatures := l.checker.GetSignaturesOfType(boundType, checker.SignatureKindCall)
	if len(boundSignatures) != 1 {
		return nil, true, l.notYet(node, "bind result without one declared function signature")
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
			return nil, true, l.notYet(node, "a bound function parameter without a value representation")
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
			return nil, true, l.notYet(node, "a bound function result without a value representation")
		}
	}
	invoke := ir.CallClosure{Bound: true, Closure: reads[0], Arguments: arguments, FunctionType: int(proven.Id()), Returns: function.Returns}
	if function.Returns == 0 {
		function.Body = []ir.Statement{ir.Evaluate{Value: invoke}, ir.Return{}}
	} else {
		function.Body = []ir.Statement{ir.Return{Value: invoke}}
	}
	l.result.Functions[wrapper] = function
	l.closureRecords = append(l.closureRecords, closureRecord{proven: boundType, function: wrapper, node: node})
	made, err := l.programAllocation(node, l.graphAllocation(ir.BoundClosure{Function: wrapper}, node))
	if err != nil {
		return nil, true, err
	}
	l.result.Functions[maker].Parameters = environment
	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: made}}
	return ir.Call{Function: maker, Arguments: values, Returns: ir.Closure}, true, nil
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
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if l.boundSource(access.Expression, map[*ast.Symbol]bool{}) {
			switch node.Name().Text() {
			case "length":
				return l.notYet(node, "a bound function's length (bound length metadata)")
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

func init() { refusalRules.register(2890, "boundObservation", (*lowering).boundObservation) }

// Until bound adapters forward an unbounded argument vector, fail closed when a
// possible implementation observes its exact count or collects a rest tail.
func (l *lowering) checkBoundArguments() error {
	for _, record := range l.closureRecords {
		function := l.result.Functions[record.function]
		if function.Source.Role != "bound_function" {
			continue
		}
		var found error
		walk(function.Body, func(node any) bool {
			if call, ok := node.(ir.CallClosure); ok && call.Bound {
				for _, target := range l.result.FunctionTypeTargets[call.FunctionType] {
					candidate := l.result.Functions[target]
					if candidate.ReadsArguments || candidate.RestElement != 0 {
						found = l.notYet(record.node, "bind of a function that observes its argument count or collects rest arguments")
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
